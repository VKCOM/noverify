package types

import (
	"fmt"
)

type Normalizer struct {
	classFQNProvider func(string) (string, bool)
	kphp             bool
}

func NewNormalizer(classFQNProvider func(string) (string, bool), kphp bool) Normalizer {
	return Normalizer{classFQNProvider: classFQNProvider, kphp: kphp}
}

func (n Normalizer) ClassFQNProvider() func(string) (string, bool) {
	return n.classFQNProvider
}

func (n Normalizer) NormalizeTypes(typeList []Type) {
	for i := range typeList {
		n.normalizeType(&typeList[i])
	}
}

func (n Normalizer) normalizeType(typ *Type) {
	if IsTrivial(typ.Elem) {
		return
	}

	if typename, has := Alias(typ.Elem); has {
		typ.Elem = typename
		return
	}

	if n.kphp && (typ.Elem == "any" || typ.Elem == "kmixed" || typ.Elem == "future") {
		// `any` is a special KPHP type that is more-or-less
		// identical to `mixed|object`. In PHP, `mixed` already covers
		// objects, so there is no need to add `object`.
		// See https://php.watch/versions/8.0/mixed-type
		typ.Elem = "mixed"
		return
	}

	// Psalm types.
	// See https://psalm.dev/docs/annotating_code/type_syntax/scalar_types/
	switch typ.Elem {
	case "class-string", "interface-string", "trait-string", "callable-string", "numeric-string",
		"literal-string", "lowercase-string", "non-empty-string", "non-empty-lowercase-string",
		"html-escaped-string", "array-key":
		typ.Elem = "string"
		return
	case "positive-int":
		typ.Elem = "int"
		return
	case "numeric":
		typ.Elem = "float"
		return
	}

	switch typ.Elem {
	case "array":
		// Rewrite `array` to `mixed[]`.
		// If it's `array[]`, it'll become `mixed[][]`.
		typ.Dims++
		typ.Elem = "mixed"
	case "$this":
		// Handle `$this` as `static` alias in phpdoc context.
		typ.Elem = "static"
	case "static":
		// Don't replace `static` phpdoc type annotation too early
		// to make it possible to handle late static binding.
	default:
		if typ.Elem[0] == '\\' {
			return // Already FQN?
		}

		if n.classFQNProvider == nil {
			return
		}

		fullClassName, ok := n.classFQNProvider(typ.Elem)
		if !ok {
			panic(fmt.Sprintf("can't expand type name: '%s'", typ.Elem))
		}
		typ.Elem = fullClassName
	}
}
