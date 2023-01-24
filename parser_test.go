package main

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestParseArgs(t *testing.T) {
	frame := newFrame("-a 42 -v coucou.plouf -f -s dazd")
	args, middle, err := parseArgs([]string{"a", "s"}, []string{"v", "f"}, frame)
	if err != nil {
		t.Errorf(err.Error())
	}
	if middle.start != 9 {
		t.Errorf("wrong end position for left arguments : %d", middle.start)
	}
	if middle.end != 22 {
		t.Errorf("wrong start position for right arguments : %d", middle.end)
	}
	if !reflect.DeepEqual(args, map[string]string{"a": "42", "s": "dazd", "v": "", "f": ""}) {
		t.Errorf("wrong args returned : %v", args)
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
