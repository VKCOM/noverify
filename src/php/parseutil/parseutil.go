package parseutil

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/conf"
	phperrors "github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
)

// Parse combines ParseFile and ParseStmt.
//
// If input code is a proper PHP file, it's parsed with ParseFile.
// Otherwise it parsed with ParseStmt.
//
// Useful for testing.
func Parse(code []byte) (ast.Vertex, []byte, error) {
	if bytes.HasPrefix(code, []byte("<?")) || bytes.HasPrefix(code, []byte("<?php")) {
		n, err := ParseFile(code)
		return n, code, err
	}
	return ParseStmt(code)
}

// ParseFile parses PHP file sources and returns its AST root.
//
// Useful for testing.
func ParseFile(code []byte) (*ast.Root, error) {
	phpVersion, err := version.New("7.4")
	if err != nil {
		return nil, err
	}

	var parserErrors []*phperrors.Error
	rootNode, err := parser.Parse(code, conf.Config{
		Version: phpVersion,
		ErrorHandlerFunc: func(e *phperrors.Error) {
			parserErrors = append(parserErrors, e)
		},
	})
	if err != nil {
		return nil, err
	}
	if len(parserErrors) != 0 {
		return nil, errors.New(parserErrors[0].String())
	}

	if rootNode == nil {
		return nil, fmt.Errorf("file has incorrect syntax and cannot be parsed")
	}

	return rootNode.(*ast.Root), nil
}

// ParseStmt parses a single PHP statement (which can be an expression).
//
// Due to the fact that we extend the input slice with <?php tags
// so our parser can handle it, updated slice is returned.
// Result node source positions are precise only for that updated slice.
//
// Useful for testing.
func ParseStmt(code []byte) (ast.Vertex, []byte, error) {
	code = append([]byte("<?php "), code...)
	code = append(code, ';')
	root, err := ParseFile(code)
	if err != nil {
		return nil, code, err
	}
	stmts := root.Stmts
	if len(stmts) == 0 {
		return &ast.StmtNop{}, code, nil
	}
	return root.Stmts[0], code, nil
}

// ParseStmtList parses a list of PHP statement (which can be an expression).
//
// Useful for testing.
func ParseStmtList(code []byte) (*ast.Root, error) {
	code = append([]byte("<?php "), code...)
	root, err := ParseFile(code)
	if err != nil {
		return nil, err
	}
	return root, nil
}
