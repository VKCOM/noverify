package meta

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/types"
)

type FuncFlags uint8

const (
	FuncStatic FuncFlags = 1 << iota
	FuncPure
	FuncAbstract
	FuncFinal
)

type PhpDocInfo struct {
	Deprecated      bool
	DeprecationNote string
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
	Pos            ElementPosition
	Name           string
	Params         []FuncParam
	MinParamsCnt   int
	Typ            types.Map
	AccessLevel    AccessLevel
	Flags          FuncFlags
	ExitFlags      int  // if function has exit/die/throw, then ExitFlags will be <> 0
	FromAnnotation bool // if the method is described in the annotation for the class
	Doc            PhpDocInfo
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
	Pos            ElementPosition
	Typ            types.Map
	AccessLevel    AccessLevel
	FromAnnotation bool // if the property is described in the annotation for the class
}

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

// TODO: rename it; it's not only class-related.
type ClassParseState struct {
	Info *Info

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
