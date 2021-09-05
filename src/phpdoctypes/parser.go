package phpdoctypes

import (
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/types"
)

type Param struct {
	Optional bool
	Typ      types.Map
}

type ParamsMap map[string]Param

type ParseResult struct {
	ReturnType  types.Map
	ParamTypes  ParamsMap
	Deprecation meta.DeprecationInfo
	Inherit     bool

	Shapes   types.ShapesMap
	Closures types.ClosureMap
}

func Parse(doc phpdoc.Comment, actualParams []ir.Node, normalizer types.Normalizer) (result ParseResult) {
	if doc.Raw == "" {
		return result
	}

	result.Shapes = make(types.ShapesMap)
	result.Closures = make(types.ClosureMap)
	result.ParamTypes = make(ParamsMap)

	var curParam int

	for _, rawPart := range doc.Parsed {

		if rawPart.Name() == "deprecated" {
			part := rawPart.(*phpdoc.RawCommentPart)
			result.Deprecation.Deprecated = true
			result.Deprecation.Reason = part.ParamsText
			continue
		}

		if rawPart.Name() == "removed" {
			part := rawPart.(*phpdoc.RawCommentPart)
			result.Deprecation.Removed = true
			result.Deprecation.RemovedReason = part.ParamsText
			continue
		}

		if rawPart.Name() == "see" {
			part := rawPart.(*phpdoc.RawCommentPart)
			if result.Deprecation.Deprecated {
				if result.Deprecation.Replacement != "" {
					result.Deprecation.Replacement += " or " + part.ParamsText
				} else {
					result.Deprecation.Replacement = part.ParamsText
				}
			}
		}

		if rawPart.Name() == "return" {
			part := rawPart.(*phpdoc.TypeCommentPart)

			converted := ToRealType(normalizer.ClassFQNProvider(), part.Type)
			for name, shape := range converted.Shapes {
				result.Shapes[name] = shape
			}
			for name, closure := range converted.Closures {
				result.Closures[name] = closure
			}

			result.ReturnType = types.NewMapWithNormalization(normalizer, converted.Types)
			continue
		}

		// Rest is for @param handling.

		if rawPart.Name() != "param" {
			continue
		}

		part := rawPart.(*phpdoc.TypeVarCommentPart)
		optional := strings.Contains(part.Rest, "[optional]")

		variable := part.Var
		if !strings.HasPrefix(variable, "$") {
			if len(actualParams) > curParam {
				variable = actualParams[curParam].(*ir.Parameter).Variable.Name
			}
		}

		curParam++

		converted := ToRealType(normalizer.ClassFQNProvider(), part.Type)
		for name, shape := range converted.Shapes {
			result.Shapes[name] = shape
		}
		for name, closure := range converted.Closures {
			result.Closures[name] = closure
		}

		var param Param
		param.Typ = types.NewMapWithNormalization(normalizer, converted.Types)
		param.Optional = optional

		variable = strings.TrimPrefix(variable, "$")
		result.ParamTypes[variable] = param
	}

	result.ReturnType = result.ReturnType.Immutable()
	result.Inherit = doc.Inherit
	return result
}
