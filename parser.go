package main

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var commandDispatch map[string]func(frame Frame) (node, *ParserError)
var lsCommands = []string{"lsten", "lssite", "lsbldg", "lsroom", "lsrack", "lsdev", "lsac",
	"lspanel", "lscabinet", "lscorridor", "lssensor"}
var noArgsCommands map[string]node

var manCommands = []string{
	"get", "getu", "getslot",
	"+", "-", "=",
	".cmds", ".template", ".var",
	"ui", "camera",
	"link", "unlink",
	"lsten", "lssite", "lsbldg", "lsroom", "lsrack", "lsdev", "lsac",
	"lspanel", "lscabinet", "lscorridor", "lssensor", "lsenterprise",
	"drawable", "draw", "undraw",
	"tree", "lsog", "env", "cd", "pwd", "clear", "grep", "ls", "exit", "len", "man", "hc",
	"print", "unset", "selection",
}

var wordRegex = `([A-Za-z_][A-Za-z0-9_\-]*)`
var valueRegex = `((\S*)|(".*")|(\(.*\)))`

func sliceContains(slice []string, s string) bool {
	for _, str := range slice {
		if str == s {
			return true
		}
	}
	return false
}

func indexOf(arr []string, val string) int {
	for pos, v := range arr {
		if v == val {
			return pos
		}
	}
	return -1
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
	frames   []Frame
	messages []string
}

func buildPointerString(frame Frame) string {
	pointerBytes := []byte{}
	for i := 0; i < len(frame.buf); i++ {
		pointerBytes = append(pointerBytes, ' ')
	}
	if frame.start < len(frame.buf) {
		pointerBytes[frame.start] = '^'
	}
	for i := frame.start + 1; i < frame.end; i++ {
		pointerBytes[i] = '^'
	}
	pointerString := string(pointerBytes)
	if frame.start == frame.end {
		pointerString += "(empty string)"
	}
	return pointerString
}

func buildColoredFrame(frame Frame) string {
	result := ""
	result += frame.buf[0:frame.start]
	result += "\033[31m"
	result += frame.buf[frame.start:frame.end]
	result += "\033[0m"
	result += frame.buf[frame.end:]
	return result
}

func (err *ParserError) Error() string {
	errorString := ""
	for i := len(err.messages) - 1; i >= 0; i-- {
		frame := err.frames[i]
		//errorString += frame.buf + "\n"
		//errorString += buildPointerString(frame) + "\n"
		errorString += buildColoredFrame(frame) + "\n"
		errorString += err.messages[i] + "\n"
	}
	return errorString
}

func (err *ParserError) extend(frame Frame, message string) *ParserError {
	return &ParserError{append(err.frames, frame), append(err.messages, message)}
}

func (err *ParserError) extendMessage(message string) *ParserError {
	currentMessage := err.messages[len(err.messages)-1]
	err.messages[len(err.messages)-1] = message + " : " + currentMessage
	return err
}

func newParserError(frame Frame, message string) *ParserError {
	return &ParserError{[]Frame{frame}, []string{message}}
}

type Frame struct {
	buf   string
	start int
	end   int
}

func newFrame(buffer string) Frame {
	return Frame{buffer, 0, len(buffer)}
}

func (frame Frame) new(start int, end int) Frame {
	return Frame{frame.buf, start, end}
}

func (frame Frame) until(end int) Frame {
	return frame.new(frame.start, end)
}

func (frame Frame) from(start int) Frame {
	return frame.new(start, frame.end)
}

func (frame Frame) empty(pos int) Frame {
	return frame.new(pos, pos)
}

func (frame Frame) strUntil(end int) string {
	return frame.buf[frame.start:end]
}

func (frame Frame) strFrom(start int) string {
	return frame.buf[start:frame.end]
}

func (frame Frame) full() string {
	return frame.buf[frame.start:frame.end]
}

func skipWhiteSpaces(frame Frame) int {
	i := frame.start
	for i < frame.end && (frame.buf[i] == ' ' || frame.buf[i] == '\t') {
		i += 1
	}
	return i
}

func findNext(substring string, frame Frame) int {
	idx := strings.Index(frame.full(), substring)
	if idx != -1 {
		return frame.start + idx
	}
	return frame.end
}

func findNextAmong(substringList []string, frame Frame) int {
	minIdx := frame.end
	for _, s := range substringList {
		idx := strings.Index(frame.full(), s)
		if idx < minIdx {
			minIdx = idx
		}
	}
	return minIdx
}

func findClosing(frame Frame) int {
	openToClose := map[byte]byte{'(': ')', '{': '}', '[': ']'}
	open := frame.buf[frame.start]
	close, ok := openToClose[open]
	if !ok {
		panic("invalid opening character")
	}
	stackCount := 0
	for cursor := frame.start; cursor < frame.end; cursor++ {
		if frame.buf[cursor] == open {
			stackCount++
		}
		if frame.buf[cursor] == close {
			stackCount--
		}
		if stackCount == 0 {
			return cursor
		}
	}
	return frame.end
}

func splitFrameOn(sep string, frame Frame) []Frame {
	bufs := strings.Split(frame.buf, sep)
	frames := []Frame{}
	cursor := frame.start
	for _, buf := range bufs {
		frames = append(frames, frame.new(cursor, cursor+len(buf)))
		cursor += len(sep)
	}
	return frames
}

func parseExact(word string, frame Frame) (bool, int) {
	if frame.start+len(word) < frame.end && frame.strUntil(frame.start+len(word)) == word {
		return true, frame.start + len(word)
	}
	return false, frame.start
}

func isCommandPrefix(prefix string) bool {
	for command := range commandDispatch {
		if strings.HasPrefix(command, prefix) {
			return true
		}
	}
	for command := range noArgsCommands {
		if strings.HasPrefix(command, prefix) {
			return true
		}
	}
	for _, command := range lsCommands {
		if strings.HasPrefix(command, prefix) {
			return true
		}
	}
	return false
}

func parseCommandKeyWord(frame Frame) (string, int, *ParserError) {
	commandEnd := frame.start
	for commandEnd < frame.end && isCommandPrefix(frame.strUntil(commandEnd+1)) {
		commandEnd++
	}
	if commandEnd == frame.start {
		return "", 0, newParserError(frame, "command name expected")
	}
	return frame.strUntil(commandEnd), commandEnd, nil
}

func parseWord(frame Frame) (string, int, *ParserError) {
	cursor := frame.end
	for cursor > frame.start && !regexMatch(wordRegex, frame.strUntil(cursor)) {
		cursor--
	}
	if cursor == frame.start {
		return "", 0, newParserError(frame, "invalid word")
	}
	return frame.strUntil(cursor), cursor, nil
}

func parseSeparatedWords(sep string, frame Frame) ([]string, *ParserError) {
	frames := splitFrameOn(sep, frame)
	words := []string{}
	for _, subframe := range frames {
		wordStart := skipWhiteSpaces(subframe)
		word, _, err := parseWord(subframe.from(wordStart))
		if err != nil {
			return nil, err.extend(frame, "parsing list of words")
		}
		words = append(words, word)
	}
	return words, nil
}

func parseSeparatedPaths(sep string, frame Frame) ([]node, *ParserError) {
	frames := splitFrameOn(sep, frame)
	paths := []node{}
	for _, subframe := range frames {
		pathStart := skipWhiteSpaces(subframe)
		path, _, err := parsePath(subframe.from(pathStart))
		if err != nil {
			return nil, err.extend(frame, "parsing list of words")
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func parseInt(frame Frame) (int, *ParserError) {
	val, err := strconv.Atoi(frame.full())
	if err != nil {
		return 0, newParserError(frame, "integer expected")
	}
	return val, nil
}

func parseString(frame Frame) (node, *ParserError) {
	var varName string
	var err *ParserError
	nodesToConcat := []node{}
	cursor := frame.start
	for cursor < frame.end {
		varIndex := findNext("&{", frame.from(cursor))
		if varIndex == frame.end {
			break
		}
		leftStr := frame.buf[cursor:varIndex]
		if leftStr != "" {
			nodesToConcat = append(nodesToConcat, &strLeaf{leftStr})
		}
		endVarIndex := findNext("}", frame.from(varIndex))
		if endVarIndex == frame.end {
			return nil, newParserError(frame.from(varIndex), "&{ opened but never closed").
				extend(frame, "parsing string")
		}
		varName, cursor, err = parseWord(frame.new(varIndex+2, endVarIndex))
		if err != nil {
			return nil, err.extend(frame, "parsing string")
		}
		nodesToConcat = append(nodesToConcat, &symbolReferenceNode{varName})
		cursor++
	}
	if len(nodesToConcat) == 0 {
		return &strLeaf{frame.full()}, nil
	} else if len(nodesToConcat) == 1 {
		return nodesToConcat[0], nil
	}
	return &concatNode{nodes: nodesToConcat}, nil
}

func parseGenericPath(mode PathMode, frame Frame) (node, int, *ParserError) {
	endPath := findNext(" ", frame)
	if frame.start == endPath {
		return &pathNode{&strLeaf{"."}, STD}, frame.end, nil
	}
	path, err := parseString(frame.until(endPath))
	if err != nil {
		return nil, 0, err.extend(frame, "parsing path")
	}
	cursor := skipWhiteSpaces(frame.from(endPath))
	return &pathNode{path, mode}, cursor, nil
}

func parsePath(frame Frame) (node, int, *ParserError) {
	return parseGenericPath(STD, frame)
}

func parsePhysicalPath(frame Frame) (node, int, *ParserError) {
	return parseGenericPath(PHYSICAL, frame)
}

func parseArgValue(frame Frame) (string, int, *ParserError) {
	if frame.buf[frame.start] == '(' {
		close := findClosing(frame)
		if close == frame.end {
			return "", 0, newParserError(frame, "( opened but never closed")
		}
		return frame.strUntil(close + 1), close + 1, nil
	} else if frame.buf[frame.start] == '"' {
		endQuote := findNext("\"", frame)
		return frame.strUntil(endQuote), endQuote + 1, nil
	}
	endValue := findNext(" ", frame)
	endValueAndSpaces := skipWhiteSpaces(frame.from(endValue))
	return frame.strUntil(endValue), endValueAndSpaces, nil
}

func parseSingleArg(allowedArgs []string, allowedFlags []string, frame Frame) (
	string, string, int, *ParserError,
) {
	cursor := frame.start
	cursor++ // skip dash
	cursor = skipWhiteSpaces(frame.from(cursor))
	wordEnd := findNext(" ", frame.from(cursor))
	arg, cursor, err := parseWord(frame.new(cursor, wordEnd))
	if err != nil {
		return "", "", 0, err.extendMessage("parsing arg name").
			extend(frame, "parsing argument")
	}
	cursor = skipWhiteSpaces(frame.from(cursor))
	var value string
	if sliceContains(allowedArgs, arg) {
		value, cursor, err = parseArgValue(frame.from(cursor))
		if err != nil {
			return "", "", 0, err.extendMessage("pasing arg value").
				extend(frame, "parsing argument")
		}
	} else if sliceContains(allowedFlags, arg) {
		value = ""
	} else {
		panic("unexpected argument")
	}

	cursor = skipWhiteSpaces(frame.from(cursor))

	return arg, value, cursor, nil
}

func parseArgsNoCommand(allowedArgs []string, allowedFlags []string, frame Frame) (
	map[string]string, *ParserError,
) {
	args := map[string]string{}
	cursor := frame.start
	for cursor < frame.end && frame.buf[cursor] == '-' {
		arg, value, newCursor, err := parseSingleArg(allowedArgs, allowedFlags, frame.from(cursor))
		if err != nil {
			return nil, err
		}
		args[arg] = value
		cursor = newCursor
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

func buildMultipleArgsRegex(allowedArgs []string, allowedFlags []string) string {
	singleArgRegex := buildSingleArgRegex(allowedArgs)
	singleFlagRegex := buildSingleFlagRegex(allowedFlags)
	return "((" + singleArgRegex + ")|(" + singleFlagRegex + "))*"
}

func parseArgs(allowedArgs []string, allowedFlags []string, frame Frame) (
	map[string]string, int, int, *ParserError,
) {
	multipleArgsRegex := buildMultipleArgsRegex(allowedArgs, allowedFlags)

	endArgsLeft := frame.end
	for endArgsLeft > frame.start && !regexMatch(multipleArgsRegex, frame.strUntil(endArgsLeft)) {
		endArgsLeft--
	}
	startArgsRight := endArgsLeft
	for startArgsRight < frame.end && !regexMatch(multipleArgsRegex, frame.strFrom(startArgsRight)) {
		startArgsRight++
	}
	argsBuffer := frame.strUntil(endArgsLeft) + frame.strFrom(startArgsRight)
	argsFrame := newFrame(argsBuffer)
	args, err := parseArgsNoCommand(allowedArgs, allowedFlags, argsFrame)
	if err != nil {
		return nil, 0, 0, err.extend(frame, "parsing args")
	}
	return args, endArgsLeft, startArgsRight, nil
}

func parseLsObj(lsIdx int, frame Frame) (node, *ParserError) {
	args, pathStart, _, err := parseArgs([]string{"s", "f"}, []string{"r"}, frame)

	if err != nil {
		return nil, err.extendMessage("parsing lsobj arguments")
	}
	path, _, err := parsePath(frame.from(pathStart))
	if err != nil {
		return nil, err.extendMessage("pasing lsobj path")
	}
	_, recursive := args["r"]
	sort := args["s"]

	//msg := "Please provide a quote enclosed string for '-f' with arguments separated by ':'. Or provide an argument with printf formatting (ie -f (\"%d\",arg1))"

	var attrList []string
	var format string
	if formatArg, ok := args["f"]; ok {
		if regexMatch(`\(\s*".*"\s*,.+\)`, formatArg) {
			formatFrame := Frame{formatArg, 1, len(formatArg)}
			startFormat := findNext("\"", formatFrame)
			endFormat := findNext("\"", formatFrame.from(startFormat+1))
			format = formatArg[startFormat+1 : endFormat]
			cursor := findNext(",", formatFrame.from(endFormat)) + 1
			attrList, err = parseSeparatedWords(",", formatFrame.new(cursor, len(formatArg)-1))
			if err != nil {
				return nil, err.extendMessage("parsing lsobj format")
			}
		} else {
			formatFrame := newFrame(formatArg)
			attrList, err = parseSeparatedWords(":", formatFrame)
			if err != nil {
				return nil, err.extendMessage("parsing lsobj format")
			}
		}
	}
	println(path, lsIdx, recursive, sort, attrList, format)
	return &lsObjNode{path, lsIdx, recursive, sort, attrList, format}, nil
}

func parseLs(frame Frame) (node, *ParserError) {
	args, pathStart, _, err := parseArgs([]string{"s", "f"}, []string{"r"}, frame)
	if err != nil {
		return nil, err.extendMessage("parsing ls arguments")
	}
	path, _, err := parsePath(frame.from(pathStart))
	if err != nil {
		return nil, err.extendMessage("parsing ls path")
	}
	if attr, ok := args["s"]; ok {
		return &lsAttrGenericNode{path, attr}, nil
	}
	return &lsNode{path}, nil
}

func parseGet(frame Frame) (node, *ParserError) {
	path, _, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing get path")
	}
	return &getObjectNode{path}, nil
}

func parseGetU(frame Frame) (node, *ParserError) {
	path, cursor, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing getu path")
	}
	if cursor == frame.end {
		return &getUNode{path, 0}, nil
	}
	u, err := parseInt(frame.from(cursor))
	if err != nil {
		return nil, err.extendMessage("parsing getu depth")
	}
	return &getUNode{path, u}, nil
}

func parseGetSlot(frame Frame) (node, *ParserError) {
	path, cursor, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing getslot path")
	}
	slotName, err := parseString(frame.from(cursor))
	if err != nil {
		return nil, err.extendMessage("parsing getslot slot name")
	}
	return &getSlotNode{path, slotName}, nil
}

func parseUndraw(frame Frame) (node, *ParserError) {
	if frame.start == frame.end {
		return &undrawNode{nil}, nil
	}
	path, _, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing undraw path")
	}
	return &undrawNode{path}, nil
}

func parseDraw(frame Frame) (node, *ParserError) {
	args, cursor, rightArgsStart, err := parseArgs([]string{"f"}, []string{}, frame)
	if err != nil {
		return nil, err.extendMessage("parsing draw arguments")
	}
	path, cursor, err := parsePath(frame.new(cursor, rightArgsStart))
	if err != nil {
		return nil, err.extendMessage("parsing draw path")
	}
	depth := 0
	if cursor < rightArgsStart {
		depth, err = parseInt(frame.from(cursor))
		if err != nil {
			return nil, err.extendMessage("parsing draw depth")
		}
	}
	return &drawNode{path, depth, args}, nil
}

func parseDrawable(frame Frame) (node, *ParserError) {
	path, cursor, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing drawable path")
	}
	if cursor == frame.end {
		return &isEntityDrawableNode{path}, nil
	}
	attrName, _, err := parseWord(frame.from(cursor))
	if err != nil {
		return nil, err.extendMessage("parsing drawable attribute name")
	}
	return &isAttrDrawableNode{path, attrName}, nil
}

func parseHc(frame Frame) (node, *ParserError) {
	path, cursor, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing hc path")
	}
	if cursor == frame.end {
		return &hierarchyNode{path, 1}, nil
	}
	depth, err := parseInt(frame.from(cursor))
	if err != nil {
		return nil, err.extendMessage("parsing hc depth")
	}
	return &hierarchyNode{path, depth}, nil
}

func parseUnset(frame Frame) (node, *ParserError) {
	args, cursor, rightArgsStart, err := parseArgs([]string{"f", "v"}, []string{}, frame)
	if err != nil {
		return nil, err.extendMessage("parsing unset arguments")
	}
	if len(args) == 0 {
		path, _, err := parsePath(frame.new(cursor, rightArgsStart))
		if err != nil {
			return nil, err.extendMessage("parsing unset path")
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

func parseEnv(frame Frame) (node, *ParserError) {
	endArg := findNextAmong([]string{" ", "="}, frame)
	arg, cursor, err := parseWord(frame.until(endArg))
	if err != nil {
		return nil, err.extendMessage("parsing env variable name")
	}
	cursor = skipWhiteSpaces(frame.from(cursor))
	if frame.buf[cursor] != '=' {
		return nil, newParserError(frame.empty(cursor), "= expected")
	}
	cursor++
	value, err := parseString(frame.from(cursor))
	if err != nil {
		return nil, err.extendMessage("parsing env variable value")
	}
	return &setEnvNode{arg, value}, nil
}

func parseDelete(frame Frame) (node, *ParserError) {
	deleteSelection, _ := parseExact("selection", frame)
	if deleteSelection {
		return &deleteSelectionNode{}, nil
	}
	path, _, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing deletion path")
	}
	return &deleteObjNode{path}, nil
}

func parseEqual(frame Frame) (node, *ParserError) {
	if frame.buf[frame.start] == '{' {
		endBracket := findClosing(frame)
		if endBracket == frame.end {
			return nil, newParserError(frame, "{ opened but never closed")
		}
		paths, err := parseSeparatedPaths(",", frame.new(frame.start+1, endBracket))
		if err != nil {
			return nil, err.extendMessage("parsing selection paths")
		}
		return &selectChildrenNode{paths}, nil
	}
	path, _, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing selection path")
	}
	return &selectObjectNode{path}, nil
}

func parseVar(frame Frame) (node, *ParserError) {
	varName, cursor, err := parseWord(frame)
	if err != nil {
		return nil, err.extendMessage("parsing variable name")
	}
	cursor = skipWhiteSpaces(frame.from(cursor))
	if frame.buf[cursor] != '=' {
		return nil, newParserError(frame.empty(cursor), "= expected").
			extend(frame, "parsing variable assignment")
	}
	cursor = skipWhiteSpaces(frame.from(cursor + 1))
	commandExpr, _ := parseExact("$(", frame.from(cursor))
	endCommandExpr := findClosing(frame.from(cursor + 1))
	if endCommandExpr == frame.end {
		return nil, newParserError(frame.from(cursor+1), "$( opened but never closed").
			extend(frame, "parsing variable assignment")
	}
	if commandExpr {
		value, err := parseCommand(frame.new(cursor+2, endCommandExpr))
		if err != nil {
			return nil, err.extendMessage("parsing variable value (command expression)")
		}
		return &assignNode{varName, value}, nil
	}
	value, err := parseString(frame.from(cursor))
	if err != nil {
		return nil, err.extendMessage("parsing variable value")
	}
	return &assignNode{varName, value}, nil
}

func parseLoad(frame Frame) (node, *ParserError) {
	filePath, err := parseString(frame)
	if err != nil {
		return nil, err.extendMessage("parsing file path")
	}
	return &loadNode{filePath}, nil
}

func parseTemplate(frame Frame) (node, *ParserError) {
	filePath, err := parseString(frame)
	if err != nil {
		return nil, err.extendMessage("parsing file path")
	}
	return &loadTemplateNode{filePath}, nil
}

func parseLen(frame Frame) (node, *ParserError) {
	varName, _, err := parseWord(frame)
	if err != nil {
		return nil, err.extendMessage("parsing variable name")
	}
	return &lenNode{varName}, nil
}

func parseLink(frame Frame) (node, *ParserError) {
	frames := splitFrameOn("@", frame)
	if len(frames) < 2 || len(frames) > 3 {
		return nil, newParserError(frame, "too many fields given (separated by @)")
	}
	sourcePath, _, err := parsePhysicalPath(frames[0])
	if err != nil {
		return nil, err.extendMessage("parsing source path (physical)")
	}
	destPath, _, err := parsePhysicalPath(frames[1])
	if err != nil {
		return nil, err.extendMessage("parsing destination path (physical)")
	}
	if len(frames) == 3 {
		slot, err := parseString(frames[2])
		if err != nil {
			return nil, err.extendMessage("parsing slot name")
		}
		return &linkObjectNode{sourcePath, destPath, slot}, nil
	}
	return &linkObjectNode{sourcePath, destPath, nil}, nil
}

func parseUnlink(frame Frame) (node, *ParserError) {
	frames := splitFrameOn("@", frame)
	if len(frames) < 1 || len(frames) > 2 {
		return nil, newParserError(frame, "too many fields given (separated by @)")
	}
	sourcePath, _, err := parsePhysicalPath(frames[0])
	if err != nil {
		return nil, err.extendMessage("parsing source path (physical)")
	}
	if len(frames) == 2 {
		destPath, _, err := parsePhysicalPath(frames[1])
		if err != nil {
			return nil, err.extendMessage("parsing destination path (physical)")
		}
		return &unlinkObjectNode{sourcePath, destPath}, nil
	}
	return &unlinkObjectNode{sourcePath, nil}, nil
}

func parsePrint(frame Frame) (node, *ParserError) {
	str, err := parseString(frame)
	if err != nil {
		return nil, err.extendMessage("parsing message to print")
	}
	return &printNode{str}, nil
}

func parseMan(frame Frame) (node, *ParserError) {
	if frame.start == frame.end {
		return &helpNode{""}, nil
	}
	endCommandName := findNext(" ", frame)
	commandName := frame.buf[frame.start:endCommandName]
	if !sliceContains(manCommands, commandName) {
		return nil, newParserError(frame, "unknown command")
	}
	return &helpNode{commandName}, nil
}

func parseCd(frame Frame) (node, *ParserError) {
	if frame.start == frame.end {
		return &cdNode{strLeaf{"/"}}, nil
	}
	path, _, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing path")
	}
	return &cdNode{path}, nil
}

func parseTree(frame Frame) (node, *ParserError) {
	if frame.start == frame.end {
		return &treeNode{&pathNode{&strLeaf{"."}, STD}, 0}, nil
	}
	path, cursor, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing tree path")
	}
	if cursor == frame.end {
		return &treeNode{path, 0}, nil
	}
	u, err := parseInt(frame.from(cursor))
	if err != nil {
		return nil, err.extendMessage("parsing tree depth")
	}
	return &treeNode{path, u}, nil
}

func parseCommand(frame Frame) (node, *ParserError) {
	cursor := skipWhiteSpaces(frame)
	commandKeyWord, cursor, err := parseCommandKeyWord(frame.from(cursor))
	if err != nil {
		return nil, err.extendMessage("parsing command keyword")
	}
	cursor = skipWhiteSpaces(frame.from(cursor))
	if lsIdx := indexOf(lsCommands, commandKeyWord); lsIdx != -1 {
		return parseLsObj(lsIdx, frame.from(cursor))
	}
	parseFunc, ok := commandDispatch[commandKeyWord]

	if ok {
		return parseFunc(frame.from(cursor))
	}
	result, ok := noArgsCommands[commandKeyWord]
	if ok {
		return result, nil
	}

	return nil, newParserError(frame, "unknown command")
}

func firstNonAscii(frame Frame) int {
	for i := frame.start; i < frame.end; i++ {
		if frame.buf[i] > unicode.MaxASCII {
			return i
		}
	}
	return frame.end
}

func Parse(buffer string) (node, *ParserError) {
	frame := newFrame(buffer)
	commandDispatch = map[string]func(frame Frame) (node, *ParserError){
		"ls":         parseLs,
		"get":        parseGet,
		"getu":       parseGetU,
		"getslot":    parseGetSlot,
		"undraw":     parseUndraw,
		"draw":       parseDraw,
		"drawable":   parseDrawable,
		"hc":         parseHc,
		"unset":      parseUnset,
		"env":        parseEnv,
		"-":          parseDelete,
		"=":          parseEqual,
		".var:":      parseVar,
		".cmds:":     parseLoad,
		".template:": parseTemplate,
		"len":        parseLen,
		"link:":      parseLink,
		"unlink":     parseUnlink,
		"print":      parsePrint,
		"man":        parseMan,
		"cd":         parseCd,
		"tree":       parseTree,
	}
	noArgsCommands = map[string]node{
		"selection":    &selectNode{},
		"clear":        &clrNode{},
		"grep":         &grepNode{},
		"lsog":         &lsogNode{},
		"lsenterprise": &lsenterpriseNode{},
		"env":          &envNode{},
		"pwd":          &pwdNode{},
		"exit":         &exitNode{},
	}
	firstNonAsciiIdx := firstNonAscii(frame)
	if firstNonAsciiIdx < frame.end {
		return nil, newParserError(
			frame.new(firstNonAsciiIdx, firstNonAsciiIdx+1),
			"command should only contain ascii characters",
		)
	}
	return parseCommand(frame)
}