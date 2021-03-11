package normalize

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/VKCOM/noverify/src/constfold"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irfmt"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/meta"
)

type Config struct {
	NormalizeMore bool
}

func FuncBody(st *meta.ClassParseState, conf Config, params, statements []ir.Node) []ir.Node {
	norm := normalizer{
		out:     irutil.NodeSliceClone(statements),
		conf:    conf,
		st:      st,
		globals: make(map[string]struct{}),
	}

	norm.runPre()

	norm.run(params)
	changed := !irutil.NodeSliceEqual(norm.out, statements)
	if changed {
		norm.run(params)
	}
	return norm.out
}

type normalizer struct {
	conf     Config
	st       *meta.ClassParseState
	out      []ir.Node
	varNames map[string]int
	globals  map[string]struct{}
}

func (norm *normalizer) runPre() {
	norm.out = norm.sortStatements(norm.out)

	// TODO:
	// `$x = $y = 10` => `$y = 10; $x = $y;`
	// `global $x, $y` => `global $x; global $y;`
}

func (norm *normalizer) run(params []ir.Node) {
	norm.varNames = make(map[string]int)
	for _, p := range params {
		p := p.(*ir.Parameter)
		norm.internVarName(p.Variable.Name)
	}

	for _, n := range norm.out {
		n.Walk(norm)
	}
}

func (norm *normalizer) sortStatements(statements []ir.Node) []ir.Node {
	var simpleExpr func(n ir.Node) bool
	simpleExpr = func(n ir.Node) bool {
		if n == nil {
			return true
		}
		switch n := n.(type) {
		case *ir.ArrayItemExpr:
			return simpleExpr(n.Key) && simpleExpr(n.Val)
		case *ir.ArrayExpr:
			for _, item := range n.Items {
				if !simpleExpr(item) {
					return false
				}
			}
			return true
		case *ir.String, *ir.Lnumber, *ir.Dnumber, *ir.ConstFetchExpr:
			return true
		}
		return false
	}

	var sortKey func(n ir.Node) string
	sortKey = func(n ir.Node) string {
		switch n := n.(type) {
		case *ir.ImportExpr, *ir.GlobalStmt:
			// Since we don't allow variables and function calls inside
			// reordered blocks, requires and global statements can be
			// considered to be reorderable.
			return irfmt.Node(n)
		case *ir.ExpressionStmt:
			return sortKey(n.Expr)
		case *ir.Assign:
			_, ok := n.Variable.(*ir.SimpleVar)
			if ok && simpleExpr(n.Expr) {
				return irfmt.Node(n.Expr)
			}
		}
		return ""
	}

	out := statements[:0]
	for len(statements) > 0 {
		type sortableStatement struct {
			key string
			n   ir.Node
		}
		var block []sortableStatement
		for _, n := range statements {
			k := sortKey(n)
			if k == "" {
				break
			}
			block = append(block, sortableStatement{key: k, n: n})
		}
		if len(block) == 0 {
			out = append(out, statements[0])
			statements = statements[1:]
		} else {
			sort.SliceStable(block, func(i, j int) bool {
				return block[i].key < block[j].key
			})
			for i := range block {
				out = append(out, block[i].n)
			}
			statements = statements[len(block):]
		}
	}
	return out
}

func (norm *normalizer) LeaveNode(n ir.Node) {}

func (norm *normalizer) EnterNode(n ir.Node) bool {
	switch n := n.(type) {
	case *ir.GlobalStmt:
		for _, v := range n.Vars {
			v, ok := v.(*ir.SimpleVar)
			if !ok {
				continue
			}
			norm.globals[v.Name] = struct{}{}
		}
		return false

	case *ir.IfStmt:
		norm.normalizeIf(n)

	case *ir.SimpleVar:
		norm.renameVar(n)

	case *ir.StmtList:
		n.Stmts = norm.sortStatements(n.Stmts)

	case *ir.ReturnStmt:
		n.Expr = norm.normalizedExpr(n.Expr)
	case *ir.Argument:
		n.Expr = norm.normalizedExpr(n.Expr)
	case *ir.ExpressionStmt:
		n.Expr = norm.normalizedStmtExpr(n.Expr)

	case *ir.Assign:
		n.Expr = norm.normalizedExpr(n.Expr)
		n.Variable = norm.normalizedExpr(n.Variable)
	case *ir.AssignPlus:
		n.Expr = norm.normalizedExpr(n.Expr)
		n.Variable = norm.normalizedExpr(n.Variable)
	case *ir.AssignConcat:
		n.Expr = norm.normalizedExpr(n.Expr)
		n.Variable = norm.normalizedExpr(n.Variable)

	case *ir.StaticPropertyFetchExpr:
		n.Class = norm.normalizedExpr(n.Class)
		return false // To avoid var renaming

	case *ir.StaticCallExpr:
		n.Class = norm.normalizedExpr(n.Class)
	case *ir.ClassConstFetchExpr:
		n.Class = norm.normalizedExpr(n.Class)

	case *ir.ForStmt:
		for i := range n.Init {
			n.Init[i] = norm.normalizedStmtExpr(n.Init[i])
		}
		for i := range n.Loop {
			n.Loop[i] = norm.normalizedStmtExpr(n.Loop[i])
		}

	case *ir.EchoStmt:
		for i := range n.Exprs {
			n.Exprs[i] = norm.normalizedExpr(n.Exprs[i])
		}
	}

	return true
}

func (norm *normalizer) normalizeIf(n *ir.IfStmt) {
	// `if (!$x) { $a } else { $b }` => `if ($x) { $b } else { $a }`
	if len(n.ElseIf) != 0 || n.Else == nil {
		return
	}
	not, ok := n.Cond.(*ir.BooleanNotExpr)
	if !ok {
		return
	}
	elseNode := n.Else.(*ir.ElseStmt)
	n.Stmt, elseNode.Stmt = elseNode.Stmt, n.Stmt
	n.Cond = not.Expr
}

func (norm *normalizer) normalizedStmtExpr(e ir.Node) ir.Node {
	// We can be less conservative if expression result is unused.
	// For example, in loop pos-statement both `$i++` and `++$i` are identical.

	switch e := e.(type) {
	case *ir.PostIncExpr:
		// `$x++` => `++$x`
		return &ir.PreIncExpr{Variable: e.Variable}
	case *ir.PostDecExpr:
		// `$x--` => `--$x`
		return &ir.PreDecExpr{Variable: e.Variable}
	case *ir.FunctionCallExpr:
		funcName, ok := e.Function.(*ir.Name)
		if !ok {
			break
		}
		// `array_push($a, $e)` => `$a[] = $e`
		if len(e.Args) == 2 && funcName.Value == "array_push" {
			a := e.Arg(0).Expr
			e := e.Arg(1).Expr
			return &ir.Assign{
				Variable: &ir.ArrayDimFetchExpr{
					Variable: norm.normalizedExpr(a),
				},
				Expr: norm.normalizedExpr(e),
			}
		}
	}

	return norm.normalizedExpr(e)
}

func (norm *normalizer) normalizedExpr(e ir.Node) ir.Node {
	constFolded := constfold.Eval(norm.st, e)
	if constFolded.IsValid() {
		if e2 := constToIR(constFolded); e2 != nil {
			return e2
		}
	}

	switch e := e.(type) {
	case *ir.Name:
		if e.Value == "self" {
			return &ir.Name{Value: norm.st.CurrentClass}
		}

	case *ir.NewExpr:
		// `new T` => `new T()`
		if e.Args == nil {
			e.Args = []ir.Node{}
		}

	case *ir.BooleanNotExpr:
		// `!!$x` => `(bool)$x`
		if x, ok := e.Expr.(*ir.BooleanNotExpr); ok {
			return &ir.TypeCastExpr{
				Type: "bool",
				Expr: norm.normalizedExpr(x.Expr),
			}
		}

	case *ir.Encapsed:
		// `"$x"` => `'' . $x`
		if len(e.Parts) == 1 {
			return &ir.ConcatExpr{
				Left:  emptyStringNode,
				Right: encapsedPartToConcatArg(e.Parts[0]),
			}
		}
		// `"$x$y"` => `$x . $y`
		// `"$x$y$z"` => `$x . $y . $z`
		concat := &ir.ConcatExpr{
			Left:  encapsedPartToConcatArg(e.Parts[0]),
			Right: encapsedPartToConcatArg(e.Parts[1]),
		}
		for _, p := range e.Parts[2:] {
			concat = &ir.ConcatExpr{
				Left:  concat,
				Right: encapsedPartToConcatArg(p),
			}
		}
		return concat

	case *ir.PlusExpr:
		norm.sortCommutative(&e.Left, &e.Right)
	case *ir.MulExpr:
		norm.sortCommutative(&e.Left, &e.Right)
	case *ir.EqualExpr:
		norm.sortCommutative(&e.Left, &e.Right)
	case *ir.NotEqualExpr:
		norm.sortCommutative(&e.Left, &e.Right)
	case *ir.IdenticalExpr:
		norm.sortCommutative(&e.Left, &e.Right)
	case *ir.NotIdenticalExpr:
		norm.sortCommutative(&e.Left, &e.Right)

	case *ir.String:
		// `"abc"` => `'abc'`
		e.DoubleQuotes = false

	case *ir.TernaryExpr:
		// `$x ? $x : $y` => `$x ?: $y`
		if sideEffectFree(e.Condition) && irutil.NodeEqual(e.Condition, e.IfTrue) {
			e.IfTrue = nil
			return e
		}
		// TODO: `isset($x) ? $x : $y` => `$x ?? $y`

	case *ir.ConstFetchExpr:
		// `NULL` => `null`
		constNameString := e.Constant.Value
		switch {
		case strings.EqualFold(constNameString, `null`):
			if constNameString != `null` {
				e.Constant = &ir.Name{Value: `null`}
			}
		case strings.EqualFold(constNameString, `true`):
			if constNameString != `true` {
				e.Constant = &ir.Name{Value: `true`}
			}
		case strings.EqualFold(constNameString, `false`):
			if constNameString != `false` {
				e.Constant = &ir.Name{Value: `false`}
			}
		}

	case *ir.FunctionCallExpr:
		funcName, ok := e.Function.(*ir.Name)
		if !ok {
			break
		}
		// `is_null($x)` => `$x === null`
		if funcName.Value == "is_null" {
			return &ir.IdenticalExpr{
				Left:  e.Arg(0).Expr,
				Right: nullConstNode,
			}
		}

		// Replace aliased functions.
		alias, ok := funcAliases[funcName.Value]
		if ok {
			e.Function = alias
		}

	case *ir.Assign:
		if !sideEffectFree(e.Variable) {
			break
		}
		// `$x = $x <op> $y` => `$x <op>= $y`
		switch rhs := e.Expr.(type) {
		case *ir.PlusExpr:
			if irutil.NodeEqual(e.Variable, rhs.Left) {
				return &ir.AssignPlus{Variable: e.Variable, Expr: rhs.Right}
			}
		case *ir.MinusExpr:
			if irutil.NodeEqual(e.Variable, rhs.Left) {
				return &ir.AssignMinus{Variable: e.Variable, Expr: rhs.Right}
			}
		case *ir.ConcatExpr:
			if irutil.NodeEqual(e.Variable, rhs.Left) {
				return &ir.AssignConcat{Variable: e.Variable, Expr: rhs.Right}
			}
		}

	case *ir.AssignPlus:
		// `$x += 1` => `++$x`
		if literalValue(e.Expr) == `1` {
			return &ir.PreIncExpr{Variable: e.Variable}
		}
	case *ir.AssignMinus:
		// `$x -= 1` => `--$x`
		if literalValue(e.Expr) == `1` {
			return &ir.PreDecExpr{Variable: e.Variable}
		}

	case *ir.ArrayExpr:
		// `array(...)` => `[...]`
		e.ShortSyntax = true
	case *ir.ListExpr:
		// `list(...)` => `[...]`
		e.ShortSyntax = true
		// `list($x, $y)` => `list(0 => $x, 1 => $y)`
		// `list(, $x)` => `list(1 => $x)`,
		items := e.Items[:0]
		for i, item := range e.Items {
			if item.Val == nil {
				continue
			}
			if item.Key == nil {
				item.Key = &ir.Lnumber{Value: strconv.Itoa(i)}
			}
			items = append(items, item)
		}
		e.Items = items
	}

	return e
}

func (norm *normalizer) internVarName(name string) int {
	id, ok := norm.varNames[name]
	if !ok {
		id = len(norm.varNames)
		norm.varNames[name] = id
	}
	return id
}

func (norm *normalizer) sortCommutative(left, right *ir.Node) {
	if !(sideEffectFree(*left) && sideEffectFree(*right)) {
		return
	}

	if irfmt.Node(*left) > irfmt.Node(*right) {
		*left, *right = *right, *left
	}
}

func (norm *normalizer) renameVar(v *ir.SimpleVar) {
	if _, ok := norm.globals[v.Name]; ok {
		return
	}
	id := norm.internVarName(v.Name)
	v.Name = fmt.Sprintf("v%d", id)
}

var (
	nullConstNode   = &ir.ConstFetchExpr{Constant: &ir.Name{Value: "null"}}
	emptyStringNode = &ir.String{}
)

var funcAliases = map[string]*ir.Name{
	// See https://www.php.net/manual/ru/aliases.php

	`doubleval`: {Value: `floatval`},

	`ini_alter`:    {Value: `ini_set`},
	`is_integer`:   {Value: `is_int`},
	`is_long`:      {Value: `is_int`},
	`is_real`:      {Value: `is_float`},
	`is_double`:    {Value: `is_float`},
	`is_writeable`: {Value: `is_writable`},

	`join`:       {Value: `implode`},
	`chop`:       {Value: `rtrim`},
	`strchr`:     {Value: `strstr`},
	`pos`:        {Value: `current`},
	`key_exists`: {Value: `array_key_exists`},
	`sizeof`:     {Value: `count`},

	`close`:                {Value: `closedir`},
	`fputs`:                {Value: `fwrite`},
	`magic_quotes_runtime`: {Value: `set_magic_quotes_runtime`},
	`show_source`:          {Value: `highlight_file`},
}

func literalValue(e ir.Node) string {
	switch e := e.(type) {
	case *ir.Lnumber:
		return e.Value
	default:
		return ""
	}
}

func encapsedPartToConcatArg(n ir.Node) ir.Node {
	switch n := n.(type) {
	case *ir.EncapsedStringPart:
		return &ir.String{Value: n.Value}
	default:
		return n
	}
}

func constToIR(v meta.ConstValue) ir.Node {
	value := v.Value
	switch v.Type {
	case meta.Integer:
		return &ir.Lnumber{Value: fmt.Sprint(value)}
	case meta.Float:
		return &ir.Dnumber{Value: fmt.Sprint(value)}
	case meta.String:
		return &ir.String{Value: value.(string)}
	case meta.Bool:
		return &ir.Name{Value: fmt.Sprint(value)}
	default:
		return nil
	}
}
