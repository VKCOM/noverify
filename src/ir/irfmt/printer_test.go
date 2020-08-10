package irfmt

import (
	"bytes"
	"testing"

	"github.com/VKCOM/noverify/src/ir"
)

func TestPrintFile(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "\t")
	p.Print(&ir.Root{
		Stmts: []ir.Node{
			&ir.NamespaceStmt{
				NamespaceName: &ir.Name{
					Parts: []ir.Node{
						&ir.NamePart{Value: "Foo"},
					},
				},
			},
			&ir.ClassStmt{
				Modifiers: []*ir.Identifier{{Value: "abstract"}},
				ClassName: &ir.Identifier{Value: "Bar"},
				Extends: &ir.ClassExtendsStmt{
					ClassName: &ir.Name{
						Parts: []ir.Node{
							&ir.NamePart{Value: "Baz"},
						},
					},
				},
				Stmts: []ir.Node{
					&ir.ClassMethodStmt{
						Modifiers:  []*ir.Identifier{{Value: "public"}},
						MethodName: &ir.Identifier{Value: "greet"},
						Stmt: &ir.StmtList{
							Stmts: []ir.Node{
								&ir.EchoStmt{
									Exprs: []ir.Node{
										&ir.String{Value: "'Hello world'"},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	expected := `<?php
namespace Foo;
abstract class Bar extends Baz
{
	public function greet()
	{
		echo 'Hello world';
	}
}
`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintFileInlineHtml(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.Root{
		Stmts: []ir.Node{
			&ir.InlineHTMLStmt{Value: "<div>HTML</div>"},
			&ir.ExpressionStmt{
				Expr: &ir.Heredoc{
					Label: "<<<\"LBL\"\n",
					Parts: []ir.Node{
						&ir.EncapsedStringPart{Value: "hello world\n"},
					},
				},
			},
		},
	})

	expected := `<div>HTML</div><?php
<<<"LBL"
hello world
LBL;
`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

// node

func TestPrintIdentifier(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.Identifier{Value: "test"})

	if o.String() != `test` {
		t.Errorf("TestPrintIdentifier is failed\n")
	}
}

func TestPrintParameter(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.Parameter{
		ByRef:        false,
		Variadic:     true,
		VariableType: &ir.FullyQualifiedName{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
		Variable:     &ir.SimpleVar{Name: "var"},
		DefaultValue: &ir.String{Value: "'default'"},
	})

	expected := "\\Foo ...$var = 'default'"
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintNullable(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.Nullable{
		Expr: &ir.Parameter{
			ByRef:        false,
			Variadic:     true,
			VariableType: &ir.FullyQualifiedName{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
			Variable:     &ir.SimpleVar{Name: "var"},
			DefaultValue: &ir.String{Value: "'default'"},
		},
	})

	expected := "?\\Foo ...$var = 'default'"
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintArgument(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.Argument{
		IsReference: false,
		Variadic:    true,
		Expr:        &ir.SimpleVar{Name: "var"},
	})

	expected := "...$var"
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}
func TestPrintArgumentByRef(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.Argument{
		IsReference: true,
		Variadic:    false,
		Expr:        &ir.SimpleVar{Name: "var"},
	})

	expected := "&$var"
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

// name

func TestPrintNameNamePart(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamePart{
		Value: "foo",
	})

	expected := "foo"
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintNameName(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.Name{
		Parts: []ir.Node{
			&ir.NamePart{
				Value: "Foo",
			},
			&ir.NamePart{
				Value: "Bar",
			},
		},
	})

	expected := "Foo\\Bar"
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintNameFullyQualified(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.FullyQualifiedName{
		Parts: []ir.Node{
			&ir.NamePart{
				Value: "Foo",
			},
			&ir.NamePart{
				Value: "Bar",
			},
		},
	})

	expected := "\\Foo\\Bar"
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintNameRelative(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.RelativeName{
		Parts: []ir.Node{
			&ir.NamePart{
				Value: "Foo",
			},
			&ir.NamePart{
				Value: "Bar",
			},
		},
	})

	expected := "namespace\\Foo\\Bar"
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

// scalar

func TestPrintScalarLNumber(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.Lnumber{Value: "1"})

	if o.String() != `1` {
		t.Errorf("TestPrintScalarLNumber is failed\n")
	}
}

func TestPrintScalarDNumber(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.Dnumber{Value: ".1"})

	if o.String() != `.1` {
		t.Errorf("TestPrintScalarDNumber is failed\n")
	}
}

func TestPrintScalarString(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.String{Value: "'hello world'"})

	expected := `'hello world'`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintScalarEncapsedStringPart(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.EncapsedStringPart{Value: "hello world"})

	if o.String() != `hello world` {
		t.Errorf("TestPrintScalarEncapsedStringPart is failed\n")
	}
}

func TestPrintScalarEncapsed(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.Encapsed{
		Parts: []ir.Node{
			&ir.EncapsedStringPart{Value: "hello "},
			&ir.SimpleVar{Name: "var"},
			&ir.EncapsedStringPart{Value: " world"},
		},
	})

	if o.String() != `"hello {$var} world"` {
		t.Errorf("TestPrintScalarEncapsed is failed\n")
	}
}

func TestPrintScalarHeredoc(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.Heredoc{
		Label: "<<<LBL\n",
		Parts: []ir.Node{
			&ir.EncapsedStringPart{Value: "hello "},
			&ir.SimpleVar{Name: "var"},
			&ir.EncapsedStringPart{Value: " world\n"},
		},
	})

	expected := `<<<LBL
hello {$var} world
LBL`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintScalarNowdoc(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.Heredoc{
		Label: "<<<'LBL'\n",
		Parts: []ir.Node{
			&ir.EncapsedStringPart{Value: "hello world\n"},
		},
	})

	expected := `<<<'LBL'
hello world
LBL`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintScalarMagicConstant(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.MagicConstant{Value: "__DIR__"})

	if o.String() != `__DIR__` {
		t.Errorf("TestPrintScalarMagicConstant is failed\n")
	}
}

// assign

func TestPrintAssign(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.Assign{
		Variable:   &ir.SimpleVar{Name: "a"},
		Expression: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a = $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintReference(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.AssignReference{
		Variable:   &ir.SimpleVar{Name: "a"},
		Expression: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a =& $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAssignBitwiseAnd(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.AssignBitwiseAnd{
		Variable:   &ir.SimpleVar{Name: "a"},
		Expression: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a &= $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAssignBitwiseOr(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.AssignBitwiseOr{
		Variable:   &ir.SimpleVar{Name: "a"},
		Expression: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a |= $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAssignBitwiseXor(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.AssignBitwiseXor{
		Variable:   &ir.SimpleVar{Name: "a"},
		Expression: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a ^= $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAssignConcat(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.AssignConcat{
		Variable:   &ir.SimpleVar{Name: "a"},
		Expression: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a .= $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAssignDiv(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.AssignDiv{
		Variable:   &ir.SimpleVar{Name: "a"},
		Expression: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a /= $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAssignMinus(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.AssignMinus{
		Variable:   &ir.SimpleVar{Name: "a"},
		Expression: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a -= $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAssignMod(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.AssignMod{
		Variable:   &ir.SimpleVar{Name: "a"},
		Expression: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a %= $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAssignMul(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.AssignMul{
		Variable:   &ir.SimpleVar{Name: "a"},
		Expression: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a *= $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAssignPlus(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.AssignPlus{
		Variable:   &ir.SimpleVar{Name: "a"},
		Expression: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a += $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAssignPow(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.AssignPow{
		Variable:   &ir.SimpleVar{Name: "a"},
		Expression: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a **= $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAssignShiftLeft(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.AssignShiftLeft{
		Variable:   &ir.SimpleVar{Name: "a"},
		Expression: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a <<= $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAssignShiftRight(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.AssignShiftRight{
		Variable:   &ir.SimpleVar{Name: "a"},
		Expression: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a >>= $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

// binary

func TestPrintBinaryBitwiseAnd(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.BitwiseAndExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a & $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryBitwiseOr(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.BitwiseOrExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a | $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryBitwiseXor(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.BitwiseXorExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a ^ $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryBooleanAnd(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.BooleanAndExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a && $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryBooleanOr(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.BooleanOrExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a || $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryCoalesce(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.CoalesceExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a ?? $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryConcat(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ConcatExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a . $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryDiv(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.DivExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a / $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryEqual(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.EqualExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a == $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryGreaterOrEqual(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.GreaterOrEqualExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a >= $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryGreater(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.GreaterExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a > $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryIdentical(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.IdenticalExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a === $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryLogicalAnd(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.LogicalAndExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a and $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryLogicalOr(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.LogicalOrExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a or $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryLogicalXor(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.LogicalXorExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a xor $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryMinus(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.MinusExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a - $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryMod(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ModExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a % $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryMul(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.MulExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a * $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryNotEqual(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NotEqualExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a != $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryNotIdentical(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NotIdenticalExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a !== $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryPlus(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.PlusExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a + $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryPow(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.PowExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a ** $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryShiftLeft(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ShiftLeftExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a << $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinaryShiftRight(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ShiftRightExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a >> $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinarySmallerOrEqual(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.SmallerOrEqualExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a <= $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinarySmaller(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.SmallerExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a < $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBinarySpaceship(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.SpaceshipExpr{
		Left:  &ir.SimpleVar{Name: "a"},
		Right: &ir.SimpleVar{Name: "b"},
	})

	expected := `$a <=> $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

// cast

func TestPrintArray(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.TypeCastExpr{
		Type: "array",
		Expr: &ir.SimpleVar{Name: "var"},
	})

	expected := `(array)$var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintBool(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.TypeCastExpr{
		Type: "bool",
		Expr: &ir.SimpleVar{Name: "var"},
	})

	expected := `(bool)$var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintDouble(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.TypeCastExpr{
		Type: "float",
		Expr: &ir.SimpleVar{Name: "var"},
	})

	expected := `(float)$var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintInt(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.TypeCastExpr{
		Type: "int",
		Expr: &ir.SimpleVar{Name: "var"},
	})

	expected := `(int)$var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintObject(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.TypeCastExpr{
		Type: "object",
		Expr: &ir.SimpleVar{Name: "var"},
	})

	expected := `(object)$var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintString(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.TypeCastExpr{
		Type: "string",
		Expr: &ir.SimpleVar{Name: "var"},
	})

	expected := `(string)$var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintUnset(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.UnsetCastExpr{
		Expr: &ir.SimpleVar{Name: "var"},
	})

	expected := `(unset)$var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

// expr

func TestPrintExprArrayDimFetch(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ArrayDimFetchExpr{
		Variable: &ir.SimpleVar{Name: "var"},
		Dim:      &ir.Lnumber{Value: "1"},
	})

	expected := `$var[1]`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintExprArrayItemWithKey(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ArrayItemExpr{
		Key: &ir.String{Value: "'Hello'"},
		Val: &ir.SimpleVar{Name: "world"},
	})

	expected := `'Hello' => $world`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintExprArrayItem(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ArrayItemExpr{
		Val: &ir.ReferenceExpr{Variable: &ir.SimpleVar{Name: "world"}},
	})

	expected := `&$world`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintExprArray(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ArrayExpr{
		Items: []*ir.ArrayItemExpr{
			{
				Key: &ir.String{Value: "'Hello'"},
				Val: &ir.SimpleVar{Name: "world"},
			},
			{
				Key: &ir.Lnumber{Value: "2"},
				Val: &ir.ReferenceExpr{Variable: &ir.SimpleVar{Name: "var"}},
			},
			{
				Val: &ir.SimpleVar{Name: "var"},
			},
		},
	})

	expected := `array('Hello' => $world, 2 => &$var, $var)`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintExprBitwiseNot(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.BitwiseNotExpr{
		Expr: &ir.SimpleVar{Name: "var"},
	})

	expected := `~$var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintExprBooleanNot(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.BooleanNotExpr{
		Expr: &ir.SimpleVar{Name: "var"},
	})

	expected := `!$var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintExprClassConstFetch(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ClassConstFetchExpr{
		Class:        &ir.SimpleVar{Name: "var"},
		ConstantName: &ir.Identifier{Value: "CONST"},
	})

	expected := `$var::CONST`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintExprClone(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.CloneExpr{
		Expr: &ir.SimpleVar{Name: "var"},
	})

	expected := `clone $var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintExprClosureUse(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ClosureUseExpr{
		Uses: []ir.Node{
			&ir.ReferenceExpr{Variable: &ir.SimpleVar{Name: "foo"}},
			&ir.SimpleVar{Name: "bar"},
		},
	})

	expected := `use (&$foo, $bar)`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintExprClosure(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.ClosureExpr{
				Static:     true,
				ReturnsRef: true,
				Params: []ir.Node{
					&ir.Parameter{
						ByRef:    true,
						Variadic: false,
						Variable: &ir.SimpleVar{Name: "var"},
					},
				},
				ClosureUse: &ir.ClosureUseExpr{
					Uses: []ir.Node{
						&ir.ReferenceExpr{Variable: &ir.SimpleVar{Name: "a"}},
						&ir.SimpleVar{Name: "b"},
					},
				},
				ReturnType: &ir.FullyQualifiedName{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
				Stmts: []ir.Node{
					&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
				},
			},
		},
	})

	expected := `namespace {
    static function &(&$var) use (&$a, $b): \Foo {
        $a;
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintExprConstFetch(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ConstFetchExpr{
		Constant: &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "null"}}},
	})

	expected := "null"
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintEmpty(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.EmptyExpr{Expr: &ir.SimpleVar{Name: "var"}})

	expected := `empty($var)`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrettyPrinterrorSuppress(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ErrorSuppressExpr{Expr: &ir.SimpleVar{Name: "var"}})

	expected := `@$var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintEval(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.EvalExpr{Expr: &ir.SimpleVar{Name: "var"}})

	expected := `eval($var)`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintExit(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ExitExpr{Die: false, Expr: &ir.SimpleVar{Name: "var"}})

	expected := `exit($var)`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintDie(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ExitExpr{Die: true, Expr: &ir.SimpleVar{Name: "var"}})

	expected := `die($var)`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintFunctionCall(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.FunctionCallExpr{
		Function: &ir.SimpleVar{Name: "var"},
		ArgumentList: &ir.ArgumentList{
			Arguments: []ir.Node{
				&ir.Argument{
					IsReference: true,
					Expr:        &ir.SimpleVar{Name: "a"},
				},
				&ir.Argument{
					Variadic: true,
					Expr:     &ir.SimpleVar{Name: "b"},
				},
				&ir.Argument{
					Expr: &ir.SimpleVar{Name: "c"},
				},
			},
		},
	})

	expected := `$var(&$a, ...$b, $c)`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintInclude(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.IncludeExpr{Expr: &ir.String{Value: "'path'"}})

	expected := `include 'path'`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintIncludeOnce(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.IncludeOnceExpr{Expr: &ir.String{Value: "'path'"}})

	expected := `include_once 'path'`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintInstanceOf(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.InstanceOfExpr{
		Expr:  &ir.SimpleVar{Name: "var"},
		Class: &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
	})

	expected := `$var instanceof Foo`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintIsset(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.IssetExpr{
		Variables: []ir.Node{
			&ir.SimpleVar{Name: "a"},
			&ir.SimpleVar{Name: "b"},
		},
	})

	expected := `isset($a, $b)`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintList(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ListExpr{
		Items: []*ir.ArrayItemExpr{
			{
				Val: &ir.SimpleVar{Name: "a"},
			},
			{
				Val: &ir.ListExpr{
					Items: []*ir.ArrayItemExpr{
						{
							Val: &ir.SimpleVar{Name: "b"},
						},
						{
							Val: &ir.SimpleVar{Name: "c"},
						},
					},
				},
			},
		},
	})

	expected := `list($a, list($b, $c))`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintMethodCall(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.MethodCallExpr{
		Variable: &ir.SimpleVar{Name: "foo"},
		Method:   &ir.Identifier{Value: "bar"},
		ArgumentList: &ir.ArgumentList{
			Arguments: []ir.Node{
				&ir.Argument{
					Expr: &ir.SimpleVar{Name: "a"},
				},
				&ir.Argument{
					Expr: &ir.SimpleVar{Name: "b"},
				},
			},
		},
	})

	expected := `$foo->bar($a, $b)`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintNew(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NewExpr{
		Class: &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
		ArgumentList: &ir.ArgumentList{
			Arguments: []ir.Node{
				&ir.Argument{
					Expr: &ir.SimpleVar{Name: "a"},
				},
				&ir.Argument{
					Expr: &ir.SimpleVar{Name: "b"},
				},
			},
		},
	})

	expected := `new Foo($a, $b)`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintPostDec(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.PostDecExpr{
		Variable: &ir.SimpleVar{Name: "var"},
	})

	expected := `$var--`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintPostInc(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.PostIncExpr{
		Variable: &ir.SimpleVar{Name: "var"},
	})

	expected := `$var++`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintPreDec(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.PreDecExpr{
		Variable: &ir.SimpleVar{Name: "var"},
	})

	expected := `--$var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintPreInc(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.PreIncExpr{
		Variable: &ir.SimpleVar{Name: "var"},
	})

	expected := `++$var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintPrint(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.PrintExpr{Expr: &ir.SimpleVar{Name: "var"}})

	expected := `print($var)`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintPropertyFetch(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.PropertyFetchExpr{
		Variable: &ir.SimpleVar{Name: "foo"},
		Property: &ir.Identifier{Value: "bar"},
	})

	expected := `$foo->bar`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintExprReference(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ReferenceExpr{
		Variable: &ir.SimpleVar{Name: "foo"},
	})

	expected := `&$foo`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintRequire(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.RequireExpr{Expr: &ir.String{Value: "'path'"}})

	expected := `require 'path'`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintRequireOnce(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.RequireOnceExpr{Expr: &ir.String{Value: "'path'"}})

	expected := `require_once 'path'`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintShellExec(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ShellExecExpr{
		Parts: []ir.Node{
			&ir.EncapsedStringPart{Value: "hello "},
			&ir.SimpleVar{Name: "world"},
			&ir.EncapsedStringPart{Value: "!"},
		},
	})

	expected := "`hello {$world}!`"
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintExprShortArray(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ArrayExpr{
		ShortSyntax: true,
		Items: []*ir.ArrayItemExpr{
			{
				Key: &ir.String{Value: "'Hello'"},
				Val: &ir.SimpleVar{Name: "world"},
			},
			{
				Key: &ir.Lnumber{Value: "2"},
				Val: &ir.ReferenceExpr{Variable: &ir.SimpleVar{Name: "var"}},
			},
			{
				Val: &ir.SimpleVar{Name: "var"},
			},
		},
	})

	expected := `['Hello' => $world, 2 => &$var, $var]`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintShortList(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ListExpr{
		ShortSyntax: true,
		Items: []*ir.ArrayItemExpr{
			{
				Val: &ir.SimpleVar{Name: "a"},
			},
			{
				Val: &ir.ListExpr{
					Items: []*ir.ArrayItemExpr{
						{
							Val: &ir.SimpleVar{Name: "b"},
						},
						{
							Val: &ir.SimpleVar{Name: "c"},
						},
					},
				},
			},
		},
	})

	expected := `[$a, list($b, $c)]`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStaticCall(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.StaticCallExpr{
		Class: &ir.Identifier{Value: "Foo"},
		Call:  &ir.Identifier{Value: "bar"},
		ArgumentList: &ir.ArgumentList{
			Arguments: []ir.Node{
				&ir.Argument{
					Expr: &ir.SimpleVar{Name: "a"},
				},
				&ir.Argument{
					Expr: &ir.SimpleVar{Name: "b"},
				},
			},
		},
	})

	expected := `Foo::bar($a, $b)`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStaticPropertyFetch(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.StaticPropertyFetchExpr{
		Class:    &ir.Identifier{Value: "Foo"},
		Property: &ir.SimpleVar{Name: "bar"},
	})

	expected := `Foo::$bar`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintTernary(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.TernaryExpr{
		Condition: &ir.SimpleVar{Name: "a"},
		IfFalse:   &ir.SimpleVar{Name: "b"},
	})

	expected := `$a ?: $b`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintTernaryFull(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.TernaryExpr{
		Condition: &ir.SimpleVar{Name: "a"},
		IfTrue:    &ir.SimpleVar{Name: "b"},
		IfFalse:   &ir.SimpleVar{Name: "c"},
	})

	expected := `$a ? $b : $c`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintUnaryMinus(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.UnaryMinusExpr{
		Expr: &ir.SimpleVar{Name: "var"},
	})

	expected := `-$var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintUnaryPlus(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.UnaryPlusExpr{
		Expr: &ir.SimpleVar{Name: "var"},
	})

	expected := `+$var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintVariable(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.Var{Expr: &ir.SimpleVar{Name: "var"}})

	expected := `$$var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintYieldFrom(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.YieldFromExpr{
		Expr: &ir.SimpleVar{Name: "var"},
	})

	expected := `yield from $var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintYield(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.YieldExpr{
		Value: &ir.SimpleVar{Name: "var"},
	})

	expected := `yield $var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintYieldFull(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.YieldExpr{
		Key:   &ir.SimpleVar{Name: "k"},
		Value: &ir.SimpleVar{Name: "var"},
	})

	expected := `yield $k => $var`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

// stmt

func TestPrintAltElseIf(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ElseIfStmt{
		AltSyntax: true,
		Cond:      &ir.SimpleVar{Name: "a"},
		Stmt: &ir.StmtList{
			Stmts: []ir.Node{
				&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "b"}},
			},
		},
	})

	expected := `elseif ($a) :
    $b;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAltElseIfEmpty(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ElseIfStmt{
		AltSyntax: true,
		Cond:      &ir.SimpleVar{Name: "a"},
		Stmt:      &ir.StmtList{},
	})

	expected := `elseif ($a) :`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAltElse(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ElseStmt{
		AltSyntax: true,
		Stmt: &ir.StmtList{
			Stmts: []ir.Node{
				&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "b"}},
			},
		},
	})

	expected := `else :
    $b;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAltElseEmpty(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ElseStmt{
		AltSyntax: true,
		Stmt:      &ir.StmtList{},
	})

	expected := `else :`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAltFor(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.ForStmt{
				AltSyntax: true,
				Init: []ir.Node{
					&ir.SimpleVar{Name: "a"},
				},
				Cond: []ir.Node{
					&ir.SimpleVar{Name: "b"},
				},
				Loop: []ir.Node{
					&ir.SimpleVar{Name: "c"},
				},
				Stmt: &ir.StmtList{
					Stmts: []ir.Node{
						&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "d"}},
					},
				},
			},
		},
	})

	expected := `namespace {
    for ($a; $b; $c) :
        $d;
    endfor;
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAltForeach(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.ForeachStmt{
				AltSyntax: true,
				Expr:      &ir.SimpleVar{Name: "var"},
				Key:       &ir.SimpleVar{Name: "key"},
				Variable:  &ir.ReferenceExpr{Variable: &ir.SimpleVar{Name: "val"}},
				Stmt: &ir.StmtList{
					Stmts: []ir.Node{
						&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "d"}},
					},
				},
			},
		},
	})

	expected := `namespace {
    foreach ($var as $key => &$val) :
        $d;
    endforeach;
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAltIf(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.IfStmt{
				AltSyntax: true,
				Cond:      &ir.SimpleVar{Name: "a"},
				Stmt: &ir.StmtList{
					Stmts: []ir.Node{
						&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "d"}},
					},
				},
				ElseIf: []ir.Node{
					&ir.ElseIfStmt{
						AltSyntax: true,
						Cond:      &ir.SimpleVar{Name: "b"},
						Stmt: &ir.StmtList{
							Stmts: []ir.Node{
								&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "b"}},
							},
						},
					},
					&ir.ElseIfStmt{
						AltSyntax: true,
						Cond:      &ir.SimpleVar{Name: "c"},
						Stmt:      &ir.StmtList{},
					},
				},
				Else: &ir.ElseStmt{
					AltSyntax: true,
					Stmt: &ir.StmtList{
						Stmts: []ir.Node{
							&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "b"}},
						},
					},
				},
			},
		},
	})

	expected := `namespace {
    if ($a) :
        $d;
    elseif ($b) :
        $b;
    elseif ($c) :
    else :
        $b;
    endif;
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtAltSwitch(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.SwitchStmt{
				AltSyntax: true,
				Cond:      &ir.SimpleVar{Name: "var"},
				CaseList: &ir.CaseListStmt{
					Cases: []ir.Node{
						&ir.CaseStmt{
							Cond: &ir.String{Value: "'a'"},
							Stmts: []ir.Node{
								&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
							},
						},
						&ir.CaseStmt{
							Cond: &ir.String{Value: "'b'"},
							Stmts: []ir.Node{
								&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "b"}},
							},
						},
					},
				},
			},
		},
	})

	expected := `namespace {
    switch ($var) :
        case 'a':
            $a;
        case 'b':
            $b;
    endswitch;
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintAltWhile(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.WhileStmt{
				AltSyntax: true,
				Cond:      &ir.SimpleVar{Name: "a"},
				Stmt: &ir.StmtList{
					Stmts: []ir.Node{
						&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "b"}},
					},
				},
			},
		},
	})

	expected := `namespace {
    while ($a) :
        $b;
    endwhile;
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtBreak(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.BreakStmt{
		Expr: &ir.Lnumber{Value: "1"},
	})

	expected := "break 1;"
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtCase(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.CaseStmt{
		Cond: &ir.SimpleVar{Name: "a"},
		Stmts: []ir.Node{
			&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
		},
	})

	expected := `case $a:
    $a;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtCaseEmpty(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.CaseStmt{
		Cond:  &ir.SimpleVar{Name: "a"},
		Stmts: []ir.Node{},
	})

	expected := "case $a:"
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtCatch(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.CatchStmt{
				Types: []ir.Node{
					&ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Exception"}}},
					&ir.FullyQualifiedName{Parts: []ir.Node{&ir.NamePart{Value: "RuntimeException"}}},
				},
				Variable: &ir.SimpleVar{Name: "e"},
				Stmts: []ir.Node{
					&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
				},
			},
		},
	})

	expected := `namespace {
    catch (Exception | \RuntimeException $e) {
        $a;
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtClassMethod(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ClassMethodStmt{
		Modifiers:  []*ir.Identifier{{Value: "public"}},
		ReturnsRef: true,
		MethodName: &ir.Identifier{Value: "foo"},
		Params: []ir.Node{
			&ir.Parameter{
				ByRef:        true,
				VariableType: &ir.Nullable{Expr: &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "int"}}}},
				Variable:     &ir.SimpleVar{Name: "a"},
				DefaultValue: &ir.ConstFetchExpr{Constant: &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "null"}}}},
			},
			&ir.Parameter{
				Variadic: true,
				Variable: &ir.SimpleVar{Name: "b"},
			},
		},
		ReturnType: &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "void"}}},
		Stmt: &ir.StmtList{
			Stmts: []ir.Node{
				&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
			},
		},
	})

	expected := `public function &foo(?int &$a = null, ...$b): void
{
    $a;
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}
func TestPrintStmtAbstractClassMethod(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ClassMethodStmt{
		Modifiers:  []*ir.Identifier{{Value: "public"}},
		ReturnsRef: true,
		MethodName: &ir.Identifier{Value: "foo"},
		Params: []ir.Node{
			&ir.Parameter{
				ByRef:        true,
				VariableType: &ir.Nullable{Expr: &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "int"}}}},
				Variable:     &ir.SimpleVar{Name: "a"},
				DefaultValue: &ir.ConstFetchExpr{Constant: &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "null"}}}},
			},
			&ir.Parameter{
				Variadic: true,
				Variable: &ir.SimpleVar{Name: "b"},
			},
		},
		ReturnType: &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "void"}}},
		Stmt:       &ir.NopStmt{},
	})

	expected := `public function &foo(?int &$a = null, ...$b): void;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtClass(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.ClassStmt{
				Modifiers: []*ir.Identifier{{Value: "abstract"}},
				ClassName: &ir.Identifier{Value: "Foo"},
				Extends: &ir.ClassExtendsStmt{
					ClassName: &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Bar"}}},
				},
				Implements: &ir.ClassImplementsStmt{
					InterfaceNames: []ir.Node{
						&ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Baz"}}},
						&ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Quuz"}}},
					},
				},
				Stmts: []ir.Node{
					&ir.ClassConstListStmt{
						Modifiers: []*ir.Identifier{{Value: "public"}},
						Consts: []ir.Node{
							&ir.ConstantStmt{
								ConstantName: &ir.Identifier{Value: "FOO"},
								Expr:         &ir.String{Value: "'bar'"},
							},
						},
					},
				},
			},
		},
	})

	expected := `namespace {
    abstract class Foo extends Bar implements Baz, Quuz
    {
        public const FOO = 'bar';
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtAnonymousClass(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.ClassStmt{
				Modifiers: []*ir.Identifier{{Value: "abstract"}},
				ArgumentList: &ir.ArgumentList{
					Arguments: []ir.Node{
						&ir.Argument{
							Expr: &ir.SimpleVar{Name: "a"},
						},
						&ir.Argument{
							Expr: &ir.SimpleVar{Name: "b"},
						},
					},
				},
				Extends: &ir.ClassExtendsStmt{
					ClassName: &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Bar"}}},
				},
				Implements: &ir.ClassImplementsStmt{
					InterfaceNames: []ir.Node{
						&ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Baz"}}},
						&ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Quuz"}}},
					},
				},
				Stmts: []ir.Node{
					&ir.ClassConstListStmt{
						Modifiers: []*ir.Identifier{{Value: "public"}},
						Consts: []ir.Node{
							&ir.ConstantStmt{
								ConstantName: &ir.Identifier{Value: "FOO"},
								Expr:         &ir.String{Value: "'bar'"},
							},
						},
					},
				},
			},
		},
	})

	expected := `namespace {
    abstract class($a, $b) extends Bar implements Baz, Quuz
    {
        public const FOO = 'bar';
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtClassConstList(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ClassConstListStmt{
		Modifiers: []*ir.Identifier{{Value: "public"}},
		Consts: []ir.Node{
			&ir.ConstantStmt{
				ConstantName: &ir.Identifier{Value: "FOO"},
				Expr:         &ir.String{Value: "'a'"},
			},
			&ir.ConstantStmt{
				ConstantName: &ir.Identifier{Value: "BAR"},
				Expr:         &ir.String{Value: "'b'"},
			},
		},
	})

	expected := `public const FOO = 'a', BAR = 'b';`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtConstant(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ConstantStmt{
		ConstantName: &ir.Identifier{Value: "FOO"},
		Expr:         &ir.String{Value: "'BAR'"},
	})

	expected := "FOO = 'BAR'"
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtContinue(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ContinueStmt{
		Expr: &ir.Lnumber{Value: "1"},
	})

	expected := `continue 1;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtDeclareStmts(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.StmtList{
		Stmts: []ir.Node{
			&ir.DeclareStmt{
				Consts: []ir.Node{
					&ir.ConstantStmt{
						ConstantName: &ir.Identifier{Value: "FOO"},
						Expr:         &ir.String{Value: "'bar'"},
					},
				},
				Stmt: &ir.StmtList{
					Stmts: []ir.Node{
						&ir.NopStmt{},
					},
				},
			},
		},
	})

	expected := `{
    declare(FOO = 'bar') {
        ;
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtDeclareExpr(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.StmtList{
		Stmts: []ir.Node{
			&ir.DeclareStmt{
				Consts: []ir.Node{
					&ir.ConstantStmt{
						ConstantName: &ir.Identifier{Value: "FOO"},
						Expr:         &ir.String{Value: "'bar'"},
					},
				},
				Stmt: &ir.ExpressionStmt{Expr: &ir.String{Value: "'bar'"}},
			},
		},
	})

	expected := `{
    declare(FOO = 'bar')
        'bar';
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtDeclareNop(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.DeclareStmt{
		Consts: []ir.Node{
			&ir.ConstantStmt{
				ConstantName: &ir.Identifier{Value: "FOO"},
				Expr:         &ir.String{Value: "'bar'"},
			},
		},
		Stmt: &ir.NopStmt{},
	})

	expected := `declare(FOO = 'bar');`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtDefalut(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.DefaultStmt{
		Stmts: []ir.Node{
			&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
		},
	})

	expected := `default:
    $a;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtDefalutEmpty(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.DefaultStmt{
		Stmts: []ir.Node{},
	})

	expected := `default:`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtDo_Expression(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.DoStmt{
				Cond: &ir.Lnumber{Value: "1"},
				Stmt: &ir.ExpressionStmt{
					Expr: &ir.SimpleVar{Name: "a"},
				},
			},
		},
	})

	expected := `namespace {
    do
        $a;
    while (1);
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtDo_StmtList(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.DoStmt{
				Cond: &ir.Lnumber{Value: "1"},
				Stmt: &ir.StmtList{
					Stmts: []ir.Node{
						&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
					},
				},
			},
		},
	})

	expected := `namespace {
    do {
        $a;
    } while (1);
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtEcho(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.EchoStmt{
		Exprs: []ir.Node{
			&ir.SimpleVar{Name: "a"},
			&ir.SimpleVar{Name: "b"},
		},
	})

	expected := `echo $a, $b;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtElseIfStmts(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ElseIfStmt{
		Cond: &ir.SimpleVar{Name: "a"},
		Stmt: &ir.StmtList{
			Stmts: []ir.Node{
				&ir.NopStmt{},
			},
		},
	})

	expected := `elseif ($a) {
    ;
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtElseIfExpr(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ElseIfStmt{
		Cond: &ir.SimpleVar{Name: "a"},
		Stmt: &ir.ExpressionStmt{Expr: &ir.String{Value: "'bar'"}},
	})

	expected := `elseif ($a)
    'bar';`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtElseIfNop(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ElseIfStmt{
		Cond: &ir.SimpleVar{Name: "a"},
		Stmt: &ir.NopStmt{},
	})

	expected := `elseif ($a);`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtElseStmts(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ElseStmt{
		Stmt: &ir.StmtList{
			Stmts: []ir.Node{
				&ir.NopStmt{},
			},
		},
	})

	expected := `else {
    ;
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtElseExpr(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ElseStmt{
		Stmt: &ir.ExpressionStmt{Expr: &ir.String{Value: "'bar'"}},
	})

	expected := `else
    'bar';`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtElseNop(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ElseStmt{
		Stmt: &ir.NopStmt{},
	})

	expected := `else;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintExpression(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}})

	expected := `$a;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtFinally(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.FinallyStmt{
				Stmts: []ir.Node{
					&ir.NopStmt{},
				},
			},
		},
	})

	expected := `namespace {
    finally {
        ;
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtForStmts(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.ForStmt{
				Init: []ir.Node{
					&ir.SimpleVar{Name: "a"},
					&ir.SimpleVar{Name: "b"},
				},
				Cond: []ir.Node{
					&ir.SimpleVar{Name: "c"},
					&ir.SimpleVar{Name: "d"},
				},
				Loop: []ir.Node{
					&ir.SimpleVar{Name: "e"},
					&ir.SimpleVar{Name: "f"},
				},
				Stmt: &ir.StmtList{
					Stmts: []ir.Node{
						&ir.NopStmt{},
					},
				},
			},
		},
	})

	expected := `namespace {
    for ($a, $b; $c, $d; $e, $f) {
        ;
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtForExpr(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.ForStmt{
				Init: []ir.Node{
					&ir.SimpleVar{Name: "a"},
				},
				Cond: []ir.Node{
					&ir.SimpleVar{Name: "b"},
				},
				Loop: []ir.Node{
					&ir.SimpleVar{Name: "c"},
				},
				Stmt: &ir.ExpressionStmt{Expr: &ir.String{Value: "'bar'"}},
			},
		},
	})

	expected := `namespace {
    for ($a; $b; $c)
        'bar';
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtForNop(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ForStmt{
		Init: []ir.Node{
			&ir.SimpleVar{Name: "a"},
		},
		Cond: []ir.Node{
			&ir.SimpleVar{Name: "b"},
		},
		Loop: []ir.Node{
			&ir.SimpleVar{Name: "c"},
		},
		Stmt: &ir.NopStmt{},
	})

	expected := `for ($a; $b; $c);`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtForeachStmts(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.ForeachStmt{
				Expr:     &ir.SimpleVar{Name: "a"},
				Variable: &ir.SimpleVar{Name: "b"},
				Stmt: &ir.StmtList{
					Stmts: []ir.Node{
						&ir.NopStmt{},
					},
				},
			},
		},
	})

	expected := `namespace {
    foreach ($a as $b) {
        ;
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtForeachExpr(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.ForeachStmt{
				Expr:     &ir.SimpleVar{Name: "a"},
				Key:      &ir.SimpleVar{Name: "k"},
				Variable: &ir.SimpleVar{Name: "v"},
				Stmt:     &ir.ExpressionStmt{Expr: &ir.String{Value: "'bar'"}},
			},
		},
	})

	expected := `namespace {
    foreach ($a as $k => $v)
        'bar';
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtForeachNop(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ForeachStmt{
		Expr:     &ir.SimpleVar{Name: "a"},
		Key:      &ir.SimpleVar{Name: "k"},
		Variable: &ir.ReferenceExpr{Variable: &ir.SimpleVar{Name: "v"}},
		Stmt:     &ir.NopStmt{},
	})

	expected := `foreach ($a as $k => &$v);`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtFunction(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.FunctionStmt{
				ReturnsRef:   true,
				FunctionName: &ir.Identifier{Value: "foo"},
				Params: []ir.Node{
					&ir.Parameter{
						ByRef:    true,
						Variadic: false,
						Variable: &ir.SimpleVar{Name: "var"},
					},
				},
				ReturnType: &ir.FullyQualifiedName{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
				Stmts: []ir.Node{
					&ir.NopStmt{},
				},
			},
		},
	})

	expected := `namespace {
    function &foo(&$var): \Foo {
        ;
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtGlobal(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.GlobalStmt{
		Vars: []ir.Node{
			&ir.SimpleVar{Name: "a"},
			&ir.SimpleVar{Name: "b"},
		},
	})

	expected := `global $a, $b;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtGoto(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.GotoStmt{
		Label: &ir.Identifier{Value: "FOO"},
	})

	expected := `goto FOO;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtGroupUse(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.GroupUseStmt{
		UseType: &ir.Identifier{Value: "function"},
		Prefix:  &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
		UseList: []ir.Node{
			&ir.UseStmt{
				Use:   &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Bar"}}},
				Alias: &ir.Identifier{Value: "Baz"},
			},
			&ir.UseStmt{
				Use: &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Quuz"}}},
			},
		},
	})

	expected := `use function Foo\{Bar as Baz, Quuz};`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintHaltCompiler(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.HaltCompilerStmt{})

	expected := `__halt_compiler();`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintIfExpression(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.IfStmt{
				Cond: &ir.SimpleVar{Name: "a"},
				Stmt: &ir.ExpressionStmt{
					Expr: &ir.SimpleVar{Name: "b"},
				},
				ElseIf: []ir.Node{
					&ir.ElseIfStmt{
						Cond: &ir.SimpleVar{Name: "c"},
						Stmt: &ir.StmtList{
							Stmts: []ir.Node{
								&ir.ExpressionStmt{
									Expr: &ir.SimpleVar{Name: "d"},
								},
							},
						},
					},
					&ir.ElseIfStmt{
						Cond: &ir.SimpleVar{Name: "e"},
						Stmt: &ir.NopStmt{},
					},
				},
				Else: &ir.ElseStmt{
					Stmt: &ir.ExpressionStmt{
						Expr: &ir.SimpleVar{Name: "f"},
					},
				},
			},
		},
	})

	expected := `namespace {
    if ($a)
        $b;
    elseif ($c) {
        $d;
    }
    elseif ($e);
    else
        $f;
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintIfStmtList(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.IfStmt{
				Cond: &ir.SimpleVar{Name: "a"},
				Stmt: &ir.StmtList{
					Stmts: []ir.Node{
						&ir.ExpressionStmt{
							Expr: &ir.SimpleVar{Name: "b"},
						},
					},
				},
			},
		},
	})

	expected := `namespace {
    if ($a) {
        $b;
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintIfNop(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.IfStmt{
		Cond: &ir.SimpleVar{Name: "a"},
		Stmt: &ir.NopStmt{},
	})

	expected := `if ($a);`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintInlineHtml(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.InlineHTMLStmt{
		Value: "test",
	})

	expected := `?>test<?php`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintInterface(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.InterfaceStmt{
				InterfaceName: &ir.Identifier{Value: "Foo"},
				Extends: &ir.InterfaceExtendsStmt{
					InterfaceNames: []ir.Node{
						&ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Bar"}}},
						&ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Baz"}}},
					},
				},
				Stmts: []ir.Node{
					&ir.ClassMethodStmt{
						Modifiers:  []*ir.Identifier{{Value: "public"}},
						MethodName: &ir.Identifier{Value: "foo"},
						Params:     []ir.Node{},
						Stmt: &ir.StmtList{
							Stmts: []ir.Node{
								&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
							},
						},
					},
				},
			},
		},
	})

	expected := `namespace {
    interface Foo extends Bar, Baz
    {
        public function foo()
        {
            $a;
        }
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintLabel(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.LabelStmt{
		LabelName: &ir.Identifier{Value: "FOO"},
	})

	expected := `FOO:`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintNamespace(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		NamespaceName: &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
	})

	expected := `namespace Foo;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintNamespaceWithStmts(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.StmtList{
		Stmts: []ir.Node{
			&ir.NamespaceStmt{
				NamespaceName: &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
				Stmts: []ir.Node{
					&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
				},
			},
		},
	})

	expected := `{
    namespace Foo {
        $a;
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintNop(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NopStmt{})

	expected := `;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintPropertyList(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.PropertyListStmt{
		Modifiers: []*ir.Identifier{
			{Value: "public"},
			{Value: "static"},
		},
		Properties: []ir.Node{
			&ir.PropertyStmt{
				Variable: &ir.SimpleVar{Name: "a"},
			},
			&ir.PropertyStmt{
				Variable: &ir.SimpleVar{Name: "b"},
			},
		},
	})

	expected := `public static $a, $b;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintProperty(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.PropertyStmt{
		Variable: &ir.SimpleVar{Name: "a"},
		Expr:     &ir.Lnumber{Value: "1"},
	})

	expected := `$a = 1`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintReturn(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ReturnStmt{
		Expr: &ir.Lnumber{Value: "1"},
	})

	expected := `return 1;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStaticVar(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.StaticVarStmt{
		Variable: &ir.SimpleVar{Name: "a"},
		Expr:     &ir.Lnumber{Value: "1"},
	})

	expected := `$a = 1`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStatic(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.StaticStmt{
		Vars: []ir.Node{
			&ir.StaticVarStmt{
				Variable: &ir.SimpleVar{Name: "a"},
			},
			&ir.StaticVarStmt{
				Variable: &ir.SimpleVar{Name: "b"},
			},
		},
	})

	expected := `static $a, $b;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtList(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.StmtList{
		Stmts: []ir.Node{
			&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
			&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "b"}},
		},
	})

	expected := `{
    $a;
    $b;
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtListNested(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.StmtList{
		Stmts: []ir.Node{
			&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
			&ir.StmtList{
				Stmts: []ir.Node{
					&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "b"}},
					&ir.StmtList{
						Stmts: []ir.Node{
							&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "c"}},
						},
					},
				},
			},
		},
	})

	expected := `{
    $a;
    {
        $b;
        {
            $c;
        }
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtSwitch(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.StmtList{
		Stmts: []ir.Node{
			&ir.SwitchStmt{
				Cond: &ir.SimpleVar{Name: "var"},
				CaseList: &ir.CaseListStmt{
					Cases: []ir.Node{
						&ir.CaseStmt{
							Cond: &ir.String{Value: "'a'"},
							Stmts: []ir.Node{
								&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
							},
						},
						&ir.CaseStmt{
							Cond: &ir.String{Value: "'b'"},
							Stmts: []ir.Node{
								&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "b"}},
							},
						},
					},
				},
			},
		},
	})

	expected := `{
    switch ($var) {
        case 'a':
            $a;
        case 'b':
            $b;
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtThrow(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.ThrowStmt{
		Expr: &ir.SimpleVar{Name: "var"},
	})

	expected := `throw $var;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtTraitMethodRef(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.TraitMethodRefStmt{
		Trait:  &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
		Method: &ir.Identifier{Value: "a"},
	})

	expected := `Foo::a`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtTraitUseAlias(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.TraitUseAliasStmt{
		Ref: &ir.TraitMethodRefStmt{
			Trait:  &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
			Method: &ir.Identifier{Value: "a"},
		},
		Modifier: &ir.Identifier{Value: "public"},
		Alias:    &ir.Identifier{Value: "b"},
	})

	expected := `Foo::a as public b;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtTraitUsePrecedence(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.TraitUsePrecedenceStmt{
		Ref: &ir.TraitMethodRefStmt{
			Trait:  &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
			Method: &ir.Identifier{Value: "a"},
		},
		Insteadof: []ir.Node{
			&ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Bar"}}},
			&ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Baz"}}},
		},
	})

	expected := `Foo::a insteadof Bar, Baz;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtTraitUse(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.TraitUseStmt{
		Traits: []ir.Node{
			&ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
			&ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Bar"}}},
		},
	})

	expected := `use Foo, Bar;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtTraitAdaptations(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.TraitUseStmt{
				Traits: []ir.Node{
					&ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
					&ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Bar"}}},
				},
				TraitAdaptationList: &ir.TraitAdaptationListStmt{
					Adaptations: []ir.Node{
						&ir.TraitUseAliasStmt{
							Ref: &ir.TraitMethodRefStmt{
								Trait:  &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
								Method: &ir.Identifier{Value: "a"},
							},
							Alias: &ir.Identifier{Value: "b"},
						},
					},
				},
			},
		},
	})

	expected := `namespace {
    use Foo, Bar {
        Foo::a as b;
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintTrait(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.TraitStmt{
				TraitName: &ir.Identifier{Value: "Foo"},
				Stmts: []ir.Node{
					&ir.ClassMethodStmt{
						Modifiers:  []*ir.Identifier{{Value: "public"}},
						MethodName: &ir.Identifier{Value: "foo"},
						Params:     []ir.Node{},
						Stmt: &ir.StmtList{
							Stmts: []ir.Node{
								&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
							},
						},
					},
				},
			},
		},
	})

	expected := `namespace {
    trait Foo
    {
        public function foo()
        {
            $a;
        }
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtTry(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.TryStmt{
				Stmts: []ir.Node{
					&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
				},
				Catches: []ir.Node{
					&ir.CatchStmt{
						Types: []ir.Node{
							&ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Exception"}}},
							&ir.FullyQualifiedName{Parts: []ir.Node{&ir.NamePart{Value: "RuntimeException"}}},
						},
						Variable: &ir.SimpleVar{Name: "e"},
						Stmts: []ir.Node{
							&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "b"}},
						},
					},
				},
				Finally: &ir.FinallyStmt{
					Stmts: []ir.Node{
						&ir.NopStmt{},
					},
				},
			},
		},
	})

	expected := `namespace {
    try {
        $a;
    }
    catch (Exception | \RuntimeException $e) {
        $b;
    }
    finally {
        ;
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtUset(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.UnsetStmt{
		Vars: []ir.Node{
			&ir.SimpleVar{Name: "a"},
			&ir.SimpleVar{Name: "b"},
		},
	})

	expected := `unset($a, $b);`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintStmtUseList(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.UseListStmt{
		UseType: &ir.Identifier{Value: "function"},
		Uses: []ir.Node{
			&ir.UseStmt{
				Use:   &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
				Alias: &ir.Identifier{Value: "Bar"},
			},
			&ir.UseStmt{
				Use: &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Baz"}}},
			},
		},
	})

	expected := `use function Foo as Bar, Baz;`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintUse(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.UseStmt{
		UseType: &ir.Identifier{Value: "function"},
		Use:     &ir.Name{Parts: []ir.Node{&ir.NamePart{Value: "Foo"}}},
		Alias:   &ir.Identifier{Value: "Bar"},
	})

	expected := `function Foo as Bar`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintWhileStmtList(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.WhileStmt{
				Cond: &ir.SimpleVar{Name: "a"},
				Stmt: &ir.StmtList{
					Stmts: []ir.Node{
						&ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
					},
				},
			},
		},
	})

	expected := `namespace {
    while ($a) {
        $a;
    }
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintWhileExpression(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.NamespaceStmt{
		Stmts: []ir.Node{
			&ir.WhileStmt{
				Cond: &ir.SimpleVar{Name: "a"},
				Stmt: &ir.ExpressionStmt{Expr: &ir.SimpleVar{Name: "a"}},
			},
		},
	})

	expected := `namespace {
    while ($a)
        $a;
}`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}

func TestPrintWhileNop(t *testing.T) {
	o := bytes.NewBufferString("")

	p := NewPrettyPrinter(o, "    ")
	p.Print(&ir.WhileStmt{
		Cond: &ir.SimpleVar{Name: "a"},
		Stmt: &ir.NopStmt{},
	})

	expected := `while ($a);`
	actual := o.String()

	if expected != actual {
		t.Errorf("\nexpected: %s\ngot: %s\n", expected, actual)
	}
}
