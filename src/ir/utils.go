package ir

import (
	"github.com/VKCOM/php-parser/pkg/token"
)

// traverseToken calls the passed function with the passed token and
// recursively traverse all FreeFloating of the passed token.
//
// If the passed function returns false during the token traverse,
// then FreeFloating is not traversed.
//
// If the passed function returns false when traversing FreeFloating,
// then all other tokens from FreeFloating will not be traversed.
func traverseToken(t *token.Token, cb func(*token.Token) bool) (continueTraverse bool) {
	if t == nil {
		return true
	}

	if !cb(t) {
		return false
	}

	for _, ff := range t.FreeFloating {
		if !traverseToken(ff, cb) {
			return false
		}
	}

	return true
}

func FindParent[T any](v any) (*T, bool) {
	for v != nil {
		if el, ok := v.(*T); ok {
			return el, true
		}
		v = v.(Node).Parent()
	}
	return nil, false
}

func FindParentInterface[T any](v any) (*T, bool) {
	for v != nil {
		if el, ok := v.(T); ok {
			return &el, true
		}
		v = v.(Node).Parent()
	}
	return nil, false
}
