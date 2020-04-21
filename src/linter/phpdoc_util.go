package linter

import (
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
)

func fixPHPDocType(typ string) (fixed, notice string) {
	var fixer phpdocTypeFixer
	return fixer.Fix(typ)
}

type phpdocTypeFixer struct {
	notice string
}

// Fix tries to return a corrected version of typ.
// If typ was already correct, it returned unchanged.
// Returns first correction notice in addition to the fixed type.
func (f *phpdocTypeFixer) Fix(typ string) (fixed, notice string) {
	f.notice = ""
	fixedTyp := f.fix(typ)
	return fixedTyp, f.notice
}

func (f *phpdocTypeFixer) fix(typ string) string {
	// Check commonly misspelled types and other unfortunate cases.
	switch typ {
	case "callback":
		f.noticef("use callable type instead of callback")
		return "callable"
	case "boolean":
		f.noticef("use bool type instead of boolean")
		return "bool"
	case "double", "real":
		f.noticef("use float type instead of " + typ)
		return "float"
	case "long", "integer":
		f.noticef("use int type instead of " + typ)
		return "int"
	case "[]":
		f.noticef("[] is not a valid type, mixed[] implied")
		return "mixed[]"
	case "array":
		return "mixed[]"
	case "-":
		// This happens when either of these formats is used:
		// `* @param $name - description`
		// `* @param - $name description`
		// We don't want to make "-" slip as a type name.
		f.noticef("expected a type, found '-'; if you want to express 'any' type, use 'mixed'")
		return "mixed"
	case "":
		return "mixed"
	}

	// Fix []T -> T[]
	if strings.HasPrefix(typ, "[]") && typ != "[]" {
		f.noticef("%s type syntax: use [] after the type, e.g. T[]", typ)
		typ = strings.TrimPrefix(typ, "[]")
		typ += "[]"
		return f.fix(typ)
	}

	if strings.HasSuffix(typ, "[]") && typ != "[]" {
		typ = f.fix(strings.TrimSuffix(typ, "[]"))
		return typ + "[]"
	}

	return typ
}

func (f *phpdocTypeFixer) noticef(format string, args ...interface{}) {
	if f.notice == "" {
		f.notice = fmt.Sprintf(format, args...)
	}
}

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

func parseClassPHPDoc(st *meta.ClassParseState, doc string) classPhpDocParseResult {
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
			parseClassPHPDocProperty(st, &result, part)
		case "method":
			parseClassPHPDocMethod(st, &result, part)
		}
	}

	return result
}

func parseClassPHPDocMethod(st *meta.ClassParseState, result *classPhpDocParseResult, part phpdoc.CommentPart) {
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

	typ, err := fixPHPDocType(params[0])
	if err != "" {
		result.errs.pushType("%s on line %d", err, part.Line)
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
		Typ:          meta.NewTypesMap(normalizeType(st, typ)),
		Name:         methodName,
		Flags:        funcFlags,
		MinParamsCnt: 0, // TODO: parse signature and assign a proper value
		AccessLevel:  meta.Public,
	})
}

func parseClassPHPDocProperty(st *meta.ClassParseState, result *classPhpDocParseResult, part phpdoc.CommentPart) {
	// The syntax is:
	//	@property [Type] [name] [<description>]
	// Type and name are mandatory.

	if len(part.Params) < 2 {
		result.errs.pushLint("line %d: @property requires type and property name fields", part.Line)
		return
	}

	typ := part.Params[0]
	var nm string
	if len(part.Params) >= 2 {
		nm = part.Params[1]
	} else {
		// Either type or var name is missing.
		if strings.HasPrefix(typ, "$") {
			result.errs.pushLint("malformed @property %s tag (maybe type is missing?) on line %d",
				part.Params[0], part.Line)
			return
		}
		result.errs.pushLint("malformed @property tag (maybe field name is missing?) on line %d", part.Line)
	}

	if len(part.Params) >= 2 && strings.HasPrefix(typ, "$") && !strings.HasPrefix(nm, "$") {
		result.errs.pushLint("non-canonical order of name and type on line %d", part.Line)
		nm, typ = typ, nm
	}

	typ, err := fixPHPDocType(typ)
	if err != "" {
		result.errs.pushType("%s on line %d", err, part.Line)
	}

	if !strings.HasPrefix(nm, "$") {
		result.errs.pushLint("@property %s field name must start with '$' on line %d", nm, part.Line)
		return
	}

	result.properties[nm[len("$"):]] = meta.PropertyInfo{
		Typ:         meta.NewTypesMap(normalizeType(st, typ)),
		AccessLevel: meta.Public,
	}
}

// parseAngleBracketedType converts types like "array<k1,array<k2,v2>>" (no spaces) to an internal representation.
func parseAngleBracketedType(st *meta.ClassParseState, t string) string {
	if len(t) == 0 {
		return "[error_empty_type]"
	}

	idx := strings.IndexByte(t, '<')
	if idx == -1 {
		return t
	}
	if idx == 0 {
		return "[error_empty_container_name]"
	}
	if t[len(t)-1] != '>' {
		return "[unbalanced_angled_bracket]"
	}

	// e.g. container: "array", rest: "k1,array<k2,v2>"
	container, rest := t[0:idx], t[idx+1:len(t)-1]

	switch container {
	case "array":
		commaIdx := strings.IndexByte(rest, ',')
		if commaIdx == -1 {
			return meta.WrapArrayOf(normalizeType(st, rest))
		}

		ktype, vtype := rest[0:commaIdx], rest[commaIdx+1:]
		if ktype == "" {
			return "[empty_array_key_type]"
		}
		if vtype == "" {
			return "[empty_array_value_type]"
		}

		return meta.WrapArray2(ktype, normalizeType(st, vtype))
	case "list", "non-empty-list":
		return meta.WrapArrayOf(normalizeType(st, rest))
	}

	// unknown container type, just ignoring
	return ""
}
