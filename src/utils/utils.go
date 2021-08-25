package utils

import (
	"path/filepath"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
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

var typeCheckFunctions = map[string]bool{
	"is_bool":    true,
	"is_double":  true,
	"is_float":   true,
	"is_real":    true,
	"is_int":     true,
	"is_integer": true,
	"is_long":    true,
	"is_object":  true,
	"is_string":  true,
	"is_null":    true,
}

func IsTypeCheckFunctions(n *ir.FunctionCallExpr) bool {
	name, ok := n.Function.(*ir.Name)
	if !ok {
		return false
	}

	return typeCheckFunctions[name.Value]
}
