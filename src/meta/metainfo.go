package meta

import (
	"strings"
	"sync"

	"github.com/VKCOM/noverify/src/ir"
)

var (
	internalFunctions         FunctionsMap
	internalFunctionOverrides FunctionsOverrideMap
	internalClasses           ClassesMap

	indexingComplete bool
	loadingStubs     bool

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
		allTraits:             NewClassesMap(),
		allClasses:            NewClassesMap(),
		allFunctions:          NewFunctionsMap(),
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
	allTraits             ClassesMap
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

func (i *info) GetConstant(nm string) (res ConstInfo, ok bool) {
	res, ok = i.allConstants[nm]
	return res, ok
}

func (i *info) NumConstants() int {
	return len(i.allConstants)
}

func (i *info) GetClass(nm string) (res ClassInfo, ok bool) {
	return i.allClasses.Get(nm)
}

func (i *info) GetTrait(nm string) (res ClassInfo, ok bool) {
	return i.allTraits.Get(nm)
}

func (i *info) GetClassOrTrait(nm string) (res ClassInfo, ok bool) {
	res, ok = i.allClasses.Get(nm)
	if ok {
		return res, true
	}
	res, ok = i.allTraits.Get(nm)
	return res, ok
}

func (i *info) NumClasses() int {
	return i.allClasses.Len()
}

func (i *info) GetFunction(nm string) (res FuncInfo, ok bool) {
	res, ok = i.allFunctions.Get(nm)
	return res, ok
}

func (i *info) GetFunctionOverride(nm string) (res FuncInfoOverride, ok bool) {
	res, ok = i.allFunctionsOverrides[nm]
	return res, ok
}

func (i *info) NumFunctions() int {
	return i.allFunctions.Len()
}

func (i *info) NumFilesWithFunctions() int {
	return len(i.perFileFunctions)
}

func (i *info) FindFunctions(substr string) (res []string) {
	for _, fn := range i.allFunctions.H {
		if strings.HasPrefix(fn.Name, substr) {
			res = append(res, fn.Name)
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

func (i *info) InitKphpStubs() {
	internalFunctions.H[`\array_first_element`] = FuncInfo{
		Name:         `\array_first_element`,
		Params:       []FuncParam{{Name: "el"}},
		MinParamsCnt: 1,
		Typ:          NewTypesMap("mixed"),
	}
	internalFunctions.H[`\array_last_element`] = FuncInfo{
		Name:         `\array_last_element`,
		Params:       []FuncParam{{Name: "el"}},
		MinParamsCnt: 1,
		Typ:          NewTypesMap("mixed"),
	}
	internalFunctions.H[`\instance_deserialize`] = FuncInfo{
		Name:         `\instance_deserialize`,
		Params:       []FuncParam{{Name: "packed_str"}, {Name: "type_of_instance"}},
		MinParamsCnt: 2,
		Typ:          NewTypesMap("object|null"),
	}

	internalFunctionOverrides[`\array_first_element`] = FuncInfoOverride{
		OverrideType: OverrideElementType,
		ArgNum:       0,
	}
	internalFunctionOverrides[`\array_last_element`] = FuncInfoOverride{
		OverrideType: OverrideElementType,
		ArgNum:       0,
	}
	internalFunctionOverrides[`\instance_deserialize`] = FuncInfoOverride{
		OverrideType: OverrideClassType,
		ArgNum:       1,
	}
}

func (i *info) InitStubs() {
	i.Lock()
	defer i.Unlock()

	{
		internalFunctions = NewFunctionsMap()
		h := make(map[lowercaseString]FuncInfo, len(i.allFunctions.H))
		for k, v := range i.allFunctions.H {
			h[k] = v
		}
		internalFunctions.H = h
	}

	{
		internalClasses = NewClassesMap()
		h := make(map[lowercaseString]ClassInfo, len(i.allClasses.H))
		for k, v := range i.allClasses.H {
			h[k] = v
		}
		internalClasses.H = h
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

	for f := range oldClasses.H {
		delete(i.allClasses.H, f)
	}

	oldTraits := i.perFileTraits[filename]
	delete(i.perFileTraits, filename)

	for f := range oldTraits.H {
		delete(i.allTraits.H, f)
	}

	oldFunctions := i.perFileFunctions[filename]
	delete(i.perFileFunctions, filename)

	{
		allFuncs := i.allFunctions.H
		for f, oldFn := range oldFunctions.H {
			fn, ok := allFuncs[f]
			if !ok || oldFn.Pos.Length != fn.Pos.Length {
				continue
			}
			delete(allFuncs, f)
		}
	}

	oldConstants := i.perFileConstants[filename]
	delete(i.perFileConstants, filename)

	for f := range oldConstants {
		delete(i.allConstants, f)
	}
}

func (i *info) AddClassesNonLocked(filename string, m ClassesMap) {
	i.perFileClasses[filename] = m

	allClasses := i.allClasses.H
	for k, v := range m.H {
		prevClass, ok := allClasses[k]
		if !ok || v.Pos.Length > prevClass.Pos.Length {
			allClasses[k] = v
		}
	}
}

func (i *info) AddTraitsNonLocked(filename string, m ClassesMap) {
	i.perFileTraits[filename] = m

	allTraits := i.allTraits.H
	for k, v := range m.H {
		prevTrait, ok := allTraits[k]
		if !ok || v.Pos.Length > prevTrait.Pos.Length {
			allTraits[k] = v
		}
	}
}

func (i *info) AddFunctionsNonLocked(filename string, m FunctionsMap) {
	i.perFileFunctions[filename] = m

	allFuncs := i.allFunctions.H
	for k, v := range m.H {
		prevFn, ok := allFuncs[k]
		if !ok || v.Pos.Length > prevFn.Pos.Length {
			allFuncs[k] = v
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
		// This may cause a name conflict if we have several
		// constants with the same name inside the project.
		// When we'll store a list of symbols for the every name,
		// it won't be a problem anymore.
		i.allConstants[k] = v
	}
}

func (i *info) AddToGlobalScopeNonLocked(filename string, sc *Scope) {
	sc.Iterate(func(nm string, typ TypesMap, flags VarFlags) {
		i.AddVarName(nm, typ, "global", flags)
	})
}

type FuncParam struct {
	IsRef bool
	Name  string
	Typ   TypesMap
}

type PhpDocInfo struct {
	Deprecated      bool
	DeprecationNote string
}

type FuncFlags uint8

const (
	FuncStatic FuncFlags = 1 << iota
	FuncPure
	FuncAbstract
	FuncFinal
)

type FuncInfo struct {
	Pos          ElementPosition
	Name         string
	Params       []FuncParam
	MinParamsCnt int
	Typ          TypesMap
	AccessLevel  AccessLevel
	Flags        FuncFlags
	ExitFlags    int // if function has exit/die/throw, then ExitFlags will be <> 0
	Doc          PhpDocInfo
}

func (info *FuncInfo) IsStatic() bool   { return info.Flags&FuncStatic != 0 }
func (info *FuncInfo) IsAbstract() bool { return info.Flags&FuncAbstract != 0 }
func (info *FuncInfo) IsPure() bool     { return info.Flags&FuncPure != 0 }

type OverrideType int

const (
	// OverrideArgType means that return type of a function is the same as the type of the argument
	OverrideArgType OverrideType = iota
	// OverrideElementType means that return type of a function is the same as the type of an element of the argument
	OverrideElementType
	// OverrideClassType means that return type of a function is the same as the type represented by the class name.
	OverrideClassType
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
	Typ         TypesMap
	AccessLevel AccessLevel
}

type ConstInfo struct {
	Pos         ElementPosition
	Typ         TypesMap
	AccessLevel AccessLevel
	Value       ConstValue
}

type ClassFlags uint8

const (
	ClassAbstract ClassFlags = 1 << iota
	ClassFinal
	ClassShape
)

type ClassInfo struct {
	Pos              ElementPosition
	Name             string
	Flags            ClassFlags
	Parent           string
	ParentInterfaces []string // interfaces allow multiple inheritance
	Traits           map[string]struct{}
	Interfaces       map[string]struct{}
	Methods          FunctionsMap
	Properties       PropertiesMap // both instance and static properties are inside. Static properties have "$" prefix
	Constants        ConstantsMap
	Mixins           []string
}

func (info *ClassInfo) IsAbstract() bool { return info.Flags&ClassAbstract != 0 }
func (info *ClassInfo) IsShape() bool    { return info.Flags&ClassShape != 0 }

type ClassParseState struct {
	IsTrait                 bool
	Namespace               string
	FunctionUses            map[string]string
	Uses                    map[string]string
	CurrentFile             string
	CurrentClass            string
	CurrentParentClass      string
	CurrentParentInterfaces []string // interfaces allow for multiple inheritance...
	CurrentFunction         string   // current method or function name
}

type FunctionsOverrideMap map[string]FuncInfoOverride
type PropertiesMap map[string]PropertyInfo
type ConstantsMap map[string]ConstInfo

type ElementPosition struct {
	Filename  string
	Line      int32
	EndLine   int32
	Character int32
	Length    int32 // body length
}

func IsInternalClass(className string) bool {
	_, ok := internalClasses.Get(className)
	return ok
}

func GetInternalFunctionInfo(fn string) (info FuncInfo, ok bool) {
	return internalFunctions.Get(fn)
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

// SetLoadingStubs changes IsLoadingStubs() return value.
//
// Should be only called from linter.InitStubs() function.
func SetLoadingStubs(loading bool) {
	loadingStubs = loading
}

// IsLoadingStubs reports whether we're parsing stub files right now.
func IsLoadingStubs() bool {
	return loadingStubs
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

// NameNodeToString converts nodes of *name.Name, and *node.Identifier to string.
// This function is a helper function to aid printing function names, not for actual code analysis.
func NameNodeToString(n ir.Node) string {
	switch n := n.(type) {
	case *ir.Name:
		return n.Value
	case *ir.Identifier:
		return n.Value
	case *ir.SimpleVar:
		return "$" + n.Name
	case *ir.Var:
		return "$" + NameNodeToString(n.Expr)
	default:
		return "<expression>"
	}
}

// NameNodeEquals checks whether n node name value is identical to s.
func NameNodeEquals(n ir.Node, s string) bool {
	switch n := n.(type) {
	case *ir.Name:
		return n.Value == s
	case *ir.Identifier:
		return n.Value == s
	default:
		return false
	}
}
