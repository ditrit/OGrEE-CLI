package controllers

//This file contains code associated with initialising the Shell

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	l "cli/logger"
	"cli/models"
	"cli/readline"
)

//Intialises env map with .env file
func LoadEnvFile(env map[string]interface{}, path string) {
	file, err := os.Open(path)
	defer file.Close()
	if err == nil {
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanWords) // use scanwords
		for scanner.Scan() {

			splitArr := strings.SplitN(scanner.Text(), "=", 2)
			key := splitArr[0]
			val := splitArr[1]
			env[key] = val
		}
	} else {
		fmt.Println(err.Error())
		l.GetErrorLogger().Println("Error at initialisation:" +
			err.Error())
	}
}

//Intialise the ShellState
func InitState(flags, env map[string]interface{}) {
	State.DebugLvl = flags["v"].(int)
	State.ClipBoard = nil
	State.TreeHierarchy = &(Node{})
	(*(State.TreeHierarchy)).Entity = -1
	State.TreeHierarchy.PID = ""
	State.CurrPath = "/Physical"
	State.PrevPath = "/Physical"
	State.LineNumber = 0

	//Send login notification
	data := map[string]interface{}{"api_url": State.APIURL, "api_token": GetKey()}
	req := map[string]interface{}{"type": "login", "data": data}
	e := models.ContactUnity("POST", State.UnityClientURL, req, State.Timeout)
	if e != nil {
		l.GetWarningLogger().Println("Note: Unity Client (" + State.UnityClientURL + ") Unreachable")
		fmt.Println("Note: Unity Client (" + State.UnityClientURL + ") Unreachable ")
		State.UnityClientAvail = false
	} else {
		fmt.Println("Unity Client is Reachable!")
		State.UnityClientAvail = true
	}
	//Set the filter attributes setting
	State.FilterDisplay = false
	//Set the Analyser setting to ON for now
	State.Analyser = true

	phys := &Node{}
	phys.Name = "Physical"
	phys.PID = ""
	phys.ID = "-2"
	State.TreeHierarchy.Nodes.PushBack(phys)

	stray := &Node{}
	stray.Name = "Stray"
	stray.PID = "-2"
	stray.ID = "-3"
	stray.Path = "/Physical/"
	SearchAndInsert(&State.TreeHierarchy, stray, "/Physical")

	strayDev := &Node{}
	strayDev.Name = "Device"
	strayDev.PID = "-3"
	strayDev.ID = "-4"
	strayDev.Path = "/Physical/Stray"
	SearchAndInsert(&State.TreeHierarchy, strayDev, "/Physical/Stray")

	straySens := &Node{}
	straySens.Name = "Sensor"
	straySens.PID = "-3"
	straySens.ID = "-5"
	straySens.Path = "/Physical/Stray"
	SearchAndInsert(&State.TreeHierarchy, straySens, "/Physical/Stray")

	// SETUP LOGICAL HIERARCHY START
	// TODO: PUT THIS SECTION IN A LOOP
	logique := &Node{}
	logique.ID = "0"
	logique.Name = "Logical"
	logique.Path = "/"
	State.TreeHierarchy.Nodes.PushBack(logique)

	oTemplate := &Node{}
	oTemplate.ID = "1"
	oTemplate.PID = "0"
	oTemplate.Entity = -1
	oTemplate.Name = "ObjectTemplates"
	oTemplate.Path = "/Logical"
	SearchAndInsert(&State.TreeHierarchy, oTemplate, "/Logical")

	rTemplate := &Node{}
	rTemplate.ID = "2"
	rTemplate.PID = "0"
	rTemplate.Entity = -1
	rTemplate.Name = "RoomTemplates"
	rTemplate.Path = "/Logical"
	SearchAndInsert(&State.TreeHierarchy, rTemplate, "/Logical")

	group := &Node{}
	group.ID = "3"
	group.PID = "0"
	group.Entity = -1
	group.Name = "Groups"
	group.Path = "/Logical"
	SearchAndInsert(&State.TreeHierarchy, group, "/Logical")

	//SETUP LOGICAL HIERARCHY END

	//SETUP DOMAIN/ENTERPRISE
	organisation := &Node{}
	organisation.ID = "5"
	organisation.Name = "Organisation"
	organisation.Path = "/"
	State.TreeHierarchy.Nodes.PushBack(organisation)

	domain := &Node{}
	domain.Name = "Domain"
	domain.PID = "5"
	domain.ID = "-6"
	domain.Path = "/Organisation"
	SearchAndInsert(&State.TreeHierarchy, domain, "/Organisation")

	enterprise := &Node{}
	enterprise.ID = "0"
	enterprise.PID = "5"
	enterprise.Name = "Enterprise"
	enterprise.Path = "/Organisation"
	SearchAndInsert(&State.TreeHierarchy, enterprise, "/Organisation")

	//Set which objects Unity will be notified about
	State.ObjsForUnity = SetObjsForUnity("updates", env)
	State.DrawableObjs = SetObjsForUnity("drawable", env)
	State.DrawableJsons = make(map[string]map[string]interface{}, 16)

	for i := TENANT; i < GROUP+1; i++ {
		ent := EntityToString(i)
		State.DrawableJsons[ent] = SetDrawableTemplate(ent, env)
	}
}

//It is useful to have the state to hold
//a pointer to our readline terminal
func SetStateReadline(rl *readline.Instance) {
	State.Terminal = &rl
}

//Startup the go routine for listening
func TriggerListen(rl *readline.Instance, addr string) {
	go models.ListenForUnity(rl, addr)
}

func SetListener(flags, env map[string]interface{}) {
	if flags["listen_port"] != nil && flags["listen_port"] != "" {
		portInt := flags["listen_port"].(int)
		port := strconv.Itoa(portInt)
		State.ListenAddr = "0.0.0.0:" + port
		return
	}

	if env["listenPORT"] != nil {
		State.ListenAddr = "0.0.0.0:" + env["listenPORT"].(string)
		return
	}

	println("Falling back to default Listening Port")
	l.GetListenInfoLogger().Println("Falling back to default Listening Port")
	State.ListenAddr = "0.0.0.0:5501"
}

func InitTimeout(env map[string]interface{}) {
	if env["unityDeadline"] != nil && env["unityDeadline"] != "" {
		var timeLen int
		var durationType string
		duration := env["unityDeadline"].(string)
		fmt.Sscanf(duration, "%d%s", &timeLen, &durationType)
		switch durationType {
		case "ns":
			State.Timeout = time.Nanosecond * time.Duration(timeLen)
		case "us":
			State.Timeout = time.Microsecond * time.Duration(timeLen)
		case "ms":
			State.Timeout = time.Millisecond * time.Duration(timeLen)
		case "s":
			State.Timeout = time.Second * time.Duration(timeLen)
		default:
			l.GetWarningLogger().Println("Invalid duration unit found. Resorting to default of ms")
			println("Invalid duration unit found in env file. Resorting to default of ms")
			State.Timeout = time.Millisecond * time.Duration(timeLen)
		}
		return
	}

	l.GetWarningLogger().Println("Unity deadline not found. Resorting to default time duration of 10 ms")
	println("Warning: Unity deadline not found in env file. Resorting to default of 10 ms")
	State.Timeout = time.Millisecond * time.Duration(10)
	return
}

func InitKey(flags, env map[string]interface{}) string {
	if flags["api_key"] != nil && flags["api_key"] != "" {
		State.APIKEY = flags["api_key"].(string)
		return State.APIKEY
	}

	if env["apikey"] != nil {
		State.APIKEY = env["apikey"].(string)
		return State.APIKEY
	}

	fmt.Println("Error: No API Key Found")
	l.GetErrorLogger().Println(
		"No API Key provided in env file nor as argument")
	State.APIKEY = ""
	return ""

}

func GetEmail() string {
	file, err := os.Open("./.resources/.env")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords) // use scanwords
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "user=") {
			return scanner.Text()[5:]
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		l.GetErrorLogger().Println(err.Error())
	}
	return ""
}

//Automatically assign Unity and API URLs
func GetURLs(flags, env map[string]interface{}) {
	if flags["api_url"] != nil && flags["api_url"] != "" {
		State.APIURL = flags["api_url"].(string)
	}
	if flags["unity_url"] != nil && flags["unity_url"] != "" {
		State.UnityClientURL = flags["unity_url"].(string)
	}

	if State.UnityClientURL == "" {
		if env["unityURL"] != nil {
			State.UnityClientURL = env["unityURL"].(string)
		}
	}

	if State.APIURL == "" {
		if env["apiURL"] != nil {
			State.APIURL = env["apiURL"].(string)
		}
	}

	if State.APIURL == "" {
		fmt.Println("Falling back to default API URL:" +
			"http://localhost:3001")
		l.GetInfoLogger().Println("Falling back to default API URL:" +
			"http://localhost:3001")
		State.APIURL = "http://localhost:3001"
	}

	if State.UnityClientURL == "" {
		fmt.Println("Falling back to default Unity URL:" +
			"http://localhost:5500")
		l.GetInfoLogger().Println("Falling back to default Unity URL:" +
			"http://localhost:5500")
		State.APIURL = "http://localhost:5500"
	}

}

//Helper for InitState will
//insert objs
func SetObjsForUnity(x string, env map[string]interface{}) []int {
	res := []int{}
	allDetected := false

	if env[x] != nil && env[x] != "" {
		//ObjStr is equal to everything after 'updates='
		objStr := env[x].(string)
		arr := strings.Split(objStr, ",")

		for i := range arr {
			arr[i] = strings.ToLower(arr[i])

			if val := EntityStrToInt(arr[i]); val != -1 {
				res = append(res, val)

			} else if arr[i] == "all" {
				//Exit the loop and use default code @ end of function
				allDetected = true
				i = len(arr)
			}
		}
	}

	//Use default values
	//Set the array to all and exit
	//GROUP is the greatest value int enum type
	//So we use that for the cond guard
	if allDetected || len(res) == 0 {
		if len(res) == 0 && !allDetected {
			l.GetWarningLogger().Println(x + " key not found, going to use defaults")
			println(x + " key not found, going to use defaults")
		}
		for idx := 0; idx < GROUP; idx++ {
			res = append(res, idx)
		}
	}
	return res
}

func SetDrawableTemplate(entity string, env map[string]interface{}) map[string]interface{} {
	var objStr string
	templateKey := entity + "DrawableJson"
	if env[templateKey] != nil && env[templateKey] != "" {
		objStr = strings.Trim(objStr, "'\"")
		//Now retrieve file
		ans := map[string]interface{}{}
		f, e := ioutil.ReadFile(objStr)
		if e == nil {
			json.Unmarshal(f, &ans)
			return ans
		}

	}

	l.GetWarningLogger().Println("Specified template for " + entity + " not found")
	if State.DebugLvl > 2 {
		println("Specified template for " + entity +
			" not found, resorting to defaults")
	}
	return nil
}

func CreateCredentials() (string, string) {
	var tp map[string]interface{}
	var key string
	client := &http.Client{}

	user, _ := readline.Line("Please Enter desired user email: ")
	pass, _ := readline.Password("Please Enter desired password: ")

	buf, _ := json.Marshal(map[string]interface{}{"email": user,
		"password": pass})

	req, _ := http.NewRequest("POST",
		State.APIURL+"/api/user",
		bytes.NewBuffer(buf))

	resp, e := client.Do(req)
	if e != nil || resp.StatusCode != http.StatusCreated {
		println("Error while creating credentials on server! Now exiting")
		l.GetErrorLogger().Println("Error while creating credentials on server! Now exiting")
		os.Exit(-1)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		readline.Line("Error: " + err.Error() + " Now Exiting")
		l.GetErrorLogger().Println("Error while trying to read server response: ", err)
		os.Exit(-1)
	}
	json.Unmarshal(bodyBytes, &tp)
	key = (tp["account"].(map[string]interface{}))["token"].(string)

	os.Mkdir(".resources", 0755)
	os.WriteFile("./.resources/.env",
		[]byte("user="+user+"\n"+"apikey="+key),
		0666)

	l.GetInfoLogger().Println("Credentials created")
	return user, key
}

func CheckKeyIsValid(key string) bool {
	client := &http.Client{}

	req, _ := http.NewRequest("GET",
		State.APIURL+"/api/token/valid", nil)

	req.Header.Set("Authorization", "Bearer "+key)

	resp, e := client.Do(req)
	if e != nil || resp.StatusCode != 200 {
		//readline.Line(e.Error())
		if resp != nil {
			readline.Line("Status code" + strconv.Itoa(resp.StatusCode))
		} else {
			l.GetErrorLogger().Println("Unable to connect to API: ", State.APIURL)
			println("Unable to connect to API: ", State.APIURL)
		}

		return false
	}
	return true
}

func Login(env map[string]interface{}) (string, string) {
	var user, key string

	if env["user"] == nil || env["apikey"] == nil ||
		env["user"] == "" || env["apikey"] == "" {
		l.GetInfoLogger().Println("Key not found, going to generate..")
		user, key = CreateCredentials()
	} else {
		user = env["user"].(string)
		key = env["apikey"].(string)
	}

	if !CheckKeyIsValid(key) {
		println("Error while checking key. Now exiting")
		l.GetErrorLogger().Println("Error while checking key. Now exiting")
		os.Exit(-1)
	}

	//println("Checking credentials...")
	//println(CheckKeyIsValid(key))

	user = (strings.Split(user, "@"))[0]
	l.GetInfoLogger().Println("Successfully Logged In")
	return user, key
}