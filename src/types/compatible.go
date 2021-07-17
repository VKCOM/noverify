package types

import (
	"fmt"
	"strings"
)

type ClassData struct {
	Name        string
	Parent      string
	Interfaces  map[string]struct{}
	IsInterface bool
}

type Compatible struct {
	// Checks that T2 is compatible with at least one type of T1.
	// Otherwise, the union types must match strictly.
	OneInMany bool

	ClassDataProvider func(name string) (ClassData, bool)
}

type CompatibleResult struct {
	T1 Map
	T2 Map

	IsCompatible bool

	// If T1 is ?Foo, and T2 is Foo
	MoreGeneralType bool
	// If T1 is Foo, and T2 is ?Foo
	MoreSpecificType bool

	// If T1 is ?int, and T2 is int
	ExtraNullable bool
	// If T1 is int, and T2 is ?int
	LostNullable bool

	// If T1 is float, and T2 is int
	FloatInt bool
	// If T1 is int, and T2 is float
	IntFloat bool

	// If T1 is bool, and T2 is false
	BoolFalse bool
	// If T1 is false, and T2 is bool
	FalseBool bool

	// If T1 is some T, and T2 is array<P>
	TypeAndArray bool
	// If T1 is array<P>, and T2 is some T
	ArrayAndType bool

	// If T1 is array<T>, and T2 is array<P>
	ArraysTypeMismatch bool
	ArrayCheckResult   *CompatibleResult

	// If T1 is T implements I, and T2 is I
	ClassAndInterface bool
	// If T1 is I, and T2 is T implements I
	InterfaceAndClass bool

	// If T1 is T extends P, and T2 is P
	ClassAndParent bool
	// If T1 is P, and T2 is T extends P
	ParentAndClass bool

	// If T1 is Class, and T2 is Not Class
	ClassAndNotClass bool
	// If T1 is Not Class, and T2 is Class
	NotClassAndClass bool

	// All relationships from T1 to T2.
	Description string
}

func (c *Compatible) CompatibleTypes(t1, t2 Map) CompatibleResult {
	res := compatibleTypes(t1, t2, c)
	res.T1 = t1
	res.T2 = t2
	return res
}

func (c *Compatible) CompatibleType(t1, t2 string) CompatibleResult {
	res := compatibleType(t1, t2, c)
	res.T1 = NewMap(t1)
	res.T2 = NewMap(t2)
	return res
}

func compatibleTypes(t1, t2 Map, c *Compatible) (res CompatibleResult) {
	if t1.Empty() || t2.Empty() {
		return CompatibleResult{IsCompatible: true}
	}

	var needNext bool
	res, needNext = compatibleOneWithOne(t1, t2, c)
	if !needNext {
		return res
	}

	// res = compatibleOneWithMany(t1, t2, c)
	// if !res.IsCompatible {
	// 	return res
	// }
	//
	// res = compatibleOneWithMany(t2, t1, c)
	// if !res.IsCompatible {
	// 	return res
	// }

	res = compatibleManyWithMany(t1, t2, c)
	if !res.IsCompatible {
		return res
	}

	return res
}

func compatibleManyWithMany(t1 Map, t2 Map, c *Compatible) (res CompatibleResult) {
	if t1.Contains("null") && !t2.Contains("null") {
		return CompatibleResult{
			ExtraNullable: true,
			Description:   fmt.Sprintf("cannot use nullable %s as %s", t1, t2),
		}
	}
	if !t1.Contains("null") && t2.Contains("null") {
		return CompatibleResult{
			LostNullable: true,
			Description:  fmt.Sprintf("cannot use %s as nullable %s", t1, t2),
		}
	}

	if c.OneInMany {
		var compatibleWithOne bool

		t1.Iterate(func(T1Typ string) {
			if compatibleWithOne {
				return
			}
			t2.Iterate(func(T2Typ string) {
				if compatibleWithOne {
					return
				}
				res = compatibleType(T1Typ, T2Typ, c)
				if res.IsCompatible {
					compatibleWithOne = true
				}
			})
		})

		return CompatibleResult{
			IsCompatible: compatibleWithOne,
		}
	}

	if t1.Len() != t2.Len() {
		return CompatibleResult{
			Description: fmt.Sprintf("cannot use %s as %s", t1, t2),
		}
	}

	return CompatibleResult{
		IsCompatible: true,
	}
}

func compatibleOneWithOne(t1 Map, t2 Map, c *Compatible) (res CompatibleResult, needNext bool) {
	if t1.Len() != 1 || t2.Len() != 1 {
		return CompatibleResult{}, true
	}

	return compatibleType(t1.String(), t2.String(), c), false
}

func compatibleOneWithMany(t1 Map, t2 Map, c *Compatible) (res CompatibleResult) {
	if t1.Len() == 1 {
		T1S := t1.String()

		if strings.Contains(T1S, "mixed") {
			return CompatibleResult{IsCompatible: true}
		}

		T2IsNullable := t2.Find(func(typ string) bool {
			return typ == "null"
		})
		if T2IsNullable {
			return CompatibleResult{
				Description: fmt.Sprintf("cannot use type %s as nullable type %s", T1S, t2.String()),
			}
		}

		var compatibleWithOne bool

		t2.Iterate(func(typ string) {
			res = compatibleType(T1S, typ, c)
			if res.IsCompatible {
				compatibleWithOne = true
			}
		})

		if !compatibleWithOne {
			return CompatibleResult{
				Description: fmt.Sprintf("none of the possible types (%s) are compatible with %s", t2.String(), T1S),
			}
		}
	}

	return CompatibleResult{IsCompatible: true}
}

func CompatibleType(t1, t2 string) CompatibleResult {
	res := compatibleType(t1, t2, nil)
	res.T1 = NewMap(t1)
	res.T2 = NewMap(t2)
	return res
}

func compatibleType(t1, t2 string, c *Compatible) (res CompatibleResult) {
	if t1 == "mixed" || t2 == "mixed" {
		return CompatibleResult{IsCompatible: true}
	}

	T1A, ok := Alias(t1)
	if ok {
		t1 = T1A
	}
	T2A, ok := Alias(t2)
	if ok {
		t2 = T2A
	}

	if IsPOD(t1) && IsPOD(t2) {
		if t1 == "int" && t2 == "float" {
			return CompatibleResult{
				IntFloat: true,
			}
		}
		if t1 == "float" && t2 == "int" {
			return CompatibleResult{
				FloatInt: true,
			}
		}

		var needNext bool
		res, needNext = compatibleBoolean(t1, t2)
		if !needNext {
			return res
		}

		if t1 != t2 {
			return CompatibleResult{IsCompatible: false}
		}

		return CompatibleResult{IsCompatible: true}
	}

	var needNext bool
	res, needNext = compatibleClass(t1, t2, c)
	if !needNext {
		return res
	}

	res = compatibleArray(t1, t2, c)
	if !res.IsCompatible {
		return res
	}

	// res = compatibleCallable(t1, t2)
	// if !res.IsCompatible {
	// 	return res
	// }
	//
	// res = compatibleCallable(t2, t1)
	// if !res.IsCompatible {
	// 	return res
	// }

	return CompatibleResult{IsCompatible: true}
}

func (c *Compatible) classExtendsClass(class, parent ClassData) bool {
	if class.Name == parent.Name {
		return true
	}

	if class.Parent == "" {
		return false
	}

	classParent, ok := c.ClassDataProvider(class.Parent)
	if ok {
		return c.classExtendsClass(classParent, parent)
	}

	return true
}

func (c *Compatible) classImplementInterface(class, iface ClassData) bool {
	if class.Parent != "" {
		parent, ok := c.ClassDataProvider(class.Parent)
		if ok {
			return c.classImplementInterface(parent, iface)
		}
	}

	for classInterface := range class.Interfaces {
		if classInterface == iface.Name {
			return true
		}

		parentInterface, ok := c.ClassDataProvider(classInterface)
		if ok {
			return c.classImplementInterface(parentInterface, iface)
		}
	}

	return false
}

func (c *Compatible) compatibleClassWithInheritance(t1, t2 ClassData) (res CompatibleResult) {
	if t1.Name == t2.Name {
		return CompatibleResult{IsCompatible: true}
	}

	if t1.IsInterface {
		implements := c.classImplementInterface(t2, t1)
		return CompatibleResult{
			IsCompatible:      implements,
			InterfaceAndClass: true,
		}
	}
	if t2.IsInterface {
		implements := c.classImplementInterface(t1, t2)
		return CompatibleResult{
			IsCompatible:      implements,
			ClassAndInterface: true,
		}
	}

	extendsTo := c.classExtendsClass(t1, t2)
	extendsFrom := c.classExtendsClass(t2, t1)

	if extendsTo {
		return CompatibleResult{
			IsCompatible:   true,
			ClassAndParent: true,
		}
	} else if extendsFrom {
		return CompatibleResult{
			IsCompatible:   true,
			ParentAndClass: true,
		}
	}

	return CompatibleResult{IsCompatible: false}
}

func compatibleClass(t1 string, t2 string, c *Compatible) (res CompatibleResult, needNext bool) {
	namesEquals := func(t1, t2 string) CompatibleResult {
		if t1 != t2 {
			return CompatibleResult{
				Description: fmt.Sprintf("cannot use class %s as class %s", t1, t2),
			}
		}

		return CompatibleResult{IsCompatible: true}
	}

	if IsClass(t1) && IsClass(t2) {
		if c == nil {
			equal := namesEquals(t1, t2)
			return equal, !equal.IsCompatible
		}

		T1Class, ok := c.ClassDataProvider(t1)
		if !ok {
			equal := namesEquals(t1, t2)
			return equal, !equal.IsCompatible
		}
		T2Class, ok := c.ClassDataProvider(t2)
		if !ok {
			equal := namesEquals(t1, t2)
			return equal, !equal.IsCompatible
		}

		res = c.compatibleClassWithInheritance(T1Class, T2Class)
		if res.IsCompatible {
			return res, false
		}

		return res, false
	}

	if IsClass(t1) {
		if t2 == "object" {
			return CompatibleResult{IsCompatible: true}, false
		}

		return CompatibleResult{
			ClassAndNotClass: true,
		}, false
	}

	if IsClass(t2) {
		if t1 == "object" {
			return CompatibleResult{IsCompatible: true}, false
		}

		return CompatibleResult{
			NotClassAndClass: true,
		}, false
	}

	return CompatibleResult{IsCompatible: true}, true
}

func compatibleArray(t1 string, t2 string, c *Compatible) (res CompatibleResult) {
	if IsArray(t1) && IsArray(t2) {
		T1El := ArrayElementType(t1)
		T2El := ArrayElementType(t2)

		resElement := compatibleType(T1El, T2El, c)
		if !resElement.IsCompatible {
			return CompatibleResult{
				ArraysTypeMismatch: true,
				ArrayCheckResult:   &resElement,
			}
		}

		return CompatibleResult{IsCompatible: true}
	}

	if IsArray(t1) {
		if t2 == "iterable" {
			return CompatibleResult{IsCompatible: true}
		}

		return CompatibleResult{
			ArrayAndType: true,
		}
	}

	if IsArray(t2) {
		if t1 == "iterable" {
			return CompatibleResult{IsCompatible: true}
		}

		return CompatibleResult{
			TypeAndArray: true,
		}
	}

	res = compatibleIterable(t1, t2)
	if !res.IsCompatible {
		return res
	}

	res = compatibleIterable(t2, t1)
	if !res.IsCompatible {
		return res
	}

	return CompatibleResult{IsCompatible: true}
}

func compatibleIterable(t1 string, t2 string) CompatibleResult {
	if t1 == "iterable" {
		if t2 == "iterable" {
			return CompatibleResult{IsCompatible: true}
		}

		if IsClass(t2) {
			return CompatibleResult{IsCompatible: true}
		}

		return CompatibleResult{
			Description: fmt.Sprintf("cannot use %s as %s", t1, t2),
		}
	}

	return CompatibleResult{IsCompatible: true}
}

func compatibleCallable(t1, t2 string) (res CompatibleResult) {
	if t1 == "callable" {
		if t2 == "string" {
			return CompatibleResult{IsCompatible: true}
		}

		if t2 == "callable" || IsClosure(t2) || IsClass(t2) {
			return CompatibleResult{IsCompatible: true}
		}

		return CompatibleResult{
			Description: fmt.Sprintf("cannot use %s as %s", t1, t2),
		}
	}

	return CompatibleResult{IsCompatible: true}
}

func compatibleBoolean(t1, t2 string) (res CompatibleResult, needNext bool) {
	if t1 == "bool" {
		if t2 == "bool" || t2 == "true" {
			return CompatibleResult{IsCompatible: true}, false
		}
		if t2 == "false" {
			return CompatibleResult{
				IsCompatible: false,
				BoolFalse:    true,
			}, false
		}
		return CompatibleResult{
			IsCompatible: false,
		}, false
	}
	if t2 == "bool" {
		if t1 == "bool" || t1 == "true" {
			return CompatibleResult{IsCompatible: true}, false
		}
		if t1 == "false" {
			return CompatibleResult{
				IsCompatible: false,
				FalseBool:    true,
			}, false
		}
		return CompatibleResult{
			IsCompatible: false,
		}, false
	}

	return CompatibleResult{}, true
}
