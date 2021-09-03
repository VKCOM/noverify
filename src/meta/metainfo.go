package meta

import (
	"strings"

	"github.com/VKCOM/noverify/src/types"
)

type FuncFlags uint8

const (
	FuncStatic FuncFlags = 1 << iota
	FuncPure
	FuncAbstract
	FuncFinal
	// FuncFromAnnotation is set if the function is described in the class annotation.
	FuncFromAnnotation
)

type PropertyFlags uint8

const (
	// PropFromAnnotation is set if the property is described in the class annotation.
	PropFromAnnotation PropertyFlags = 1 << iota
)

type DeprecationInfo struct {
	Deprecated bool
	Removed    bool

	Reason        string
	Replacement   string
	Since         string
	RemovedReason string
}

func (i *DeprecationInfo) Append(other DeprecationInfo) {
	if !i.Deprecated {
		i.Deprecated = other.Deprecated
	}
	if !i.Removed {
		i.Removed = other.Removed
	}

	if i.Replacement == "" {
		i.Replacement = other.Replacement
	}
	if i.Since == "" {
		i.Since = other.Since
	}
	if i.RemovedReason == "" {
		i.RemovedReason = other.RemovedReason
	}

	i.Reason = other.Reason
}

func (i DeprecationInfo) WithDeprecationNote() bool {
	return i.Reason != "" || i.Replacement != "" || i.Since != "" || i.RemovedReason != ""
}

func (i DeprecationInfo) String() (res string) {
	parts := make([]string, 0, 3)

	if i.Since != "" {
		parts = append(parts, "since: "+i.Since)
	}
	if i.Reason != "" {
		reason := strings.TrimRight(i.Reason, ".!,")
		parts = append(parts, "reason: "+reason)
	}
	if i.Replacement != "" {
		parts = append(parts, "use "+i.Replacement+" instead")
	}
	if i.RemovedReason != "" {
		parts = append(parts, "removed: "+i.RemovedReason)
	}

	return strings.Join(parts, ", ")
}

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

// PerFile contains all meta information about the specified file
type PerFile struct {
	Traits    ClassesMap
	Classes   ClassesMap
	Functions FunctionsMap
	Constants ConstantsMap
}

type FuncParam struct {
	IsRef bool
	Name  string
	Typ   types.Map
}

type FuncInfo struct {
	Pos          ElementPosition
	Name         string
	Params       []FuncParam
	MinParamsCnt int
	Typ          types.Map
	AccessLevel  AccessLevel
	Flags        FuncFlags
	ExitFlags    int // if function has exit/die/throw, then ExitFlags will be <> 0

	DeprecationInfo
}

func (info *FuncInfo) IsStatic() bool         { return info.Flags&FuncStatic != 0 }
func (info *FuncInfo) IsAbstract() bool       { return info.Flags&FuncAbstract != 0 }
func (info *FuncInfo) IsPure() bool           { return info.Flags&FuncPure != 0 }
func (info *FuncInfo) IsFromAnnotation() bool { return info.Flags&FuncFromAnnotation != 0 }
func (info *FuncInfo) IsFinal() bool          { return info.Flags&FuncFinal != 0 }
func (info *FuncInfo) IsDeprecated() bool     { return info.Deprecated }

type OverrideType int

const (
	// OverrideArgType means that return type of a function is the same as the type of the argument
	OverrideArgType OverrideType = iota
	// OverrideElementType means that return type of a function is the same as the type of an element of the argument
	OverrideElementType
	// OverrideClassType means that return type of a function is the same as the type represented by the class name.
	OverrideClassType
	// OverrideNullableClassType means that return type of a function is the same as the type represented by the class name, and is also nullable.
	OverrideNullableClassType
)

type OverrideProperties int

const (
	// NotNull means that the null type will be removed from the resulting type.
	NotNull OverrideProperties = 1 << iota
	// NotFalse means that the false type will be removed from the resulting type.
	NotFalse
	// ArrayOf means that the type will be converted to an array of elements of that type.
	ArrayOf
)

// FuncInfoOverride defines return type overrides based on their parameter types.
// For example, \array_slice($arr) returns type of element (OverrideElementType) of the ArgNum=0
type FuncInfoOverride struct {
	OverrideType OverrideType
	Properties   OverrideProperties
	ArgNum       int
}

type PropertyInfo struct {
	Pos         ElementPosition
	Typ         types.Map
	AccessLevel AccessLevel
	Flags       PropertyFlags
}

func (info *PropertyInfo) IsFromAnnotation() bool { return info.Flags&PropFromAnnotation != 0 }

type ConstInfo struct {
	Pos         ElementPosition
	Typ         types.Map
	AccessLevel AccessLevel
	Value       ConstValue
}

type ClassFlags uint8

const (
	ClassAbstract ClassFlags = 1 << iota
	ClassFinal
	ClassShape
	ClassInterface
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

func (info *ClassInfo) IsAbstract() bool  { return info.Flags&ClassAbstract != 0 }
func (info *ClassInfo) IsFinal() bool     { return info.Flags&ClassFinal != 0 }
func (info *ClassInfo) IsShape() bool     { return info.Flags&ClassShape != 0 }
func (info *ClassInfo) IsInterface() bool { return info.Flags&ClassInterface != 0 }

// TODO: rename it; it's not only class-related.
type ClassParseState struct {
	Info *Info

	IsTrait                 bool
	IsInterface             bool
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
