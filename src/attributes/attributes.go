package attributes

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/types"
)

const (
	LevelTypeAware       = `\JetBrains\PhpStorm\Internal\LanguageLevelTypeAware`
	ElementAvailable     = `\JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable`
	DeprecatedAnnotation = `\JetBrains\PhpStorm\Deprecated`
)

func TypeAware(groups []*ir.AttributeGroup, state *meta.ClassParseState) (typ types.Map) {
	Each(groups, func(attr *ir.Attribute) bool {
		if Name(attr, state) == LevelTypeAware && len(attr.Args) > 1 {
			defaultType := NamedStringArgument(attr, "default")
			if defaultType != "" {
				typ = types.NewMap(defaultType)
				return false
			}
		}

		return true
	})

	return typ
}

func Available(groups []*ir.AttributeGroup, state *meta.ClassParseState) (res bool) {
	res = true

	Each(groups, func(attr *ir.Attribute) bool {
		if Name(attr, state) == ElementAvailable && len(attr.Args) > 0 {
			fromVersion := NamedStringArgument(attr, "from")
			if fromVersion != "" {
				res = fromVersion != "8.0" && fromVersion != "8.1"
				return false
			}
			firstArg, ok := attr.Arg(0).Expr.(*ir.String)
			if ok {
				res = firstArg.Value != "8.0" && firstArg.Value != "8.1"
				return false
			}
		}

		return true
	})

	return res
}

func Deprecated(groups []*ir.AttributeGroup, state *meta.ClassParseState) (info meta.DeprecationInfo, ok bool) {
	Each(groups, func(attr *ir.Attribute) bool {
		if Name(attr, state) == DeprecatedAnnotation {
			info.Reason = NamedStringArgument(attr, "reason")
			info.Since = NamedStringArgument(attr, "since")
			info.Replacement = NamedStringArgument(attr, "replacement")

			info.Deprecated = true
		}
		return true
	})

	return info, info.Deprecated
}
