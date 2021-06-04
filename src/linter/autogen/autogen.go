package autogen

import (
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
)

func GenerateShapeName(props []ShapeTypeProp, cs *meta.ClassParseState) string {
	var body string
	for i, prop := range props {
		body += prop.Key
		types := ":"
		for i, typ := range prop.Types {
			types += typ.Elem + strings.Repeat("[]", typ.Dims)
			if i != len(prop.Types)-1 {
				types += ","
			}
		}

		if i != len(props)-1 {
			types += ","
		}

		body += types
	}
	// We'll probably generate names for anonymous classes in the
	// same way in future. All auto-generated names should end with "$".
	// `\shape$` prefix makes it easy to check whether a type
	// is a shape without looking it up inside classes map.
	return fmt.Sprintf(`\shape$%s$`, body)
}

func GenerateClosureName(fun *ir.ClosureExpr, cs *meta.ClassParseState) string {
	pos := ir.GetPosition(fun)
	curFunction := cs.CurrentFunction
	if curFunction != "" {
		curFunction = "," + curFunction
	}
	return fmt.Sprintf("\\Closure$(%s%s):%d$", cs.CurrentFile, curFunction, pos.StartLine)
}
