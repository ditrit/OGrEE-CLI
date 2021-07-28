package controllers

import (
	"cli/models"
	"container/list"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	TENANT = iota
	SITE
	BLDG
	ROOM
	RACK
	DEVICE
	SUBDEV
	SUBDEV1
)

type ShellState struct {
	CurrPath      string
	PrevPath      string
	sessionBuffer list.List
	TreeHierarchy *Node
}

type Node struct {
	ID     int
	PID    int
	Entity int
	Name   string
	Nodes  list.List
}

var State ShellState

//Populate hierarchy into B Tree like
//structure
func InitState() {
	State.TreeHierarchy = &(Node{})
	(*(State.TreeHierarchy)).Entity = 0
	State.TreeHierarchy.PID = -1
	State.CurrPath = "/"
	x := GetChildren(0)
	for i := range x {
		State.TreeHierarchy.Nodes.PushBack(x[i])
	}

	for i := 1; i < 8; i++ {
		time.Sleep(2 * time.Second)
		x := GetChildren(i)
		for k := range x {
			SearchAndInsert(&State.TreeHierarchy, x[k], i)
		}
	}
}

func GetChildren(curr int) []*Node {
	switch curr {
	case TENANT:
		println("TENANT")
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/tenants", GetKey(),
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, TENANT)
	case SITE:
		println("SITE")
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/sites", GetKey(),
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, SITE)
	case BLDG:
		println("BLDG")
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/buildings", GetKey(),
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, BLDG)
	case ROOM:
		println("ROOM")
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/rooms", GetKey(),
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, ROOM)
	case RACK:
		println("RACK")
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/racks", GetKey(),
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, RACK)
	case DEVICE:
		println("DEVICE")
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/devices", GetKey(),
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, DEVICE)
	case SUBDEV:
		println("SUBDEV")
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/subdevices", GetKey(),
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, SUBDEV)

	case SUBDEV1:
		println("SUBDEV1")
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/subdevice1s", GetKey(),
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, SUBDEV1)
	}
	return nil
}

func SearchAndInsert(root **Node, node *Node, dt int) {
	if root != nil {
		for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
			if node.PID == (i.Value).(*Node).ID {
				//println("NODE ", node.Name, "WITH PID: ", node.PID)
				//println("Matched with PARENT ")
				//println()
				(i.Value).(*Node).Nodes.PushBack(node)
				return
			}
			x := (i.Value).(*Node)
			SearchAndInsert(&x, node, dt+1)
		}
	}
	return
}

//Function is an abstraction of a normal exit
func Exit() {
	//writeHistoryOnExit(&State.sessionBuffer)
	//runtime.Goexit()
	os.Exit(0)
}

func makeNodeArrFromResp(resp *http.Response, entity int) []*Node {
	arr := []*Node{}
	var jsonResp map[string]interface{}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error: " + err.Error() + " Now Exiting")
		Exit()
	}
	json.Unmarshal(bodyBytes, &jsonResp)

	objs, ok := ((jsonResp["data"]).(map[string]interface{})["objects"]).([]interface{})
	sd1obj, ok1 := ((jsonResp["data"]).(map[string]interface{})["subdevices1"]).([]interface{})
	if !ok && !ok1 {
		println("Nothing found!")
		return nil
	} else if ok1 && !ok {
		objs = sd1obj
	}
	for i, _ := range objs {
		node := &Node{}
		node.Entity = entity
		node.Name = (string((objs[i].(map[string]interface{}))["name"].(string)))
		node.ID, _ = strconv.Atoi((objs[i].(map[string]interface{}))["id"].(string))
		num, ok := objs[i].(map[string]interface{})["parentId"].(float64)
		if !ok {
			node.PID, _ = strconv.Atoi((objs[i].(map[string]interface{}))["parentId"].(string))
		} else {
			node.PID = int(num)
		}
		arr = append(arr, node)
	}
	return arr
}

//func DispTree() {
//	nd := &(Node{})
//	nd.Entity = -1
//	Populate(&nd, 0)
//	println("Now viewing the tree...")
//	View(nd, 0)
//}

func View(root *Node, dt int) {
	if dt != 7 || root != nil {
		arr := (*root).Nodes
		for i := arr.Front(); i != nil; i = i.Next() {

			println("Now Printing children of: ",
				(*Node)((i.Value).(*Node)).Name)
			//println()
			View(((i.Value).(*Node)), dt+1)
		}
	}
}

func StrToStack(x string) *Stack {
	stk := Stack{}
	numPrev := 0
	sarr := strings.Split(x, "/")
	for i := len(sarr) - 1; i >= 0; i-- {
		if sarr[i] == ".." {
			numPrev += 1
		} else if sarr[i] != "" {
			if numPrev == 0 {
				stk.Push(sarr[i])
			} else {
				numPrev--
			}
		}

	}
	return &stk
}

func getNextInPath(name string, root *Node) *Node {
	for i := root.Nodes.Front(); i != nil; i = i.Next() {
		if (i.Value.(*Node)).Name == name {
			return (i.Value.(*Node))
		}
	}
	return nil
}

func DispAtLevel(root **Node, x Stack) []string {
	if x.Len() > 0 {
		name := x.Peek()
		node := getNextInPath(name.(string), *root)
		if node == nil {
			println("Name doesn't exist! ", string(name.(string)))
			return nil
		}
		x.Pop()
		return DispAtLevel(&node, x)
	} else {
		var items = make([]string, 0)
		var nm string
		println("This is what we got:")
		for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
			nm = string(i.Value.(*Node).Name)
			println(nm)
			items = append(items, nm)
		}
		return items
	}
	return nil
}

func DispAtLevelTAB(root **Node, x Stack) []string {
	if x.Len() > 0 {
		name := x.Peek()
		node := getNextInPath(name.(string), *root)
		if node == nil {
			//println("Name doesn't exist! ", string(name.(string)))
			return nil
		}
		x.Pop()
		return DispAtLevelTAB(&node, x)
	} else {
		var items = make([]string, 0)
		var nm string
		//println("This is what we got:")
		for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
			nm = string(i.Value.(*Node).Name)
			//println(nm)
			items = append(items, nm)
		}
		return items
	}
	return nil
}

func DispStk(x Stack) {
	for i := x.Pop(); i != nil; i = x.Pop() {
		println((i.(*Node)).Name)
	}
}

func GetPathStrAtPtr(root, curr **Node, path string) (bool, string) {
	if root == nil || *root == nil {
		return false, ""
	}

	if *root == *curr {
		if path == "" {
			path = "/"
		}
		return true, path
	}

	for i := (**root).Nodes.Front(); i != nil; i = i.Next() {
		nd := (*Node)((i.Value.(*Node)))
		exist, path := GetPathStrAtPtr(&nd,
			curr, path+"/"+i.Value.(*Node).Name)
		if exist == true {
			return exist, path
		}
	}
	return false, path
}

func CheckPath(root **Node, x, pstk *Stack) (bool, string) {
	if x.Len() == 0 {
		_, path := GetPathStrAtPtr(&State.TreeHierarchy, root, "")
		//println(path)
		return true, path
	}

	p := x.Pop()

	//At Root
	if pstk.Len() == 0 && string(p.(string)) == ".." {
		//Pop until p != ".."
		for ; p != nil && string(p.(string)) == ".."; p = x.Pop() {
		}
		if p == nil {
			_, path := GetPathStrAtPtr(&State.TreeHierarchy, root, "/")
			//println(path)
			return true, path
		}

		//Somewhere in tree
	} else if pstk.Len() > 0 && string(p.(string)) == ".." {
		prevNode := (pstk.Pop()).(*Node)
		return CheckPath(&prevNode, x, pstk)
	}

	nd := getNextInPath(string(p.(string)), *root)
	if nd == nil {
		return false, ""
	}

	pstk.Push(*root)
	return CheckPath(&nd, x, pstk)

}

func UpdateTree(root **Node, curr *Node) bool {
	if root == nil {
		return false
	}

	//Add only when the PID matches Parent's ID
	if (*root).ID == curr.PID && curr.Entity == (*root).Entity+1 {
		(*root).Nodes.PushBack(curr)
		return true
	}

	for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
		nxt := (i.Value).(*Node)
		x := UpdateTree(&nxt, curr)
		if x != false {
			return true
		}
	}
	return false
}

//Return extra bool so that the Parent can delete
//leaf and keep track without stack
func DeleteNodeInTree(root **Node, ID, ent int) (bool, bool) {
	if root == nil {
		return false, false
	}

	//Delete only when the PID matches Parent's ID
	if (*root).ID == ID && ent == (*root).Entity {
		return true, false
	}

	for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
		nxt := (i.Value).(*Node)
		first, deleted := DeleteNodeInTree(&nxt, ID, ent)
		if first == true && deleted == false {
			(*root).Nodes.Remove(i)
			return true, true
		}
	}
	return false, false
}

func FindNodeInTree(root **Node, path *Stack) *Node {
	if root == nil {
		return nil
	}

	if path.Len() > 0 {
		name := path.Peek()
		node := getNextInPath(name.(string), *root)
		if node == nil {
			println("Name doesn't exist! ", string(name.(string)))
			return nil
		}
		path.Pop()
		return FindNodeInTree(&node, path)
	} else {
		return *root
	}
}

func EntityToString(entity int) string {
	switch entity {
	case TENANT:
		return "tenant"
	case SITE:
		return "site"
	case BLDG:
		return "building"
	case ROOM:
		return "room"
	case RACK:
		return "rack"
	case DEVICE:
		return "device"
	case SUBDEV:
		return "subdevice"
	default:
		return "subdevice1"
	}
}

func EntityStrToInt(entity string) int {
	switch entity {
	case "tenant":
		return TENANT
	case "site":
		return SITE
	case "building":
		return BLDG
	case "room":
		return ROOM
	case "rack":
		return RACK
	case "device":
		return DEVICE
	case "subdevice":
		return SUBDEV
	default:
		return SUBDEV1
	}
}