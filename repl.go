package main

//This file inits the Readline instance
//manages the interpreter and REPL
//(read eval print loop)

//Since chzyer/readline library is not actively developed
//it may be worth switching to peterh/liner
//https://stackoverflow.com/
// questions/33025599/move-the-cursor-in-a-c-program

import (
	c "cli/controllers"
	l "cli/logger"
	"cli/readline"
	"fmt"
	"os"
	"strings"
)

func InterpretLine(str string) {
	root, parseErr := Parse(str)
	if parseErr != nil {
		fmt.Println(parseErr.Error())
		return
	}
	if root == nil {
		return
	}
	_, err := root.execute()
	if err != nil {
		l.GetErrorLogger().Println(err.Error())
		if c.State.DebugLvl > c.NONE {
			if traceErr, ok := err.(*stackTraceError); ok {
				fmt.Println(traceErr.Error())
			} else {
				fmt.Println("Error : " + err.Error())
			}
		}
	}
}

func Start(rl *readline.Instance, script, user string) {

	//Execute Script if provided as arg and exit
	if script != "" {
		if strings.Contains(script, ".ocli") {
			LoadFile(script)
			os.Exit(0)
		}
	}

	c.InitUnityCom(rl, c.State.UnityClientURL)

	// The loop of the program
	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		InterpretLine(line)
		//Update Prompt
		rl.SetPrompt("\u001b[1m\u001b[32m" + user + "@" + "OGrEE3D:" +
			"\u001b[37;1m" + c.State.CurrPath + "\u001b[1m\u001b[32m$>\u001b[0m ")
	}
}
