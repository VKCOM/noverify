package linter

import (
	"fmt"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/solver"
)

type warningString string

// typesFromPHPDoc extracts types out of the PHPDoc type string.
//
// No normalization is performed, but some PHPDoc-specific types
// are simplified to be compatible with our type system.
func typesFromPHPDoc(typ phpdoc.Type) ([]meta.Type, warningString) {
	var conv phpdocTypeConverter
	results := conv.mapType(typ.Expr)
	if conv.nullable {
		results = append(results, meta.Type{Elem: "null"})
	}
	return results, conv.warning
}

type phpdocTypeConverter struct {
	warning  warningString
	nullable bool
}

func (conv *phpdocTypeConverter) mapType(e phpdoc.TypeExpr) []meta.Type {
	switch e.Kind {
	case phpdoc.ExprInvalid, phpdoc.ExprUnknown:
		if e.Value == "-" {
			conv.warn(`expected a type, found '-'; if you want to express 'any' type, use 'mixed'`)
			return []meta.Type{{Elem: "mixed"}}
		}

	case phpdoc.ExprParen:
		return conv.mapType(e.Args[0])

	case phpdoc.ExprName:
		if suggest, ok := typeAliases[e.Value]; ok {
			conv.warn(fmt.Sprintf("use %s type instead of %s", suggest, e.Value))
		}
		return []meta.Type{{Elem: e.Value}}

	case phpdoc.ExprGeneric:
		typ := e.Args[0]
		if typ.Value == "array" {
			params := e.Args[1:]
			switch len(params) {
			case 1:
				return conv.mapArrayType(params[0])
			case 2:
				return conv.mapArrayType(params[1])
			}
		}
		return conv.mapType(typ)

	case phpdoc.ExprNullable:
		conv.nullable = true
		return conv.mapType(e.Args[0])

	case phpdoc.ExprArray:
		if e.Shape == phpdoc.ShapeArrayPrefix {
			conv.warn(e.Value + ": array syntax is T[], not []T")
		}
		return conv.mapArrayType(e.Args[0])

	case phpdoc.ExprUnion:
		types := make([]meta.Type, 0, len(e.Args))
		for _, a := range e.Args {
			types = append(types, conv.mapType(a)...)
		}
		return types
	}

	return nil
}

func (conv *phpdocTypeConverter) mapArrayType(elem phpdoc.TypeExpr) []meta.Type {
	types := conv.mapType(elem)
	if len(types) == 0 {
		return []meta.Type{{Elem: "mixed", Dims: 1}}
	}
	for i := range types {
		types[i].Dims++
	}
	return types
}

func (conv *phpdocTypeConverter) warn(msg string) {
	if conv.warning == "" {
		conv.warning = warningString(msg)
	}
}

// typesFromNode converts type hint node to meta types.
//
// No normalization is performed.
func typesFromNode(typeNode node.Node) []meta.Type {
	n := typeNode

	var results []meta.Type
	if nullable, ok := typeNode.(*node.Nullable); ok {
		n = nullable.Expr
		results = make([]meta.Type, 0, 2)
		results = append(results, meta.Type{Elem: "null"})
	} else {
		results = make([]meta.Type, 0, 1)
	}

	// There is a trick here.
	// Unlike with phpdoc types, having `integer` here
	// means that we need to force it to be interpreted as
	// `\integer`, not as `int`. This is why we prepend `\`.
	typ := meta.Type{Elem: meta.NameNodeToString(n)}
	if _, isAlias := typeAliases[typ.Elem]; isAlias {
		typ.Elem = `\` + typ.Elem
	}

	results = append(results, typ)

	return results
}

type typeNormalizer struct {
	st *meta.ClassParseState
}

func (n typeNormalizer) NormalizeTypes(types []meta.Type) {
	for i := range types {
		n.normalizeType(&types[i])
	}
}

func (n typeNormalizer) string2name(s string) *name.Name {
	// TODO: Can avoid extra work by holding 1 tmp name inside
	// typeNormalizer, since we never need more than one at the time.
	return meta.StringToName(s)
}

func (n typeNormalizer) normalizeType(typ *meta.Type) {
	if trivialTypes[typ.Elem] {
		return
	}

	if typename, ok := typeAliases[typ.Elem]; ok {
		typ.Elem = typename
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
		fullClassName, ok := solver.GetClassName(n.st, n.string2name(typ.Elem))
		if !ok {
			panic(fmt.Sprintf("can't expand type name: '%s'", typ.Elem))
		}
		typ.Elem = fullClassName
	}
}

var trivialTypes = map[string]bool{
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

var typeAliases = map[string]string{
	"integer": "int",
	"long":    "int",

	"boolean": "bool",

	"real":   "float",
	"double": "float",

	"callback": "callable",
}
