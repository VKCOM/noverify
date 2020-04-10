package php7

import (
	"strings"

	"github.com/VKCOM/noverify/src/php/parser/errors"
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/parser"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/scanner"
)

func (lval *yySymType) Token(t *scanner.Token) {
	lval.token = t
}

// Parser structure
type Parser struct {
	Lexer           scanner.Scanner
	currentToken    *scanner.Token
	positionBuilder *parser.PositionBuilder
	rootNode        *node.Root
}

// NewParser creates and returns new Parser
func NewParser(src []byte) *Parser {
	lexer := scanner.NewLexer(src)
	return &Parser{Lexer: lexer}
}

// Lex proxy to lexer Lex
func (l *Parser) Lex(lval *yySymType) int {
	t := l.Lexer.Lex(lval)
	l.currentToken = lval.token
	return t
}

func (l *Parser) Error(msg string) {
	pos := &position.Position{
		StartLine: l.currentToken.StartLine,
		EndLine:   l.currentToken.EndLine,
		StartPos:  l.currentToken.StartPos,
		EndPos:    l.currentToken.EndPos,
	}

	l.Lexer.AddError(errors.NewError(msg, pos))
}

func (l *Parser) WithFreeFloating() {
	l.Lexer.SetWithFreeFloating(true)
}

// Parse the php7 Parser entrypoint
func (l *Parser) Parse() int {
	// init
	l.Lexer.SetErrors(nil)
	l.rootNode = nil
	l.positionBuilder = &parser.PositionBuilder{}

	// parse

	return yyParse(l)
}

// GetRootNode returns root node
func (l *Parser) GetRootNode() *node.Root {
	return l.rootNode
}

// GetErrors returns errors list
func (l *Parser) GetErrors() []*errors.Error {
	return l.Lexer.GetErrors()
}

// helpers

func lastNode(nn []node.Node) node.Node {
	if len(nn) == 0 {
		return nil
	}
	return nn[len(nn)-1]
}

func firstNode(nn []node.Node) node.Node {
	return nn[0]
}

func isDollar(r rune) bool {
	return r == '$'
}

func (l *Parser) MoveFreeFloating(src node.Node, dst node.Node) {
	if !l.Lexer.GetWithFreeFloating() {
		return
	}

	if src.GetFreeFloating() == nil {
		return
	}

	l.setFreeFloating(dst, freefloating.Start, (*src.GetFreeFloating())[freefloating.Start])
	delete((*src.GetFreeFloating()), freefloating.Start)
}

func (l *Parser) setFreeFloating(dst node.Node, p freefloating.Position, strings []freefloating.String) {
	if !l.Lexer.GetWithFreeFloating() {
		return
	}

	if len(strings) == 0 {
		return
	}

	dstCollection := dst.GetFreeFloating()
	if *dstCollection == nil {
		*dstCollection = make(freefloating.Collection)
	}

	(*dstCollection)[p] = strings
}

func (l *Parser) GetFreeFloatingToken(t *scanner.Token) []freefloating.String {
	if !l.Lexer.GetWithFreeFloating() {
		return []freefloating.String{}
	}

	return t.GetFreeFloatingToken()
}

func (l *Parser) addDollarToken(v node.Node) {
	if !l.Lexer.GetWithFreeFloating() {
		return
	}

	l.setFreeFloating(v, freefloating.Dollar, []freefloating.String{
		{
			StringType: freefloating.TokenType,
			Value:      "$",
			Position: &position.Position{
				StartLine: v.GetPosition().StartLine,
				EndLine:   v.GetPosition().StartLine,
				StartPos:  v.GetPosition().StartPos,
				EndPos:    v.GetPosition().StartPos + 1,
			},
		},
	})
}

func (l *Parser) splitSemiColonAndPhpCloseTag(htmlNode node.Node, prevNode node.Node) {
	if !l.Lexer.GetWithFreeFloating() {
		return
	}

	semiColon := (*prevNode.GetFreeFloating())[freefloating.SemiColon]
	delete((*prevNode.GetFreeFloating()), freefloating.SemiColon)
	if len(semiColon) == 0 {
		return
	}

	p := semiColon[0].Position
	if semiColon[0].Value[0] == ';' {
		l.setFreeFloating(prevNode, freefloating.SemiColon, []freefloating.String{
			{
				StringType: freefloating.TokenType,
				Value:      ";",
				Position: &position.Position{
					StartLine: p.StartLine,
					EndLine:   p.StartLine,
					StartPos:  p.StartPos,
					EndPos:    p.StartPos + 1,
				},
			},
		})
	}

	vlen := len(semiColon[0].Value)
	tlen := 2
	if strings.HasSuffix(semiColon[0].Value, "?>\n") {
		tlen = 3
	}

	phpCloseTag := []freefloating.String{}
	if vlen-tlen > 1 {
		phpCloseTag = append(phpCloseTag, freefloating.String{
			StringType: freefloating.WhiteSpaceType,
			Value:      semiColon[0].Value[1 : vlen-tlen],
			Position: &position.Position{
				StartLine: p.StartLine,
				EndLine:   p.EndLine,
				StartPos:  p.StartPos + 1,
				EndPos:    p.EndPos - tlen,
			},
		})
	}

	phpCloseTag = append(phpCloseTag, freefloating.String{
		StringType: freefloating.WhiteSpaceType,
		Value:      semiColon[0].Value[vlen-tlen:],
		Position: &position.Position{
			StartLine: p.EndLine,
			EndLine:   p.EndLine,
			StartPos:  p.EndPos - tlen,
			EndPos:    p.EndPos,
		},
	})

	l.setFreeFloating(htmlNode, freefloating.Start, append(phpCloseTag, (*htmlNode.GetFreeFloating())[freefloating.Start]...))
}

func (p *Parser) returnTokenToPool(yyDollar []yySymType, yyVAL *yySymType) {
	for i := 1; i < len(yyDollar); i++ {
		if yyDollar[i].token != nil {
			p.Lexer.ReturnTokenToPool(yyDollar[i].token)
		}
		yyDollar[i].token = nil
	}
	yyVAL.token = nil
}
