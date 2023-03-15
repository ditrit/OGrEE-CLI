package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
	flag "github.com/spf13/pflag"
	"golang.org/x/exp/maps"
)

type Config struct {
	Verbose      string            `json:",omitempty"`
	APIURL       string            `json:",omitempty"`
	UnityURL     string            `json:",omitempty"`
	UnityTimeout string            `json:",omitempty"`
	ConfigPath   string            `json:",omitempty"`
	HistPath     string            `json:",omitempty"`
	Script       string            `json:",omitempty"`
	Drawable     []string          `json:",omitempty"`
	DrawableJson map[string]string `json:",omitempty"`
	DrawLimit    int               `json:",omitempty"`
	Updates      []string          `json:",omitempty"`
	User         string            `json:",omitempty"`
	APIKEY       string            `json:",omitempty"`
}

func defaultConfig() Config {
	return Config{
		Verbose:      "ERROR",
		APIURL:       "",
		UnityURL:     "",
		UnityTimeout: "10ms",
		ConfigPath:   "./config.toml",
		HistPath:     "./.history",
		Script:       "",
		Drawable:     []string{"all"},
		DrawableJson: map[string]string{},
		DrawLimit:    50,
		Updates:      []string{"all"},
		User:         "",
		APIKEY:       "",
	}
}

func defaultConfigMap() map[string]interface{} {
	return map[string]interface{}{
		"Verbose":      "ERROR",
		"APIURL":       "",
		"UnityURL":     "",
		"UnityTimeout": "10ms",
		"ConfigPath":   "./config.toml",
		"HistPath":     "./.history",
		"Script":       "",
		"Drawable":     []string{"all"},
		"DrawableJson": map[string]string{},
		"DrawLimit":    50,
		"Updates":      []string{"all"},
		"User":         "",
		"APIKEY":       "",
	}
}

// Take the defaults configuration and overwrite any defined
// parameters with environment variables from parent shell
func GetParentShellVars(defaults map[string]interface{}) {
	godotenv.Load()

	//we can use strings delimited by ':' for array arguments
	// (Drawable and Updates)
	//we can define keys of DrawableJson as separate key-values,
	//and parse all of them into a map here and check if the map
	//differs with the default map
	drawable := map[string]string{}

	for key, value := range defaults {
		if shellValue := os.Getenv(key); shellValue != "" {
			switch key {
			case "Updates", "Drawable":
				if shellValue != "" && !strings.Contains(shellValue, "all") {
					//println("DEBUG PARENT EDIT MADE @ Updates/Drawable:", key)
					//Split according to the delimiter ':'
					arr := strings.Split(shellValue, ":")
					defaults[key] = arr
				}
			default:
				if strings.HasSuffix(key, "DrawableJson") {
					if valStr, ok := value.(string); ok {
						//println("DEBUG PARENT EDIT MADE @DrawableJ:", key)
						drawable[key] = valStr
					}
				} else {
					if shellValue != "" && shellValue != value {
						//println("DEBUG PARENT EDIT MADE @ KEY:", key)
						defaults[key] = shellValue
					}
				}
			}
			//parentShellVars[value] = shellValue
		}
	}
	if len(drawable) != 0 {
		//println("DEBUG PARENT EDIT MADE")
		defaults["DrawableJson"] = drawable
	}
	//return parentShellVars
}

//func Disp(x map[string]interface{}) {
//	enc := json.NewEncoder(os.Stdout)
//	enc.SetIndent("", "    ")
//
//	if err := enc.Encode(x); err != nil {
//		log.Fatal(err)
//	}
//}

func ReadConfig() *Config {
	var confPath string
	argConf := &Config{}
	conf := defaultConfigMap()
	tomlRead := map[string]interface{}{}
	argsRead := map[string]interface{}{}

	GetParentShellVars(conf)
	//println()
	//println("DEBUG LETS VIEW Default conf after merging with parent")
	//Disp(conf)
	//println()

	flag.StringVarP(&argConf.ConfigPath, "ConfigPath", "c", "",
		"Indicate the location of the Shell's config file")

	flag.StringVarP(&argConf.Verbose, "Verbose", "v", "",
		"Indicates level of debugging messages."+
			"The levels are of in ascending order:"+
			"{NONE,ERROR,WARNING,INFO,DEBUG}.")

	flag.StringVar(&argConf.User, "User", "", "User Email Credential")
	flag.StringVarP(&argConf.UnityTimeout, "UnityTimeout", "t", "",
		" Maximum latency CLI should wait for a response from Unity before quitting")
	flag.StringVarP(&argConf.UnityURL, "UnityURL", "u", "", "Unity URL")
	flag.StringVarP(&argConf.APIURL, "APIURL", "a", "", "API URL")
	flag.StringVarP(&argConf.APIKEY, "APIKey", "k", "", "Indicate the key of the API")
	flag.StringVarP(&argConf.HistPath, "HistPath", "h", "",
		"Indicate the location of the Shell's history file")
	flag.StringVarP(&argConf.Script, "file", "f", "", "Launch the shell as an interpreter "+
		" by only executing an OCLI script file")
	flag.IntVarP(&argConf.DrawLimit, "DrawLimit", "d", 0,
		"Limit the number of objects Unity client shall draw upon success of a command ")
	flag.Parse()

	argsBytes, _ := json.Marshal(&argConf)
	json.Unmarshal(argsBytes, &argsRead)
	//println()
	//println("DEBUG LETS VIEW ArgsRead after flag parse")
	//Disp(argsRead)
	//println("DEBUG DONE VIEW ArgsRead")
	//println()

	//Get the Configuration File Path
	if confArg, ok := argsRead["ConfigPath"]; ok && confArg != "" {
		confPath = confArg.(string)
	} else {
		confPath = conf["ConfigPath"].(string)
	}

	configBytes, err := os.ReadFile(confPath)
	if err != nil {
		fmt.Println("Cannot read config file", confPath, ":", err.Error())
		fmt.Println("Please ensure that you have a properly formatted config file saved as 'config.toml' in the current directory")
		fmt.Println("\n\nFor more details please refer to: https://ogree.ditrit.io/htmls/programming.html")
		fmt.Println("View an environment file example here: https://ogree.ditrit.io/htmls/clienv.html")
	}
	_, err = toml.Decode(string(configBytes), &tomlRead)
	if err != nil {
		println("Error reading config :", err.Error())
	}

	//maps.Copy(conf, tomlRead)
	//maps.Copy(conf, argsRead)
	maps.Copy(tomlRead, argsRead)
	maps.Copy(conf, tomlRead)
	confBytes, _ := json.Marshal(conf)
	json.Unmarshal(confBytes, &argConf)

	//println()
	//Disp(conf)
	//println()
	//println("DEBUG COMPARE WITH THE CONF STRUCT")
	//DispConfig(argConf)

	return argConf
}

func UpdateConfigFile(conf *Config) error {
	configFile, err := os.Create(conf.ConfigPath)
	if err != nil {
		return fmt.Errorf("cannot open config file to edit user and key")
	}
	err = toml.NewEncoder(configFile).Encode(conf)
	if err != nil {
		panic("invalid config : " + err.Error())
	}
	return nil
}

func DispConfig(conf *Config) {
	println()
	fmt.Printf("%v", conf)
	println()
}
