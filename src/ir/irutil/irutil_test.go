package irutil_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/php/parseutil"
)

func TestParent(t *testing.T) {
	code := `<?php
  class Foo {
    public function bar() {
      $a = 1;
      $b = 2;
    }
  }
  
  function foo($с, string $d) {
    echo $с;
  }
`

	rootAst, _, err := parseutil.Parse([]byte(code))
	if err != nil {
		return
	}

	root := irconv.ConvertNode(rootAst).(*ir.Root)

	varNode := irutil.FindChild(root, func(node ir.Node) bool {
		variable, ok := node.(*ir.SimpleVar)
		return ok && variable.Name == "a"
	})

	barFuncNode := irutil.FindChild(root, func(node ir.Node) bool {
		fun, ok := node.(*ir.ClassMethodStmt)
		return ok && fun.MethodName.Value == "bar"
	})

	if barFuncNode != varNode.Parent().Parent().Parent().Parent() {
		t.Errorf("expected %v to be parent of %v", irutil.FmtNode(barFuncNode), irutil.FmtNode(varNode))
	}

	rootNew := irutil.FindParent(varNode, func(node ir.Node) bool {
		_, ok := node.(*ir.Root)
		return ok
	})

	if rootNew != root {
		t.Errorf("expected root, got %T", rootNew)
	}
}
