package main

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var commandDispatch map[string]func(frame Frame) (node, *ParserError)
var createObjDispatch map[string]func(frame Frame) (node, *ParserError)

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

func buildColoredFrame(frame Frame) string {
	result := ""
	result += frame.buf[0:frame.start]
	result += "\033[31m"
	result += "|"
	result += frame.buf[frame.start:frame.end]
	result += "\033[0m"
	result += frame.buf[frame.end:]
	return result
}

func (err *ParserError) Error() string {
	errorString := "\n"
	for i := len(err.messages) - 1; i >= 0; i-- {
		frame := err.frames[i]
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
	if start < frame.start || start > frame.end || end < frame.start || end > frame.end {
		panic("the subframe is not included in the topframe")
	}
	return Frame{frame.buf, start, end}
}

func (frame Frame) until(end int) Frame {
	return frame.new(frame.start, end)
}

func (frame Frame) from(start int) Frame {
	return frame.new(start, frame.end)
}

func (frame Frame) empty() Frame {
	return frame.until(frame.start)
}

func (frame Frame) str() string {
	return frame.buf[frame.start:frame.end]
}

func (frame Frame) char(i int) byte {
	if i < frame.start || i >= frame.end {
		panic("index outside of frame bounds")
	}
	return frame.buf[i]
}

func (frame Frame) forward(offset int) Frame {
	if frame.start+offset > frame.end {
		panic("cannot go forward")
	}
	return Frame{frame.buf, frame.start + offset, frame.end}
}

func (frame Frame) first() byte {
	return frame.char(frame.start)
}

func lexerFromFrame(frame Frame) *lexer {
	return newLexer(frame.buf, frame.start, frame.end)
}

func skipWhiteSpaces(frame Frame) Frame {
	i := frame.start
	for i < frame.end && (frame.char(i) == ' ' || frame.char(i) == '\t') {
		i += 1
	}
	return frame.from(i)
}

func findKeyword(keyword string, frame Frame) int {
	inStr := false
	for startPos := frame.start; startPos <= frame.end-len(keyword); startPos++ {
		c := frame.char(startPos)
		if c == '"' {
			inStr = !inStr
			continue
		}
		if !inStr && frame.new(startPos, startPos+len(keyword)).str() == keyword {
			return startPos
		}
	}
	return frame.end
}

func findNext(substring string, frame Frame) int {
	idx := strings.Index(frame.str(), substring)
	if idx != -1 {
		return frame.start + idx
	}
	return frame.end
}

func findNextAmong(substringList []string, frame Frame) int {
	firstIdx := -1
	for _, substring := range substringList {
		idx := findNext(substring, frame)
		if idx < firstIdx || firstIdx == -1 {
			firstIdx = idx
		}
	}
	return firstIdx
}

func findNextQuote(frame Frame) int {
	idx := strings.Index(frame.str(), "\"")
	if idx == -1 {
		return frame.end
	}
	return frame.start + idx
}

func findClosing(frame Frame) int {
	openToClose := map[byte]byte{'(': ')', '{': '}', '[': ']'}
	open := frame.first()
	close, ok := openToClose[open]
	if !ok {
		panic("invalid opening character")
	}
	stackCount := 0
	inString := false
	for cursor := frame.start; cursor < frame.end; cursor++ {
		if inString {
			if frame.char(cursor) == '"' {
				inString = false
			}
			continue
		}
		if frame.char(cursor) == '"' {
			inString = true
		} else if frame.char(cursor) == open {
			stackCount++
		} else if frame.char(cursor) == close {
			stackCount--
		}
		if stackCount == 0 {
			return cursor
		}
	}
	return frame.end
}

func splitFrameOn(sep string, frame Frame) []Frame {
	bufs := strings.Split(frame.str(), sep)
	frames := []Frame{}
	cursor := frame.start
	for _, buf := range bufs {
		frames = append(frames, frame.new(cursor, cursor+len(buf)))
		cursor += len(sep)
	}
	return frames
}

func parseExact(word string, frame Frame) (bool, Frame) {
	if frame.start+len(word) < frame.end && frame.until(frame.start+len(word)).str() == word {
		return true, frame.forward(len(word))
	}
	return false, frame
}

func isPrefix(prefix string, candidates []string) bool {
	for _, candidate := range candidates {
		if strings.HasPrefix(candidate, prefix) {
			return true
		}
	}
	return false
}

func parseKeyWord(candidates []string, frame Frame) (string, Frame) {
	commandEnd := frame.start
	for commandEnd < frame.end && isPrefix(frame.until(commandEnd+1).str(), candidates) {
		commandEnd++
	}
	if commandEnd == frame.start {
		return "", Frame{}
	}
	return frame.until(commandEnd).str(), frame.from(commandEnd)
}

func parseWord(frame Frame) (string, Frame, *ParserError) {
	l := lexerFromFrame(frame)
	tok := l.nextToken(lexExpr)
	if tok.t != tokWord {
		return "", Frame{}, newParserError(frame.empty(), "word expected")
	}
	return tok.str, frame.from(tok.end), nil
}

func parseSeparatedStuff(
	sep byte,
	frame Frame,
	parseStuff func(Frame) (any, Frame, *ParserError),
) ([]any, *ParserError) {
	topFrame := frame
	items := []any{}

	for {
		var item any
		var err *ParserError
		item, frame, err = parseStuff(frame)
		if err != nil {
			return nil, err.extend(topFrame, "parsing list of words")
		}
		items = append(items, item)
		frame = skipWhiteSpaces(frame)
		if frame.start == frame.end {
			return items, nil
		}
		if frame.first() != sep {
			return nil, newParserError(frame, string(sep)+" expected").
				extend(topFrame, "parsing list of words")
		}
		frame = skipWhiteSpaces(frame.forward(1))
	}
}

func parseSeparatedWords(sep byte, frame Frame) ([]string, *ParserError) {
	parseFunc := func(frame Frame) (any, Frame, *ParserError) {
		return parseWord(frame)
	}
	wordsAny, err := parseSeparatedStuff(sep, frame, parseFunc)
	if err != nil {
		return nil, err.extendMessage("parsing list of words")
	}
	words := []string{}
	for _, wordAny := range wordsAny {
		words = append(words, wordAny.(string))
	}
	return words, nil
}

func parseSeparatedPaths(sep byte, frame Frame) ([]node, *ParserError) {
	parseFunc := func(frame Frame) (any, Frame, *ParserError) {
		return parsePath(frame)
	}
	pathsAny, err := parseSeparatedStuff(sep, frame, parseFunc)
	if err != nil {
		return nil, err.extendMessage("parsing list of paths")
	}
	paths := []node{}
	for _, pathAny := range pathsAny {
		paths = append(paths, pathAny.(node))
	}
	return paths, nil
}

func charIsNumber(char byte) bool {
	return char >= 48 && char <= 57
}

func parseInt(frame Frame) (int, int, *ParserError) {
	end := frame.start
	for end < frame.end && charIsNumber(frame.char(end)) {
		end++
	}
	if end == frame.start {
		return 0, 0, newParserError(frame, "integer expected")
	}
	intString := frame.until(end).str()
	val, err := strconv.Atoi(intString)
	if err != nil {
		panic("cannot convert " + intString + " to integer")
	}
	return val, end, nil
}

func parseFloat(frame Frame) (float64, int, *ParserError) {
	end := frame.start
	dotseen := false
	for end < frame.end {
		if frame.char(end) == '.' {
			if dotseen {
				break
			}
			dotseen = true
		} else if !charIsNumber(frame.char(end)) {
			break
		}
		end++
	}
	if end == frame.start {
		return 0, 0, newParserError(frame, "float expected")
	}
	floatString := frame.until(end).str()
	val, err := strconv.ParseFloat(floatString, 64)
	if err != nil {
		panic("cannot convert " + floatString + " to float")
	}
	return val, end, nil
}

func parseBool(frame Frame) (bool, int, *ParserError) {
	if frame.end-frame.start >= 4 && frame.until(frame.start+4).str() == "true" {
		return true, frame.start + 4, nil
	}
	if frame.end-frame.start >= 5 && frame.until(frame.start+4).str() == "false" {
		return false, frame.start + 5, nil
	}
	return false, 0, newParserError(frame, "bool expected")
}

func parseRawText(frame Frame) (node, *ParserError) {
	l := lexerFromFrame(frame)
	s := ""
	vars := []symbolReferenceNode{}
loop:
	for {
		tok := l.nextToken(lexFormattedString)
		switch tok.t {
		case tokText:
			s += tok.str
		case tokDeref:
			s += "%v"
			vars = append(vars, symbolReferenceNode{tok.val.(string)})
		case tokEOF:
			break loop
		default:
			return nil, newParserError(frame, "unexpected token")
		}
	}
	if len(vars) == 0 {
		return &strLeaf{s}, nil
	}
	return &formatStringNode{s, vars}, nil
}

func parsePath(frame Frame) (node, Frame, *ParserError) {
	frame = skipWhiteSpaces(frame)
	endPath := findNextAmong([]string{" ", "@", ",", ":"}, frame)
	if frame.start == endPath {
		return &pathNode{&strLeaf{"."}}, frame, nil
	}
	path, err := parseRawText(frame.until(endPath))
	if err != nil {
		return nil, Frame{}, err.extend(frame, "parsing path")
	}
	return &pathNode{path}, skipWhiteSpaces(frame.from(endPath)), nil
}

func parsePathGroup(frame Frame) ([]node, Frame, *ParserError) {
	if frame.first() != '{' {
		return nil, Frame{}, newParserError(frame, "{ expected")
	}
	endBracket := findClosing(frame)
	if endBracket == frame.end {
		return nil, Frame{}, newParserError(frame, "{ opened but never closed")
	}
	paths, err := parseSeparatedPaths(',', frame.new(frame.start+1, endBracket))
	if err != nil {
		return nil, Frame{}, err.extendMessage("parsing path group")
	}
	return paths, frame.from(endBracket + 1), nil
}

func exprError(l *lexer, message string) *ParserError {
	frame := Frame{
		buf:   l.input,
		start: l.tok.start,
		end:   l.tok.end,
	}
	return newParserError(frame, message)
}

func parsePrimaryExpr(l *lexer) (node, *ParserError) {
	tok := l.tok
	l.nextToken(lexExpr)
	switch tok.t {
	case tokBool:
		return &boolLeaf{tok.val.(bool)}, nil
	case tokInt:
		return &intLeaf{tok.val.(int)}, nil
	case tokFloat:
		return &floatLeaf{tok.val.(float64)}, nil
	case tokString:
		return &strLeaf{tok.val.(string)}, nil
	case tokDeref:
		return &symbolReferenceNode{tok.val.(string)}, nil
	case tokLeftParen:
		expr, err := parseExprFromLex(l)
		if err != nil {
			return nil, err
		}
		endTok := l.tok
		if endTok.t != tokRightParen {
			return nil, exprError(l, ") expected, got "+endTok.str)
		}
		l.nextToken(lexExpr)
		return expr, nil
	case tokLeftBrac:
		exprList := []node{}
		if l.tok.t == tokRightBrac {
			l.nextToken(lexExpr)
			return &arrNode{exprList}, nil
		}
		for {
			expr, err := parseExprFromLex(l)
			if err != nil {
				return nil, err
			}
			exprList = append(exprList, expr)
			if l.tok.t == tokRightBrac {
				l.nextToken(lexExpr)
				return &arrNode{exprList}, nil
			}
			if l.tok.t == tokComma {
				l.nextToken(lexExpr)
				continue
			}
			return nil, exprError(l, "] or comma expected")
		}
	}
	return nil, exprError(l, "unexpected token : "+tok.str)
}

func parseUnaryExpr(l *lexer) (node, *ParserError) {
	switch l.tok.t {
	case tokAdd:
		l.nextToken(lexExpr)
		return parseUnaryExpr(l)
	case tokSub:
		l.nextToken(lexExpr)
		x, err := parseUnaryExpr(l)
		if err != nil {
			return nil, err
		}
		return &negateNode{x}, nil
	case tokNot:
		l.nextToken(lexExpr)
		x, err := parseUnaryExpr(l)
		if err != nil {
			return nil, err
		}
		return &negateBoolNode{x}, nil
	}
	return parsePrimaryExpr(l)
}

func parseBinaryExpr(l *lexer, leftOperand node, precedence int) (node, *ParserError) {
	var err *ParserError
	if leftOperand == nil {
		leftOperand, err = parseUnaryExpr(l)
		if err != nil {
			return nil, err
		}
	}
	for {
		operator := l.tok
		operatorPrecedence := operator.precedence()
		if operatorPrecedence < precedence {
			return leftOperand, nil
		}
		l.nextToken(lexExpr)
		rightOperand, err := parseBinaryExpr(l, nil, operatorPrecedence+1)
		if err != nil {
			return nil, err
		}
		switch operator.t {
		case tokAdd, tokSub, tokMul, tokDiv, tokMod:
			leftOperand = &arithNode{operator.str, leftOperand, rightOperand}
		case tokOr, tokAnd:
			leftOperand = &logicalNode{operator.str, leftOperand, rightOperand}
		case tokEq, tokNeq:
			leftOperand = &equalityNode{operator.str, leftOperand, rightOperand}
		case tokLeq, tokGeq, tokGtr, tokLss:
			leftOperand = &comparatorNode{operator.str, leftOperand, rightOperand}
		}
	}
}

func parseExprFromLex(l *lexer) (node, *ParserError) {
	return parseBinaryExpr(l, nil, 1)
}

func parseExpr(frame Frame) (node, Frame, *ParserError) {
	l := lexerFromFrame(frame)
	l.nextToken(lexExpr)
	expr, err := parseExprFromLex(l)
	if err != nil {
		return nil, frame.from(l.tok.start), err
	}
	if expr == nil {
		return nil, frame.from(l.tok.start), newParserError(frame, "expression expected")
	}
	return expr, frame.from(l.tok.start), nil
}

func parseAssign(frame Frame) (string, Frame, *ParserError) {
	eqIdx := findNext("=", frame)
	if eqIdx == frame.end {
		return "", Frame{}, newParserError(frame, "= expected")
	}
	varName, frame, err := parseWord(frame)
	if err != nil {
		return "", Frame{}, err.extendMessage("parsing word on the left of =")
	}
	frame = skipWhiteSpaces(frame)
	if frame.first() != '=' {
		return "", Frame{}, newParserError(skipWhiteSpaces(frame).empty(), "= expected")
	}
	return varName, frame.forward(1), nil
}

func parseIndexing(frame Frame) (node, Frame, *ParserError) {
	frame = skipWhiteSpaces(frame)
	ok, frame := parseExact("[", frame)
	if !ok {
		return nil, frame, newParserError(frame, "[ expected")
	}
	index, frame, err := parseExpr(frame)
	if err != nil {
		return nil, frame, err.extend(frame, "parsing indexing")
	}
	ok, frame = parseExact("]", frame)
	if !ok {
		return nil, frame, newParserError(frame, "] expected")
	}
	return index, frame, nil
}

func parseArgValue(frame Frame) (string, Frame, *ParserError) {
	if frame.first() == '(' {
		close := findClosing(frame)
		if close == frame.end {
			return "", Frame{}, newParserError(frame, "( opened but never closed")
		}
		return frame.until(close + 1).str(), frame.from(close + 1), nil
	} else if frame.first() == '"' {
		endQuote := findNextQuote(frame)
		return frame.until(endQuote).str(), frame.from(endQuote + 1), nil
	}
	endValue := findNext(" ", frame)
	return frame.until(endValue).str(), skipWhiteSpaces(frame.from(endValue)), nil
}

func parseSingleArg(allowedArgs []string, allowedFlags []string, frame Frame) (
	string, string, Frame, *ParserError,
) {
	topFrame := frame
	frame = skipWhiteSpaces(frame.forward(1))
	arg, frame, err := parseWord(frame)
	if err != nil {
		return "", "", Frame{}, err.extendMessage("parsing arg name").
			extend(topFrame.empty(), "parsing argument")
	}
	frame = skipWhiteSpaces(frame)
	var value string
	if sliceContains(allowedArgs, arg) {
		value, frame, err = parseArgValue(frame)
		if err != nil {
			return "", "", Frame{}, err.extendMessage("pasing arg value").
				extend(topFrame, "parsing argument")
		}
	} else if sliceContains(allowedFlags, arg) {
		value = ""
	} else {
		panic("unexpected argument")
	}
	return arg, value, skipWhiteSpaces(frame), nil
}

func parseArgsNoCommand(allowedArgs []string, allowedFlags []string, frame Frame) (
	map[string]string, *ParserError,
) {
	args := map[string]string{}
	for frame.start < frame.end && frame.first() == '-' {
		arg, value, newFrame, err := parseSingleArg(allowedArgs, allowedFlags, frame)
		if err != nil {
			return nil, err
		}
		args[arg] = value
		frame = newFrame
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
	map[string]string, Frame, *ParserError,
) {
	multipleArgsRegex := buildMultipleArgsRegex(allowedArgs, allowedFlags)
	endArgsLeft := frame.end
	for endArgsLeft > frame.start && !regexMatch(multipleArgsRegex, frame.until(endArgsLeft).str()) {
		endArgsLeft--
	}
	startArgsRight := endArgsLeft
	for startArgsRight < frame.end && !regexMatch(multipleArgsRegex, frame.from(startArgsRight).str()) {
		startArgsRight++
	}
	argsBuffer := frame.until(endArgsLeft).str() + frame.from(startArgsRight).str()
	argsFrame := newFrame(argsBuffer)
	args, err := parseArgsNoCommand(allowedArgs, allowedFlags, argsFrame)
	if err != nil {
		return nil, Frame{}, err.extend(frame, "parsing args")
	}
	return args, frame.new(endArgsLeft, startArgsRight), nil
}

func parseLsObj(lsIdx int, frame Frame) (node, *ParserError) {
	args, frame, err := parseArgs([]string{"s", "f"}, []string{"r"}, frame)
	if err != nil {
		return nil, err.extendMessage("parsing lsobj arguments")
	}
	path, _, err := parsePath(frame)
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
			startFormat := findNextQuote(formatFrame)
			endFormat := findNextQuote(formatFrame.from(startFormat + 1))
			format = formatArg[startFormat+1 : endFormat]
			cursor := findNext(",", formatFrame.from(endFormat)) + 1
			attrList, err = parseSeparatedWords(',', formatFrame.new(cursor, len(formatArg)-1))
			if err != nil {
				return nil, err.extendMessage("parsing lsobj format")
			}
		} else {
			formatFrame := newFrame(formatArg)
			attrList, err = parseSeparatedWords(':', formatFrame)
			if err != nil {
				return nil, err.extendMessage("parsing lsobj format")
			}
		}
	}
	return &lsObjNode{path, lsIdx, recursive, sort, attrList, format}, nil
}

func parseLs(frame Frame) (node, *ParserError) {
	args, frame, err := parseArgs([]string{"s", "f"}, []string{"r"}, frame)
	if err != nil {
		return nil, err.extendMessage("parsing ls arguments")
	}
	path, _, err := parsePath(frame)
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
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing getu path")
	}
	if frame.start == frame.end {
		return &getUNode{path, 0}, nil
	}
	u, _, err := parseExpr(frame)
	if err != nil {
		return nil, err.extendMessage("parsing getu depth")
	}
	return &getUNode{path, u}, nil
}

func parseGetSlot(frame Frame) (node, *ParserError) {
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing getslot path")
	}
	slotName, _, err := parseStringExpr(frame)
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
	args, frame, err := parseArgs([]string{"f"}, []string{}, frame)
	if err != nil {
		return nil, err.extendMessage("parsing draw arguments")
	}
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing draw path")
	}
	depth := 0
	if frame.start < frame.end {
		depth, _, err = parseInt(frame)
		if err != nil {
			return nil, err.extendMessage("parsing draw depth")
		}
	}
	return &drawNode{path, depth, args}, nil
}

func parseDrawable(frame Frame) (node, *ParserError) {
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing drawable path")
	}
	if frame.start == frame.end {
		return &isEntityDrawableNode{path}, nil
	}
	attrName, _, err := parseWord(frame)
	if err != nil {
		return nil, err.extendMessage("parsing drawable attribute name")
	}
	return &isAttrDrawableNode{path, attrName}, nil
}

func parseHc(frame Frame) (node, *ParserError) {
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing hc path")
	}
	if frame.start == frame.end {
		return &hierarchyNode{path, 1}, nil
	}
	depth, _, err := parseInt(frame)
	if err != nil {
		return nil, err.extendMessage("parsing hc depth")
	}
	return &hierarchyNode{path, depth}, nil
}

func parseUnset(frame Frame) (node, *ParserError) {
	args, frame, err := parseArgs([]string{"f", "v"}, []string{}, frame)
	if err != nil {
		return nil, err.extendMessage("parsing unset arguments")
	}
	if len(args) == 0 {
		path, _, err := parsePath(frame)
		if err != nil {
			return nil, err.extendMessage("parsing unset path")
		}
		ok, frame := parseExact(":", frame)
		if !ok {
			return nil, newParserError(frame, ": expected")
		}
		attr, frame, err := parseWord(frame)
		if err != nil {
			return nil, err.extend(frame, "parsing attribute name")
		}
		index, _, _ := parseIndexing(frame)
		return &unsetAttrNode{path, attr, index}, nil
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
	arg, valueFrame, err := parseAssign(frame)
	if err != nil {
		return nil, err
	}
	value, _, err := parseStringExpr(valueFrame)
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
	if frame.first() == '{' {
		paths, _, err := parsePathGroup(frame)
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
	topFrame := frame
	varName, frame, err := parseAssign(frame)
	if err != nil {
		return nil, err.extendMessage("parsing variable assignment")
	}
	frame = skipWhiteSpaces(frame)
	commandExpr, _ := parseExact("$(", frame)
	if commandExpr {
		frame = frame.forward(1)
		endCommandExpr := findClosing(frame)
		if endCommandExpr == frame.end {
			return nil, newParserError(frame, "$( opened but never closed").
				extend(topFrame, "parsing variable assignment")
		}
		value, err := parseCommand(frame.forward(1).until(endCommandExpr))
		if err != nil {
			return nil, err.extendMessage("parsing variable value (command expression)")
		}
		return &assignNode{varName, value}, nil
	}
	value, _, err := parseStringExpr(frame)
	if err != nil {
		return nil, err.extendMessage("parsing variable value")
	}
	return &assignNode{varName, value}, nil
}

func parseLoad(frame Frame) (node, *ParserError) {
	filePath, _, err := parseStringExpr(frame)
	if err != nil {
		return nil, err.extendMessage("parsing file path")
	}
	return &loadNode{filePath}, nil
}

func parseTemplate(frame Frame) (node, *ParserError) {
	filePath, _, err := parseStringExpr(frame)
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
	sourcePath, _, err := parsePath(frames[0])
	if err != nil {
		return nil, err.extendMessage("parsing source path (physical)")
	}
	destPath, _, err := parsePath(frames[1])
	if err != nil {
		return nil, err.extendMessage("parsing destination path (physical)")
	}
	if len(frames) == 3 {
		slot, _, err := parseStringExpr(frames[2])
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
	sourcePath, _, err := parsePath(frames[0])
	if err != nil {
		return nil, err.extendMessage("parsing source path (physical)")
	}
	if len(frames) == 2 {
		destPath, _, err := parsePath(frames[1])
		if err != nil {
			return nil, err.extendMessage("parsing destination path (physical)")
		}
		return &unlinkObjectNode{sourcePath, destPath}, nil
	}
	return &unlinkObjectNode{sourcePath, nil}, nil
}

func parsePrint(frame Frame) (node, *ParserError) {
	str, _, err := parseStringExpr(frame)
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
	commandName := frame.until(endCommandName).str()
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
		return &treeNode{&pathNode{&strLeaf{"."}}, 0}, nil
	}
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing tree path")
	}
	if frame.start == frame.end {
		return &treeNode{path, 0}, nil
	}
	u, _, err := parseInt(frame)
	if err != nil {
		return nil, err.extendMessage("parsing tree depth")
	}
	return &treeNode{path, u}, nil
}

func parseUi(frame Frame) (node, *ParserError) {
	key, valueFrame, err := parseAssign(frame)
	if err != nil {
		return nil, err
	}
	if key == "delay" {
		delay, _, err := parseFloat(valueFrame)
		if err != nil {
			return nil, err.extendMessage("parsing ui delay")
		}
		return &uiDelayNode{delay}, nil
	}
	if key == "debug" || key == "infos" || key == "wireframe" {
		val, _, err := parseBool(valueFrame)
		if err != nil {
			return nil, err.extendMessage("parsing ui toggle " + key)
		}
		return &uiToggleNode{key, val}, nil
	}
	if key == "highlight" || key == "hl" {
		path, _, err := parsePath(valueFrame)
		if err != nil {
			return nil, err.extendMessage("parsing ui highlight")
		}
		return &uiHighlightNode{path}, nil
	}
	return nil, newParserError(frame, "unknown ui command")
}

func parseCamera(frame Frame) (node, *ParserError) {
	key, valueFrame, err := parseAssign(frame)
	if err != nil {
		return nil, err
	}
	if key == "move" || key == "translate" {
		paramFrames := splitFrameOn("@", valueFrame)
		if len(paramFrames) != 2 {
			return nil, newParserError(valueFrame, "2 parameters expected (separated with @)")
		}
		position, _, err := parseExpr(paramFrames[0])
		if err != nil {
			return nil, err.extendMessage("parsing position vector")
		}
		rotation, _, err := parseExpr(paramFrames[1])
		if err != nil {
			return nil, err.extendMessage("parsing rotation vector")
		}
		return &cameraMoveNode{key, position, rotation}, nil
	}
	if key == "wait" {
		time, _, err := parseFloat(valueFrame)
		if err != nil {
			return nil, err.extendMessage("parsing waiting time")
		}
		return &cameraWaitNode{time}, nil
	}
	return nil, newParserError(frame, "unknown ui command")
}

func parseFocus(frame Frame) (node, *ParserError) {
	path, _, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing path")
	}
	return &focusNode{path}, nil
}

func parseWhile(frame Frame) (node, *ParserError) {
	if frame.first() != '(' {
		return nil, newParserError(frame.empty(), "( expected")
	}
	condition, frame, err := parseExpr(frame.from(frame.start + 1))
	if condition != nil {
		return nil, err.extendMessage("parsing condition")
	}
	if frame.first() != ')' {
		return nil, newParserError(frame.empty(), ") expected")
	}
	body, err := parseCommand(frame.forward(1))
	if err != nil {
		return nil, err.extendMessage("parsing while body")
	}
	return &whileNode{condition, body}, nil
}

func parseFor(frame Frame) (node, *ParserError) {
	varName, frame, err := parseWord(frame)
	if err != nil {
		return nil, err.extendMessage("parsing for loop variable")
	}
	ok, frame := parseExact("in", skipWhiteSpaces(frame))
	if !ok {
		return nil, newParserError(frame.empty(), "\"in\" expected")
	}
	ok, frame = parseExact("{", skipWhiteSpaces(frame))
	if !ok {
		return nil, newParserError(frame.empty(), "{ expected")
	}
	start, frame, err := parseExpr(frame)
	if err != nil {
		return nil, err.extendMessage("parsing for loop start index")
	}
	ok, frame = parseExact("..", frame)
	if !ok {
		return nil, newParserError(frame.empty(), ".. expected")
	}
	end, frame, err := parseExpr(frame)
	if err != nil {
		return nil, err.extendMessage("parsing for loop end index")
	}
	ok, frame = parseExact("}", frame)
	if !ok {
		return nil, newParserError(frame.empty(), "{ expected")
	}
	body, err := parseCommand(frame)
	if err != nil {
		return nil, err.extendMessage("parsing for loop body")
	}
	return &forRangeNode{varName, start, end, body}, nil
}

func parseIf(frame Frame) (node, *ParserError) {
	condition, frame, err := parseExpr(frame)
	if err != nil {
		return nil, err
	}
	ok, frame := parseExact("then", frame)
	if !ok {
		return nil, newParserError(frame, "then expected")
	}
	body, err := parseCommand(frame)
	if err != nil {
		return nil, err.extend(frame, "parsing if body")
	}
	keyword, frame := parseKeyWord([]string{"fi", "else", "elif"}, frame)
	switch keyword {
	case "fi":
		return &ifNode{condition, body, nil}, nil
	case "else":
		fiIdx := findKeyword(" fi", frame)
		if fiIdx == frame.end {
			return nil, newParserError(frame, "fi expected")
		}
		elseBody, err := parseCommand(frame.until(fiIdx))
		if err != nil {
			return nil, err.extend(frame, "")
		}
		return &ifNode{condition, body, elseBody}, nil
	case "elif":
		frame := frame.new(frame.start-2, frame.end)
		elseBody, err := parseIf(frame)
		if err != nil {
			return nil, err.extend(frame, "parsing elif body")
		}
		return &ifNode{condition, body, elseBody}, nil
	default:
		return nil, newParserError(frame, "expected fi, else or elif")
	}
}

func parseObjType(frame Frame) (string, Frame) {
	candidates := []string{}
	for command := range createObjDispatch {
		candidates = append(candidates, command)
	}
	return parseKeyWord(candidates, frame)
}

func parseCreate(frame Frame) (node, *ParserError) {
	objType, frame := parseObjType(frame)
	if objType == "" {
		return nil, newParserError(frame, "parsing object type")
	}
	frame = skipWhiteSpaces(frame)
	if objType == "orphan" {
		return parseCreateOrphan(frame)
	}
	if frame.first() != ':' {
		return nil, newParserError(frame.empty(), ": expected")
	}
	frame = skipWhiteSpaces(frame.forward(1))
	return createObjDispatch[objType](frame)
}

func parseOrientation(frame Frame) (node, Frame, *ParserError) {
	l := lexerFromFrame(frame)
	tok := l.nextToken(lexOrientation)
	if tok.t == tokOrientation {
		return &strLeaf{tok.str}, frame.from(tok.end), nil
	}
	orientation, newFrame, err := parseExpr(frame)
	if err != nil {
		return nil, Frame{}, newParserError(frame.empty(), "orientation expected")
	}
	return orientation, newFrame, nil
}

func parseKeyWordOrExpr(keywords []string, frame Frame) (node, Frame, *ParserError) {
	keyword, newFrame := parseKeyWord(keywords, frame)
	if keyword != "" {
		return &strLeaf{keyword}, newFrame, nil
	}
	expr, newFrame, err := parseExpr(frame)
	if err != nil {
		return nil, Frame{}, newParserError(frame.empty(), "keyword or expr expected")
	}
	return expr, newFrame, nil
}

func parseRackOrientation(frame Frame) (node, Frame, *ParserError) {
	return parseKeyWordOrExpr([]string{"front", "rear", "left", "right"}, frame)
}

func parseAxisOrientation(frame Frame) (node, Frame, *ParserError) {
	return parseKeyWordOrExpr([]string{"+x+y", "+x-y", "-x-y", "-x+y"}, frame)
}

func parseFloorUnit(frame Frame) (node, Frame, *ParserError) {
	return parseKeyWordOrExpr([]string{"t", "m", "f"}, frame)
}

func parseSide(frame Frame) (node, Frame, *ParserError) {
	return parseKeyWordOrExpr([]string{"front", "rear", "frontflipped", "rearflipped"}, frame)
}

func parseTemperature(frame Frame) (node, Frame, *ParserError) {
	return parseKeyWordOrExpr([]string{"cold", "warm"}, frame)
}

func parseStringExpr(frame Frame) (node, Frame, *ParserError) {
	expr, nextFrame, err := parseExpr(frame)
	if err == nil {
		return expr, nextFrame, nil
	}
	frame = skipWhiteSpaces(frame)
	endStr := findNextAmong([]string{" ", "@", ","}, frame)
	str, err := parseRawText(frame.until(endStr))
	if err != nil {
		return nil, Frame{}, err.extend(frame, "parsing string expression")
	}
	return str, skipWhiteSpaces(frame.from(endStr)), nil
}

type objParam struct {
	name string
	t    string
}

func parseObjectParams(sig []objParam, startWithAt bool, frame Frame) (map[string]node, Frame, *ParserError) {
	values := map[string]node{}
	for i, param := range sig {
		if i != 0 || startWithAt {
			ok, frame := parseExact("@", frame)
			if !ok {
				return nil, Frame{}, newParserError(frame.empty(), "@ expected")
			}
		}
		var value node
		var err *ParserError
		switch param.t {
		case "path":
			value, frame, err = parsePath(frame)
		case "expr":
			value, frame, err = parseExpr(frame)
		case "stringexpr":
			value, frame, err = parseStringExpr(frame)
		case "orientation":
			value, frame, err = parseOrientation(frame)
		case "axisOrientation":
			value, frame, err = parseAxisOrientation(frame)
		case "rackOrientation":
			value, frame, err = parseRackOrientation(frame)
		case "floorUnit":
			value, frame, err = parseFloorUnit(frame)
		case "side":
			value, frame, err = parseSide(frame)
		}
		if err != nil {
			return nil, Frame{}, err.extendMessage("parsing " + param.name)
		}
		values[param.name] = value
	}
	return values, frame, nil
}

func parseCreateTenant(frame Frame) (node, *ParserError) {
	sig := []objParam{{"path", "path"}, {"color", "expr"}}
	params, _, err := parseObjectParams(sig, false, frame)
	if err != nil {
		return nil, err.extendMessage("parsing tenant parameters")
	}
	return &createTenantNode{params["path"], params["color"]}, nil
}

func parseCreateSite(frame Frame) (node, *ParserError) {
	sig := []objParam{{"path", "path"}, {"orientation", "orientation"}}
	params, _, err := parseObjectParams(sig, false, frame)
	if err != nil {
		return nil, err.extendMessage("parsing site parameters")
	}
	return &createSiteNode{params["path"], params["orientation"]}, nil
}

func parseCreateBuilding(frame Frame) (node, *ParserError) {
	sig := []objParam{{"path", "path"}, {"posXY", "expr"}, {"rotation", "expr"}, {"sizeOrTemplate", "stringexpr"}}
	params, _, err := parseObjectParams(sig, false, frame)
	if err != nil {
		return nil, err.extendMessage("parsing building parameters")
	}
	return &createBuildingNode{params["path"], params["posXY"], params["rotation"], params["sizeOrTemplate"]}, nil
}

func parseCreateRoom(frame Frame) (node, *ParserError) {
	sig := []objParam{{"path", "path"}, {"posXY", "expr"}, {"rotation", "expr"}, {"sizeOrTemplate", "stringexpr"}}
	params1, frame, err := parseObjectParams(sig, false, frame)
	if err != nil {
		return nil, err.extendMessage("parsing room parameters")
	}
	if frame.start == frame.end {
		return &createRoomNode{
			params1["path"],
			params1["posXY"],
			params1["rotation"],
			nil, nil,
			params1["sizeOrTemplate"]}, nil
	}
	sig = []objParam{{"orientation", "axisOrientation"}}
	params2, frame, err := parseObjectParams(sig, true, frame)
	if err != nil {
		return nil, err.extendMessage("parsing room parameters")
	}
	if frame.start == frame.end {
		return &createRoomNode{
			params1["path"],
			params1["posXY"],
			params1["rotation"],
			params1["sizeOrTemplate"],
			params2["orientation"], nil}, nil
	}
	sig = []objParam{{"floorUnit", "floorUnit"}}
	params3, frame, err := parseObjectParams(sig, true, frame)
	if err != nil {
		return nil, err.extendMessage("parsing room parameters")
	}
	return &createRoomNode{
		params1["path"],
		params1["posXY"],
		params1["rotation"],
		params1["sizeOrTemplate"],
		params2["orientation"],
		params3["floorUnit"]}, nil
}

func parseCreateRack(frame Frame) (node, *ParserError) {
	sig := []objParam{{"path", "path"}, {"posXY", "expr"},
		{"sizeOrTemplate", "stringexpr"}, {"orientation", "rackOrientation"}}
	params, _, err := parseObjectParams(sig, false, frame)
	if err != nil {
		return nil, err.extendMessage("parsing rack parameters")
	}
	return &createRackNode{params["path"], params["posXY"], params["sizeOrTemplate"], params["orientation"]}, nil
}

func parseCreateDevice(frame Frame) (node, *ParserError) {
	sig := []objParam{{"path", "path"}, {"posUOrSlot", "expr"}, {"sizeUOrTemplate", "stringexpr"}}
	params1, frame, err := parseObjectParams(sig, false, frame)
	if err != nil {
		return nil, err.extendMessage("parsing device parameters")
	}
	if frame.start == frame.end {
		return &createDeviceNode{params1["path"], params1["posUOrSlot"], params1["sizeUOrTemplate"], nil}, nil
	}
	sig = []objParam{{"side", "side"}}
	params2, frame, err := parseObjectParams(sig, false, frame)
	if err != nil {
		return nil, err.extendMessage("parsing device parameters")
	}
	return &createDeviceNode{params1["path"], params1["posUOrSlot"], params1["sizeUOrTemplate"], params2["side"]}, nil
}

func parseCreateGroup(frame Frame) (node, *ParserError) {
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing group physical path")
	}
	ok, frame := parseExact("@", frame)
	if !ok {
		return nil, newParserError(frame.empty(), "@ expected")
	}
	childs, _, err := parsePathGroup(frame)
	if err != nil {
		return nil, err.extendMessage("parsing group childs")
	}
	return &createGroupNode{path, childs}, nil
}

func parseCreateCorridor(frame Frame) (node, *ParserError) {
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing group physical path")
	}
	ok, frame := parseExact("@", frame)
	if !ok {
		return nil, newParserError(frame.empty(), "@ expected")
	}
	racks, frame, err := parsePathGroup(frame)
	if err != nil {
		return nil, err.extendMessage("parsing group childs")
	}
	if len(racks) != 2 {
		return nil, err.extendMessage("only 2 racks expected")
	}
	ok, frame = parseExact("@", frame)
	if !ok {
		return nil, newParserError(frame.empty(), "@ expected")
	}
	temperature, _, err := parseTemperature(frame)
	if err != nil {
		return nil, err.extendMessage("parsing corridor temperature")
	}
	return &createCorridorNode{path, racks[0], racks[1], temperature}, nil
}

func parseCreateOrphan(frame Frame) (node, *ParserError) {
	for _, word := range []string{"device", "dv"} {
		ok, newFrame := parseExact(word, frame)
		if ok {
			return parseCreateOrphanAux(skipWhiteSpaces(newFrame), false)
		}
	}
	for _, word := range []string{"sensor", "sr"} {
		ok, newFrame := parseExact(word, frame)
		if ok {
			return parseCreateOrphanAux(skipWhiteSpaces(newFrame), true)
		}
	}
	return nil, newParserError(frame, "device or sensor keyword expected")
}

func parseCreateOrphanAux(frame Frame, sensor bool) (node, *ParserError) {
	if frame.first() != ':' {
		return nil, newParserError(frame.empty(), ": expected")
	}
	frame = skipWhiteSpaces(frame.forward(1))
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing orphan physical path")
	}
	ok, frame := parseExact("@", frame)
	if !ok {
		return nil, newParserError(frame.empty(), "@ expected")
	}
	template, _, err := parseStringExpr(frame)
	if err != nil {
		return nil, err.extendMessage("parsing orphan template")
	}
	return &createOrphanNode{path, template, sensor}, nil
}

func parseUpdate(frame Frame) (node, *ParserError) {
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, err.extendMessage("parsing update")
	}
	frame = skipWhiteSpaces(frame)
	ok, frame := parseExact(":", frame)
	if !ok {
		return nil, newParserError(frame.empty(), ": expected")
	}
	attr, frame, err := parseAssign(frame)
	if err != nil {
		return nil, err.extendMessage("parsing update")
	}
	frame = skipWhiteSpaces(frame)
	sharpe, frame := parseExact("#", frame)
	values := []node{}
	moreValues := true
	for moreValues {
		var val node
		val, frame, err = parseStringExpr(frame)
		if err != nil {
			return nil, err.extend(frame, "parsing update new value")
		}
		values = append(values, val)
		moreValues, frame = parseExact("@", frame)
	}
	return &updateObjNode{path, attr, values, sharpe}, nil
}

func parseCommandKeyWord(frame Frame) (string, Frame) {
	candidates := []string{}
	for command := range commandDispatch {
		candidates = append(candidates, command)
	}
	for command := range noArgsCommands {
		candidates = append(candidates, command)
	}
	candidates = append(candidates, lsCommands...)
	return parseKeyWord(candidates, frame)
}

func parseCommand(frame Frame) (node, *ParserError) {
	startFrame := frame
	commandKeyWord, frame := parseCommandKeyWord(skipWhiteSpaces(frame))
	println(commandKeyWord)
	if commandKeyWord == "" {
		return parseUpdate(frame)
	}
	if lsIdx := indexOf(lsCommands, commandKeyWord); lsIdx != -1 {
		return parseLsObj(lsIdx, skipWhiteSpaces(frame))
	}
	parseFunc, ok := commandDispatch[commandKeyWord]
	if ok {
		return parseFunc(frame)
	}
	result, ok := noArgsCommands[commandKeyWord]
	if ok {
		return result, nil
	}
	return parseUpdate(startFrame)
}

func firstNonAscii(frame Frame) int {
	for i := frame.start; i < frame.end; i++ {
		if frame.char(i) > unicode.MaxASCII {
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
		"+":          parseCreate,
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
		"ui.":        parseUi,
		"camera.":    parseCamera,
		">":          parseFocus,
		"while":      parseWhile,
		"for":        parseFor,
		"if":         parseIf,
	}
	createObjDispatch = map[string]func(frame Frame) (node, *ParserError){
		"tenant":   parseCreateTenant,
		"tn":       parseCreateTenant,
		"site":     parseCreateSite,
		"si":       parseCreateSite,
		"bldg":     parseCreateBuilding,
		"building": parseCreateBuilding,
		"bd":       parseCreateBuilding,
		"room":     parseCreateRoom,
		"ro":       parseCreateRoom,
		"rack":     parseCreateRack,
		"ra":       parseCreateRack,
		"device":   parseCreateDevice,
		"dv":       parseCreateDevice,
		"corridor": parseCreateCorridor,
		"co":       parseCreateCorridor,
		"group":    parseCreateGroup,
		"gr":       parseCreateGroup,
		"orphan":   parseCreateOrphan,
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
