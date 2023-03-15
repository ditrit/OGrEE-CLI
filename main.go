package main

import (
	"cli/config"
	c "cli/controllers"
	con "cli/controllers"
	log "cli/logger"
	"cli/readline"
)

func main() {

	//We have 4 Levels of precedence for initialising config parameters
	//1 -> cli args
	//2 -> env file
	//3 -> parent shell environment
	//Default if nothing found
	flags := config.ReadConfig()

	//Initialise the Shell
	log.InitLogs()
	con.InitEnvFilePath(flags.ConfigPath)
	con.InitHistoryFilePath(flags.HistPath)
	con.InitDebugLevel(flags.Verbose) //Set the Debug level

	con.InitTimeout(flags.UnityTimeout)       //Set the Unity Timeout
	con.GetURLs(flags.APIURL, flags.UnityURL) //Set the URLs
	userName, key := con.Login(flags.User, flags.APIKEY)
	con.InitEmail(flags.User) //Set the User email
	con.InitKey(key)          //Set the API Key

	con.InitState(flags)

	rl, err := readline.NewEx(&readline.Config{
		Prompt: "\u001b[1m\u001b[32m" + userName + "@" + "OGrEE3D:" +
			"\u001b[37;1m" + c.State.CurrPath + "\u001b[1m\u001b[32m$>\u001b[0m ",
		HistoryFile:       c.State.HistoryFilePath,
		AutoComplete:      GetPrefixCompleter(),
		InterruptPrompt:   "^C",
		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	//Allow the ShellState to hold a ptr to readline
	c.SetStateReadline(rl)

	//Pass control to repl.go
	Start(rl, flags.Script, userName)
}
