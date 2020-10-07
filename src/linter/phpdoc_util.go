package linter

import (
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
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

	types, warning := typesFromPHPDoc(ctx, ctx.phpdocTypeParser.Parse(params[0]))
	if warning != "" {
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
	result.methods.Set(methodName, meta.FuncInfo{
		Typ:          newTypesMap(ctx, types),
		Name:         methodName,
		Flags:        funcFlags,
		MinParamsCnt: 0, // TODO: parse signature and assign a proper value
		AccessLevel:  meta.Public,
	})
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

	types, warning := typesFromPHPDoc(ctx, part.Type)
	if warning != "" {
		result.errs.pushType("%s on line %d", warning, part.Line())
	}

	if !strings.HasPrefix(part.Var, "$") {
		result.errs.pushLint("@property %s field name must start with '$' on line %d", part.Var, part.Line())
		return
	}

	result.properties[part.Var[len("$"):]] = meta.PropertyInfo{
		Typ:         newTypesMap(ctx, types),
		AccessLevel: meta.Public,
	}
}

func parseClassPHPDocMixin(ctx *rootContext, result *classPhpDocParseResult, part *phpdoc.RawCommentPart) {
	params := part.Params

	if len(params) == 0 {
		return
	}

	param := params[0]
	if !strings.HasPrefix(param, `\`) {
		param = `\` + param
	}
	if ctx.st.Namespace != "" {
		param = ctx.st.Namespace + param
	}

	result.mixins = append(result.mixins, param)
}
