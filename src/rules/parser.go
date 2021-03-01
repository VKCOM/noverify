package rules

import (
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"sort"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/linter/lintapi"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parseutil"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/phpgrep"
)

var magicComment = regexp.MustCompile(`\* @(?:warning|error|info|maybe) `)

//go:generate stringer -type=parseMode
type parseMode int

const (
	parseNormal parseMode = iota
	parseAny
	parseSeq // To be implemented
)

type parseError struct {
	filename string
	lineNum  int
	msg      string
}

func (e *parseError) Error() string {
	return fmt.Sprintf("%s:%d: %s", e.filename, e.lineNum, e.msg)
}

// parser parses rules file into a RuleSet.
type parser struct {
	filename   string
	sources    []byte
	res        *Set
	compiler   phpgrep.Compiler
	typeParser *phpdoc.TypeParser
	names      map[string]struct{}
	mode       parseMode

	namespace string
	funcName  string
}

// Parse reads PHP code that represents a rule file from r and creates a RuleSet based on it.
func (p *parser) parse(filename string, r io.Reader) (*Set, error) {
	res := NewSet()

	// Parse PHP file.
	//
	// TODO: make phpgrep.compile accepting AST and stop
	// slurping sources here + don't parse it twice.
	sources, err := ioutil.ReadAll(r)
	if err != nil {
		return res, err
	}
	root, err := parseutil.ParseFile(sources)
	if err != nil {
		return res, err
	}
	rootIR := irconv.ConvertNode(root).(*ir.Root)

	// Convert PHP file into the rule set.
	p.filename = filename
	p.sources = sources
	p.res = res
	p.names = make(map[string]struct{})
	if err := p.parseRules(rootIR.Stmts, nil); err != nil {
		return p.res, err
	}

	res.Names = make([]string, 0, len(p.names))
	for name := range p.names {
		res.Names = append(res.Names, name)
	}
	sort.Strings(res.Names)

	return p.res, nil
}

func (p *parser) tryParseLabeledStmt(stmts []ir.Node, proto *Rule) (bool, error) {
	if len(stmts) < 2 {
		return false, nil
	}
	label, ok := stmts[0].(*ir.LabelStmt)
	if !ok {
		return false, nil
	}
	next, ok := stmts[1].(*ir.StmtList)
	if !ok {
		return false, nil
	}

	labelName := label.LabelName.Value
	var mode parseMode
	switch {
	case labelName == "any" || strings.HasPrefix(labelName, "any_"):
		mode = parseAny
	case labelName == "seq" || strings.HasPrefix(labelName, "seq_"):
		mode = parseSeq
	default:
		return false, nil
	}

	if mode == parseSeq {
		return true, p.errorf(label, "seq is not implemented yet")
	}

	nextProto, err := p.parseRuleInfo(label, proto)
	if err != nil {
		return true, err
	}

	prevMode := p.mode
	p.mode = mode
	err = p.parseRules(next.Stmts, &nextProto)
	p.mode = prevMode
	return true, err
}

func (p *parser) parseRuleInfo(st ir.Node, proto *Rule) (Rule, error) {
	var rule Rule

	comment := p.commentText(st)
	if p.mode == parseNormal && comment == "" {
		return rule, nil
	}

	if proto != nil {
		rule.Level = proto.Level
		rule.Message = proto.Message
		rule.Location = proto.Location
		rule.Path = proto.Path

		rule.Filters = make([]map[string]Filter, len(proto.Filters))
		for i, filterSet := range proto.Filters {
			rule.Filters[i] = make(map[string]Filter)
			for name, filter := range filterSet {
				rule.Filters[i][name] = filter
			}
		}
	}

	if p.funcName != "" {
		rule.Name = p.funcName
	}

	var filterSet map[string]Filter

	for _, part := range phpdoc.Parse(p.typeParser, comment) {
		part := part.(*phpdoc.RawCommentPart)
		switch part.Name() {
		case "name":
			if len(part.Params) != 1 {
				return rule, p.errorf(st, "@name expects exactly 1 param, got %d", len(part.Params))
			}
			if p.funcName != "" {
				return rule, p.errorf(st, "@name is not allowed inside a function")
			}
			rule.Name = part.Params[0]

		case "location":
			if len(part.Params) != 1 {
				return rule, p.errorf(st, "@location expects exactly 1 params, got %d", len(part.Params))
			}
			name := part.Params[0]
			if !strings.HasPrefix(name, "$") {
				return rule, p.errorf(st, "@location 2nd param must be a phpgrep variable")
			}
			rule.Location = strings.TrimPrefix(name, "$")

		case "scope":
			if len(part.Params) != 1 {
				return rule, p.errorf(st, "@scope expects exactly 1 params, got %d", len(part.Params))
			}
			switch part.Params[0] {
			case "any":
				rule.scope = "any"
			case "root":
				rule.scope = "root"
			case "local":
				rule.scope = "local"
			default:
				return rule, p.errorf(st, "unknown @scope: %s", part.Params[0])
			}

		case "error":
			rule.Level = lintapi.LevelError
			rule.Message = part.ParamsText
		case "warning":
			rule.Level = lintapi.LevelWarning
			rule.Message = part.ParamsText
		case "maybe":
			rule.Level = lintapi.LevelNotice
			rule.Message = part.ParamsText

		case "fix":
			if rule.Fix != "" {
				return rule, p.errorf(st, "duplicated @fix")
			}
			rule.Fix = part.ParamsText

		case "or":
			rule.Filters = append(rule.Filters, filterSet)
			filterSet = nil
		case "path":
			if len(part.Params) != 1 {
				return rule, p.errorf(st, "@path expects exactly 1 param, got %d", len(part.Params))
			}
			if rule.Path != "" {
				return rule, p.errorf(st, "duplicate @path constraint")
			}
			rule.Path = part.Params[0]
		case "type":
			if len(part.Params) != 2 {
				return rule, p.errorf(st, "@type expects exactly 2 params, got %d", len(part.Params))
			}
			typeString := part.Params[0]
			name := part.Params[1]
			if !strings.HasPrefix(name, "$") {
				return rule, p.errorf(st, "@type 2nd param must be a phpgrep variable")
			}
			name = strings.TrimPrefix(name, "$")
			if filterSet == nil {
				filterSet = map[string]Filter{}
			}
			filter := filterSet[name]
			if filter.Type != nil {
				return rule, p.errorf(st, "$%s: duplicate type constraint", name)
			}
			typ := p.typeParser.Parse(typeString).Clone()
			switch typ.Expr.Kind {
			case phpdoc.ExprInvalid, phpdoc.ExprUnknown:
				return rule, p.errorf(st, "$%s: parseType(%s): bad type expression", name, typ)
			}
			filter.Type = new(phpdoc.Type)
			*filter.Type = typ
			filterSet[name] = filter
		case "pure":
			if len(part.Params) != 1 {
				return rule, p.errorf(st, "@pure expects exactly 1 param, got %d", len(part.Params))
			}
			name := part.Params[0]
			if !strings.HasPrefix(name, "$") {
				return rule, p.errorf(st, "@pure param must be a phpgrep variable")
			}
			name = strings.TrimPrefix(name, "$")
			if filterSet == nil {
				filterSet = map[string]Filter{}
			}
			filter := filterSet[name]
			filter.Pure = true
			filterSet[name] = filter

		default:
			return rule, p.errorf(st, "unknown attribute @%s on line %d", part.Name(), part.Line())
		}
	}

	if rule.Name == "" {
		return rule, p.errorf(st, "missing @name attribute")
	}
	if p.namespace != "" {
		rule.Name = p.namespace + "/" + rule.Name
	}
	p.names[rule.Name] = struct{}{}

	if filterSet != nil {
		rule.Filters = append(rule.Filters, filterSet)
	}

	return rule, nil
}

func (p *parser) parseRules(stmts []ir.Node, proto *Rule) error {
	for len(stmts) > 0 {
		stmt := stmts[0]

		ok, err := p.tryParseLabeledStmt(stmts, proto)
		if err != nil {
			return err
		}
		if ok {
			stmts = stmts[2:]
			continue
		}

		if err := p.parseRule(stmt, proto); err != nil {
			return err
		}
		stmts = stmts[1:]
	}

	return nil
}

func (p *parser) parseRule(st ir.Node, proto *Rule) error {
	switch st := st.(type) {
	case *ir.FunctionStmt:
		p.funcName = st.FunctionName.Value
		if err := p.parseFuncComment(st); err != nil {
			return nil
		}
		if err := p.parseRules(st.Stmts, proto); err != nil {
			return p.errorf(st, "%s: %v", p.funcName, err)
		}
		p.funcName = ""
		return nil

	case *ir.NamespaceStmt:
		if len(st.Stmts) != 0 {
			return p.errorf(st, "namespace with body is not supported")
		}
		p.namespace = meta.NameNodeToString(st.NamespaceName)
		if strings.Contains(p.namespace, `\`) {
			return p.errorf(st, "multi-part namespace names are not supported")
		}
		return nil
	}

	rule, err := p.parseRuleInfo(st, proto)
	if err != nil {
		return err
	}

	dst := p.res.Any // Use "any" set by default
	switch rule.scope {
	case "any":
		dst = p.res.Any
	case "root":
		dst = p.res.Root
	case "local":
		dst = p.res.Local
	}

	if rulesDoc, ok := p.res.DocByName[p.funcName]; ok {
		if !rulesDoc.Fix && rule.Fix != "" {
			rulesDoc.Fix = true
			p.res.DocByName[p.funcName] = rulesDoc
		}
	}

	pos := ir.GetPosition(st)
	m, err := p.compiler.Compile(p.sources[pos.StartPos-1 : pos.EndPos])
	if err != nil {
		return p.errorf(st, "pattern compilation error: %v", err)
	}
	rule.Matcher = m

	if st2, ok := st.(*ir.ExpressionStmt); ok {
		st = st2.Expr
	}
	kind := ir.GetNodeKind(st)
	dst.Add(kind, rule)
	return nil
}

func (p *parser) parseFuncComment(fn *ir.FunctionStmt) error {
	if len(fn.PhpDoc) == 0 {
		return nil
	}
	var doc RuleDoc
	for _, part := range fn.PhpDoc {
		part := part.(*phpdoc.RawCommentPart)
		switch part.Name() {
		case "comment":
			doc.Comment = part.ParamsText
		case "before":
			doc.Before = part.ParamsText
		case "after":
			doc.After = part.ParamsText
		}
	}
	p.res.DocByName[p.funcName] = doc
	return nil
}

func (p *parser) commentText(n ir.Node) string {
	doc, found := irutil.FindPhpDoc(n)
	if !found {
		return ""
	}

	if !magicComment.MatchString(doc) {
		return ""
	}

	return doc
}

func (p *parser) errorf(n ir.Node, format string, args ...interface{}) *parseError {
	pos := ir.GetPosition(n)
	return &parseError{
		filename: p.filename,
		lineNum:  pos.StartLine,
		msg:      fmt.Sprintf(format, args...),
	}
}
