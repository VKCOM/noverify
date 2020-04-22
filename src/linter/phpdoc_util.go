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
}

func parseClassPHPDoc(ctx *rootContext, doc string) classPhpDocParseResult {
	var result classPhpDocParseResult

	if doc == "" {
		return result
	}

	// TODO: allocate maps lazily.
	// Class may not have any @property or @method annotations.
	// In that case we can handle avoid map allocations.
	result.properties = make(meta.PropertiesMap)
	result.methods = meta.NewFunctionsMap()

	for _, part := range phpdoc.Parse(doc) {
		switch part.Name {
		case "property":
			parseClassPHPDocProperty(ctx, &result, part)
		case "method":
			parseClassPHPDocMethod(ctx, &result, part)
		}
	}

	return result
}

func parseClassPHPDocMethod(ctx *rootContext, result *classPhpDocParseResult, part phpdoc.CommentPart) {
	// The syntax is:
	//	@method [[static] return type] [name]([[type] [parameter]<, ...>]) [<description>]
	// Return type and method name are mandatory.

	params := part.Params

	static := len(params) > 0 && params[0] == "static"
	if static {
		params = params[1:]
	}

	if len(params) < 2 {
		result.errs.pushLint("line %d: @method requires return type and method name fields", part.Line)
		return
	}

	types, warning := typesFromPHPDoc(ctx.phpdocTypeParser.Parse(params[0]))
	if warning != "" {
		result.errs.pushType("%s on line %d", warning, part.Line)
	}

	var methodName string
	nameEnd := strings.IndexByte(params[1], '(')
	if nameEnd != -1 {
		methodName = params[1][:nameEnd]
	} else {
		methodName = params[1] // Could be a method name without `()`.
		result.errs.pushLint("line %d: @method '(' is not found near the method name", part.Line)
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

func parseClassPHPDocProperty(ctx *rootContext, result *classPhpDocParseResult, part phpdoc.CommentPart) {
	// The syntax is:
	//	@property [Type] [name] [<description>]
	// Type and name are mandatory.

	if len(part.Params) < 2 {
		result.errs.pushLint("line %d: @property requires type and property name fields", part.Line)
		return
	}

	typeString := part.Params[0]
	var nm string
	if len(part.Params) >= 2 {
		nm = part.Params[1]
	} else {
		// Either type or var name is missing.
		if strings.HasPrefix(typeString, "$") {
			result.errs.pushLint("malformed @property %s tag (maybe type is missing?) on line %d",
				part.Params[0], part.Line)
			return
		}
		result.errs.pushLint("malformed @property tag (maybe field name is missing?) on line %d", part.Line)
	}

	if len(part.Params) >= 2 && strings.HasPrefix(typeString, "$") && !strings.HasPrefix(nm, "$") {
		result.errs.pushLint("non-canonical order of name and type on line %d", part.Line)
		nm, typeString = typeString, nm
	}

	types, warning := typesFromPHPDoc(ctx.phpdocTypeParser.Parse(typeString))
	if warning != "" {
		result.errs.pushType("%s on line %d", warning, part.Line)
	}

	if !strings.HasPrefix(nm, "$") {
		result.errs.pushLint("@property %s field name must start with '$' on line %d", nm, part.Line)
		return
	}

	result.properties[nm[len("$"):]] = meta.PropertyInfo{
		Typ:         newTypesMap(ctx, types),
		AccessLevel: meta.Public,
	}
}
