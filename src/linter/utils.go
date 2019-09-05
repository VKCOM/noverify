package linter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/solver"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/printer"
)

// FmtNode is used for debug purposes and returns string representation of a specified node.
func FmtNode(n node.Node) string {
	var b bytes.Buffer
	printer.NewPrettyPrinter(&b, " ").Print(n)
	return b.String()
}

// FlagsToString is designed for debugging flags.
func FlagsToString(f int) string {
	var res []string

	if (f & FlagReturn) == FlagReturn {
		res = append(res, "Return")
	}

	if (f & FlagDie) == FlagDie {
		res = append(res, "Die")
	}

	if (f & FlagThrow) == FlagThrow {
		res = append(res, "Throw")
	}

	return "Exit flags: " + strings.Join(res, ", ") + ", digits: " + fmt.Sprintf("%d", f)
}

func haveMagicMethod(class string, methodName string) bool {
	_, _, ok := solver.FindMethod(class, methodName)
	return ok
}

func isQuote(r rune) bool {
	return r == '"' || r == '\''
}
