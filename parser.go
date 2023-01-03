package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

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
	"lsac", "lspanel", "lscorridor", "lsenterprise",
	"tree",
	"lsog",
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

var wordRegex = `([A-Za-z_][A-Za-z0-9_\-]*)`
var valueRegex = `( (\S*) | (".*") | (\(.*\)) )`

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

func skipWhiteSpaces(buffer string, start int, end int) int {
	i := start
	for i < end && (buffer[i] == ' ' || buffer[i] == '\t') {
		i += 1
	}
	return i
}

func findNext(substring string, buffer string, start int, end int) int {
	idx := strings.Index(buffer[start:end], substring)
	if idx != -1 {
		return idx
	}
	return end
}

func findNextAmong(substringList []string, buffer string, start int, end int) int {
	minIdx := end
	for _, s := range substringList {
		idx := strings.Index(buffer[start:end], s)
		if idx < minIdx {
			minIdx = idx
		}
	}
	return minIdx
}

func findClosing(buffer string, start int, end int) int {
	openToClose := map[byte]byte{'(': ')', '{': '}', '[': ']'}
	open := buffer[start]
	close, ok := openToClose[open]
	if !ok {
		panic("invalid opening character")
	}
	cursor := start
	stackCount := 0
	for cursor < end {
		if buffer[cursor] == open {
			stackCount++
		}
		if buffer[cursor] == close {
			stackCount--
		}
		if stackCount == 0 {
			return cursor
		}
	}
	return -1
}

func parseExact(word string, buffer string, start int, end int) (bool, int) {
	if start+len(word) < end && buffer[start:start+len(word)] == word {
		return true, start + len(word)
	}
	return false, start
}

func parseCommandKeyWord(buffer string, start int) (string, int, *ParserError) {
	end := start + 1
	for end < len(buffer) && sliceContainsPrefix(commands, buffer[start:end]) {
		end++
	}
	if end == start {
		return "", 0, &ParserError{
			buffer:  buffer,
			message: "command name expected",
			start:   start,
			end:     end,
		}
	}
	return buffer[start:end], end, nil
}

func parseWord(buffer string, start int, end int) (string, int, *ParserError) {
	cursor := end
	for cursor > start && !regexMatch(wordRegex, buffer[start:cursor]) {
		cursor--
	}
	if cursor == start {
		return "", 0, &ParserError{
			buffer:  buffer,
			message: "Invalid word",
			start:   start,
			end:     end,
		}
	}
	return buffer[start:cursor], cursor, nil
}

func parseSeparatedWords(sep byte, buffer string, start int, end int) ([]string, *ParserError) {
	var word string
	var err *ParserError
	cursor := start
	words := make([]string, 0)
	for {
		word, cursor, err = parseWord(buffer, cursor, end)
		if err != nil {
			return nil, err
		}
		words = append(words, word)
		cursor = skipWhiteSpaces(buffer, cursor, end)
		if cursor == end {
			break
		}
		if buffer[cursor] != sep {
			return nil, &ParserError{
				buffer:  buffer,
				message: "comma expected",
				start:   cursor,
				end:     cursor,
			}
		}
		cursor++
	}
	return words, nil
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
	var varName string
	var err *ParserError
	nodesToConcat := []node{}
	cursor := start
	for cursor < end {
		varIndex := findNext("&{", buffer, cursor, end)
		if varIndex == end {
			break
		}
		leftStr := buffer[cursor:varIndex]
		if leftStr != "" {
			nodesToConcat = append(nodesToConcat, &strLeaf{leftStr})
		}
		endVarIndex := findNext("}", buffer, varIndex, end)
		if endVarIndex == end {
			return nil, &ParserError{
				buffer:  buffer,
				message: "&{ opened but never closed",
				start:   varIndex,
				end:     end,
			}
		}
		varName, cursor, err = parseWord(buffer, varIndex+2, endVarIndex)
		if err != nil {
			return nil, err
		}
		nodesToConcat = append(nodesToConcat, &symbolReferenceNode{varName})
		cursor++
	}
	if len(nodesToConcat) == 0 {
		return &strLeaf{buffer[start:end]}, nil
	} else if len(nodesToConcat) == 1 {
		return nodesToConcat[0], nil
	}
	return &concatNode{nodes: nodesToConcat}, nil
}

// func parseQuotedString(buffer string, start int, end int) (node, *ParserError) {
// 	if buffer[start] != '"' {
// 		return nil, &ParserError{
// 			buffer:  buffer,
// 			message: "double quote \" expected",
// 			start:   start,
// 			end:     start,
// 		}
// 	}
// 	end = findNext("\"", buffer, start+1, end)
// 	return parseString(buffer, start+1, end)
// }

func parsePath(buffer string, start int, end int) (node, int, *ParserError) {
	endPath := findNext(" ", buffer, start, end)
	if start == endPath {
		return &pathNode{&strLeaf{"."}, STD}, end, nil
	}
	path, err := parseString(buffer, start, endPath)
	if err != nil {
		return nil, 0, err
	}
	cursor := skipWhiteSpaces(buffer, endPath, end)
	return &pathNode{path, STD}, cursor, nil
}

func parseArgValue(buffer string, start int, end int) (string, int, *ParserError) {
	if buffer[start] == '(' {
		close := findClosing(buffer, start, end)
		if close == end {
			return "", 0, &ParserError{
				buffer:  buffer,
				message: "( opened but never closed",
				start:   start,
				end:     end,
			}
		}
		return buffer[start : close+1], close + 1, nil
	} else if buffer[start] == '"' {
		endQuote := findNext("\"", buffer, start, end)
		return buffer[start:endQuote], endQuote + 1, nil
	}
	endValue := findNext(" ", buffer, start, end)
	endValueAndSpaces := skipWhiteSpaces(buffer, endValue, end)
	return buffer[start:endValue], endValueAndSpaces, nil
}

func parseSingleArg(allowedArgs []string, allowedFlags []string, buffer string, start int, end int) (
	string, string, int, *ParserError,
) {
	cursor := start
	cursor++ // skip dash
	cursor = skipWhiteSpaces(buffer, cursor, end)
	wordEnd := findNext(" ", buffer, cursor, end)
	arg, cursor, err := parseWord(buffer, cursor, wordEnd)
	if err != nil {
		return "", "", 0, err
	}
	cursor = skipWhiteSpaces(buffer, cursor, end)

	var value string
	if sliceContains(allowedArgs, arg) {
		value, cursor, err = parseArgValue(buffer, cursor, end)
		if err != nil {
			return "", "", 0, err
		}
	} else if sliceContains(allowedFlags, arg) {
		value = ""
	} else {
		panic("unexpected argument")
	}

	cursor = skipWhiteSpaces(buffer, cursor, end)
	return arg, value, cursor, nil
}

func buildSingleArgRegex(allowedArgs []string) string {
	argNameRegex := strings.Join(allowedArgs, "|")
	return `([\-]\s*(` + argNameRegex + `)\s+` + valueRegex + `\s*)`
}

func buildSingleFlagRegex(allowedFlags []string) string {
	flagNameRegex := strings.Join(allowedFlags, "|")
	return `([\-]\s*(` + flagNameRegex + `)\s*)`
}

func buildMultipleArgsRegex(allowedArgs []string, allowedFlags []string) string {
	singleArgRegex := buildSingleArgRegex(allowedArgs)
	singleFlagRegex := buildSingleFlagRegex(allowedFlags)
	return "((" + singleArgRegex + ")|(" + singleFlagRegex + "))*"
}

func parseArgs(allowedArgs []string, allowedFlags []string, buffer string, start int) (
	map[string]string, int, int, *ParserError,
) {
	multipleArgsRegex := buildMultipleArgsRegex(allowedArgs, allowedFlags)

	endArgsLeft := len(buffer)
	for endArgsLeft > start && !regexMatch(multipleArgsRegex, buffer[start:endArgsLeft]) {
		endArgsLeft--
	}
	startArgsRight := endArgsLeft
	for startArgsRight < len(buffer) && !regexMatch(multipleArgsRegex, buffer[startArgsRight:]) {
		startArgsRight++
	}
	argsBuffer := buffer[start:endArgsLeft] + buffer[startArgsRight:]

	args := map[string]string{}
	cursor := start
	for cursor < len(buffer) && buffer[cursor] == '-' {
		arg, value, newCursor, err := parseSingleArg(
			allowedArgs, allowedFlags, argsBuffer, cursor, len(argsBuffer))
		if err != nil {
			return nil, 0, 0, err
		}
		args[arg] = value
		cursor = newCursor
	}
	return args, endArgsLeft, startArgsRight, nil
}

func parseLsObj(lsIdx int, buffer string, start int) (node, *ParserError) {
	args, pathStart, _, err := parseArgs([]string{"s", "f"}, []string{"r"}, buffer, start)
	if err != nil {
		return nil, err
	}
	path, _, err := parsePath(buffer, pathStart, len(buffer))
	if err != nil {
		return nil, err
	}
	_, recursive := args["r"]
	sort := args["s"]

	//msg := "Please provide a quote enclosed string for '-f' with arguments separated by ':'. Or provide an argument with printf formatting (ie -f (\"%d\",arg1))"

	var attrList []string
	var format string
	if formatArg, ok := args["f"]; ok {
		if regexMatch(`\(\s*".*"\s*,.+\)`, formatArg) {
			startFormat := findNext("\"", formatArg, 1, len(formatArg))
			endFormat := findNext("\"", formatArg, startFormat+1, len(formatArg))
			format = formatArg[startFormat+1 : endFormat]
			cursor := findNext(",", formatArg, endFormat, len(formatArg)) + 1
			attrList, err = parseSeparatedWords(',', formatArg, cursor, len(formatArg)-1)
			if err != nil {
				return nil, err
			}
		} else {
			attrList, err = parseSeparatedWords(':', formatArg, 0, len(formatArg))
			if err != nil {
				return nil, err
			}
		}
	}
	return &lsObjNode{path, lsIdx, recursive, sort, attrList, format}, nil
}

func parseLs(buffer string, start int) (node, *ParserError) {
	args, pathStart, _, err := parseArgs([]string{"s", "f"}, []string{"r"}, buffer, start)
	if err != nil {
		return nil, err
	}
	path, _, err := parsePath(buffer, pathStart, len(buffer))
	if err != nil {
		return nil, err
	}
	if attr, ok := args["s"]; ok {
		return &lsAttrGenericNode{path, attr}, nil
	}
	return &lsNode{path}, nil
}

func parseGet(buffer string, start int) (node, *ParserError) {
	path, _, err := parsePath(buffer, start, len(buffer))
	if err != nil {
		return nil, err
	}
	return &getObjectNode{path}, nil
}

func parseGetU(buffer string, start int) (node, *ParserError) {
	path, cursor, err := parsePath(buffer, start, len(buffer))
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
	path, cursor, err := parsePath(buffer, start, len(buffer))
	if err != nil {
		return nil, err
	}
	slotName, _, err := parseWord(buffer, cursor, len(buffer))
	if err != nil {
		return nil, err
	}
	return &getUNode{path, &strLeaf{slotName}}, nil
}

func parseUndraw(buffer string, start int) (node, *ParserError) {
	if start == len(buffer) {
		return &undrawNode{nil}, nil
	}
	path, _, err := parsePath(buffer, start, len(buffer))
	if err != nil {
		return nil, err
	}
	return &undrawNode{path}, nil
}

func parseDraw(buffer string, start int) (node, *ParserError) {
	args, cursor, rightArgsStart, err := parseArgs([]string{"f"}, []string{}, buffer, start)
	if err != nil {
		return nil, err
	}
	path, cursor, err := parsePath(buffer, cursor, rightArgsStart)
	if err != nil {
		return nil, err
	}
	depth := 0
	if cursor < rightArgsStart {
		depth, err = parseInt(buffer, cursor, len(buffer))
		if err != nil {
			return nil, err
		}
	}
	return &drawNode{path, depth, args}, nil
}

func parseDrawable(buffer string, start int) (node, *ParserError) {
	path, cursor, err := parsePath(buffer, start, len(buffer))
	if err != nil {
		return nil, err
	}
	if cursor == len(buffer) {
		return &isEntityDrawableNode{path}, nil
	}
	attrName, _, err := parseWord(buffer, cursor, len(buffer))
	if err != nil {
		return nil, err
	}
	return &isAttrDrawableNode{path, attrName}, nil
}

func parseHc(buffer string, start int) (node, *ParserError) {
	path, cursor, err := parsePath(buffer, start, len(buffer))
	if err != nil {
		return nil, err
	}
	if cursor == len(buffer) {
		return &hierarchyNode{path, 1}, nil
	}
	depth, err := parseInt(buffer, cursor, len(buffer))
	if err != nil {
		return nil, err
	}
	return &hierarchyNode{path, depth}, nil
}

func parseUnset(buffer string, start int) (node, *ParserError) {
	args, cursor, rightArgsStart, err := parseArgs([]string{"f", "v"}, []string{}, buffer, start)
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		path, _, err := parsePath(buffer, cursor, rightArgsStart)
		if err != nil {
			return nil, err
		}
		return &unsetAttrNode{path}, nil
	}
	if funcName, ok := args["f"]; ok {
		return &unsetFuncNode{funcName}, nil
	}
	if varName, ok := args["v"]; ok {
		return &unsetVarNode{varName}, nil
	}
	panic("unexpected argument while parsing unset command")
}

func parseEnv(buffer string, start int) (node, *ParserError) {
	endArg := findNextAmong([]string{" ", "="}, buffer, start, len(buffer))
	arg, cursor, err := parseWord(buffer, start, endArg)
	if err != nil {
		return nil, err
	}
	cursor = skipWhiteSpaces(buffer, cursor, len(buffer))
	if buffer[cursor] != '=' {
		return nil, &ParserError{
			buffer:  buffer,
			message: "= expected",
			start:   cursor,
			end:     cursor,
		}
	}
	cursor++
	value, err := parseString(buffer, cursor, len(buffer))
	if err != nil {
		return nil, err
	}
	return &setEnvNode{arg, value}, nil
}

func parseDelete(buffer string, start int) (node, *ParserError) {
	deleteSelection, _ := parseExact("selection", buffer, start, len(buffer))
	if deleteSelection {
		return &deleteSelectionNode{}, nil
	}
	path, _, err := parsePath(buffer, start, len(buffer))
	if err != nil {
		return nil, err
	}
	return &deleteObjNode{path}, nil
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

	cursor := skipWhiteSpaces(buffer, 0, len(buffer))
	commandKeyWord, cursor, err := parseCommandKeyWord(buffer, cursor)
	if err != nil {
		return nil, err
	}
	println("Command key word :", commandKeyWord)
	cursor = skipWhiteSpaces(buffer, cursor, len(buffer))

	if lsIdx, ok := lsCommands[commandKeyWord]; ok {
		return parseLsObj(lsIdx, buffer, cursor)
	}

	dispatch := map[string]func(buffer string, start int) (node, *ParserError){
		"ls":       parseLs,
		"get":      parseGet,
		"getu":     parseGetU,
		"getslot":  parseGetSlot,
		"undraw":   parseUndraw,
		"draw":     parseDraw,
		"drawable": parseDrawable,
		"hc":       parseHc,
		"unset":    parseUnset,
		"env":      parseEnv,
		"-":        parseDelete,
	}
	parseFunc, ok := dispatch[commandKeyWord]
	if ok {
		return parseFunc(buffer, cursor)
	}

	panic("command not processed")
}
