package main

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	buffer := "-a 42 -v  -s dazd  -f plouf"
	args, _, err := parseArgs([]string{"a", "s"}, []string{"v", "f"}, buffer, 0)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(args, map[string]any{"a": "42", "s": "dazd", "v": nil, "f": nil}) {
		t.Errorf("wrong args returned")
	}
}

func TestParsePath(t *testing.T) {

}

func TestParse(t *testing.T) {
	a := "⌘dsqd"
	b := []rune(a)
	println(len(a), len(b))
}
