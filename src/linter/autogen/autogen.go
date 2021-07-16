package autogen

import (
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/types"
)

func GenerateShapeName(props []types.ShapeProp) string {
	var body string
	for i, prop := range props {
		body += prop.Key
		typesList := ":"
		for i, typ := range prop.Types {
			typesList += typ.Elem + strings.Repeat("[]", typ.Dims)
			if i != len(prop.Types)-1 {
				typesList += ","
			}
		}

		if i != len(props)-1 {
			typesList += ","
		}

		body += typesList
	}
	// We'll probably generate names for anonymous classes in the
	// same way in future. All auto-generated names should end with "$".
	// `\shape$` prefix makes it easy to check whether a type
	// is a shape without looking it up inside classes map.
	return fmt.Sprintf(`\shape$%s$`, body)
}

func GenerateClosureName(fun *ir.ClosureExpr, currentFunction, currentFile string) string {
	pos := ir.GetPosition(fun)
	curFunction := currentFunction
	if curFunction != "" {
		curFunction = "," + curFunction
	}
	return fmt.Sprintf("\\Closure$(%s%s):%d$", currentFile, curFunction, pos.StartLine)
}

func TransformClosureToReadableName(name string) string {
	name = strings.TrimSuffix(name, "$")
	if types.IsClosureFromPHPDoc(name) {
		name = strings.TrimPrefix(name, `\Closure$(`)
		parts := strings.Split(name, ")")
		args := parts[0]
		returnType := strings.ReplaceAll(parts[1], "/", "|")
		returnType = strings.TrimPrefix(returnType, ":")
		if returnType == "" {
			returnType = "void"
		}
		return fmt.Sprintf("anonymous(%s): %s defined in PHPDoc", args, returnType)
	}

	index := strings.LastIndexByte(name, ':')
	if index == -1 {
		return name
	}
	return fmt.Sprintf("anonymous(...) defined on line %s", name[index+1:])
}
