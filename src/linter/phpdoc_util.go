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

type PHPDocPlace struct {
	Node ir.Node
	Line int
	Part int
	All  bool
}

type PHPDocError struct {
	Place   PHPDocPlace
	Message string
}

func NewPHPDocError(place PHPDocPlace, format string, args ...interface{}) *PHPDocError {
	return &PHPDocError{
		Place:   place,
		Message: fmt.Sprintf(format, args...),
	}
}

type PHPDocErrors struct {
	types []*PHPDocError
	lint  []*PHPDocError
}

func (e *PHPDocErrors) pushType(err *PHPDocError) {
	e.types = append(e.types, err)
}

func (e *PHPDocErrors) pushLint(err *PHPDocError) {
	e.lint = append(e.lint, err)
}

type classPhpDocParseResult struct {
	properties meta.PropertiesMap
	methods    meta.FunctionsMap
	errs       PHPDocErrors
	mixins     []string
}

func parseClassPHPDocMethod(classNode ir.Node, ctx *rootContext, result *classPhpDocParseResult, part *phpdoc.RawCommentPart) {
	// The syntax is:
	//	@method [[static] return type] [name]([[type] [parameter]<, ...>]) [<description>]
	// Return type and method name are mandatory.

	params := part.Params

	static := len(params) > 0 && params[0] == "static"
	if static {
		params = params[1:]
	}

	if len(params) < 2 {
		result.errs.pushLint(
			NewPHPDocError(
				PHPDocPlace{Node: classNode, Line: part.Line(), All: true},
				"@method requires return type and method name fields",
			),
		)
		return
	}

	typ := ctx.phpdocTypeParser.Parse(params[0])
	converted := phpdoctypes.ToRealType(ctx.typeNormalizer.ClassFQNProvider(), typ)
	moveShapesToContext(ctx, converted.Shapes)

	if converted.Warning != "" {
		result.errs.pushType(
			NewPHPDocError(
				PHPDocPlace{Node: classNode, Line: part.Line(), Part: 1},
				converted.Warning,
			),
		)
	}

	var methodName string
	nameEnd := strings.IndexByte(params[1], '(')
	if nameEnd != -1 {
		methodName = params[1][:nameEnd]
	} else {
		methodName = params[1] // Could be a method name without `()`.

		result.errs.pushLint(
			NewPHPDocError(
				PHPDocPlace{Node: classNode, Line: part.Line(), Part: 2},
				"@method missing parentheses after method name",
			),
		)
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

func parseClassPHPDocProperty(classNode ir.Node, ctx *rootContext, result *classPhpDocParseResult, part *phpdoc.TypeVarCommentPart) {
	// The syntax is:
	//	@property [Type] [name] [<description>]
	// Type and name are mandatory.

	if part.Type.IsEmpty() || part.Var == "" {
		result.errs.pushLint(
			NewPHPDocError(
				PHPDocPlace{Node: classNode, Line: part.Line(), All: true},
				"@property requires type and property name fields",
			),
		)
		return
	}

	if part.VarIsFirst {
		result.errs.pushLint(
			NewPHPDocError(
				PHPDocPlace{Node: classNode, Line: part.Line(), All: true},
				"Non-canonical order of name and type",
			),
		)
	}

	converted := phpdoctypes.ToRealType(ctx.typeNormalizer.ClassFQNProvider(), part.Type)
	moveShapesToContext(ctx, converted.Shapes)

	if converted.Warning != "" {
		result.errs.pushType(
			NewPHPDocError(
				PHPDocPlace{Node: classNode, Line: part.Line(), Part: 1},
				converted.Warning,
			),
		)
	}

	if !strings.HasPrefix(part.Var, "$") {
		result.errs.pushLint(
			NewPHPDocError(
				PHPDocPlace{Node: classNode, Line: part.Line(), Part: 2},
				"@property %s field name must start with '$'", part.Var,
			),
		)
		return
	}

	result.properties[part.Var[len("$"):]] = meta.PropertyInfo{
		Typ:         types.NewMapWithNormalization(ctx.typeNormalizer, converted.Types),
		AccessLevel: meta.Public,
		Flags:       meta.PropFromAnnotation,
	}
}

func parseClassPHPDocMixin(classNode ir.Node, cs *meta.ClassParseState, result *classPhpDocParseResult, part *phpdoc.RawCommentPart) {
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
