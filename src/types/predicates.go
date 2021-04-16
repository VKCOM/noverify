package types

import (
	"strings"
)

func IsClass(s string) bool {
	return strings.HasPrefix(s, `\`) && !IsShape(s) && !IsArray(s) && !IsClosure(s)
}

func IsShape(s string) bool {
	return strings.HasPrefix(s, `\shape$`)
}

func IsClosure(s string) bool {
	return strings.HasPrefix(s, `\Closure`)
}

func IsArray(s string) bool {
	return strings.HasSuffix(s, `[]`)
}

func IsTrivial(s string) bool {
	return trivial[s]
}

func IsAlias(s string) bool {
	_, has := aliases[s]
	return has
}

func Alias(s string) (string, bool) {
	alias, has := aliases[s]
	return alias, has
}

func ArrayType(typ string) string {
	return strings.TrimSuffix(typ, "[]")
}

var trivial = map[string]bool{
	"bool":     true,
	"callable": true,
	"float":    true,
	"int":      true,
	"mixed":    true,
	"object":   true,
	"resource": true,
	"string":   true,
	"void":     true,
	"iterable": true,

	"null":  true,
	"true":  true,
	"false": true,
}

var aliases = map[string]string{
	"integer": "int",
	"long":    "int",

	"boolean": "bool",

	"real":   "float",
	"double": "float",

	"callback": "callable",
}
