package parseutil

import (
	"bytes"
	"errors"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
)

// Parse combines ParseFile and ParseStmt.
//
// If input code is a proper PHP file, it's parsed with ParseFile.
// Otherwise it parsed with ParseStmt.
//
// Useful for testing.
func Parse(code []byte) (node.Node, []byte, error) {
	if bytes.HasPrefix(code, []byte("<?")) || bytes.HasPrefix(code, []byte("<?php")) {
		n, err := ParseFile(code)
		return n, code, err
	}
	return ParseStmt(code)
}

// ParseFile parses PHP file sources and returns its AST root.
//
// Useful for testing.
func ParseFile(code []byte) (*node.Root, error) {
	p := php7.NewParser(code)
	p.WithFreeFloating()
	p.Parse()
	if len(p.GetErrors()) != 0 {
		return nil, errors.New(p.GetErrors()[0].String())
	}
	return p.GetRootNode(), nil
}

// ParseStmt parses a single PHP statement (which can be an expression).
//
// Due to the fact that we extend the input slice with <?php tags
// so our parser can handle it, updated slice is returned.
// Result node source positions are precise only for that updated slice.
//
// Useful for testing.
func ParseStmt(code []byte) (node.Node, []byte, error) {
	code = append([]byte("<?php "), code...)
	code = append(code, ';')
	root, err := ParseFile(code)
	if err != nil {
		return nil, code, err
	}
	stmts := root.Stmts
	if len(stmts) == 0 {
		return &stmt.Nop{}, code, nil
	}
	return root.Stmts[0], code, nil
}
