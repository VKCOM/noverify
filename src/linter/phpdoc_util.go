package linter

import (
	"fmt"
	"strings"
)

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

	return typ
}

func (f *phpdocTypeFixer) noticef(format string, args ...interface{}) {
	if f.notice == "" {
		f.notice = fmt.Sprintf(format, args...)
	}
}
