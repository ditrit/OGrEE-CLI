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
var valueRegex = `((\S*)|(".*"))`

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
	for i < len(buffer) && buffer[i] != ' ' && buffer[i] != '\t' {
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
	if !regexMatch(wordRegex, word) {
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

func parseQuotedString(buffer string, start int) (node, *ParserError) {
	if buffer[start] != '"' {
		return nil, &ParserError{
			buffer:  buffer,
			message: "double quote \" expected",
			start:   start,
			end:     start,
		}
	}
	end := strings.Index(buffer[start+1:], "\"") + start + 1
	return parseString(buffer, start+1, end)
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

func parseCommaSeparatedWords(buffer string, start int) (node, *Parser) {
	
}

func parseArgsAux(allowedArgs []string, allowedFlags []string, buffer string) (
	map[string]string, *ParserError,
) {
	args := map[string]string{}
	cursor := 0
	for cursor < len(buffer) && buffer[cursor] == '-' {
		cursor++
		cursor = skipWhiteSpaces(buffer, cursor)
		wordEnd := findNextSpace(buffer, cursor)
		arg, err := parseWord(buffer, cursor, wordEnd)
		if err != nil {
			return nil, err
		}
		cursor = skipWhiteSpaces(buffer, cursor+len(arg))
		if sliceContains(allowedArgs, arg) {
			valueEnd := findNextSpace(buffer, cursor)
			value := buffer[cursor:valueEnd]
			args[arg] = value
			cursor = valueEnd
		} else if sliceContains(allowedFlags, arg) {
			args[arg] = ""
		} else {
			panic("unexpected argument")
		}
		cursor = skipWhiteSpaces(buffer, cursor)
	}
	return args, nil
}

func buildSingleArgRegex(allowedArgs []string) string {
	argNameRegex := strings.Join(allowedArgs, "|")
	return `([\-]\s*(` + argNameRegex + `)\s+` + valueRegex + `\s*)`
}

func buildSingleFlagRegex(allowedFlags []string) string {
	flagNameRegex := strings.Join(allowedFlags, "|")
	return `([\-]\s*(` + flagNameRegex + `)\s*)`
}

func parseArgs(allowedArgs []string, allowedFlags []string, buffer string, start int) (
	map[string]string, int, int, *ParserError,
) {
	singleArgRegex := buildSingleArgRegex(allowedArgs)
	singleFlagRegex := buildSingleFlagRegex(allowedFlags)
	multipleArgsRegex := "((" + singleArgRegex + ")|(" + singleFlagRegex + "))*"

	endArgsLeft := len(buffer)
	for endArgsLeft > start && !regexMatch(multipleArgsRegex, buffer[start:endArgsLeft]) {
		endArgsLeft--
	}
	startArgsRight := endArgsLeft
	for startArgsRight < len(buffer) && !regexMatch(multipleArgsRegex, buffer[startArgsRight:]) {
		startArgsRight++
	}
	argsBuffer := buffer[start:endArgsLeft] + buffer[startArgsRight:]
	args, err := parseArgsAux(allowedArgs, allowedFlags, argsBuffer)
	if err != nil {
		return nil, 0, 0, err
	}
	return args, endArgsLeft, startArgsRight, nil
}

func parseLsObj(lsIdx int, buffer string, start int) (node, *ParserError) {
	args, pathStart, _, err := parseArgs([]string{"s", "f"}, []string{"r"}, buffer, start)
	if err != nil {
		return nil, err
	}
	path, _, err := parsePath(buffer, pathStart)
	if err != nil {
		return nil, err
	}
	_, recursive := args["r"]
	_, sort := args["s"]
	format, hasFormat := args["f"]
	if !hasFormat {
		format = ""
	}
	var attrList []string
	if regexMatch(`\(".*",  \)`, format) {
		attrList = make([]string, 0)
		cursor := 1 // cursor relative to the format string
		for cursor < len(format) && 
	}
	return &lsObjNode{path, lsIdx, recursive, sort, format}, nil // TODO : Adapt lsObjNode
}

func parseLs(buffer string, start int) (node, *ParserError) {
	args, pathStart, _, err := parseArgs([]string{"s", "f"}, []string{"r"}, buffer, start)
	if err != nil {
		return nil, err
	}
	path, _, err := parsePath(buffer, pathStart)
	if err != nil {
		return nil, err
	}
	if attr, ok := args["s"]; ok {
		return &lsAttrGenericNode{path, attr}, nil
	}
	return &lsNode{path}, nil
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

func parseUndraw(buffer string, start int) (node, *ParserError) {
	if start == len(buffer) {
		return &undrawNode{nil}, nil
	}
	path, _, err := parsePath(buffer, start)
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
	path, cursor, err := parsePath(buffer, cursor)
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

func parseHc(buffer string, start int) (node, *ParserError) {
	path, cursor, err := parsePath(buffer, start)
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
	args, cursor, _, err := parseArgs([]string{"f", "v"}, []string{}, buffer, start)
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		path, _, err := parsePath(buffer, cursor)
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
		"undraw":  parseUndraw,
		"draw":    parseDraw,
		"hc":      parseHc,
		"unset":   parseUnset,
	}
	parseFunc, ok := dispatch[commandKeyWord]
	if ok {
		return parseFunc(buffer, cursor)
	}

	panic("command not processed")
}
