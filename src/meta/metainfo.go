package meta

import (
	"strings"
	"sync"

	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/name"
)

var (
	internalFunctions         FunctionsMap
	internalFunctionOverrides FunctionsOverrideMap
	internalClasses           ClassesMap

	indexingComplete bool

	// Info contains global meta information for all classes, functions, etc.
	Info info
)

func init() {
	ResetInfo()
}

// ResetInfo creates empty meta info
func ResetInfo() {
	Info = info{
		Scope:                 NewScope(),
		allFiles:              make(map[string]bool),
		allTraits:             make(TraitsMap),
		allClasses:            make(ClassesMap),
		allFunctions:          make(FunctionsMap),
		allConstants:          make(ConstantsMap),
		allFunctionsOverrides: make(FunctionsOverrideMap),
		perFileTraits:         make(map[string]ClassesMap),
		perFileClasses:        make(map[string]ClassesMap),
		perFileFunctions:      make(map[string]FunctionsMap),
		perFileConstants:      make(map[string]ConstantsMap),
	}

	indexingComplete = false
}

type info struct {
	sync.Mutex
	*Scope
	allFiles              map[string]bool
	allTraits             TraitsMap
	allClasses            ClassesMap
	allFunctions          FunctionsMap
	allConstants          ConstantsMap
	allFunctionsOverrides FunctionsOverrideMap
	perFileTraits         map[string]ClassesMap
	perFileClasses        map[string]ClassesMap
	perFileFunctions      map[string]FunctionsMap
	perFileConstants      map[string]ConstantsMap
}

// PerFile contains all meta information about the specified file
type PerFile struct {
	Traits    ClassesMap
	Classes   ClassesMap
	Functions FunctionsMap
	Constants ConstantsMap
}

func (i *info) GetConstant(nm string) (res ConstantInfo, ok bool) {
	res, ok = i.allConstants[nm]
	return res, ok
}

func (i *info) NumConstants() int {
	return len(i.allConstants)
}

func (i *info) GetClass(nm string) (res ClassInfo, ok bool) {
	res, ok = i.allClasses[nm]
	return res, ok
}

func (i *info) GetTrait(nm string) (res ClassInfo, ok bool) {
	res, ok = i.allTraits[nm]
	return res, ok
}

func (i *info) NumClasses() int {
	return len(i.allClasses)
}

func (i *info) GetFunction(nm string) (res FuncInfo, ok bool) {
	res, ok = i.allFunctions[nm]
	return res, ok
}

func (i *info) GetFunctionOverride(nm string) (res FuncInfoOverride, ok bool) {
	res, ok = i.allFunctionsOverrides[nm]
	return res, ok
}

func (i *info) NumFunctions() int {
	return len(i.allFunctions)
}

func (i *info) NumFilesWithFunctions() int {
	return len(i.perFileFunctions)
}

func (i *info) FindFunctions(substr string) (res []string) {
	for f := range i.allFunctions {
		if strings.HasPrefix(f, substr) {
			res = append(res, f)
		}
	}
	return res
}

func (i *info) FindConstants(substr string) (res []string) {
	for c := range i.allConstants {
		if strings.HasPrefix(c, substr) {
			res = append(res, c)
		}
	}
	return res
}

func (i *info) InitStubs() {
	i.Lock()
	defer i.Unlock()

	internalFunctions = make(FunctionsMap)
	for k, v := range i.allFunctions {
		internalFunctions[k] = v
	}

	internalClasses = make(ClassesMap)
	for k, v := range i.allClasses {
		internalClasses[k] = v
	}

	internalFunctionOverrides = make(FunctionsOverrideMap)
	for k, v := range i.allFunctionsOverrides {
		internalFunctionOverrides[k] = v
	}
}

func (i *info) AddFilenameNonLocked(filename string) {
	i.allFiles[filename] = true
}

func (i *info) FileExists(filename string) bool {
	return i.allFiles[filename]
}

func (i *info) GetMetaForFile(filename string) (res PerFile) {
	if t, ok := i.perFileTraits[filename]; ok {
		res.Traits = t
	}

	if c, ok := i.perFileConstants[filename]; ok {
		res.Constants = c
	}

	if f, ok := i.perFileFunctions[filename]; ok {
		res.Functions = f
	}

	if c, ok := i.perFileClasses[filename]; ok {
		res.Classes = c
	}

	return res
}

func (i *info) DeleteMetaForFileNonLocked(filename string) {
	oldClasses := i.perFileClasses[filename]
	delete(i.allFiles, filename)
	delete(i.perFileClasses, filename)

	for f := range oldClasses {
		delete(i.allClasses, f)
	}

	oldTraits := i.perFileTraits[filename]
	delete(i.perFileTraits, filename)

	for f := range oldTraits {
		delete(i.allTraits, f)
	}

	oldFunctions := i.perFileFunctions[filename]
	delete(i.perFileFunctions, filename)

	for f, oldFn := range oldFunctions {
		fn, ok := i.allFunctions[f]
		if !ok || oldFn.Pos.Length != fn.Pos.Length {
			continue
		}
		delete(i.allFunctions, f)
	}

	oldConstants := i.perFileConstants[filename]
	delete(i.perFileConstants, filename)

	for f := range oldConstants {
		delete(i.allConstants, f)
	}
}

func (i *info) AddClassesNonLocked(filename string, m ClassesMap) {
	i.perFileClasses[filename] = m
	for k, v := range m {
		// TODO: resolve duplicate class conflicts
		i.allClasses[k] = v
	}
}

func (i *info) AddTraitsNonLocked(filename string, m ClassesMap) {
	i.perFileTraits[filename] = m
	for k, v := range m {
		// TODO: resolve duplicate trait conflicts
		i.allTraits[k] = v
	}
}

func (i *info) AddFunctionsNonLocked(filename string, m FunctionsMap) {
	i.perFileFunctions[filename] = m

	for k, v := range m {
		prevFn, ok := i.allFunctions[k]
		if !ok || v.Pos.Length > prevFn.Pos.Length {
			i.allFunctions[k] = v
		}
	}
}

func (i *info) AddFunctionsOverridesNonLocked(filename string, m FunctionsOverrideMap) {
	// TODO: support filename map

	for k, v := range m {
		i.allFunctionsOverrides[k] = v
	}
}

func (i *info) AddConstantsNonLocked(filename string, m ConstantsMap) {
	i.perFileConstants[filename] = m

	for k, v := range m {
		i.allConstants[k] = v
	}
}

func (i *info) AddToGlobalScopeNonLocked(filename string, sc *Scope) {
	sc.Iterate(func(nm string, typ *TypesMap, alwaysDefined bool) {
		i.AddVarName(nm, typ, "global", alwaysDefined)
	})
}

type FuncParam struct {
	IsRef bool
	Name  string
	Typ   *TypesMap
}

type FuncInfo struct {
	Pos          ElementPosition
	Params       []FuncParam
	MinParamsCnt int
	Typ          *TypesMap
	AccessLevel  AccessLevel
	ExitFlags    int // if function has exit/die/throw, then ExitFlags will be <> 0
}

type OverrideType int

const (
	// OverrideArgType means that return type of a function is the same as the type of the argument
	OverrideArgType OverrideType = iota
	// OverrideElementType means that return type of a function is the same as the type of an element of the argument
	OverrideElementType
)

type AccessLevel int

const (
	Public AccessLevel = iota
	Protected
	Private
)

func (l AccessLevel) String() string {
	switch l {
	case Public:
		return "public"
	case Protected:
		return "protected"
	case Private:
		return "private"
	}

	panic("Invalid access level")
}

// FuncInfoOverride defines return type overrides based on their parameter types.
// For example, \array_slice($arr) returns type of element (OverrideElementType) of the ArgNum=0
type FuncInfoOverride struct {
	OverrideType OverrideType
	ArgNum       int
}

type PropertyInfo struct {
	Pos         ElementPosition
	Typ         *TypesMap
	AccessLevel AccessLevel
}

type ConstantInfo struct {
	Pos         ElementPosition
	Typ         *TypesMap
	AccessLevel AccessLevel
}

type ClassInfo struct {
	Pos              ElementPosition
	Parent           string
	ParentInterfaces []string // interfaces allow multiple inheritance
	Traits           map[string]struct{}
	Interfaces       map[string]struct{}
	Methods          FunctionsMap
	Properties       PropertiesMap // both instance and static properties are inside. Static properties have "$" prefix
	Constants        ConstantsMap
}

type ClassParseState struct {
	IsTrait                 bool
	Namespace               string
	FunctionUses            map[string]string
	Uses                    map[string]string
	CurrentClass            string
	CurrentParentClass      string
	CurrentParentInterfaces []string // interfaces allow for multiple inheritance...
}

type TraitsMap map[string]ClassInfo
type ClassesMap map[string]ClassInfo
type FunctionsMap map[string]FuncInfo
type FunctionsOverrideMap map[string]FuncInfoOverride
type PropertiesMap map[string]PropertyInfo
type ConstantsMap map[string]ConstantInfo

type ElementPosition struct {
	Filename  string
	Line      int32
	EndLine   int32
	Character int32
	Length    int32 // body length
}

func GetInternalFunctionInfo(fn string) (info FuncInfo, ok bool) {
	info, ok = internalFunctions[fn]
	return info, ok
}

func GetInternalFunctionOverrideInfo(fn string) (info FuncInfoOverride, ok bool) {
	info, ok = internalFunctionOverrides[fn]
	return info, ok
}

var onCompleteCallbacks []func()

func OnIndexingComplete(cb func()) {
	if indexingComplete {
		cb()
	} else {
		onCompleteCallbacks = append(onCompleteCallbacks, cb)
	}
}

func SetIndexingComplete(complete bool) {
	indexingComplete = complete

	if complete {
		for _, cb := range onCompleteCallbacks {
			cb()
		}
	}
}

func IsIndexingComplete() bool {
	return indexingComplete
}

func FullyQualifiedToString(n *name.FullyQualified) string {
	s := make([]string, 1, len(n.Parts)+1)
	for _, v := range n.Parts {
		s = append(s, v.(*name.NamePart).Value)
	}
	return strings.Join(s, `\`)
}

// NameToString returns string like 'NS\SomeClass' for given name node
func NameToString(n *name.Name) string {
	return NamePartsToString(n.Parts)
}

// StringToName creates name node that can be analyzed using solver
func StringToName(nm string) *name.Name {
	var parts []node.Node
	for _, p := range strings.Split(nm, `\`) {
		parts = append(parts, name.NewNamePart(p))
	}
	return name.NewName(parts)
}

// NamePartsToString converts slice of *name.NamePart to string
func NamePartsToString(parts []node.Node) string {
	s := make([]string, 0, len(parts))
	for _, v := range parts {
		s = append(s, v.(*name.NamePart).Value)
	}
	return strings.Join(s, `\`)
}

// NameNodeToString converts nodes of *name.Name, *name.FullyQualified and *node.Identifier to string.
// This function is a helper function to aid printing function names, not for actual code analysis.
func NameNodeToString(n node.Node) string {
	switch n := n.(type) {
	case *name.Name:
		return NameToString(n)
	case *name.FullyQualified:
		return FullyQualifiedToString(n)
	case *node.Identifier:
		return n.Value
	case *expr.Variable:
		return "$" + NameNodeToString(n.VarName)
	default:
		return "<expression>"
	}
}

func NameEquals(n *name.Name, s string) bool {
	if len(n.Parts) != strings.Count(s, `\`)+1 {
		return false
	}

	sParts := strings.Split(s, `\`)

	for i, part := range n.Parts {
		p, ok := part.(*name.NamePart)
		if !ok {
			// d.debug("Unrecognized name part: %T", p)
			return false
		}

		if p.Value != sParts[i] {
			return false
		}
	}

	return true
}
