package main

import (
	"flag"
	"os"
	"strconv"

	con "cli/controllers"
	log "cli/logger"

	"github.com/joho/godotenv"
)

type Flags struct {
	verbose    string
	unityURL   string
	User       string
	APIURL     string
	APIKEY     string
	listenPort int
	envPath    string
	histPath   string
	script     string
	timeout    string
}

//We have 4 Levels of precedence for initialising vars
//1 -> cli args (single letter)
//2 -> cli args (full word)
//3 -> env file
//4 -> parent shell environment
//Defaults if nothing found

func InitParams[T comparable](letter, word, env, envDefault, parent, parentDefault, defaultValue T) T {
	if letter != defaultValue {
		return letter
	} else if word != defaultValue {
		return word
	} else if env != defaultValue && env != envDefault {
		return env
	} else if parent != defaultValue && parent != parentDefault {
		return parent
	} else {
		return defaultValue
	}
}

func main() {
	var listenPORT, l int
	var verboseLevel, v, unityURL, u, APIURL, a, user, c, APIKEY, k,
		envPath, e, histPath, h, file, f, t, timeout string

	flag.StringVar(&v, "v", "ERROR",
		"Indicates level of debugging messages."+
			"The levels are of in ascending order:"+
			"{NONE,ERROR,WARNING,INFO,DEBUG}.")

	flag.StringVar(&verboseLevel, "verbose", "ERROR",
		"Indicates level of debugging messages."+
			"The levels are of in ascending order:"+
			"{NONE,ERROR,WARNING,INFO,DEBUG}.")

	flag.StringVar(&unityURL, "unity_url", "", "Unity URL")
	flag.StringVar(&u, "u", "", "Unity URL")

	flag.StringVar(&APIURL, "api_url", "", "API URL")
	flag.StringVar(&a, "a", "", "API URL")

	flag.IntVar(&listenPORT, "listen_port", 0,
		"Indicates which port to communicate to Unity")
	flag.IntVar(&l, "l", 0,
		"Indicates which port to communicate to Unity")

	flag.StringVar(&user, "user", "",
		"Indicate the user email to access the API")
	flag.StringVar(&c, "c", "",
		"Indicate the user email to access the API")

	flag.StringVar(&APIKEY, "api_key", "", "Indicate the key of the API")
	flag.StringVar(&k, "k", "", "Indicate the key of the API")

	flag.StringVar(&envPath, "env_path", "./.env",
		"Indicate the location of the Shell's env file")
	flag.StringVar(&e, "e", "./.env",
		"Indicate the location of the Shell's env file")

	flag.StringVar(&histPath, "history_path", "./.history",
		"Indicate the location of the Shell's history file")
	flag.StringVar(&h, "h", "./.history",
		"Indicate the location of the Shell's history file")

	flag.StringVar(&file, "file", "", "Launch the shell as an interpreter "+
		" by only executing an OCLI script file")
	flag.StringVar(&f, "f", "", "Launch the shell as an interpreter "+
		" by only executing an OCLI script file")

	flag.Parse()

	var flags Flags
	flags.envPath = InitParams(e, envPath, "./.env", os.Getenv("envPath"), "./.env", "./.env", "./.env")

	//Test if env path is good otherwise
	env, envErr := godotenv.Read(flags.envPath)
	if envErr != nil {
		//Print
		println("env file was not found")
		env = nil
	}

	//env file does not have verboseLevel parameter at this time
	flags.verbose = InitParams(v, verboseLevel, "ERROR", "ERROR", os.Getenv("verboseLevel"), "", "ERROR")

	flags.unityURL = InitParams(u, unityURL, env["unityURL"], "", os.Getenv("unityURL"), "", "")

	flags.User = InitParams(c, user, env["user"], "", os.Getenv("user"), "", "")

	flags.APIURL = InitParams(a, APIURL, env["apiURL"], "", os.Getenv("apiURL"), "", "")

	flags.APIKEY = InitParams(k, APIKEY, env["apiKey"], "", os.Getenv("apiKey"), "", "")

	envlistenPort, convErr := strconv.Atoi(env["listenPort"])
	if convErr != nil {
		envlistenPort = 0
	}

	oslistenPort, convErr1 := strconv.Atoi(env["listenPort"])
	if convErr1 != nil {
		oslistenPort = 0
	}
	flags.listenPort = InitParams(l, listenPORT, envlistenPort, 0, oslistenPort, 0, 0)

	flags.histPath = InitParams(h, histPath, env["history"], "", os.Getenv("history"), "", "./.history")

	//Doesn't make sense for env file and OS Env
	//to have OCLI assigned
	flags.script = InitParams(f, file, "", "", "", "", "")

	flags.timeout = InitParams(t, timeout, env["unityTimeout"], "", os.Getenv("unityTimeout"), "", "10ms")

	log.InitLogs()
	con.InitEnvFilePath(flags.envPath)
	con.InitHistoryFilePath(flags.histPath)
	con.InitDebugLevel(flags.verbose) //Set the Debug level

	con.InitTimeout(flags.timeout)            //Set the Unity Timeout
	con.GetURLs(flags.APIURL, flags.unityURL) //Set the URLs
	userName, key := con.Login(flags.User, flags.APIKEY)
	con.InitEmail(flags.User) //Set the User email
	con.InitKey(key)          //Set the API Key

	con.InitState(env)

	//Pass control to repl.go
	Start(&flags, userName)
}
