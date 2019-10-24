package binary_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/binary"
	"github.com/VKCOM/noverify/src/php/parser/walker"
	"gotest.tools/assert"
)

var nodesToTest = []struct {
	node                node.Node // node
	expectedVisitedKeys []string  // visited keys
	expectedAttributes  map[string]interface{}
}{
	{
		&binary.BitwiseAnd{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.BitwiseOr{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.BitwiseXor{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.BooleanAnd{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.BooleanOr{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.Coalesce{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.Concat{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.Div{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.Equal{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.GreaterOrEqual{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.Greater{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.Identical{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.LogicalAnd{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.LogicalOr{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.LogicalXor{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.Minus{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.Mod{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.Mul{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.NotEqual{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.NotIdentical{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.Plus{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.Pow{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.ShiftLeft{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.ShiftRight{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.SmallerOrEqual{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.Smaller{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
		nil,
	},
	{
		&binary.Spaceship{
			Left:  &node.Variable{},
			Right: &node.Variable{},
		},
		[]string{"Left", "Right"},
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
