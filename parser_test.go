package main

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestFindClosing(t *testing.T) {
	frame := newFrame("(a(a)(a()\")\"a))aa")
	i := findClosing(frame)
	if i != 14 {
		t.Errorf("cannot find the closing parenthesis")
	}
	frame = newFrame("(a(a)(a()\")\"a)aa")
	i = findClosing(frame)
	if i != 16 {
		t.Errorf("closing parenthesis should not be found")
	}
}

func TestParseExact(t *testing.T) {
	frame := newFrame("testabc")
	ok, nextFrame := parseExact("test", frame)
	if !ok {
		t.Errorf("parseExact should return true")
	}
	if nextFrame.str() != "abc" {
		t.Errorf("parseExact returns the wrong next frame")
	}
	frame = newFrame("abctest")
	ok, nextFrame = parseExact("test", frame)
	if ok {
		t.Errorf("parseExact should return false")
	}
	if nextFrame.str() != "abctest" {
		t.Errorf("parseExact should return the same next frame")
	}
}

func TestParseWord(t *testing.T) {
	frame := newFrame("test abc")
	word, nextFrame, err := parseWord(frame)
	if err != nil {
		t.Errorf(err.Error())
	}
	if word != "test" {
		t.Errorf("wrong word parsed")
	}
	if nextFrame.str() != " abc" {
		t.Errorf("wrong next frame")
	}
}

func TestParsePathGroup(t *testing.T) {
	s := "{ test.plouf.plaf , test.plaf.plouf } a"
	frame := newFrame(s)
	paths, nextFrame, err := parsePathGroup(frame)
	if err != nil {
		t.Errorf(err.Error())
	}
	firstNode := &pathNode{&strLeaf{"test.plouf.plaf"}}
	secondNode := &pathNode{&strLeaf{"test.plaf.plouf"}}
	if !reflect.DeepEqual(paths, []node{firstNode, secondNode}) {
		t.Errorf("wrong path group parsed : %s", spew.Sdump(paths))
	}
	if nextFrame.str() != " a" {
		t.Errorf("wrong next frame")
	}
}

func TestParseWordSingleLetter(t *testing.T) {
	frame := newFrame("a 42")
	word, nextFrame, err := parseWord(frame)
	if err != nil {
		t.Errorf(err.Error())
	}
	if word != "a" {
		t.Errorf("wrong word parsed")
	}
	if nextFrame.str() != " 42" {
		t.Errorf("wrong next frame")
	}
}

func TestParseArgs(t *testing.T) {
	frame := newFrame("-a 42 -v coucou.plouf -f -s dazd")
	args, middle, err := parseArgs([]string{"a", "s"}, []string{"v", "f"}, frame)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if middle.start != 9 {
		t.Errorf("wrong end position for left arguments : %d", middle.start)
		return
	}
	if middle.end != 22 {
		t.Errorf("wrong start position for right arguments : %d", middle.end)
		return
	}
	if !reflect.DeepEqual(args, map[string]string{"a": "42", "s": "dazd", "v": "", "f": ""}) {
		t.Errorf("wrong args returned : %v", args)
		return
	}
}

func TestParseExpr(t *testing.T) {
	frame := newFrame("\"plouf\" + (3 - 4.2) * ${a}  42")
	expr, cursor, err := parseExpr(frame)
	if err != nil {
		t.Errorf(err.Error())
	}
	expectedExpr := &arithNode{
		op:   "+",
		left: &strLeaf{"plouf"},
		right: &arithNode{
			op: "*",
			left: &arithNode{
				op:    "-",
				left:  &intLeaf{3},
				right: &floatLeaf{4.2},
			},
			right: &symbolReferenceNode{"a"},
		},
	}
	if !reflect.DeepEqual(expr, expectedExpr) {
		t.Errorf("unexpected expression : \n%s", spew.Sdump(expr))
	}
	if cursor.start != 28 {
		t.Errorf("unexpected cursor : %d", cursor.start)
	}
}

func TestParseExpr2(t *testing.T) {
	frame := newFrame("42..48")
	expr, _, _ := parseExpr(frame)
	expected := &intLeaf{42}
	if !reflect.DeepEqual(expr, expected) {
		t.Errorf("unexpected expression : \n%s", spew.Sdump(expr))
	}
}

func TestParseAssign(t *testing.T) {
	frame := newFrame("test= plouf")
	va, nextFrame, err := parseAssign(frame)
	if err != nil {
		t.Errorf("cannot parse assign : %s", err.Error())
	}
	if va != "test" {
		t.Errorf("wrong variable parserd : %s", va)
	}
	if nextFrame.str() != " plouf" {
		t.Errorf("wrong next frame : %s", nextFrame.str())
	}
}

func TestParseLsObj(t *testing.T) {
	buffer := "lsbldg -s height plouf.plaf - f attr1:attr2 -r"
	n, err := Parse(buffer)
	if err != nil {
		t.Errorf("cannot parse lsobj : %s", err.Error())
	}
	path := &pathNode{&strLeaf{"plouf.plaf"}}
	entity := 2
	recursive := true
	sort := "height"
	attrList := []string{"attr1", "attr2"}
	format := ""
	expected := &lsObjNode{path, entity, recursive, sort, attrList, format}
	if !reflect.DeepEqual(n, expected) {
		t.Errorf("unexpected lsobj parsing : \n%s", spew.Sdump(n))
	}
	buffer = "lsbldg -s height plouf.plaf - f (\"height is %s\", height) -r"
	n, err = Parse(buffer)
	if err != nil {
		t.Errorf("cannot parse lsobj : %s", err.Error())
	}
	attrList = []string{"height"}
	format = "height is %s"
	expected = &lsObjNode{path, entity, recursive, sort, attrList, format}
	if !reflect.DeepEqual(n, expected) {
		t.Errorf("unexpected lsobj parsing : \n%s", spew.Sdump(n))
	}
}

func TestParseLs(t *testing.T) {
	buffer := "ls"
	n, err := Parse(buffer)
	if err != nil {
		t.Errorf("cannot parse ls : %s", err.Error())
	}
	expected := &lsNode{&pathNode{&strLeaf{"."}}}
	if !reflect.DeepEqual(n, expected) {
		t.Errorf("unexpected parsing : \n%s", spew.Sdump(n))
	}
}

func TestParseUpdate(t *testing.T) {
	buffer := "coucou.plouf : attr = #val1 @ val2"
	n, err := Parse(buffer)
	if err != nil {
		t.Errorf("cannot parse update : %s", err.Error())
	}
	expected := &updateObjNode{
		&pathNode{&strLeaf{"coucou.plouf"}},
		"attr",
		[]node{&strLeaf{"val1"}, &strLeaf{"val2"}},
		true,
	}
	if !reflect.DeepEqual(n, expected) {
		t.Errorf("unexpected parsing : \n%s", spew.Sdump(n))
	}
}
