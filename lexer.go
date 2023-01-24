package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type tokenType int

const (
	tokEOF tokenType = iota
	tokError
	tokWord       // identifier
	tokDeref      // variable dereferenciation
	tokInt        // integer constant
	tokFloat      // float constant
	tokBool       // boolean constant
	tokString     // quoted string
	tokLeftParen  // '('
	tokRightParen // ')'
	tokNot        // '!'
	tokAdd        // '+'
	tokSub        // '-'
	tokMul        // '*'
	tokDiv        // '/'
	tokMod        // '%'
	tokOr         // '||'
	tokAnd        // '&&'
	tokEq         // '=='
	tokNeq        // '!='
	tokLeq        // '<=
	tokGeq        // '>='
	tokGtr        // '>'
	tokLss        // '<'
	tokOrientation
)

func (s tokenType) String() string {
	return [...]string{
		"eof", "word", "deref", "int", "float", "bool", "string",
		"leftParen", "rightParen", "not", "add", "sub", "mul", "div",
		"mod", "or", "and", "eq", "neq", "leq", "geq", "gtr", "lss", "orientation"}[s]
}

type token struct {
	t     tokenType
	start int
	end   int
	str   string
	val   interface{}
}

func (t token) precedence() int {
	switch t.t {
	case tokOr:
		return 1
	case tokAnd:
		return 2
	case tokEq, tokNeq, tokLss, tokLeq, tokGtr, tokGeq:
		return 3
	case tokAdd, tokSub:
		return 4
	case tokMul, tokDiv, tokMod:
		return 5
	case tokNot:
		return 6
	}
	return 0
}

const eof = 0

type lexer struct {
	input string
	pos   int
	start int
	end   int
	tok   token
	atEOF bool
}

type stateFn func(*lexer) stateFn

func (l *lexer) emit(t tokenType, val interface{}) stateFn {
	str := l.input[l.start:l.pos]
	if t == tokEOF {
		str = "eof"
	}
	l.tok = token{
		t:     t,
		start: l.start,
		end:   l.pos,
		str:   str,
		val:   val,
	}
	l.start = l.pos
	return nil
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	val := fmt.Sprintf(format, args...)
	l.tok = token{
		t:     tokError,
		start: l.start,
		str:   val,
		val:   val,
	}
	l.start = 0
	l.pos = 0
	l.input = l.input[:0]
	return nil
}

func (l *lexer) next() byte {
	if l.pos >= l.end {
		l.atEOF = true
		return eof
	}
	char := l.input[l.pos]
	l.pos++
	return char
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	if l.pos > 0 && !l.atEOF {
		l.pos--
	}
}

func (l *lexer) accept(valid string) bool {
	if strings.Contains(valid, string(l.next())) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.Contains(valid, string(l.next())) {
	}
	l.backup()
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t'
}

func isLetter(c byte) bool {
	return c == '_' || unicode.IsLetter(rune(c))
}

func isDigit(c byte) bool {
	return unicode.IsDigit(rune(c))
}

func isAlphaNumeric(c byte) bool {
	return isLetter(c) || isDigit(c)
}

func lexExpr(l *lexer) stateFn {
	c := l.next()
	switch c {
	case ' ', '\t':
		l.ignore()
		return lexExpr
	case '$':
		return lexDeref
	case '"':
		return lexString
	case '(':
		return l.emit(tokLeftParen, nil)
	case ')':
		return l.emit(tokRightParen, nil)
	case '+':
		return l.emit(tokAdd, nil)
	case '-':
		return l.emit(tokSub, nil)
	case '*':
		return l.emit(tokMul, nil)
	case '/':
		return l.emit(tokDiv, nil)
	case '%':
		return l.emit(tokMod, nil)
	case '|':
		if l.next() == '|' {
			return l.emit(tokOr, nil)
		}
	case '&':
		if l.next() == '&' {
			return l.emit(tokAnd, nil)
		}
	case '=':
		if l.next() == '=' {
			return l.emit(tokEq, nil)
		}
	case '!':
		c = l.next()
		if c == '=' {
			return l.emit(tokNeq, nil)
		}
		if c == eof {
			return l.emit(tokNot, nil)
		}
	case '<':
		c = l.next()
		if c == '=' {
			return l.emit(tokLeq, nil)
		}
		if c == eof {
			return l.emit(tokLss, nil)
		}
	case '>':
		c = l.next()
		if c == '=' {
			return l.emit(tokGeq, nil)
		}
		if c == eof {
			return l.emit(tokGtr, nil)
		}
	}
	if isDigit(c) {
		return lexNumber
	}
	if c == '.' {
		if l.accept(".") {
			return l.emit(tokEOF, nil)
		}
		l.backup()
		return lexNumber
	}

	if isLetter(c) {
		return lexAlphaNumeric
	}
	return l.emit(tokEOF, nil)
}

func lexDeref(l *lexer) stateFn {
	if l.next() != '{' {
		return l.errorf("{ expected")
	}
	for isSpace(l.next()) {
	}
	l.backup()
	if !isLetter(l.next()) {
		return l.errorf("letter expected")
	}
	for isAlphaNumeric(l.next()) {
	}
	l.backup()
	for isSpace(l.next()) {
	}
	l.backup()
	if l.next() != '}' {
		return l.errorf("} expected")
	}
	return l.emit(tokDeref, l.input[l.start+2:l.pos-1])
}

func lexString(l *lexer) stateFn {
	for {
		switch l.next() {
		case eof:
			return l.errorf("unterminated string")
		case '"':
			return l.emit(tokString, l.input[l.start+1:l.pos-1])
		}
	}
}

func lexNumber(l *lexer) stateFn {
	digits := "0123456789_"
	l.acceptRun(digits)
	isFloat := false
	if l.accept(".") {
		if l.accept(".") {
			l.backup()
			l.backup()
			val, _ := strconv.Atoi(l.input[l.start:l.pos])
			return l.emit(tokInt, val)
		}
		isFloat = true
		l.acceptRun(digits)
	}
	if isFloat {
		val, _ := strconv.ParseFloat(l.input[l.start:l.pos], 64)
		return l.emit(tokFloat, val)
	}
	val, _ := strconv.Atoi(l.input[l.start:l.pos])
	return l.emit(tokInt, val)
}

func lexAlphaNumeric(l *lexer) stateFn {
	for {
		c := l.next()
		if isAlphaNumeric(c) {
			continue
		}
		l.backup()
		word := l.input[l.start:l.pos]
		if word == "true" || word == "false" {
			return l.emit(tokBool, nil)
		}
		return l.emit(tokWord, nil)
	}
}

func lexOrientation(l *lexer) stateFn {
	if l.accept("+-") {
		if l.accept("EW") {
			l.accept("+-")
			l.accept("NW")
		} else if l.accept("NW") {
			l.accept("+-")
			l.accept("EW")
		} else {
			return l.errorf("invalid orientation")
		}
	} else if l.accept("EW") {
		l.accept("NW")
	} else if l.accept("NW") {
		l.accept("EW")
	} else {
		return l.errorf("invalid orientation")
	}
	return l.emit(tokOrientation, nil)
}

func (l *lexer) nextToken(state stateFn) token {
	if l.next() == eof {
		state = l.emit(tokEOF, nil)
	} else {
		l.backup()
	}
	for {
		if state == nil {
			return l.tok
		}
		state = state(l)
	}
}

func lex(input string, start int, end int) *lexer {
	return &lexer{input: input, start: start, end: end, pos: start}
}
