package types

import (
	"fmt"
	"strings"
)

type ClassData struct {
	Name       string
	Parent     string
	Interfaces map[string]struct{}
}

type Compatible struct {
	GetClassByType func(name string) (ClassData, bool)
}

// CompatibleTypes T1 is Wanted, T2 is Gotten
func (c *Compatible) CompatibleTypes(T1, T2 Map) (ok bool, desc string) {
	return compatibleTypes(T1, T2, c)
}

func (c *Compatible) CompatibleType(T1, T2 string) (ok bool, desc string) {
	return compatibleType(T1, T2, c)
}

func (c *Compatible) getClassByType(name string) (*ClassData, bool) {
	class, ok := c.GetClassByType(name)
	if !ok {
		return nil, false
	}

	return &class, true
}

// CompatibleTypes T1 is Wanted, T2 is Gotten
func compatibleTypes(T1, T2 Map, c *Compatible) (ok bool, desc string) {
	if T1.Empty() || T2.Empty() {
		return true, ""
	}

	ok, desc = compatibleOneWithMany(T1, T2, c)
	if !ok {
		return false, desc
	}

	ok, desc = compatibleOneWithMany(T2, T1, c)
	if !ok {
		return false, desc
	}

	ok, desc = compatibleManyWithMany(T1, T2, c)
	if !ok {
		return false, desc
	}

	return true, ""
}

func compatibleManyWithMany(T1 Map, T2 Map, c *Compatible) (ok bool, desc string) {
	if T1.Len() == 1 || T2.Len() == 1 {
		return true, ""
	}

	if T1.Contains("null") && !T2.Contains("null") {
		return false, fmt.Sprintf("cannot use nullable %s as %s", T1, T2)
	}
	if !T1.Contains("null") && T2.Contains("null") {
		return false, fmt.Sprintf("cannot use %s as nullable %s", T1, T2)
	}

	var compatibleWithOne bool

	T1.Iterate(func(T1Typ string) {
		T2.Iterate(func(T2Typ string) {
			ok, _ = compatibleType(T1Typ, T2Typ, c)
			if ok {
				compatibleWithOne = true
			}
		})
	})

	return compatibleWithOne, ""
}

func compatibleOneWithMany(T1 Map, T2 Map, c *Compatible) (ok bool, desc string) {
	if T1.Len() == 1 {
		if T2.Len() == 1 {
			return compatibleType(T1.String(), T2.String(), c)
		}

		T1S := T1.String()

		if strings.Contains(T1S, "mixed") {
			return true, ""
		}

		T2IsNullable := T2.Find(func(typ string) bool {
			return typ == "null"
		})
		if T2IsNullable {
			return false, fmt.Sprintf("cannot use type %s as nullable type %s", T1S, T2.String())
		}

		var compatibleWithOne bool

		T2.Iterate(func(typ string) {
			ok, desc = compatibleType(T1S, typ, c)
			if ok {
				compatibleWithOne = true
			}
		})

		if !compatibleWithOne {
			return false, fmt.Sprintf("none of the possible types (%s) are compatible with %s", T2.String(), T1S)
		}
	}

	return true, ""
}

func CompatibleType(T1, T2 string) (ok bool, desc string) {
	return compatibleType(T1, T2, nil)
}

func compatibleType(T1, T2 string, c *Compatible) (ok bool, desc string) {
	if T1 == "mixed" || T2 == "mixed" {
		return true, ""
	}

	if IsPOD(T1) && IsPOD(T2) {
		if compatibleBoolean(T1, T2) {
			return true, ""
		}

		if T1 != T2 {
			return false, fmt.Sprintf("cannot use %s as %s", T1, T2)
		}
	}

	T1A, ok := Alias(T1)
	if ok {
		T1 = T1A
	}
	T2A, ok := Alias(T1)
	if ok {
		T2 = T2A
	}

	ok, desc = compatibleClass(T1, T2, c)
	if !ok {
		return false, desc
	}

	ok, desc = compatibleArray(T1, T2, c)
	if !ok {
		return false, desc
	}

	ok, desc = compatibleCallable(T1, T2)
	if !ok {
		return false, desc
	}

	ok, desc = compatibleCallable(T2, T1)
	if !ok {
		return false, desc
	}

	return true, ""
}

func (c *Compatible) compatibleClassWithInheritance(T1, T2 ClassData) (ok bool, desc string) {
	if T1.Name == T2.Name {
		return true, ""
	}

	if T1.Parent != "" {
		T1Parent, ok := c.GetClassByType(T1.Parent)
		if ok {
			ok, desc = c.compatibleClassWithInheritance(T1Parent, T2)
			if ok {
				return true, ""
			}
		}
	}

	if len(T1.Interfaces) > 0 {
		for iface := range T1.Interfaces {
			T1Iface, ok := c.GetClassByType(iface)
			if ok {
				ok, desc = c.compatibleClassWithInheritance(T1Iface, T2)
				if ok {
					return true, ""
				}
			}
		}
	}

	return false, fmt.Sprintf("cannot use class %s as class %s", T1.Name, T2.Name)
}

func compatibleClass(T1 string, T2 string, c *Compatible) (ok bool, desc string) {
	namesEquals := func(T1, T2 string) (ok bool, desc string) {
		if T1 != T2 {
			return false, fmt.Sprintf("cannot use class %s as class %s", T1, T2)
		}

		return true, ""
	}

	if IsClass(T1) && IsClass(T2) {
		if c == nil {
			return namesEquals(T1, T2)
		}

		T1Class, ok := c.GetClassByType(T1)
		if !ok {
			return namesEquals(T1, T2)
		}
		T2Class, ok := c.GetClassByType(T2)
		if !ok {
			return namesEquals(T1, T2)
		}

		ok, desc = c.compatibleClassWithInheritance(T1Class, T2Class)
		if ok {
			return true, ""
		}

		ok, desc = c.compatibleClassWithInheritance(T2Class, T1Class)
		if ok {
			return true, ""
		}

		return false, fmt.Sprintf("cannot use class %s as class %s", T1, T2)
	}

	if IsClass(T1) {
		if T2 == "object" {
			return true, ""
		}

		return false, fmt.Sprintf("cannot use class %s as %s", T1, T2)
	}

	if IsClass(T2) {
		if T1 == "object" {
			return true, ""
		}

		return false, fmt.Sprintf("cannot use %s as class %s", T1, T2)
	}

	return true, ""
}

func compatibleArray(T1 string, T2 string, c *Compatible) (ok bool, desc string) {
	if IsArray(T1) && IsArray(T2) {
		T1El := ArrayElementType(T1)
		T2El := ArrayElementType(T2)

		comp, des := compatibleType(T1El, T2El, c)
		if !comp {
			return false, fmt.Sprintf("cannot use array of %s as array of %s: %s", T1El, T2El, des)
		}

		return true, ""
	}

	if IsArray(T1) {
		if T2 == "iterable" {
			return true, ""
		}

		return false, fmt.Sprintf("cannot use array of %s as %s", T1, T2)
	}

	if IsArray(T2) {
		if T1 == "iterable" {
			return true, ""
		}

		return false, fmt.Sprintf("cannot use %s as array of %s", T1, T2)
	}

	ok, desc = compatibleIterable(T1, T2)
	if !ok {
		return false, desc
	}

	ok, desc = compatibleIterable(T2, T1)
	if !ok {
		return false, desc
	}

	return true, ""
}

func compatibleIterable(T1 string, T2 string) (ok bool, desc string) {
	if T1 == "iterable" {
		if T2 == "iterable" {
			return true, ""
		}

		if IsClass(T2) {
			return true, ""
		}

		return false, fmt.Sprintf("cannot use %s as %s", T1, T2)
	}

	return true, ""
}

func compatibleCallable(T1, T2 string) (ok bool, desc string) {
	if T1 == "callable" {
		if T2 == "string" {
			return true, ""
		}

		if T2 == "callable" || IsClosure(T2) || IsClass(T2) {
			return true, ""
		}

		return false, fmt.Sprintf("cannot use %s as %s", T1, T2)
	}

	return true, ""
}

func compatibleBoolean(T1, T2 string) bool {
	if T1 == "bool" {
		return T2 == "bool" || T2 == "true" || T2 == "false"
	}
	if T2 == "bool" {
		return T1 == "bool" || T1 == "true" || T1 == "false"
	}
	return false
}
