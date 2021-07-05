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

	if typ.Elem == "any" && n.kphp {
		// `any` is a special KPHP type that is more-or-less
		// identical to `mixed|object`. In PHP, `mixed` already covers
		// objects, so there is no need to add `object`.
		// See https://php.watch/versions/8.0/mixed-type
		typ.Elem = "mixed"
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
