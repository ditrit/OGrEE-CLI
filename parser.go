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

func findNext(buffer string, start int, substringList []string) int {
	minIdx := len(buffer)
	for _, s := range substringList {
		idx := strings.Index(buffer[start:], s)
		if idx < minIdx {
			minIdx = idx
		}
	}
	return minIdx
}

func findNextSpace(buffer string, start int) int {
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

func parseWord(buffer string, start int, end int) (string, *ParserError) {
	word := buffer[start:end]
	if !regexMatch(`[A-Za-z_][A-Za-z0-9_\-]*`, word) {
		return "", &ParserError{
			buffer:  buffer,
			message: "Invalid word : " + word,
			start:   start,
			end:     end,
		}
	}
	return word, nil
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
		varName, err := parseWord(buffer, varIndex+2, endVarIndex)
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
	end := findNextSpace(buffer, start)
	if start == end {
		return &pathNode{&strLeaf{"."}, STD}, end, nil
	}
	path, err := parseString(buffer, start, end)
	if err != nil {
		return nil, 0, err
	}
	end = skipWhiteSpaces(buffer, end)
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
		wordEnd := findNextSpace(buffer, cursor)
		arg, err := parseWord(buffer, cursor, wordEnd)
		if err != nil {
			return nil, 0, err
		}
		cursor = skipWhiteSpaces(buffer, cursor+len(arg))
		if sliceContains(allowedArgs, arg) {
			valueEnd := findNextSpace(buffer, cursor)
			value := buffer[cursor:valueEnd]
			args[arg] = value
			cursor = valueEnd
		} else if sliceContains(allowedFlags, arg) {
			args[arg] = nil
		} else {
			return nil, 0, &ParserError{
				buffer:  buffer,
				message: fmt.Sprintf("unexpected argument : %s", arg),
				start:   cursor,
				end:     wordEnd,
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

func parseGetU(buffer string, start int) (node, *ParserError) {
	path, cursor, err := parsePath(buffer, start)
	if err != nil {
		return nil, err
	}
	if cursor == len(buffer) {
		return &getUNode{path, &intLeaf{0}}, nil
	}
	u, err := parseInt(buffer, cursor, len(buffer))
	if err != nil {
		return nil, err
	}
	return &getUNode{path, &intLeaf{u}}, nil
}

func parseGetSlot(buffer string, start int) (node, *ParserError) {
	path, cursor, err := parsePath(buffer, start)
	if err != nil {
		return nil, err
	}
	slotName, err := parseWord(buffer, cursor, len(buffer))
	return &getUNode{path, &strLeaf{slotName}}, nil
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
		"ls":      parseLs,
		"get":     parseGet,
		"getu":    parseGetU,
		"getslot": parseGetSlot,
	}
	parseFunc, ok := dispatch[commandKeyWord]
	if ok {
		return parseFunc(buffer, cursor)
	}

	panic("command not processed")
}
