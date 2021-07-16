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
	classDataProvider func(name string) (ClassData, bool)
}

func (c *Compatible) CompatibleTypes(t1, t2 Map) (ok bool, desc string) {
	return compatibleTypes(t1, t2, c)
}

func (c *Compatible) CompatibleType(t1, t2 string) (ok bool, desc string) {
	return compatibleType(t1, t2, c)
}

func compatibleTypes(t1, t2 Map, c *Compatible) (ok bool, desc string) {
	if t1.Empty() || t2.Empty() {
		return true, ""
	}

	ok, desc = compatibleOneWithMany(t1, t2, c)
	if !ok {
		return false, desc
	}

	ok, desc = compatibleOneWithMany(t2, t1, c)
	if !ok {
		return false, desc
	}

	ok, desc = compatibleManyWithMany(t1, t2, c)
	if !ok {
		return false, desc
	}

	return true, ""
}

func compatibleManyWithMany(t1 Map, t2 Map, c *Compatible) (ok bool, desc string) {
	if t1.Len() == 1 || t2.Len() == 1 {
		return true, ""
	}

	if t1.Contains("null") && !t2.Contains("null") {
		return false, fmt.Sprintf("cannot use nullable %s as %s", t1, t2)
	}
	if !t1.Contains("null") && t2.Contains("null") {
		return false, fmt.Sprintf("cannot use %s as nullable %s", t1, t2)
	}

	var compatibleWithOne bool

	t1.Iterate(func(T1Typ string) {
		t2.Iterate(func(T2Typ string) {
			ok, _ = compatibleType(T1Typ, T2Typ, c)
			if ok {
				compatibleWithOne = true
			}
		})
	})

	return compatibleWithOne, ""
}

func compatibleOneWithMany(t1 Map, t2 Map, c *Compatible) (ok bool, desc string) {
	if t1.Len() == 1 {
		if t2.Len() == 1 {
			return compatibleType(t1.String(), t2.String(), c)
		}

		T1S := t1.String()

		if strings.Contains(T1S, "mixed") {
			return true, ""
		}

		T2IsNullable := t2.Find(func(typ string) bool {
			return typ == "null"
		})
		if T2IsNullable {
			return false, fmt.Sprintf("cannot use type %s as nullable type %s", T1S, t2.String())
		}

		var compatibleWithOne bool

		t2.Iterate(func(typ string) {
			ok, desc = compatibleType(T1S, typ, c)
			if ok {
				compatibleWithOne = true
			}
		})

		if !compatibleWithOne {
			return false, fmt.Sprintf("none of the possible types (%s) are compatible with %s", t2.String(), T1S)
		}
	}

	return true, ""
}

func CompatibleType(t1, t2 string) (ok bool, desc string) {
	return compatibleType(t1, t2, nil)
}

func compatibleType(t1, t2 string, c *Compatible) (ok bool, desc string) {
	if t1 == "mixed" || t2 == "mixed" {
		return true, ""
	}

	if IsPOD(t1) && IsPOD(t2) {
		if compatibleBoolean(t1, t2) {
			return true, ""
		}

		if t1 != t2 {
			return false, fmt.Sprintf("cannot use %s as %s", t1, t2)
		}
	}

	T1A, ok := Alias(t1)
	if ok {
		t1 = T1A
	}
	T2A, ok := Alias(t1)
	if ok {
		t2 = T2A
	}

	ok, desc = compatibleClass(t1, t2, c)
	if !ok {
		return false, desc
	}

	ok, desc = compatibleArray(t1, t2, c)
	if !ok {
		return false, desc
	}

	ok, desc = compatibleCallable(t1, t2)
	if !ok {
		return false, desc
	}

	ok, desc = compatibleCallable(t2, t1)
	if !ok {
		return false, desc
	}

	return true, ""
}

func (c *Compatible) compatibleClassWithInheritance(t1, t2 ClassData) (ok bool, desc string) {
	if t1.Name == t2.Name {
		return true, ""
	}

	if t1.Parent != "" {
		T1Parent, ok := c.classDataProvider(t1.Parent)
		if ok {
			ok, desc = c.compatibleClassWithInheritance(T1Parent, t2)
			if ok {
				return true, ""
			} else {
				return false, desc
			}
		}
	}

	if len(t1.Interfaces) > 0 {
		for iface := range t1.Interfaces {
			T1Iface, ok := c.classDataProvider(iface)
			if ok {
				ok, desc = c.compatibleClassWithInheritance(T1Iface, t2)
				if ok {
					return true, ""
				} else {
					return false, desc
				}
			}
		}
	}

	return false, fmt.Sprintf("cannot use class %s as class %s", t1.Name, t2.Name)
}

func compatibleClass(t1 string, t2 string, c *Compatible) (ok bool, desc string) {
	namesEquals := func(t1, t2 string) (ok bool, desc string) {
		if t1 != t2 {
			return false, fmt.Sprintf("cannot use class %s as class %s", t1, t2)
		}

		return true, ""
	}

	if IsClass(t1) && IsClass(t2) {
		if c == nil {
			return namesEquals(t1, t2)
		}

		T1Class, ok := c.classDataProvider(t1)
		if !ok {
			return namesEquals(t1, t2)
		}
		T2Class, ok := c.classDataProvider(t2)
		if !ok {
			return namesEquals(t1, t2)
		}

		ok, _ = c.compatibleClassWithInheritance(T1Class, T2Class)
		if ok {
			return true, ""
		}

		ok, _ = c.compatibleClassWithInheritance(T2Class, T1Class)
		if ok {
			return true, ""
		}

		return false, fmt.Sprintf("cannot use class %s as class %s", t1, t2)
	}

	if IsClass(t1) {
		if t2 == "object" {
			return true, ""
		}

		return false, fmt.Sprintf("cannot use class %s as %s", t1, t2)
	}

	if IsClass(t2) {
		if t1 == "object" {
			return true, ""
		}

		return false, fmt.Sprintf("cannot use %s as class %s", t1, t2)
	}

	return true, ""
}

func compatibleArray(t1 string, t2 string, c *Compatible) (ok bool, desc string) {
	if IsArray(t1) && IsArray(t2) {
		T1El := ArrayElementType(t1)
		T2El := ArrayElementType(t2)

		comp, des := compatibleType(T1El, T2El, c)
		if !comp {
			return false, fmt.Sprintf("cannot use array of %s as array of %s: %s", T1El, T2El, des)
		}

		return true, ""
	}

	if IsArray(t1) {
		if t2 == "iterable" {
			return true, ""
		}

		return false, fmt.Sprintf("cannot use array of %s as %s", t1, t2)
	}

	if IsArray(t2) {
		if t1 == "iterable" {
			return true, ""
		}

		return false, fmt.Sprintf("cannot use %s as array of %s", t1, t2)
	}

	ok, desc = compatibleIterable(t1, t2)
	if !ok {
		return false, desc
	}

	ok, desc = compatibleIterable(t2, t1)
	if !ok {
		return false, desc
	}

	return true, ""
}

func compatibleIterable(t1 string, t2 string) (ok bool, desc string) {
	if t1 == "iterable" {
		if t2 == "iterable" {
			return true, ""
		}

		if IsClass(t2) {
			return true, ""
		}

		return false, fmt.Sprintf("cannot use %s as %s", t1, t2)
	}

	return true, ""
}

func compatibleCallable(t1, t2 string) (ok bool, desc string) {
	if t1 == "callable" {
		if t2 == "string" {
			return true, ""
		}

		if t2 == "callable" || IsClosure(t2) || IsClass(t2) {
			return true, ""
		}

		return false, fmt.Sprintf("cannot use %s as %s", t1, t2)
	}

	return true, ""
}

func compatibleBoolean(t1, t2 string) bool {
	if t1 == "bool" {
		return t2 == "bool" || t2 == "true" || t2 == "false"
	}
	if t2 == "bool" {
		return t1 == "bool" || t1 == "true" || t1 == "false"
	}
	return false
}
