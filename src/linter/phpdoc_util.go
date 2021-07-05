package linter

import (
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/phpdoctypes"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/types"
)

type phpdocErrors struct {
	phpdocLint []string
	phpdocType []string
}

func (e *phpdocErrors) pushLint(format string, args ...interface{}) {
	e.phpdocLint = append(e.phpdocLint, fmt.Sprintf(format, args...))
}

func (e *phpdocErrors) pushType(format string, args ...interface{}) {
	e.phpdocType = append(e.phpdocType, fmt.Sprintf(format, args...))
}

type classPhpDocParseResult struct {
	properties meta.PropertiesMap
	methods    meta.FunctionsMap
	errs       phpdocErrors
	mixins     []string
}

func parseClassPHPDocMethod(ctx *rootContext, result *classPhpDocParseResult, part *phpdoc.RawCommentPart) {
	// The syntax is:
	//	@method [[static] return type] [name]([[type] [parameter]<, ...>]) [<description>]
	// Return type and method name are mandatory.

	params := part.Params

	static := len(params) > 0 && params[0] == "static"
	if static {
		params = params[1:]
	}

	if len(params) < 2 {
		result.errs.pushLint("line %d: @method requires return type and method name fields", part.Line())
		return
	}

	typ := ctx.phpdocTypeParser.Parse(params[0])
	converted := phpdoctypes.ToRealType(ctx.typeNormalizer.ClassFQNProvider(), typ)
	moveShapesToContext(ctx, converted.Shapes)
	for _, warning := range converted.Warnings {
		result.errs.pushType("%s on line %d", warning, part.Line())
	}

	var methodName string
	nameEnd := strings.IndexByte(params[1], '(')
	if nameEnd != -1 {
		methodName = params[1][:nameEnd]
	} else {
		methodName = params[1] // Could be a method name without `()`.
		result.errs.pushLint("line %d: @method '(' is not found near the method name", part.Line())
	}

	var funcFlags meta.FuncFlags
	if static {
		funcFlags |= meta.FuncStatic
	}
	funcFlags |= meta.FuncFromAnnotation
	result.methods.Set(methodName, meta.FuncInfo{
		Typ:          types.NewMapWithNormalization(ctx.typeNormalizer, converted.Types),
		Name:         methodName,
		Flags:        funcFlags,
		MinParamsCnt: 0, // TODO: parse signature and assign a proper value
		AccessLevel:  meta.Public,
	})
}

func moveShapesToContext(ctx *rootContext, shapes types.ShapesMap) {
	for name, shape := range shapes {
		ctx.shapes[name] = shape
	}
}

func parseClassPHPDocProperty(ctx *rootContext, result *classPhpDocParseResult, part *phpdoc.TypeVarCommentPart) {
	// The syntax is:
	//	@property [Type] [name] [<description>]
	// Type and name are mandatory.

	if part.Type.IsEmpty() || part.Var == "" {
		result.errs.pushLint("line %d: @property requires type and property name fields", part.Line())
		return
	}

	if part.VarIsFirst {
		result.errs.pushLint("non-canonical order of name and type on line %d", part.Line())
	}

	converted := phpdoctypes.ToRealType(ctx.typeNormalizer.ClassFQNProvider(), part.Type)
	moveShapesToContext(ctx, converted.Shapes)
	for _, warning := range converted.Warnings {
		result.errs.pushType("%s on line %d", warning, part.Line())
	}

	if !strings.HasPrefix(part.Var, "$") {
		result.errs.pushLint("@property %s field name must start with '$' on line %d", part.Var, part.Line())
		return
	}

	result.properties[part.Var[len("$"):]] = meta.PropertyInfo{
		Typ:         types.NewMapWithNormalization(ctx.typeNormalizer, converted.Types),
		AccessLevel: meta.Public,
		Flags:       meta.PropFromAnnotation,
	}
}

func parseClassPHPDocMixin(cs *meta.ClassParseState, result *classPhpDocParseResult, part *phpdoc.RawCommentPart) {
	params := part.Params
	if len(params) == 0 {
		return
	}

	name, ok := solver.GetClassName(cs, &ir.Name{
		Value: params[0],
	})

	if !ok {
		return
	}

	result.mixins = append(result.mixins, name)
}
