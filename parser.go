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

var lsCommands = map[string]int{
	"lsten":      0,
	"lssite":     1,
	"lsbldg":     2,
	"lsroom":     3,
	"lsrack":     4,
	"lsdev":      5,
	"lsac":       6,
	"lspanel":    7,
	"lscabinet":  8,
	"lscorridor": 9,
	"lssensor":   10,
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

type ParserError struct {
	buffer  string
	message string
	start   int
	end     int
}

func (err *ParserError) Error() string {
	return fmt.Sprintf("%s, start : %d, end : %d",
		err.message, err.start, err.end,
	)
}

func skipWhiteSpaces(buffer string, start int) int {
	i := start
	for i < len(buffer) && (buffer[i] == ' ' || buffer[i] == '\t') {
		i += 1
	}
	return i
}

func parseCommandKeyWord(buffer string, start int) (string, *ParserError) {
	end := start + 1
	for end < len(buffer) && listHasPrefix(commands, buffer[start:end]) {
		end++
	}
	end--
	if end == start {
		err := &ParserError{
			buffer:  buffer,
			message: "command name expected",
			start:   start,
			end:     end,
		}
		return "", err
	}
	return buffer[start:end], nil
}

func parseIdentifier(buffer string, start int, end int) (string, *ParserError) {
	identifier := buffer[start:end]
	if !regexMatch(`[A-Za-z0-9_][A-Za-z0-9_\-]*`, identifier) {
		err := &ParserError{
			buffer:  buffer,
			message: "Invalid identifier name : " + identifier,
			start:   start,
			end:     end,
		}
		return "", err
	}
	return identifier, nil
}

func parseString(buffer string, start int, end int) (node, *ParserError) {
	nodesToConcat := []node{}
	cursor := start
	for cursor < end {
		varIndex := strings.Index(buffer[cursor:end], "&{")
		if varIndex == -1 {
			break
		}
		varIndex += cursor
		leftStr := buffer[cursor:varIndex]
		if leftStr != "" {
			nodesToConcat = append(nodesToConcat, &strLeaf{leftStr})
		}
		endVarIndex := strings.Index(buffer[varIndex+2:end], "}")
		if endVarIndex == -1 {
			err := &ParserError{
				buffer:  buffer,
				message: "&{ opened but never closed",
				start:   varIndex,
				end:     end,
			}
			return nil, err
		}
		varName, err := parseIdentifier(buffer, varIndex+2, endVarIndex)
		if err != nil {
			return nil, err
		}
		nodesToConcat = append(nodesToConcat, &symbolReferenceNode{varName})
		cursor = endVarIndex + 1
	}
	if len(nodesToConcat) == 0 {
		return &strLeaf{buffer[start:end]}, nil
	} else if len(nodesToConcat) == 1 {
		return nodesToConcat[0], nil
	}
	return &concatNode{nodes: nodesToConcat}, nil
}

func parsePath(buffer string, start int, end int) (node, *ParserError) {
	if start == end {
		return &pathNode{&strLeaf{"."}, STD}, nil
	}
	path, err := parseString(buffer, start, end)
	if err != nil {
		return nil, err
	}
	return &pathNode{path, STD}, nil
}

func parseArgs(buffer string, start int, end int) (map[string]string, *ParserError) {
	cursor := start
	args := map[string]string{}
	for cursor < end {
		cursor = skipWhiteSpaces(buffer, cursor)
		identifier, err := parseIdentifier(buffer, cursor, end)
		if err != nil {
			return nil, err
		}
		cursor = skipWhiteSpaces(buffer, cursor+len(identifier))
		value, err := parseString()
		args[identifier] = 
	}
	return nil, nil
}

func Parse(buffer string) (node, *ParserError) {
	cursor := skipWhiteSpaces(buffer, 0)
	commandKeyWord, err := parseCommandKeyWord(buffer, cursor)
	if err != nil {
		return nil, err
	}
	println("Command key word :", commandKeyWord)
	cursor = skipWhiteSpaces(buffer, cursor+len(commandKeyWord))

	if lsIdx, ok := lsCommands[commandKeyWord]; ok {
		path, err := parsePath(buffer, cursor, len(buffer))
		if err != nil {
			return nil, err
		}
		return &lsObjNode{path, lsIdx, nil}, nil
	}

	switch commandKeyWord {
	case "get":
		path, err := parsePath(buffer, cursor, len(buffer))
		if err != nil {
			return nil, err
		}
		return &getObjectNode{path}, nil
	case "ls":
		path, err := parsePath(buffer, cursor, len(buffer))
		if err != nil {
			return nil, err
		}
		return &lsNode{path}, nil

	}
	panic("command not processed")
}
