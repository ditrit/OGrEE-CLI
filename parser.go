package main

import (
	"fmt"
	"regexp"
	"strings"
)

var eof = rune(0)

var commands = []string{
	"get", "getu", "getslot",
	"hc",
	"+", "-", "=",
	"print",
	"unset",
	"selection",
	".cmds",
	".template",
	".var",
	"ui.delay", "ui.wireframe", "ui.infos", "ui.debug", "ui.highlight", "ui.hl",
	"camera.move", "camera.wait", "camera.translate",
	"link", "unlink",
	"lsten", "lssite", "lsbldg", "lsroom", "lsrack", "lsdev", "lscabinet", "lssensor",
	"lsac", "lspanel", "lscorridor", "lsenterprise", "tree", "lsog",
	"env",
	"cd",
	"pwd",
	"clear",
	"grep",
	"ls",
	"exit",
	"len",
	"man",
	"drawable", "draw", "undraw",
}

func listHasPrefix(list []string, prefix string) bool {
	for _, str := range list {
		if strings.HasPrefix(str, prefix) {
			return true
		}
	}
	return false
}

func regexMatch(regex string, str string) bool {
	reg, err := regexp.Compile(regex)
	if err != nil {
		panic("Regexp compilation error")
	}
	reg.Longest()
	return str == reg.FindString(str)
}

type Parser struct {
	buffer string
}

type Cursor struct {
	start   int
	current int
	end     int
}

type ParserError struct {
	message string
	cursor  *Cursor
}

func (err *ParserError) Error() string {
	return fmt.Sprintf("%s, start : %d, current : %d",
		err.message, err.cursor.start, err.cursor.current,
	)
}

func (parser *Parser) Error(message string, cursor *Cursor) *ParserError {
	return &ParserError{message, cursor}
}

func (cursor *Cursor) forward(n int) {
	cursor.current += n
	if cursor.current > cursor.end {
		panic("Cursor went too far right")
	}
}

func (cursor *Cursor) backward(n int) {
	cursor.current -= n
	if cursor.current < 0 {
		panic("Cursor went too far left")
	}
}

func (cursor *Cursor) isStart() bool {
	return cursor.current == cursor.start
}

func (cursor *Cursor) isEnd() bool {
	return cursor.current == cursor.end
}

func (cursor *Cursor) left() *Cursor {
	var left Cursor = *cursor
	left.end = left.current
	left.current = left.start
	return &left
}

func (cursor *Cursor) right() *Cursor {
	var right Cursor = *cursor
	right.start = right.current
	return &right
}

func (parser *Parser) read(cursor *Cursor) string {
	return parser.buffer[cursor.start:cursor.current]
}

func (parser *Parser) readAll(cursor *Cursor) string {
	return parser.buffer[cursor.start:cursor.end]
}

func (parser *Parser) readRemaining(cursor *Cursor) string {
	return parser.buffer[cursor.current:cursor.end]
}

func (parser *Parser) expectCommandKeyWord(cursor *Cursor) (string, *ParserError) {
	for !cursor.isEnd() && listHasPrefix(commands, parser.read(cursor)) {
		cursor.forward(1)
	}
	cursor.backward(1)
	commandKeyWord := parser.read(cursor)
	if commandKeyWord == "" {
		return "", parser.Error("command name expected", cursor)
	}
	return commandKeyWord, nil
}

func (parser *Parser) expectVarName(cursor *Cursor) (string, *ParserError) {
	varName := parser.readAll(cursor)
	if !regexMatch(`[A-Za-z0-9_][A-Za-z0-9_\-]*`, varName) {
		return "", parser.Error("Invalid variable name : "+varName, cursor)
	}
	return varName, nil
}

func (parser *Parser) expectString(cursor *Cursor) (node, *ParserError) {
	nodesToConcat := []node{}
	for {
		remainingCaracters := parser.readRemaining(cursor)
		varIndex := strings.Index(remainingCaracters, "&{")
		if varIndex == -1 {
			break
		}
		cursor.forward(varIndex)
		leftStr := parser.read(cursor)
		if leftStr != "" {
			nodesToConcat = append(nodesToConcat, &strLeaf{leftStr})
		}
		cursor.forward(2)
		*cursor = *cursor.right()
		remainingCaracters = parser.readRemaining(cursor)
		endVarIndex := strings.Index(remainingCaracters, "}")
		cursor.forward(endVarIndex)
		varName, err := parser.expectVarName(cursor.left())
		if err != nil {
			return nil, err
		}
		nodesToConcat = append(nodesToConcat, &symbolReferenceNode{varName})
		cursor.forward(1)
	}
	if len(nodesToConcat) == 0 {
		return &strLeaf{parser.readAll(cursor)}, nil
	}
	return &concatNode{nodes: nodesToConcat}, nil
}

func (parser *Parser) expectPath(cursor *Cursor) (node, *ParserError) {
	path, err := parser.expectString(cursor)
	if err != nil {
		return nil, err
	}
	return &pathNode{path, STD}, nil
}

func (parser *Parser) rootParse() (node, *ParserError) {
	commandKeyWord, err := parser.expectCommandKeyWord()
	if err != nil {
		return nil, err
	}
	println("Command key word :", commandKeyWord)

	switch commandKeyWord {
	case "get":
		path, err := parser.expectPath()
		if err != nil {
			return nil, err
		}
		return &getObjectNode{path}, nil
	case "ls":

	}

	panic("command not processed")
}

func Parse(command string) node {
	buffer := []rune(command)
	parser := Parser{
		buffer:        buffer,
		startCursor:   0,
		currentCursor: 0,
		endCursor:     len(buffer),
	}

	root, err := parser.rootParse()
	if err != nil {
		println("Error :", err.Error())
	}

	return
}
