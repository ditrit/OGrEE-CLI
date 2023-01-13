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
	tokWord
	tokLeftParen  // '('
	tokInt        // integer constant
	tokFloat      // float constant
	tokBool       // boolean constant
	tokString     // quoted string
	tokRightParen // ')'
	tokDeref      // variable dereferenciation
	tokAdd        // '+'
	tokSub        // '-'
)

type token struct {
	t   tokenType
	pos int
	val interface{}
}

const eof = 0

type lexer struct {
	input string
	pos   int
	start int
	end   int
	t     token
}

type stateFn func(*lexer) stateFn

func (l *lexer) emit(t tokenType) stateFn {
	val := l.input[l.start:l.pos]
	newToken := token{
		t:   t,
		pos: l.start,
		val: val,
	}
	return l.emitToken(newToken)
}

func (l *lexer) emitToken(t token) stateFn {
	l.t = t
	l.start = l.pos
	return nil
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	val := fmt.Sprintf(format, args...)
	l.t = token{
		t:   tokError,
		pos: l.start,
		val: val,
	}
	l.start = 0
	l.pos = 0
	l.input = l.input[:0]
	return nil
}

func (l *lexer) next() byte {
	if l.pos >= l.end {
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
	l.start = l.pos
}

func (l *lexer) peek() byte {
	char := l.next()
	l.backup()
	return char
}

func (l *lexer) accept(valid string) bool {
	if strings.Index(valid, string(l.next())) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.Index(valid, string(l.next())) >= 0 {
	}
	l.backup()
}

func (l *lexer) atTerminator() bool {
	r := l.peek()
	switch r {
	case eof, ' ', '\t', ')', '}':
		return true
	}
	return false
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
	for {
		c := l.next()
		switch c {
		case eof:
			return nil
		case ' ', '\t':
			l.ignore()
		case '$':
			return lexDeref
		case '"':
			return lexString
		case '(':
			return l.emit(tokLeftParen)
		case ')':
			return l.emit(tokRightParen)
		case '+':
			return l.emit(tokAdd)
		case '-':
			return l.emit(tokSub)
		}
		if isDigit(c) || c == '.' {
			return lexNumber
		}
		if isLetter(c) {
			return lexAlphaNumeric
		}
	}
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
	return l.emitToken(token{t: tokDeref, pos: l.start, val: l.input[l.start+2 : l.pos-1]})
}

func lexString(l *lexer) stateFn {
	for {
		switch l.next() {
		case eof:
			return l.errorf("unterminated string")
		case '"':
			return l.emitToken(token{t: tokString, pos: l.start, val: l.input[l.start+1 : l.pos-1]})
		}
	}
}

func lexNumber(l *lexer) stateFn {
	digits := "0123456789_"
	l.acceptRun(digits)
	isFloat := false
	if l.accept(".") {
		isFloat = true
		l.acceptRun(digits)
	}
	if isFloat {
		val, _ := strconv.ParseFloat(l.input[l.start:l.pos])
		newToken := token{
			t:   tokFloat,
			pos: l.start,
			val: val,
		}
		return l.emitToken(newToken)
	}
	val, _ := strconv.ParseInt(l.input[l.start:l.pos])
	newToken := token{
		t:   tokInt,
		pos: l.start,
		val: val,
	}
	return l.emitToken(newToken)
}

func lexAlphaNumeric(l *lexer) stateFn {
	for {
		c := l.next()
		if isAlphaNumeric(c) {
			continue
		}
		l.backup()
		if !l.atTerminator() {
			return l.errorf("bad character %c", c)
		}
		word := l.input[l.start:l.pos]
		if word == "true" || word == "false" {
			return l.emit(tokBool)
		}
		return l.emit(tokWord)
	}
}

func (l *lexer) nextToken() token {
	l.t = token{
		t:   tokEOF,
		pos: l.pos,
		val: "EOF",
	}
	state := lexExpr
	for {
		state = state(l)
		if state == nil {
			return l.t
		}
	}
}

func lex(input string, start int, end int) *lexer {
	return &lexer{input: input, start: start, end: end, pos: start}
}
