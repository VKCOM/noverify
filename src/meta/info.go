package meta

import (
	"strings"
	"sync"

	"github.com/VKCOM/noverify/src/types"
)

// Info contains meta information for all classes, functions, etc.
type Info struct {
	indexingComplete   bool
	loadingStubs       bool
	onIndexingComplete []func(*Info)

	sync.Mutex
	*Scope
	allFiles                  map[string]bool
	allTraits                 ClassesMap
	allClasses                ClassesMap
	allFunctions              FunctionsMap
	allConstants              ConstantsMap
	allFunctionsOverrides     FunctionsOverrideMap
	perFileTraits             map[string]ClassesMap
	perFileClasses            map[string]ClassesMap
	perFileFunctions          map[string]FunctionsMap
	perFileConstants          map[string]ConstantsMap
	internalFunctions         FunctionsMap
	internalFunctionOverrides FunctionsOverrideMap
	internalClasses           ClassesMap
}

func NewInfo() *Info {
	return &Info{
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
}

func (i *Info) OnIndexingComplete(cb func(*Info)) {
	if i.indexingComplete {
		cb(i)
	} else {
		i.onIndexingComplete = append(i.onIndexingComplete, cb)
	}
}

// IsLoadingStubs reports whether we're parsing stub files right now.
func (i *Info) IsLoadingStubs() bool {
	return i.loadingStubs
}

// SetLoadingStubs changes IsLoadingStubs() return value.
//
// Should be only called from linter.InitStubs() function.
func (i *Info) SetLoadingStubs(isLoading bool) {
	i.loadingStubs = isLoading
}

func (i *Info) IsIndexingComplete() bool {
	return i.indexingComplete
}

func (i *Info) SetIndexingComplete(complete bool) {
	i.indexingComplete = complete

	if complete {
		for _, cb := range i.onIndexingComplete {
			cb(i)
		}
	}
}

func (i *Info) GetConstant(nm string) (res ConstInfo, ok bool) {
	res, ok = i.allConstants[nm]
	return res, ok
}

func (i *Info) NumConstants() int {
	return len(i.allConstants)
}

func (i *Info) GetClass(nm string) (res ClassInfo, ok bool) {
	return i.allClasses.Get(nm)
}

func (i *Info) GetTrait(nm string) (res ClassInfo, ok bool) {
	return i.allTraits.Get(nm)
}

func (i *Info) GetClassOrTrait(nm string) (res ClassInfo, ok bool) {
	res, ok = i.allClasses.Get(nm)
	if ok {
		return res, true
	}
	res, ok = i.allTraits.Get(nm)
	return res, ok
}

func (i *Info) NumClasses() int {
	return i.allClasses.Len()
}

func (i *Info) GetFunction(nm string) (res FuncInfo, ok bool) {
	res, ok = i.allFunctions.Get(nm)
	return res, ok
}

func (i *Info) GetFunctionOverride(nm string) (res FuncInfoOverride, ok bool) {
	res, ok = i.allFunctionsOverrides[nm]
	return res, ok
}

func (i *Info) NumFunctions() int {
	return i.allFunctions.Len()
}

func (i *Info) NumFilesWithFunctions() int {
	return len(i.perFileFunctions)
}

func (i *Info) FindFunctions(substr string) (res []string) {
	for _, fn := range i.allFunctions.H {
		if strings.HasPrefix(fn.Name, substr) {
			res = append(res, fn.Name)
		}
	}
	return res
}

func (i *Info) FindConstants(substr string) (res []string) {
	for c := range i.allConstants {
		if strings.HasPrefix(c, substr) {
			res = append(res, c)
		}
	}
	return res
}

func (i *Info) InitKphpStubs() {
	i.internalFunctions.H[`\array_first_element`] = FuncInfo{
		Name:         `\array_first_element`,
		Params:       []FuncParam{{Name: "el"}},
		MinParamsCnt: 1,
		Typ:          types.NewMap("mixed"),
	}
	i.internalFunctions.H[`\array_last_element`] = FuncInfo{
		Name:         `\array_last_element`,
		Params:       []FuncParam{{Name: "el"}},
		MinParamsCnt: 1,
		Typ:          types.NewMap("mixed"),
	}
	i.internalFunctions.H[`\instance_deserialize`] = FuncInfo{
		Name:         `\instance_deserialize`,
		Params:       []FuncParam{{Name: "packed_str"}, {Name: "type_of_instance"}},
		MinParamsCnt: 2,
		Typ:          types.NewMap("object|null"),
	}

	i.internalFunctionOverrides[`\array_first_element`] = FuncInfoOverride{
		OverrideType: OverrideElementType,
		ArgNum:       0,
	}
	i.internalFunctionOverrides[`\array_last_element`] = FuncInfoOverride{
		OverrideType: OverrideElementType,
		ArgNum:       0,
	}
	i.internalFunctionOverrides[`\instance_deserialize`] = FuncInfoOverride{
		OverrideType: OverrideClassType,
		ArgNum:       1,
	}
}

func (i *Info) InitStubs() {
	i.Lock()
	defer i.Unlock()

	{
		i.internalFunctions = NewFunctionsMap()
		h := make(map[lowercaseString]FuncInfo, len(i.allFunctions.H))
		for k, v := range i.allFunctions.H {
			h[k] = v
		}
		i.internalFunctions.H = h
	}

	{
		i.internalClasses = NewClassesMap()
		h := make(map[lowercaseString]ClassInfo, len(i.allClasses.H))
		for k, v := range i.allClasses.H {
			h[k] = v
		}
		i.internalClasses.H = h
	}

	i.internalFunctionOverrides = make(FunctionsOverrideMap)
	for k, v := range i.allFunctionsOverrides {
		i.internalFunctionOverrides[k] = v
	}
}

func (i *Info) AddFilenameNonLocked(filename string) {
	i.allFiles[filename] = true
}

func (i *Info) FileExists(filename string) bool {
	return i.allFiles[filename]
}

func (i *Info) GetMetaForFile(filename string) (res PerFile) {
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

func (i *Info) DeleteMetaForFileNonLocked(filename string) {
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

func (i *Info) AddClassesNonLocked(filename string, m ClassesMap) {
	i.perFileClasses[filename] = m

	allClasses := i.allClasses.H
	for k, v := range m.H {
		prevClass, ok := allClasses[k]
		if !ok || v.Pos.Length > prevClass.Pos.Length {
			allClasses[k] = v
		}
	}
}

func (i *Info) AddTraitsNonLocked(filename string, m ClassesMap) {
	i.perFileTraits[filename] = m

	allTraits := i.allTraits.H
	for k, v := range m.H {
		prevTrait, ok := allTraits[k]
		if !ok || v.Pos.Length > prevTrait.Pos.Length {
			allTraits[k] = v
		}
	}
}

func (i *Info) AddFunctionsNonLocked(filename string, m FunctionsMap) {
	i.perFileFunctions[filename] = m

	allFuncs := i.allFunctions.H
	for k, v := range m.H {
		prevFn, ok := allFuncs[k]
		if !ok || v.Pos.Length > prevFn.Pos.Length {
			allFuncs[k] = v
		}
	}
}

func (i *Info) AddFunctionsOverridesNonLocked(filename string, m FunctionsOverrideMap) {
	// TODO: support filename map

	for k, v := range m {
		i.allFunctionsOverrides[k] = v
	}
}

func (i *Info) AddConstantsNonLocked(filename string, m ConstantsMap) {
	i.perFileConstants[filename] = m

	for k, v := range m {
		// This may cause a name conflict if we have several
		// constants with the same name inside the project.
		// When we'll store a list of symbols for the every name,
		// it won't be a problem anymore.
		i.allConstants[k] = v
	}
}

func (i *Info) AddToGlobalScopeNonLocked(filename string, sc *Scope) {
	sc.Iterate(func(nm string, typ types.Map, flags VarFlags) {
		i.AddVarName(nm, typ, "global", flags)
	})
}

func (i *Info) IsInternalClass(className string) bool {
	_, ok := i.internalClasses.Get(className)
	return ok
}

func (i *Info) GetInternalFunctionInfo(fn string) (info FuncInfo, ok bool) {
	return i.internalFunctions.Get(fn)
}

func (i *Info) GetInternalFunctionOverrideInfo(fn string) (info FuncInfoOverride, ok bool) {
	info, ok = i.internalFunctionOverrides[fn]
	return info, ok
}
