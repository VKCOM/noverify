package assign_test

import (
	"testing"

	"gotest.tools/assert"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/assign"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

var nodesToTest = []struct {
	node                node.Node // node
	expectedVisitedKeys []string  // visited keys
	expectedAttributes  map[string]interface{}
}{
	{
		&assign.Reference{
			Variable:   &node.Variable{},
			Expression: &node.Variable{},
		},
		[]string{"Variable", "Expression"},
		nil,
	},
	{
		&assign.Assign{
			Variable:   &node.Variable{},
			Expression: &node.Variable{},
		},
		[]string{"Variable", "Expression"},
		nil,
	},
	{
		&assign.BitwiseAnd{
			Variable:   &node.Variable{},
			Expression: &node.Variable{},
		},
		[]string{"Variable", "Expression"},
		nil,
	},
	{
		&assign.BitwiseOr{
			Variable:   &node.Variable{},
			Expression: &node.Variable{},
		},
		[]string{"Variable", "Expression"},
		nil,
	},
	{
		&assign.BitwiseXor{
			Variable:   &node.Variable{},
			Expression: &node.Variable{},
		},
		[]string{"Variable", "Expression"},
		nil,
	},
	{
		&assign.Concat{
			Variable:   &node.Variable{},
			Expression: &node.Variable{},
		},
		[]string{"Variable", "Expression"},
		nil,
	},
	{
		&assign.Div{
			Variable:   &node.Variable{},
			Expression: &node.Variable{},
		},
		[]string{"Variable", "Expression"},
		nil,
	},
	{
		&assign.Minus{
			Variable:   &node.Variable{},
			Expression: &node.Variable{},
		},
		[]string{"Variable", "Expression"},
		nil,
	},
	{
		&assign.Mod{
			Variable:   &node.Variable{},
			Expression: &node.Variable{},
		},
		[]string{"Variable", "Expression"},
		nil,
	},
	{
		&assign.Mul{
			Variable:   &node.Variable{},
			Expression: &node.Variable{},
		},
		[]string{"Variable", "Expression"},
		nil,
	},
	{
		&assign.Plus{
			Variable:   &node.Variable{},
			Expression: &node.Variable{},
		},
		[]string{"Variable", "Expression"},
		nil,
	},
	{
		&assign.Pow{
			Variable:   &node.Variable{},
			Expression: &node.Variable{},
		},
		[]string{"Variable", "Expression"},
		nil,
	},
	{
		&assign.ShiftLeft{
			Variable:   &node.Variable{},
			Expression: &node.Variable{},
		},
		[]string{"Variable", "Expression"},
		nil,
	},
	{
		&assign.ShiftRight{
			Variable:   &node.Variable{},
			Expression: &node.Variable{},
		},
		[]string{"Variable", "Expression"},
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
		v := &visitorMock{true, []string{}}
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
