package main

import (
	//"flag"

	"cli/config"
)

type Flags struct {
	verbose    string
	apiURL     string
	unityURL   string
	timeout    string
	envPath    string
	histPath   string
	script     string
	User       string
	apiKey     string
	listenPort int
}

//We have 4 Levels of precedence for initialising every var
//1 -> cli args
//2 -> env file
//3 -> parent shell environment
//Default if nothing found

func InitParams[T comparable](word, env, envDefault, parent, parentDefault, defaultValue T) T {
	if word != defaultValue {
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
	/*var listenPORT int
	var verbose, unityURL, apiURL, user, apiKey,
		envPath, histPath, file, timeout string

	flag.StringVarP(&verbose, "verbose", "v", "ERROR",
		"Indicates level of debugging messages."+
			"The levels are of in ascending order:"+
			"{NONE,ERROR,WARNING,INFO,DEBUG}.")

	flag.StringVarP(&unityURL, "unity_url", "u", "", "Unity URL")
	flag.StringVarP(&apiURL, "api_url", "a", "", "API URL")

	flag.IntVarP(&listenPORT, "listen_port", "l", 0,
		"Indicates which port to communicate to Unity")

	flag.StringVar(&user, "user", "",
		"Indicate the user email to access the API")

	flag.StringVarP(&apiKey, "api_key", "k", "", "Indicate the key of the API")

	flag.StringVarP(&envPath, "env_path", "e", "./.env",
		"Indicate the location of the Shell's env file")

	flag.StringVarP(&histPath, "history_path", "h", "./.history",
		"Indicate the location of the Shell's history file")

	flag.StringVarP(&file, "file", "f", "", "Launch the shell as an interpreter "+
		" by only executing an OCLI script file")

	flag.StringVarP(&timeout, "unity_timeout", "t", "10ms",
		" Maximum latency CLI should wait for a response from Unity before quitting")

	flag.Parse()

	var flags Flags
	flags.envPath = InitParams(envPath, "./.env", "./.env", os.Getenv("envPath"), "./.env", "./.env")

	//Test if env path is good otherwise
	env, envErr := godotenv.Read(flags.envPath)
	if envErr != nil {
		//Print
		println("env file was not found")
		env = nil
	}

	//env file does not have verbose parameter at this time
	flags.verbose = InitParams(verbose, env["verbose"], "ERROR", os.Getenv("verbose"), "", "ERROR")

	flags.unityURL = InitParams(unityURL, env["unityURL"], "", os.Getenv("unityURL"), "", "")

	flags.User = InitParams(user, env["user"], "", os.Getenv("user"), "", "")

	flags.apiURL = InitParams(apiURL, env["apiURL"], "", os.Getenv("apiURL"), "", "")

	flags.apiKey = InitParams(apiKey, env["apiKey"], "", os.Getenv("apiKey"), "", "")

	envlistenPort, convErr := strconv.Atoi(env["listenPort"])
	if convErr != nil {
		envlistenPort = 0
	}

	oslistenPort, convErr1 := strconv.Atoi(env["listenPort"])
	if convErr1 != nil {
		oslistenPort = 0
	}
	flags.listenPort = InitParams(listenPORT, envlistenPort, 0, oslistenPort, 0, 0)

	flags.histPath = InitParams(histPath, env["history"], "", os.Getenv("history"), "", "./.history")

	//Doesn't make sense for env file and OS Env
	//to have OCLI assigned
	flags.script = InitParams(file, "", "", "", "", "")

	flags.timeout = InitParams(timeout, env["unityTimeout"], "", os.Getenv("unityTimeout"), "", "10ms")
	*/
	config.ReadConfig()
	/*log.InitLogs()
	con.InitEnvFilePath(flags.envPath)
	con.InitHistoryFilePath(flags.histPath)
	con.InitDebugLevel(flags.verbose) //Set the Debug level

	con.InitTimeout(flags.timeout)            //Set the Unity Timeout
	con.GetURLs(flags.apiURL, flags.unityURL) //Set the URLs
	userName, key := con.Login(flags.User, flags.apiKey)
	con.InitEmail(flags.User) //Set the User email
	con.InitKey(key)          //Set the API Key

	con.InitState(env)

	//Pass control to repl.go
	Start(&flags, userName)*/
}
