package irfmt

import (
	"bytes"

	"github.com/VKCOM/noverify/src/ir"
)

func Node(n ir.Node) string {
	var b bytes.Buffer
	NewPrettyPrinter(&b, " ").Print(n)
	return b.String()
}
