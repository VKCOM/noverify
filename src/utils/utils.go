package utils

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
)

// NameNodeToString converts nodes of *name.Name, and *node.Identifier to string.
// This function is a helper function to aid printing function names, not for actual code analysis.
func NameNodeToString(n ir.Node) string {
	switch n := n.(type) {
	case *ir.Name:
		return n.Value
	case *ir.Identifier:
		return n.Value
	case *ir.SimpleVar:
		return "$" + n.Name
	case *ir.Var:
		return "$" + NameNodeToString(n.Expr)
	default:
		return "<expression>"
	}
}

// NameNodeEquals checks whether n node name value is identical to s.
func NameNodeEquals(n ir.Node, s string) bool {
	switch n := n.(type) {
	case *ir.Name:
		return n.Value == s
	case *ir.Identifier:
		return n.Value == s
	default:
		return false
	}
}

// IsSpecialClassName checks if the passed node is a special class name.
func IsSpecialClassName(n ir.Node) bool {
	name := NameNodeToString(n)
	return name == "static" || name == "self" || name == "parent"
}

func InVendor(path string) bool {
	return strings.Contains(filepath.ToSlash(path), "/vendor/")
}

func InCoalesceOrIsset(path irutil.NodePath) bool {
	inIsset := false
	_, inCoalesce := path.NthParent(1).(*ir.CoalesceExpr)
	call, inFuncCall := path.NthParent(1).(*ir.FunctionCallExpr)
	if inFuncCall {
		name, ok := call.Function.(*ir.Name)
		if ok {
			inIsset = name.Value == "isset"
		}
	}

	return inCoalesce || inIsset
}

// FindVarNode returns expression variable node root.
// If expression doesn't start from a variable, returns nil.
func FindVarNode(n ir.Node) ir.Node {
	switch n := n.(type) {
	case *ir.Var, *ir.SimpleVar:
		return n
	case *ir.PropertyFetchExpr:
		return FindVarNode(n.Variable)
	case *ir.ArrayDimFetchExpr:
		return FindVarNode(n.Variable)
	default:
		return nil
	}
}

func IsQuote(r rune) bool {
	return r == '"' || r == '\''
}

func VarToString(v ir.Node) string {
	switch t := v.(type) {
	case *ir.SimpleVar:
		return t.Name
	case *ir.Var:
		return "$" + VarToString(t.Expr)
	case *ir.FunctionCallExpr:
		// TODO: support function calls here :)
		return ""
	case *ir.String:
		// Things like ${"x"}
		return "${" + t.Value + "}"
	default:
		return ""
	}
}

// IsCapitalized reports whether s starts with an upper case letter.
func IsCapitalized(s string) bool {
	ch, _ := utf8.DecodeRuneInString(s)
	return unicode.IsUpper(ch)
}

// FindPHPFiles traverses the file system starting with the
// passed folder and finding all files with a .php extension.
func FindPHPFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".php") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

func BinaryOpString(n ir.Node) string {
	switch n.(type) {
	case *ir.BitwiseAndExpr:
		return "&"
	case *ir.BitwiseOrExpr:
		return "|"
	case *ir.BitwiseXorExpr:
		return "^"
	case *ir.LogicalAndExpr:
		return "and"
	case *ir.BooleanAndExpr:
		return "&&"
	case *ir.LogicalOrExpr:
		return "or"
	case *ir.BooleanOrExpr:
		return "||"
	case *ir.LogicalXorExpr:
		return "xor"
	case *ir.PlusExpr:
		return "+"
	case *ir.MinusExpr:
		return "-"
	case *ir.MulExpr:
		return "*"
	case *ir.DivExpr:
		return "/"
	case *ir.ModExpr:
		return "%"
	case *ir.PowExpr:
		return "**"
	case *ir.EqualExpr:
		return "=="
	case *ir.NotEqualExpr:
		return "!="
	case *ir.IdenticalExpr:
		return "==="
	case *ir.NotIdenticalExpr:
		return "!=="
	case *ir.SmallerExpr:
		return "<"
	case *ir.SmallerOrEqualExpr:
		return "<="
	case *ir.GreaterExpr:
		return ">"
	case *ir.GreaterOrEqualExpr:
		return ">="
	case *ir.SpaceshipExpr:
		return "<=>"

	default:
		return ""
	}
}

func IsPHPKeyword(word string) bool {
	return phpKeywords[word]
}

// List taken from https://wiki.php.net/rfc/context_sensitive_lexer
var phpKeywords = map[string]bool{
	"callable":     true,
	"class":        true,
	"trait":        true,
	"extends":      true,
	"implements":   true,
	"static":       true,
	"abstract":     true,
	"final":        true,
	"public":       true,
	"protected":    true,
	"private":      true,
	"const":        true,
	"enddeclare":   true,
	"endfor":       true,
	"endforeach":   true,
	"endif":        true,
	"endwhile":     true,
	"and":          true,
	"global":       true,
	"goto":         true,
	"instanceof":   true,
	"insteadof":    true,
	"interface":    true,
	"namespace":    true,
	"new":          true,
	"or":           true,
	"xor":          true,
	"try":          true,
	"use":          true,
	"var":          true,
	"exit":         true,
	"list":         true,
	"clone":        true,
	"include":      true,
	"include_once": true,
	"throw":        true,
	"array":        true,
	"print":        true,
	"echo":         true,
	"require":      true,
	"require_once": true,
	"return":       true,
	"else":         true,
	"elseif":       true,
	"default":      true,
	"break":        true,
	"continue":     true,
	"switch":       true,
	"yield":        true,
	"function":     true,
	"if":           true,
	"endswitch":    true,
	"finally":      true,
	"for":          true,
	"foreach":      true,
	"declare":      true,
	"case":         true,
	"do":           true,
	"while":        true,
	"as":           true,
	"catch":        true,
	"die":          true,
	"self":         true,
	"parent":       true,
}
