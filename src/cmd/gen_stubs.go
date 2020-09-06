// +build generate

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/workspace"
)

// TODO: set overrides.
// TODO: dump meta.Info.Scope as well.
// TODO: uncomment the cleanup code.
// TODO: refactor writeEntryPoints.

func main() {
	var g genStubs
	g.stubsDir = filepath.Join("stubs", "phpstorm-stubs")

	steps := []struct {
		name string
		fn   func() error
	}{
		{"init", g.init},
		{"stubs indexing", g.indexStubs},
		{"collect files", g.collectFiles},
		{"sort", g.sort},
		// {"remove old files", g.removeOldFiles},
		{"dump stubs", g.dumpStubs},
		{"write entry points", g.writeEntryPoints},
		{"write version file", g.writeVersionFile},
	}

	for _, step := range steps {
		if err := step.fn(); err != nil {
			log.Fatalf("%s: %v", step.name, err)
		}
	}
}

type versionInfo struct {
	Commit string
	Files  []string
}

type stubsFile struct {
	ID         string
	GoFilename string
	FuncName   string
	Filename   string

	Classes   []classInfoEntry
	Functions []funcInfoEntry
	Constants []constInfoEntry
}

type classInfoEntry struct {
	Key     string
	Val     meta.ClassInfo
	Methods []funcInfoEntry
}

type funcInfoEntry struct {
	Key string
	Val meta.FuncInfo
}

type constInfoEntry struct {
	Key string
	Val meta.ConstInfo
}

type stubsDumper struct {
	dirPrefix string
	indent    int
	out       *bytes.Buffer
	f         *stubsFile
	nameSeq   int
}

type genStubs struct {
	oldVersion versionInfo

	stubsCommit string

	workDir  string
	stubsDir string

	stubsInfo meta.StubsInfo
	files     []*stubsFile
}

func (g *genStubs) init() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %v", err)
	}
	if !strings.HasSuffix(wd, string(filepath.Separator)) {
		wd += string(filepath.Separator)
	}

	gitDir := filepath.Join(wd, "stubs", "phpstorm-stubs", ".git")
	gitOut, err := exec.Command("git", "--git-dir", gitDir, "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git: %v: %s", err, gitOut)
	}

	g.workDir = wd
	g.oldVersion = g.readVersionFile()
	g.stubsCommit = strings.TrimSpace(string(gitOut))
	return nil
}

func (g *genStubs) indexStubs() error {
	go linter.MemoryLimiterThread()

	readCallback := workspace.ReadFilenames([]string{g.stubsDir}, nil)
	g.stubsInfo = linter.InitStubsFromReadCallback(readCallback)

	return nil
}

func (g *genStubs) collectFiles() error {
	wd := g.workDir

	byFile := map[string]*stubsFile{}

	getStubsFile := func(pos meta.ElementPosition) *stubsFile {
		filename := strings.TrimPrefix(pos.Filename, wd)
		f, ok := byFile[filename]
		if !ok {
			id := filenameToID(filename)
			f = &stubsFile{
				ID:         id,
				FuncName:   "Load_" + filepath.Base(strings.Replace(id, "-", "_", -1)),
				GoFilename: id + ".go",
				Filename:   filename,
			}
			byFile[filename] = f
		}
		return f
	}

	for className, class := range g.stubsInfo.Classes.H {
		f := getStubsFile(class.Pos)
		entry := classInfoEntry{Key: string(className), Val: class}
		entry.Methods = make([]funcInfoEntry, 0, class.Methods.Len())
		for methodName, m := range class.Methods.H {
			entry.Methods = append(entry.Methods, funcInfoEntry{
				Key: string(methodName),
				Val: m,
			})
		}
		f.Classes = append(f.Classes, entry)
	}

	for funcName, fn := range g.stubsInfo.Functions.H {
		f := getStubsFile(fn.Pos)
		entry := funcInfoEntry{Key: string(funcName), Val: fn}
		f.Functions = append(f.Functions, entry)
	}

	for constName, c := range g.stubsInfo.Constants {
		f := getStubsFile(c.Pos)
		entry := constInfoEntry{Key: constName, Val: c}
		f.Constants = append(f.Constants, entry)
	}

	files := make([]*stubsFile, 0, len(byFile))
	for _, f := range byFile {
		files = append(files, f)
	}

	g.files = files
	return nil
}

func (g *genStubs) sort() error {
	sortedFiles := make([]*stubsFile, 0, len(g.files))
	for _, f := range g.files {
		sort.SliceStable(f.Classes, func(i, j int) bool {
			return f.Classes[i].Key < f.Classes[j].Key
		})
		for _, entry := range f.Classes {
			sort.SliceStable(entry.Methods, func(i, j int) bool {
				return entry.Methods[i].Key < entry.Methods[j].Key
			})
		}

		sort.SliceStable(f.Functions, func(i, j int) bool {
			return f.Functions[i].Key < f.Functions[j].Key
		})

		sortedFiles = append(sortedFiles, f)
	}
	sort.SliceStable(sortedFiles, func(i, j int) bool {
		return sortedFiles[i].Filename < sortedFiles[j].Filename
	})

	g.files = sortedFiles
	return nil
}

func (g *genStubs) removeOldFiles() error {
	for _, filename := range g.oldVersion.Files {
		fullFilename := filepath.Join(g.workDir, "stubs", filename)
		if err := os.Remove(fullFilename); err != nil {
			return err
		}
	}
	return nil
}

func (g *genStubs) dumpStubs() error {
	for _, f := range g.files {
		var buf bytes.Buffer
		dumper := &stubsDumper{
			dirPrefix: g.workDir,
			out:       &buf,
		}
		dumper.dumpFile(f)
		g.writeGoFile(f.GoFilename, buf.Bytes())
	}

	return nil
}

func packageNameForFile(f string) string {
	f = filepath.Dir(f)
	f = strings.ReplaceAll(f, string(os.PathSeparator), "_")
	f = strings.ReplaceAll(f, ".", "_")
	f = strings.ReplaceAll(f, " ", "_")
	f = strings.ReplaceAll(f, "-", "_")
	return f
}

func (g *genStubs) writeEntryPoints() error {
	var buf bytes.Buffer
	buf.WriteString("// Code generated by the `cmd/gen_stubs.go`. DO NOT EDIT.\n")
	buf.WriteString("package stubs\n")

	pkgs := make(map[string]bool)
	for _, f := range g.files {
		p := packageNameForFile(f.Filename)
		if !pkgs[p] {
			fmt.Fprintf(&buf, "import %s \"github.com/VKCOM/noverify/src/cmd/stubs/%s\"\n", p, filepath.Dir(f.GoFilename))
		}
		pkgs[p] = true
	}

	buf.WriteString("func Load() {\n")
	for _, f := range g.files {
		fmt.Fprintf(&buf, "  %s.%s()\n", packageNameForFile(f.Filename), f.FuncName)
	}
	buf.WriteString("}\n")
	buf.WriteString("\n")
	buf.WriteString("func LoadByName(name string) bool {\n")
	buf.WriteString("  switch name {")
	for _, f := range g.files {
		fmt.Fprintf(&buf, "  case %q: %s.%s()\n", f.Filename, packageNameForFile(f.Filename), f.FuncName)
	}
	buf.WriteString("  default: return false")
	buf.WriteString("  }\n")
	buf.WriteString("  return true")
	buf.WriteString("}\n")
	g.writeGoFile("stubs.go", buf.Bytes())

	return nil
}

func (g *genStubs) writeVersionFile() error {
	filenames := make([]string, len(g.files))
	for i, f := range g.files {
		filenames[i] = f.GoFilename
	}
	newVersion := versionInfo{
		Commit: g.stubsCommit,
		Files:  filenames,
	}
	jsonData, err := json.MarshalIndent(newVersion, "", "\t")
	if err != nil {
		return fmt.Errorf("json encode: %v", err)
	}
	jsonData = append(jsonData, '\n')
	if err := g.writeFile("version.json", jsonData); err != nil {
		return fmt.Errorf("write file: %v", err)
	}
	return nil
}

func (g *genStubs) readVersionFile() versionInfo {
	var info versionInfo
	filename := filepath.Join(g.workDir, "stubs", "version.json")
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return info
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return info
	}
	return info
}

func (g *genStubs) writeFile(filename string, data []byte) error {
	dstFilename := filepath.Join(g.workDir, "stubs", filename)
	if err := os.MkdirAll(filepath.Dir(dstFilename), 0777); err != nil {
		return err
	}
	return ioutil.WriteFile(dstFilename, data, 0666)
}

func (g *genStubs) writeGoFile(filename string, data []byte) {
	pretty, err := format.Source(data)

	if err != nil {
		pretty = data
	}

	if err := g.writeFile(filename, pretty); err != nil {
		panic(err.Error())
	}
	if err != nil {
		fmt.Println(string(data))
		panic(fmt.Errorf("gofmt: %s: %v", filename, err))
	}
}

func filenameToID(filename string) string {
	id := filename
	id = strings.TrimSuffix(id, ".php")
	id = strings.ReplaceAll(id, " ", "")
	id = strings.ReplaceAll(id, "_", "-")
	id = strings.ReplaceAll(id, ".", "-")
	id = filepath.Dir(id) + "/s" + filepath.Base(id)
	return id
}

var tabs = strings.Repeat("\t", 32)

func (d *stubsDumper) printf(format string, args ...interface{}) {
	d.out.WriteString(tabs[:d.indent])
	fmt.Fprintf(d.out, format, args...)
	d.newline()
}

func (d *stubsDumper) newline() {
	d.out.WriteByte('\n')
}

func (d *stubsDumper) gensym(prefix string) string {
	id := d.nameSeq
	d.nameSeq++
	return prefix + "_" + strconv.Itoa(id)
}

func (d *stubsDumper) dumpFile(f *stubsFile) {
	d.f = f

	d.printf("// Code generated by the `cmd/gen_stubs.go`. DO NOT EDIT.")
	d.printf("package stubs")
	d.printf(`import "github.com/VKCOM/noverify/src/meta"`)
	d.printf(`import "github.com/VKCOM/noverify/src/cmd/stubs/stubsutil"`)

	d.printf("func %s() {", f.FuncName)
	d.indent++

	if len(f.Classes) != 0 {
		d.printf("fileClasses := meta.NewClassesMap()")
	}
	if len(f.Functions) != 0 {
		d.printf("fileFunctions := meta.NewFunctionsMap()")
	}
	if len(f.Constants) != 0 {
		d.printf("fileConstants := make(meta.ConstantsMap)")
	}
	for _, class := range f.Classes {
		d.dumpClass(class)
		d.newline()
	}
	for _, fn := range f.Functions {
		d.dumpFunc(fn)
		d.newline()
	}
	for _, c := range f.Constants {
		lhs := fmt.Sprintf("fileConstants[%q]", c.Key)
		d.dumpConst(lhs, c.Val)
	}
	if len(f.Classes) != 0 {
		d.printf("meta.Info.AddClassesNonLocked(%q, fileClasses)", f.Filename)
	}
	if len(f.Functions) != 0 {
		d.printf("meta.Info.AddFunctionsNonLocked(%q, fileFunctions)", f.Filename)
	}
	if len(f.Constants) != 0 {
		d.printf("meta.Info.AddConstantsNonLocked(%q, fileConstants)", f.Filename)
	}

	d.indent--
	d.printf("}")
}

var posType = reflect.TypeOf(meta.ElementPosition{})

func (d *stubsDumper) dumpPos(lhs string, pos meta.ElementPosition) {
	// positions are only used in language server
	return

	// d.printf("stubsutil.InitPos(&%s, %q, %d, %d, %d, %d)",
	// lhs, d.f.Filename, pos.Line, pos.EndLine, pos.Character, pos.Length)

	// posVal := reflect.ValueOf(rhsVal)
	// for j := 0; j < posType.NumField(); j++ {
	// 	field := posType.Field(j)
	// 	if field.Name == "Filename" {
	// 		continue
	// 	}
	// 	d.printf("%s.%s = %#v", lhs, field.Name, posVal.Field(j).Interface())
	// }
	// // Filename is assignment separately because we want to hide
	// // the original abs path of the stubs files.
	// d.printf("%s.Filename = %q", lhs, d.f.Filename)
}

func (d *stubsDumper) typeInitializer(m meta.TypesMap) string {
	switch m.String() {
	case "string":
		return "meta.StringType"
	case "bool":
		return "meta.BoolType"
	case "int":
		return "meta.IntType"
	case "mixed", "":
		return "meta.MixedType"
	case "void":
		return "meta.VoidType"
	default:
		types := make(map[string]struct{}, m.Len())
		m.Iterate(func(typ string) {
			types[typ] = struct{}{}
		})
		// fmt package prints map literals with keys sorted.
		return fmt.Sprintf("meta.NewTypesMapFromMap(%#v).Immutable()", types)
	}
}

func (d *stubsDumper) dumpType(lhs string, m meta.TypesMap) {
	d.printf("%s = %s", lhs, d.typeInitializer(m))
}

func (d *stubsDumper) dumpParam(lhs string, param meta.FuncParam) {
	typ := d.typeInitializer(param.Typ)
	if param.IsRef {
		d.printf("%s = stubsutil.NewRefFuncParam(%q, %s)", lhs, param.Name, typ)
	} else {
		d.printf("%s = stubsutil.NewFuncParam(%q, %s)", lhs, param.Name, typ)
	}
	// d.dumpBoolAssign(lhs+".IsRef", param.IsRef)
	// d.printf("%s.Name = %q", lhs, param.Name)
	// d.dumpType(lhs+".Typ", param.Typ)
}

func (d *stubsDumper) dumpIntAssign(lhs string, v int64) {
	if v != 0 {
		d.printf("%s = %d", lhs, v)
	}
}

func (d *stubsDumper) dumpStringAssign(lhs string, v string) {
	if v != "" {
		d.printf("%s = %q", lhs, v)
	}
}

func (d *stubsDumper) dumpBoolAssign(lhs string, v bool) {
	if v {
		d.printf("%s = true", lhs)
	}
}

func (d *stubsDumper) dumpProp(lhs string, prop meta.PropertyInfo) {
	d.printf("{")
	d.indent++

	d.printf("prop := meta.PropertyInfo{Typ: %s}", d.typeInitializer(prop.Typ))
	d.dumpPos("prop.Pos", prop.Pos)
	d.dumpIntAssign("prop.AccessLevel", int64(prop.AccessLevel))
	d.printf("%s = prop", lhs)

	// TODO: guard against new fields in PropertyInfo.

	d.indent--
	d.printf("}")
}

func (d *stubsDumper) dumpConst(lhs string, c meta.ConstInfo) {
	// d.printf("{")
	// d.indent++

	rhs := d.gensym("c")
	d.printf("%s := meta.ConstInfo{Value: %#v, Typ: %s}", rhs, c.Value, d.typeInitializer(c.Typ))
	d.dumpPos(rhs+".Pos", c.Pos)
	d.dumpIntAssign(rhs+".AccessLevel", int64(c.AccessLevel))
	d.printf("%s = %s", lhs, rhs)

	// TODO: guard against new fields in ConstInfo.

	// d.indent--
	// d.printf("}")
}

func (d *stubsDumper) dumpFnVariable(m funcInfoEntry) string {
	rv := reflect.ValueOf(m.Val)
	typ := rv.Type()

	name := d.gensym("fn")

	d.printf("%s := meta.FuncInfo{Name: %q, Flags: %d, MinParamsCnt: %d, Typ: %s}",
		name, m.Val.Name, m.Val.Flags, m.Val.MinParamsCnt, d.typeInitializer(m.Val.Typ))
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		switch field.Name {
		case "Name", "Typ", "Flags", "MinParamsCnt":
			// Part of the literal initialization.
		case "Pos":
			d.dumpPos(name+".Pos", m.Val.Pos)
		case "AccessLevel":
			d.dumpIntAssign(name+".AccessLevel", int64(m.Val.AccessLevel))
		case "ExitFlags":
			d.dumpIntAssign(name+".ExitFlags", int64(m.Val.ExitFlags))
		case "Doc":
			doc := m.Val.Doc
			if doc.Deprecated || doc.DeprecationNote != "" {
				d.printf(name+".Doc = %#v", m.Val.Doc)
			}
		case "Params":
			if len(m.Val.Params) == 0 {
				continue
			}
			d.printf("%s.Params = make([]meta.FuncParam, %d)", name, len(m.Val.Params))
			for i, param := range m.Val.Params {
				lhs := fmt.Sprintf("%s.Params[%d]", name, i)
				d.dumpParam(lhs, param)
			}
		default:
			panic(fmt.Sprintf("can't dump FuncInfo.%s field", field.Name))
		}
	}

	return name
}

func (d *stubsDumper) dumpFunc(m funcInfoEntry) {
	// d.printf("{")
	// d.indent++
	fn := d.dumpFnVariable(m)
	d.printf("fileFunctions.H[%q] = %s", m.Key, fn)
	// d.indent--
	// d.printf("}")
}

func (d *stubsDumper) dumpMethod(class classInfoEntry, m funcInfoEntry) {
	// d.printf("{")
	// d.indent++
	fn := d.dumpFnVariable(m)
	d.printf("class.Methods.H[%q] = %s", m.Key, fn)
	// d.indent--
	// d.printf("}")
}

func (d *stubsDumper) dumpClass(class classInfoEntry) {
	rv := reflect.ValueOf(class.Val)
	typ := rv.Type()
	d.printf("{")
	d.indent++
	d.printf("class := meta.ClassInfo{Name: %q}", class.Val.Name)
	if class.Val.Methods.Len() != 0 {
		d.printf("class.Methods = meta.NewFunctionsMap()")
	}
	if len(class.Val.Constants) != 0 {
		d.printf("class.Constants = make(meta.ConstantsMap)")
	}
	if len(class.Val.Properties) != 0 {
		d.printf("class.Properties = make(meta.PropertiesMap)")
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		switch field.Name {
		case "Name":
			// Part of the literal initialization.

		case "Pos":
			d.dumpPos("class.Pos", rv.Field(i).Interface().(meta.ElementPosition))

		case "Methods":
			for _, m := range class.Methods {
				d.dumpMethod(class, m)
			}

		case "Properties":
			for propName, p := range class.Val.Properties {
				lhs := fmt.Sprintf("class.Properties[%q]", propName)
				d.dumpProp(lhs, p)
			}

		case "Interfaces":
			if len(class.Val.Interfaces) != 0 {
				d.printf("class.Interfaces = %#v", class.Val.Interfaces)
			}
		case "Traits":
			if len(class.Val.Traits) != 0 {
				d.printf("class.Traits = %#v", class.Val.Traits)
			}

		case "Parent":
			d.dumpStringAssign("class.Parent", class.Val.Parent)

		case "Flags":
			d.dumpIntAssign("class.Flags", int64(class.Val.Flags))

		case "ParentInterfaces":
			if len(class.Val.ParentInterfaces) != 0 {
				d.printf("class.ParentInterfaces = %#v", class.Val.ParentInterfaces)
			}

		case "Constants":
			for constName, c := range class.Val.Constants {
				lhs := fmt.Sprintf("class.Constants[%q]", constName)
				d.dumpConst(lhs, c)
			}

		default:
			panic(fmt.Sprintf("can't dump ClassInfo.%s field", field.Name))
		}
	}
	d.printf("fileClasses.H[%q] = class", class.Key)

	d.indent--
	d.printf("}")
}
