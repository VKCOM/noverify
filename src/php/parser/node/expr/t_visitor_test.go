package expr_test

import (
	"testing"

	"gotest.tools/assert"

	"github.com/VKCOM/noverify/src/php/parser/node/stmt"

	"github.com/VKCOM/noverify/src/php/parser/node/name"

	"github.com/VKCOM/noverify/src/php/parser/node/scalar"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

var nodesToTest = []struct {
	node                node.Node // node
	expectedVisitedKeys []string  // visited keys
	expectedAttributes  map[string]interface{}
}{
	{
		&expr.ArrayDimFetch{
			Variable: &node.Variable{},
			Dim:      &scalar.Lnumber{Value: "1"},
		},
		[]string{"Variable", "Dim"},
		nil,
	},
	{
		&expr.ArrayItem{
			Key: &scalar.String{Value: "key"},
			Val: &scalar.Lnumber{Value: "1"},
		},
		[]string{"Key", "Val"},
		nil,
	},
	{
		&expr.Array{
			Items: []*expr.ArrayItem{
				&expr.ArrayItem{},
			},
		},
		[]string{"Items"},
		nil,
	},
	{
		&expr.BitwiseNot{
			Expr: &node.Variable{},
		},
		[]string{"Expr"},
		nil,
	},
	{
		&expr.BooleanNot{
			Expr: &node.Variable{},
		},
		[]string{"Expr"},
		nil,
	},
	{
		&expr.ClassConstFetch{
			Class:        &node.Variable{},
			ConstantName: &node.Identifier{Value: "foo"},
		},
		[]string{"Class", "ConstantName"},
		nil,
	},
	{
		&expr.Clone{
			Expr: &node.Variable{},
		},
		[]string{"Expr"},
		nil,
	},
	{
		&expr.ClosureUse{
			Uses: []node.Node{
				&node.Variable{},
			},
		},
		[]string{"Uses"},
		nil,
	},
	{
		&expr.Closure{
			ReturnsRef:    true,
			Static:        false,
			PhpDocComment: "",
			Params:        []node.Node{&node.Parameter{}},
			ClosureUse:    &expr.ClosureUse{},
			ReturnType:    &name.Name{},
			Stmts:         []node.Node{&stmt.Nop{}},
		},
		[]string{"Params", "ClosureUse", "ReturnType", "Stmts"},
		map[string]interface{}{"ReturnsRef": true, "Static": false, "PhpDocComment": ""},
	},
	{
		&expr.ConstFetch{
			Constant: &node.Identifier{Value: "foo"},
		},
		[]string{"Constant"},
		nil,
	},
	{
		&expr.Empty{
			Expr: &node.Variable{},
		},
		[]string{"Expr"},
		nil,
	},
	{
		&expr.ErrorSuppress{
			Expr: &node.Variable{},
		},
		[]string{"Expr"},
		nil,
	},
	{
		&expr.Eval{
			Expr: &node.Variable{},
		},
		[]string{"Expr"},
		nil,
	},
	{
		&expr.Exit{
			Die:  true,
			Expr: &node.Variable{},
		},
		[]string{"Expr"},
		map[string]interface{}{"Die": true},
	},
	{
		&expr.FunctionCall{
			Function:     &node.Variable{},
			ArgumentList: &node.ArgumentList{},
		},
		[]string{"Function", "ArgumentList"},
		nil,
	},
	{
		&expr.IncludeOnce{
			Expr: &node.Variable{},
		},
		[]string{"Expr"},
		nil,
	},
	{
		&expr.Include{
			Expr: &node.Variable{},
		},
		[]string{"Expr"},
		nil,
	},
	{
		&expr.InstanceOf{
			Expr:  &node.Variable{},
			Class: &name.Name{},
		},
		[]string{"Expr", "Class"},
		nil,
	},
	{
		&expr.Isset{
			Variables: []node.Node{
				&node.Variable{},
			},
		},
		[]string{"Variables"},
		nil,
	},
	{
		&expr.List{
			Items: []*expr.ArrayItem{
				&expr.ArrayItem{},
			},
		},
		[]string{"Items"},
		nil,
	},
	{
		&expr.MethodCall{
			Variable:     &node.Variable{},
			Method:       &node.Identifier{Value: "foo"},
			ArgumentList: &node.ArgumentList{},
		},
		[]string{"Variable", "Method", "ArgumentList"},
		nil,
	},
	{
		&expr.New{
			Class:        &name.Name{},
			ArgumentList: &node.ArgumentList{},
		},
		[]string{"Class", "ArgumentList"},
		nil,
	},
	{
		&expr.PostDec{
			Variable: &node.Variable{},
		},
		[]string{"Variable"},
		nil,
	},
	{
		&expr.PostInc{
			Variable: &node.Variable{},
		},
		[]string{"Variable"},
		nil,
	},
	{
		&expr.PreDec{
			Variable: &node.Variable{},
		},
		[]string{"Variable"},
		nil,
	},
	{
		&expr.PreInc{
			Variable: &node.Variable{},
		},
		[]string{"Variable"},
		nil,
	},
	{
		&expr.Print{
			Expr: &node.Variable{},
		},
		[]string{"Expr"},
		nil,
	},
	{
		&expr.PropertyFetch{
			Variable: &node.Variable{},
			Property: &node.Identifier{Value: "foo"},
		},
		[]string{"Variable", "Property"},
		nil,
	},
	{
		&expr.Reference{
			Variable: &node.Variable{},
		},
		[]string{"Variable"},
		nil,
	},
	{
		&expr.RequireOnce{
			Expr: &node.Variable{},
		},
		[]string{"Expr"},
		nil,
	},
	{
		&expr.Require{
			Expr: &node.Variable{},
		},
		[]string{"Expr"},
		nil,
	},
	{
		&expr.ShellExec{
			Parts: []node.Node{
				&scalar.EncapsedStringPart{},
			},
		},
		[]string{"Parts"},
		nil,
	},
	{
		&expr.StaticCall{
			Class:        &name.Name{},
			Call:         &node.Identifier{Value: "foo"},
			ArgumentList: &node.ArgumentList{},
		},
		[]string{"Class", "Call", "ArgumentList"},
		nil,
	},
	{
		&expr.StaticPropertyFetch{
			Class:    &name.Name{},
			Property: &node.Identifier{Value: "foo"},
		},
		[]string{"Class", "Property"},
		nil,
	},
	{
		&expr.Ternary{
			Condition: &node.Variable{},
			IfTrue:    &node.Variable{},
			IfFalse:   &node.Variable{},
		},
		[]string{"Condition", "IfTrue", "IfFalse"},
		nil,
	},
	{
		&expr.UnaryMinus{
			Expr: &node.Variable{},
		},
		[]string{"Expr"},
		nil,
	},
	{
		&expr.UnaryPlus{
			Expr: &node.Variable{},
		},
		[]string{"Expr"},
		nil,
	},
	{
		&node.Variable{VarName: &node.Identifier{Value: "a"}},
		[]string{"VarName"},
		nil,
	},
	{
		&expr.YieldFrom{
			Expr: &node.Variable{},
		},
		[]string{"Expr"},
		nil,
	},
	{
		&expr.Yield{
			Key:   &node.Variable{},
			Value: &node.Variable{},
		},
		[]string{"Key", "Value"},
		nil,
	},
}

type visitorMock struct {
	visitChildren bool
	visitedKeys   []string
}

func (v *visitorMock) EnterNode(n walker.Walkable) bool { return v.visitChildren }
func (v *visitorMock) LeaveNode(n walker.Walkable)      {}
func (v *visitorMock) EnterChildNode(key string, w walker.Walkable) {
	v.visitedKeys = append(v.visitedKeys, key)
}
func (v *visitorMock) LeaveChildNode(key string, w walker.Walkable) {}
func (v *visitorMock) EnterChildList(key string, w walker.Walkable) {
	v.visitedKeys = append(v.visitedKeys, key)
}
func (v *visitorMock) LeaveChildList(key string, w walker.Walkable) {}

func TestVisitorDisableChildren(t *testing.T) {
	for _, tt := range nodesToTest {
		v := &visitorMock{false, []string{}}
		tt.node.Walk(v)

		expected := []string{}
		actual := v.visitedKeys

		assert.DeepEqual(t, expected, actual)
	}
}

func TestVisitor(t *testing.T) {
	for _, tt := range nodesToTest {
		v := &visitorMock{true, nil}
		tt.node.Walk(v)

		expected := tt.expectedVisitedKeys
		actual := v.visitedKeys

		assert.DeepEqual(t, expected, actual)
	}
}

// test Attributes()

func TestNameAttributes(t *testing.T) {
	for _, tt := range nodesToTest {
		expected := tt.expectedAttributes
		actual := tt.node.Attributes()

		assert.DeepEqual(t, expected, actual)
	}
}
