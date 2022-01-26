package main

import (
	"bufio"
	c "cli/controllers"
	"cli/readline"
	"io/ioutil"
	"os"
	"strings"
)

//This file inits the State and
//manages the interpreter and REPL
//(read eval print loop)

//Since readline hasn't been updated since 2018
//it may be worth switching to peterh/liner
//https://stackoverflow.com/
// questions/33025599/move-the-cursor-in-a-c-program

func InterpretLine(str *string) {
	lex := NewLexer(strings.NewReader(*str))
	yyParse(lex)
	if root != nil {
		root.execute()
		root = nil
	}

	return
}

func loadFile() {
	file, err := os.Open(c.State.ScriptPath)
	if err != nil {
		println("Error:", err.Error())
		c.WarningLogger.Println("Error:", err)
	}
	defer file.Close()
	fullcom := ""
	keepScanning := false
	scanner := bufio.NewScanner(file)
	c.State.LineNumber = 1 //Indicate Line Number

	for scanner.Scan() {
		x := scanner.Text()
		if len(x) > 0 {
			if commentIdx := strings.Index(x, "//"); commentIdx != -1 { //Comment found
				fullcom += x[:commentIdx]
			} else if string(x[len(x)-1]) == "\\" {
				fullcom += x
				keepScanning = true
			} else if keepScanning == true {
				fullcom += x
				InterpretLine(&fullcom)
				keepScanning = false
				fullcom = ""
			} else {
				if strings.ContainsAny(x, ".") == true {
					x = strings.ReplaceAll(x, ".", "/")
				}
				InterpretLine(&x)

			}
		}
		c.State.LineNumber++ //Increment
	}

	//Reset
	c.State.LineNumber = 0
	c.State.ScriptCalled = false
}

//Functions for autocompleter

func listEntities(path string) func(string) []string {
	return func(line string) []string {

		if strings.TrimSpace(line[2:]) == "" {
			path = c.State.CurrPath
		} else {
			path = TrimToSlash(line[3:])
			if len(line) > 4 {
				if strings.TrimSpace(line[2:])[:2] == ".." || strings.TrimSpace(line[2:])[:1] != "/" {
					path = c.State.CurrPath + "/" + path
				}
			}

			if path == "" {
				path = c.State.CurrPath
			} /*else if path == "/" {
				path = "/"
			}*/
		}

		items := c.DispAtLevelTAB(&c.State.TreeHierarchy,
			*c.StrToStack(path))
		return items
	}
}

func listLocal(path string) func(string) []string {
	return func(line string) []string {
		cmdLength := len(".cmds:")

		q := strings.Index(line, ":")
		if q > 0 {
			path = strings.TrimSpace(line[cmdLength:])

			if path == "" {
				path = "./"
			}

		} else {
			path = "./"
		}

		//If line[len(".cmds:"):] has only spaces
		//then use path
		//otherwise use line[len(".cmds:"):]

		println("PATH:", path)

		names := make([]string, 0)
		files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}

func TrimToSlash(x string) string {
	idx := strings.LastIndex(x, "/")
	return x[:idx+1]
}

//End of Functions for autocompleter

//Init the Shell
func Start(verboseLevel int) {

	c.InitLogs()
	c.GetURLs() //Set the URLs
	user, _ := c.Login()

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "\u001b[32m\u001b[1m" + user + "@" + "OGrEE3D:$> " + "\u001b[0m",
		HistoryFile:     ".resources/.history",
		AutoComplete:    getPrefixCompleter(),
		InterruptPrompt: "^C",
		//EOFPrompt:       "exit",

		HistorySearchFold: true,
		//FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}

	defer rl.Close()

	println("Caching data... please wait")
	c.InitState(verboseLevel)

	Repl(rl, user)

}

//The loop of the program
func Repl(rl *readline.Instance, user string) {
	for {
		if c.State.ScriptCalled == true {
			//Load the path and
			//call interpret line
			loadFile()
			c.State.ScriptCalled = false
		} else {
			line, err := rl.Readline()
			if err != nil { // io.EOF
				break
			}
			InterpretLine(&line)
		}

		//c.UpdateSessionState(&line)
		//Update Prompt
		rl.SetPrompt("\u001b[1m\u001b[32m" + user + "@" + "OGrEE3D:" +
			"\u001b[37;1m" + c.State.CurrPath + "\u001b[1m\u001b[32m$>\u001b[0m ")
	}
}

//Helper function that returns the prefix completer
//It is placed in a helper function to maintain readability
//and organisation in the Start() func
func getPrefixCompleter() *readline.PrefixCompleter {
	return readline.NewPrefixCompleter(false,
		readline.PcItem("cd", true,
			readline.PcItemDynamic(listEntities(""), false)),
		readline.PcItem("pwd", false),
		readline.PcItem("clear", false),
		readline.PcItem("exit", false),
		readline.PcItem("grep", false),
		readline.PcItem("ls", true,
			readline.PcItemDynamic(listEntities(""), false)),
		readline.PcItem("man", false,
			readline.PcItem("pwd", false),
			readline.PcItem("print", false),
			readline.PcItem("clear", false),
			readline.PcItem("grep", false),
			readline.PcItem("ls", true),
			readline.PcItem("cd", false),
			readline.PcItem("tree", false),
			readline.PcItem("selection", false),
			readline.PcItem("if", false),
			readline.PcItem("for", false),
			readline.PcItem("while", false),
			readline.PcItem(".cmds", false),
			readline.PcItem("lsog", false),
			readline.PcItem("lsten", false),
			readline.PcItem("lssite", false),
			readline.PcItem("lsbldg", false),
			readline.PcItem("lsroom", false),
			readline.PcItem("lsrack", false),
			readline.PcItem("lsdev", false),
			readline.PcItem("lsrow", false),
			readline.PcItem("lstile", false),
			readline.PcItem("lscabinet", false),
			readline.PcItem("lscorridor", false),
			readline.PcItem("lsac", false),
			readline.PcItem("lspanel", false),
			readline.PcItem("lsseparator", false),
			readline.PcItem("lssensor", false),
			readline.PcItem("create", false),
			readline.PcItem("gt", false),
			readline.PcItem("update", false),
			readline.PcItem("camera", false),
			readline.PcItem("ui", false),
			readline.PcItem(".template", false),
			readline.PcItem(".var", false),
			readline.PcItem("unset", false),
			readline.PcItem("=", false),
			readline.PcItem("-", false),
			readline.PcItem("+", false),
			readline.PcItem("delete", false)),
		readline.PcItem("+", false,
			readline.PcItem("tn:", false),
			readline.PcItem("si:", false),
			readline.PcItem("bd:", false),
			readline.PcItem("ro:", false),
			readline.PcItem("rk:", false),
			readline.PcItem("dv:", false),
			readline.PcItem("gr:", false),
			readline.PcItem("co:", false),
			readline.PcItem("wa:", false)),

		readline.PcItem("create", false,
			readline.PcItem("tenant", false),
			readline.PcItem("site", false),
			readline.PcItem("building", false),
			readline.PcItem("room", false),
			readline.PcItem("rack", false),
			readline.PcItem("device", false),
			readline.PcItem("separator", false),
			readline.PcItem("corridor", false),
			readline.PcItem("group", false),
			readline.PcItem("row", false),
			readline.PcItem("tile", false),
			readline.PcItem("panel", false),
			readline.PcItem("cabinet", false),
			readline.PcItem("sensor", false),
			readline.PcItem("obj_template", false),
			readline.PcItem("room_template", false),
		),

		readline.PcItem("gt", true,
			readline.PcItem("tenant", false),
			readline.PcItem("site", false),
			readline.PcItem("building", false),
			readline.PcItem("room", false),
			readline.PcItem("rack", false),
			readline.PcItem("device", false),
			readline.PcItemDynamic(listEntities(""), false)),
		readline.PcItem("update", false),
		readline.PcItem("delete", false),
		readline.PcItem("selection", false),
		readline.PcItem(".cmds:", false),
		readline.PcItem(".template:", false),
		readline.PcItem(".var:", false),
		readline.PcItem("tree", false),
		readline.PcItem("lsten", false),
		readline.PcItem("lssite", false),
		readline.PcItem("lsbldg", false),
		readline.PcItem("lsroom", false),
		readline.PcItem("lsrack", false),
		readline.PcItem("lsdev", false),
		readline.PcItem("lsrow", false),
		readline.PcItem("lstile", false),
		readline.PcItem("lscabinet", false),
		readline.PcItem("lscorridor", false),
		readline.PcItem("lsac", false),
		readline.PcItem("lspanel", false),
		readline.PcItem("lsseparator", false),
		readline.PcItem("lssensor", false),
		readline.PcItem("lsog", false),
		readline.PcItem("print", false),
		readline.PcItem("unset", false,
			readline.PcItem("-v", false),
			readline.PcItem("-f", false)),
		readline.PcItem("while", false,
			readline.PcItem("done", false),
		),
		readline.PcItem("for", false,
			readline.PcItem("in", false),
			readline.PcItem("done", false),
		),
		readline.PcItem("if", false,
			readline.PcItem("then", false),
			readline.PcItem("elif", false),
			readline.PcItem("else", false),
			readline.PcItem("fi", false),
		),
		readline.PcItem("camera", false,
			readline.PcItem(".", false,
				readline.PcItem("move", false),
				readline.PcItem("translate", false),
				readline.PcItem("wait", false)),
		),
		readline.PcItem("ui", false,
			readline.PcItem(".", false,
				readline.PcItem("highlight", false),
				readline.PcItem("hl", false),
				readline.PcItem("debug", false),
				readline.PcItem("infos", false),
				readline.PcItem("wireframe", false),
				readline.PcItem("delay", false)),
		),
		/*readline.PcItem("gt", false,
			readline.PcItem("tenant", false),
			readline.PcItem("site", false),
			readline.PcItem("building", false),
			readline.PcItem("room", false),
			readline.PcItem("rack", false),
			readline.PcItem("device", false),
			readline.PcItem("subdevice", false),
			readline.PcItem("subdevice1", false),
		),*/
		readline.PcItem("=", true,
			readline.PcItemDynamic(listEntities(""), false)),
	)
}