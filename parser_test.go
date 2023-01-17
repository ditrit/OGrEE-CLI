package main

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	frame := newFrame("-a 42 -v coucou.plouf -f -s dazd")
	args, endLeft, startRight, err := parseArgs([]string{"a", "s"}, []string{"v", "f"}, frame)
	if err != nil {
		t.Errorf(err.Error())
	}
	if endLeft != 9 {
		t.Errorf("wrong end position for left arguments : %d", endLeft)
	}
	if startRight != 22 {
		t.Errorf("wrong start position for right arguments : %d", startRight)
	}
	if !reflect.DeepEqual(args, map[string]string{"a": "42", "s": "dazd", "v": "", "f": ""}) {
		t.Errorf("wrong args returned : %v", args)
	}
}

func TestParseExpr(t *testing.T) {
	frame := newFrame("42 + (3 - 4) * 6")
	expr, _, err := parseExpr(frame)
	if err != nil {
		t.Errorf(err.Error())
	}
	expectedExpr := &arithNode{
		op:   "+",
		left: &intLeaf{42},
		right: &arithNode{
			op: "*",
			left: &arithNode{
				op:    "-",
				left:  &intLeaf{3},
				right: &intLeaf{4},
			},
			right: &intLeaf{6},
		},
	}
	if !reflect.DeepEqual(expr, expectedExpr) {
		t.Errorf("unexpected expression")
	}
}

func TestParseExpr2(t *testing.T) {
	frame := newFrame("3 - 4) * 6")
	expr, _, err := parseExpr(frame)
	if err != nil {
		t.Errorf(err.Error())
	}
	expectedExpr := &arithNode{
		op:    "-",
		left:  &intLeaf{3},
		right: &intLeaf{4},
	}
	if !reflect.DeepEqual(expr, expectedExpr) {
		t.Errorf("unexpected expression")
	}
}
