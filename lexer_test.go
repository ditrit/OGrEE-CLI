package main

import (
	"testing"
)

func TestLex(t *testing.T) {
	str := "false42 + (3 - 4) * plouf42 || false"
	expected := []tokenType{tokWord, tokAdd, tokLeftParen, tokInt, tokSub, tokInt, tokRightParen, tokMul, tokWord, tokOr, tokBool, tokEOF}
	l := lex(str, 0, len(str))
	for i := 0; i < len(expected); i++ {
		tok := l.nextToken()
		if tok.t != expected[i] {
			t.Errorf("Unexpected token")
		}
	}
}
