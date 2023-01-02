package main

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	buffer := "-a 42 -v coucou.plouf -f -s dazd"
	args, endLeft, startRight, err := parseArgs([]string{"a", "s"}, []string{"v", "f"}, buffer, 0)
	if err != nil {
		t.Errorf(err.Error())
	}
	if endLeft != 9 {
		t.Errorf("wrong end position for left arguments : %d", endLeft)
	}
	if startRight != 22 {
		t.Errorf("wrong start position for right arguments : %d", startRight)
	}
	if !reflect.DeepEqual(args, map[string]any{"a": "42", "s": "dazd", "v": nil, "f": nil}) {
		t.Errorf("wrong args returned : %v", args)
	}
}
