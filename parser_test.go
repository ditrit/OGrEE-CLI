package main

import (
	"fmt"
	"testing"
)

func TestParseArsg(t *testing.T) {
	buffer := "-a 42 -s dazd"
	args, err := parseArgs(buffer, 0, len(buffer))
	if err != nil {
		t.Errorf(err.Error())
	}
	fmt.Printf("%v", args)
}

func TestParse(t *testing.T) {
	Parse("get coucou")
}
