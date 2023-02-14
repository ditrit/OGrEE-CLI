package main

import (
	"testing"
)

func checkTokSequence(expected []tokenType, str string, t *testing.T) {
	l := newLexer(str, 0, len(str))
	for i := 0; i < len(expected); i++ {
		tok := l.nextToken(lexExpr)
		if tok.t != expected[i] {
			t.Errorf("Unexpected token : %s instead of %s", tok.t.String(), expected[i].String())
		}
	}
}

func TestLex(t *testing.T) {
	str := "false42 + (3 - 4) * plouf42 + \"plouf\" || false"
	expected := []tokenType{tokWord, tokAdd, tokLeftParen, tokInt, tokSub, tokInt, tokRightParen,
		tokMul, tokWord, tokAdd, tokString, tokOr, tokBool, tokEOF}
	checkTokSequence(expected, str, t)
}

func TestLexDoubleDot(t *testing.T) {
	str := "42.."
	expected := []tokenType{tokInt, tokEOF}
	checkTokSequence(expected, str, t)
}

func TestLexFormattedString(t *testing.T) {
	str := "${a}a"
	expected := []tokenType{tokDeref, tokWord}
	checkTokSequence(expected, str, t)
}
