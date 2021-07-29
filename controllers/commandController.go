package controllers

import (
	"cli/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func Execute() {
	println("Congratulations, nobody cares")
	return
}

func PWD() {
	println(State.CurrPath)
}

func Disp(x map[string]interface{}) {
	/*for i, k := range x {
		println("We got: ", i, " and ", k)
	}*/

	jx, _ := json.Marshal(x)

	println("JSON: ", string(jx))
}

func PostObj(entity, path string, data map[string]interface{}) {
	var respMap map[string]interface{}
	resp, e := models.Send("POST",
		"https://ogree.chibois.net/api/user/"+entity+"s", GetKey(), data)

	if e != nil {
		println("There was an error!")
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error: " + err.Error() + " Now Exiting")
		os.Exit(-1)
	}

	json.Unmarshal(bodyBytes, &respMap)
	println(string(respMap["message"].(string)) /*bodyBytes*/)
	if resp.StatusCode == http.StatusCreated && respMap["status"].(bool) == true {
		//Insert object into tree
		node := &Node{}
		node.ID, _ = strconv.Atoi(respMap["data"].(map[string]interface{})["id"].(string))
		node.Name = respMap["data"].(map[string]interface{})["name"].(string)
		_, ok := respMap["data"].(map[string]interface{})["parentId"].(float64)
		//node.PID = int(respMap["data"].(map[string]interface{})["parentId"].(float64))
		if ok {
			node.PID = int(respMap["data"].(map[string]interface{})["parentId"].(float64))
		} else {
			node.PID, _ = strconv.Atoi(respMap["data"].(map[string]interface{})["parentId"].(string))
		}
		switch entity {
		case "tenant":
			node.Entity = TENANT
			State.TreeHierarchy.Nodes.PushBack(node)
		case "site":
			node.Entity = SITE
			UpdateTree(&State.TreeHierarchy, node)

		case "building":
			node.Entity = BLDG
			UpdateTree(&State.TreeHierarchy, node)

		case "room":
			node.Entity = ROOM
			UpdateTree(&State.TreeHierarchy, node)

		case "rack":
			node.Entity = RACK
			UpdateTree(&State.TreeHierarchy, node)

		case "device":
			node.Entity = DEVICE
			UpdateTree(&State.TreeHierarchy, node)

		case "subdevice":
			node.Entity = SUBDEV
			UpdateTree(&State.TreeHierarchy, node)

		case "subdevice1":
			node.Entity = SUBDEV1
			UpdateTree(&State.TreeHierarchy, node)

		}

	}
	return
}

func DeleteObj(path string) {
	URL := "https://ogree.chibois.net/api/user/"
	nd := new(*Node)

	switch path {
	case "":
		nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(State.CurrPath))
	default:
		if path[0] != '/' && len(State.CurrPath) > 1 {
			nd = FindNodeInTree(&State.TreeHierarchy,
				StrToStack(State.CurrPath+"/"+path))
		} else {
			nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(path))
		}
	}

	if nd == nil {
		println("Error finding Object from given path!")
		return
	}

	URL += EntityToString((*nd).Entity) + "s/" + strconv.Itoa((*nd).ID)
	resp, e := models.Send("DELETE", URL, GetKey(), nil)
	if e != nil {
		println("Error while obtaining Object details!")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		DeleteNodeInTree(&State.TreeHierarchy, (*nd).ID, (*nd).Entity)
		println("Success")
	} else {
		println("Error while object!")
		//json.Unmarshal()
	}

	return
}

//Search for objects
func SearchObjects(entity string, data map[string]interface{}) {
	var jsonResp map[string]interface{}
	URL := "https://ogree.chibois.net/api/user/" + entity + "s?"

	for i, k := range data {
		if i == "attributes" {
			for j, _ := range k.(map[string]string) {
				URL = URL + "&" + j + "=" + data[i].(map[string]string)[j]
			}
		} else {
			URL = URL + "&" + i + "=" + string(data[i].(string))
		}
	}

	println("Here is URL: ", URL)

	resp, e := models.Send("GET", URL, GetKey(), nil)
	println("Response Code: ", resp.Status)
	if e != nil {
		println("There was an error!")
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error: " + err.Error() + " Now Exiting")
		os.Exit(-1)
	}
	//println(string(bodyBytes))
	json.Unmarshal(bodyBytes, &jsonResp)
	if resp.StatusCode == http.StatusOK {
		obj := jsonResp["data"].(map[string]interface{})["objects"].([]interface{})
		for idx := range obj {
			println()
			println()
			println("OBJECT: ", idx)
			displayObject(obj[idx].(map[string]interface{}))
			println()
		}

	}
}

func GetObject(path string) {
	URL := "https://ogree.chibois.net/api/user/"
	nd := new(*Node)
	var data map[string]interface{}

	switch path {
	case "":
		nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(State.CurrPath))
	default:
		if path[0] != '/' && len(State.CurrPath) > 1 {
			nd = FindNodeInTree(&State.TreeHierarchy,
				StrToStack(State.CurrPath+"/"+path))
		} else {
			nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(path))
		}
	}

	if nd == nil {
		println("Error finding Object from given path!")
		return
	}

	URL += EntityToString((*nd).Entity) + "s/" + strconv.Itoa((*nd).ID)
	resp, e := models.Send("GET", URL, GetKey(), nil)
	if e != nil {
		println("Error while obtaining Object details!")
		return
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error while reading response!")
		return
	}
	json.Unmarshal(bodyBytes, &data)
	if resp.StatusCode == http.StatusOK {
		if data["data"] != nil {
			obj := data["data"].(map[string]interface{})
			displayObject(obj)
		}
	}

}

func UpdateObj(path string, data map[string]interface{}) {
	println("OK. Attempting to update...")
	if data != nil {
		var respJson map[string]string
		nd := new(*Node)
		switch path {
		case "":
			nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(State.CurrPath))
		default:
			if path[0] != '/' && len(State.CurrPath) > 1 {
				nd = FindNodeInTree(&State.TreeHierarchy,
					StrToStack(State.CurrPath+"/"+path))
			} else {
				nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(path))
			}
		}

		if nd == nil {
			println("Error finding Object from given path!")
			return
		}

		URL := "https://ogree.chibois.net/api/user/" +
			EntityToString((*nd).Entity) + "s/" + strconv.Itoa((*nd).ID)

		resp, e := models.Send("PUT", URL, GetKey(), data)
		//println("Response Code: ", resp.Status)
		if e != nil {
			println("There was an error!")
		}
		defer resp.Body.Close()
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			println("Error: " + err.Error() + " Now Exiting")
			os.Exit(-1)
		}
		json.Unmarshal(bodyBytes, &respJson)
		println(respJson["message"])
		if resp.StatusCode == http.StatusOK && data["name"] != nil {
			//Need to update name of Obj in tree
			(*nd).Name = string(data["name"].(string))
		}
		//println(string(bodyBytes))
	} else {
		println("Error! Please enter desired parameters of Object to be updated")
	}

}

func LS(x string) {
	if x == "" || x == "." {
		DispAtLevel(&State.TreeHierarchy, *StrToStack(State.CurrPath))
	} else if string(x[0]) == "/" {
		DispAtLevel(&State.TreeHierarchy, *StrToStack(x))
	} else {
		DispAtLevel(&State.TreeHierarchy, *StrToStack(State.CurrPath + "/" + x))
	}
}

func CD(x string) {
	if x == ".." {
		lastIdx := strings.LastIndexByte(State.CurrPath, '/')
		if lastIdx != -1 {
			if lastIdx == 0 {
				lastIdx += 1
			}
			State.PrevPath = State.CurrPath
			State.CurrPath =
				State.CurrPath[0:lastIdx]
		}

	} else if x == "" {
		State.PrevPath = State.CurrPath
		State.CurrPath = "/"
	} else if x == "." {
		//Do nothing
	} else if x == "-" {
		//Change to previous path
		tmp := State.CurrPath
		State.CurrPath = State.PrevPath
		State.PrevPath = tmp
	} else if strings.Count(x, "/") >= 1 {
		exist := false
		var pth string

		if string(x[0]) != "/" {
			exist, pth = CheckPath(&State.TreeHierarchy, StrToStack(State.CurrPath+"/"+x), New())
		} else {
			exist, pth = CheckPath(&State.TreeHierarchy, StrToStack(x), New())
		}
		if exist == true {
			println("THE PATH: ", pth)
			State.PrevPath = State.CurrPath
			State.CurrPath = pth
		} else {
			println("Path does not exist")
		}
	} else {
		if len(State.CurrPath) != 1 {
			if exist, _ := CheckPath(&State.TreeHierarchy,
				StrToStack(State.CurrPath+"/"+x), New()); exist == true {
				State.PrevPath = State.CurrPath
				State.CurrPath += "/" + x
			} else {
				println("OGREE: ", x, " : No such object")
			}

		} else {
			if exist, _ := CheckPath(&State.TreeHierarchy,
				StrToStack(State.CurrPath+x), New()); exist == true {
				State.PrevPath = State.CurrPath
				State.CurrPath += x
			} else {
				println("OGREE: ", x, " : No such object")
			}

		}

	}

}

func Help(entry string) {
	switch entry {
	case "ls":
		fmt.Println(`Usage: `, entry, "[PATH] (optional)")
		fmt.Println(`Displays objects in a given directory`)
	case "cd":
		fmt.Println(`Usage: `, entry, "[PATH] (optional)")
		fmt.Println(`Changes current directory`)
	case "create":
		fmt.Println(`Usage: `, entry, "ENTITY [PATH](optional)  [ATTRIBUTES]")
		fmt.Println(`Creates an object in a given directory`)
		printAttributeOptions()
	case "gt":
		fmt.Println(`Usage: `, entry, "ENTITY (optional) [PATH](optional)  [ATTRIBUTES](optional)")
		fmt.Println(`Obtains object(s) details. 
				If ENTITY is specified then it will enter a 'search mode' 
				and at least 1 ATTRIBUTE must be specified. Otherwise an 
				object's details will be retrieved`)
		printAttributeOptions()
	case "update":
		fmt.Println(`Usage: `, entry, "[PATH](optional)  [ATTRIBUTES]")
		fmt.Println(`Modify an object by specifying new attribute values`)
		printAttributeOptions()
	case "delete":
		fmt.Println(`Usage: `, entry, "[PATH]")
		fmt.Println(`Delete an object`)
	case "grep":
		fmt.Println("NOT YET IMPLEMENTED")
	default:
		fmt.Printf(`A Shell interface to the API and your datacenter visualisation solution`)
	}

}

func displayObject(obj map[string]interface{}) {
	for i := range obj {
		if i == "attributes" {
			for q := range obj[i].(map[string]interface{}) {
				val := string(obj[i].(map[string]interface{})[q].(string))
				if val == "" {
					println(q, ":", "NONE")
				} else {
					println(q, ":", val)
				}
			}
		} else {
			if i == "description" {
				print(i)
				inf := obj[i].([]interface{})
				for idx := range inf {
					println(inf[idx].(string))
				}
			} else if val, ok := obj[i].(string); ok == true {
				if val == "" {
					println(i, ":", "NONE")
				} else {
					println(i, ":", val)
				}
			} else {
				println(obj[i].(float64))
			}
		}

	}
}

func printAttributeOptions() {
	attrArr := []string{"address", "category", "city", "color",
		"country", "description", "domain", "gps", "height",
		"heightUnit", "id", "mainContact", "mainEmail", "mainPhone",
		"model", "name", "nbFloors", "orientation", "parentId", "posU",
		"posXY", "posXYUnit", "posZ", "posZUnit", "reserved", "reservedColor",
		"serial", "size", "sizeU", "sizeUnit", "slot", "technical",
		"technicalColor", "template", "token", "type", "usableColor",
		"vendor", "zipcode"}
	fmt.Println("Attributes: ")
	//fmt.Println("")
	for i := range attrArr {
		fmt.Println("", attrArr[i])
	}
}
