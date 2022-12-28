package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
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

func sliceContains(slice []string, s string) bool {
	for _, str := range slice {
		if str == s {
			return true
		}
	}
	return false
}

func sliceContainsPrefix(slice []string, prefix string) bool {
	for _, str := range slice {
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

func findNextWhiteSpace(buffer string, start int) int {
	i := start
	for buffer[i] != ' ' && buffer[i] != '\t' {
		i += 1
	}
	return i
}

func parseCommandKeyWord(buffer string, start int) (string, *ParserError) {
	end := start + 1
	for end < len(buffer) && sliceContainsPrefix(commands, buffer[start:end]) {
		end++
	}
	end--
	if end == start {
		return "", &ParserError{
			buffer:  buffer,
			message: "command name expected",
			start:   start,
			end:     end,
		}
	}
	return buffer[start:end], nil
}

func parseIdentifier(buffer string, start int, end int) (string, *ParserError) {
	identifier := buffer[start:end]
	if !regexMatch(`[A-Za-z_][A-Za-z0-9_\-]*`, identifier) {
		return "", &ParserError{
			buffer:  buffer,
			message: "Invalid identifier name : " + identifier,
			start:   start,
			end:     end,
		}
	}
	return identifier, nil
}

func parseInt(buffer string, start int, end int) (int, *ParserError) {
	val, err := strconv.Atoi(buffer[start:end])
	if err != nil {
		return 0, &ParserError{
			buffer:  buffer,
			message: "integer expected",
			start:   start,
			end:     end,
		}
	}
	return val, nil
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
			return nil, &ParserError{
				buffer:  buffer,
				message: "&{ opened but never closed",
				start:   varIndex,
				end:     end,
			}
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

func parsePath(buffer string, start int) (node, int, *ParserError) {
	end := findNextWhiteSpace(buffer, start)
	if start == end {
		return &pathNode{&strLeaf{"."}, STD}, end, nil
	}
	path, err := parseString(buffer, start, end)
	if err != nil {
		return nil, 0, err
	}
	return &pathNode{path, STD}, end, nil
}

func parseArgs(allowedArgs []string, allowedFlags []string, buffer string, start int) (
	map[string]any, int, *ParserError,
) {
	args := map[string]any{}
	cursor := start
	for buffer[cursor] == '-' {
		cursor++
		cursor = skipWhiteSpaces(buffer, cursor)
		identifierEnd := findNextWhiteSpace(buffer, cursor)
		identifier, err := parseIdentifier(buffer, cursor, identifierEnd)
		if err != nil {
			return nil, 0, err
		}
		cursor = skipWhiteSpaces(buffer, cursor+len(identifier))
		if sliceContains(allowedArgs, identifier) {
			valueEnd := findNextWhiteSpace(buffer, cursor)
			args[identifier] = buffer[cursor:valueEnd]
			cursor = valueEnd
		} else if sliceContains(allowedFlags, identifier) {
			args[identifier] = nil
		} else {
			return nil, 0, &ParserError{
				buffer:  buffer,
				message: fmt.Sprintf("unexpected argument : %s", identifier),
				start:   cursor,
				end:     identifierEnd,
			}
		}
		cursor = skipWhiteSpaces(buffer, cursor)
	}
	return args, cursor, nil
}

func parseLsObj(lsIdx int, buffer string, start int) (node, *ParserError) {
	args, cursor, err := parseArgs([]string{"s", "f"}, []string{"r"}, buffer, start)
	if err != nil {
		return nil, err
	}
	path, cursor, err := parsePath(buffer, cursor)
	if err != nil {
		return nil, err
	}
	argsAfter, cursor, err := parseArgs([]string{"s", "f"}, []string{"r"}, buffer, cursor)
	if err != nil {
		return nil, err
	}
	for arg, value := range argsAfter {
		args[arg] = value
	}
	return &lsObjNode{path, lsIdx, args}, nil // TODO : Adapt lsObjNode
}

func parseLs(buffer string, start int) (node, *ParserError) {
	args, cursor, err := parseArgs([]string{"s"}, []string{}, buffer, start)
	if err != nil {
		return nil, err
	}
	path, cursor, err := parsePath(buffer, cursor)
	if err != nil {
		return nil, err
	}
	afterArgs, cursor, err := parseArgs([]string{"s"}, []string{}, buffer, cursor)
	if err != nil {
		return nil, err
	}
	for arg, value := range afterArgs {
		args[arg] = value
	}
	return &lsAttrGenericNode{path, args}, nil
}

func parseGet(buffer string, start int) (node, *ParserError) {
	path, _, err := parsePath(buffer, start)
	if err != nil {
		return nil, err
	}
	return &getObjectNode{path}, nil
}

func ParseGetU(buffer string, start int) (node, *ParserError) {
	path, cursor, err := parsePath(buffer, start)
	if err != nil {
		return nil, err
	}
	u, err := parseInt(buffer, cursor, len(buffer))
	if err != nil {
		return nil, err
	}
	return &getUNode{path, &intLeaf{u}}, nil
}

func firstNonAscii(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return i
		}
	}
	return -1
}

func Parse(buffer string) (node, *ParserError) {
	firstNonAsciiIdx := firstNonAscii(buffer)
	if firstNonAsciiIdx != -1 {
		return nil, &ParserError{
			buffer:  buffer,
			message: "command should only contain ascii characters",
			start:   firstNonAsciiIdx,
			end:     firstNonAsciiIdx,
		}
	}

	cursor := skipWhiteSpaces(buffer, 0)
	commandKeyWord, err := parseCommandKeyWord(buffer, cursor)
	if err != nil {
		return nil, err
	}
	println("Command key word :", commandKeyWord)
	cursor = skipWhiteSpaces(buffer, cursor+len(commandKeyWord))

	if lsIdx, ok := lsCommands[commandKeyWord]; ok {
		return parseLsObj(lsIdx, buffer, cursor)
	}

	dispatch := map[string]func(buffer string, start int) (node, *ParserError){
		"ls":   parseLs,
		"get":  parseGet,
		"getu": ParseGetU,
	}
	parseFunc, ok := dispatch[commandKeyWord]
	if !ok {
		panic("command not processed")
	}
	return parseFunc(buffer, cursor)
}
