package main

import (
	"bytes"
	cmd "cli/controllers"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

var dynamicSymbolTable = make(map[string]interface{})
var funcTable = make(map[string]interface{})

func GetFuncTable() map[string]interface{} {
	return funcTable
}

func GetDynamicSymbolTable() map[string]interface{} {
	return dynamicSymbolTable
}

type node interface {
	execute() (interface{}, error)
}

type valueNode struct {
	val interface{}
}

func (n *valueNode) execute() (interface{}, error) {
	return n.val, nil
}

type ast struct {
	statements []node
}

func (a *ast) execute() (interface{}, error) {
	for i, _ := range a.statements {
		if a.statements[i] != nil {
			_, err := a.statements[i].execute()
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

type funcDefNode struct {
	name string
	body node
}

func (n *funcDefNode) execute() (interface{}, error) {
	funcTable[n.name] = n.body
	if cmd.State.DebugLvl >= 3 {
		println("New function ", n.name)
	}
	return nil, nil
}

type funcCallNode struct {
	name string
}

func (n *funcCallNode) execute() (interface{}, error) {
	val, ok := funcTable[n.name]
	if !ok {
		return nil, fmt.Errorf("undefined function %s", n.name)
	}
	body, ok := val.(node)
	if !ok {
		return nil, fmt.Errorf("variable %s does not contain a function", n.name)
	}
	return body.execute()
}

// At this time arrays are all []floats
type arrNode struct {
	nodes []node
}

func (n *arrNode) execute() (interface{}, error) {
	var r []interface{}
	for i := range n.nodes {
		v, err := n.nodes[i].execute()
		if err != nil {
			return nil, err
		}
		val, err := getFloat(v)
		if err != nil {
			return nil, fmt.Errorf("Array should contain floats")
		}
		r = append(r, val)
	}
	return r, nil
}

type lenNode struct {
	variable string
}

func (n *lenNode) execute() (interface{}, error) {
	val, ok := dynamicSymbolTable[n.variable]
	if !ok {
		return nil, fmt.Errorf("Undefined variable %s", n.variable)
	}
	arr, ok := val.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Variable %s does not contain an array.", n.variable)
	}
	return len(arr), nil
}

type postObjNode struct {
	entity string
	data   map[string]interface{}
}

func (n *postObjNode) execute() (interface{}, error) {
	return cmd.PostObj(cmd.EntityStrToInt(n.entity), n.entity, n.data)
}

type easyPostNode struct {
	entity string
	path   node
}

func (n *easyPostNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}

	data := make(map[string]interface{})
	/*x, e := ioutil.ReadFile(n.path)
	if e != nil {
		println("Error while opening file! " + e.Error())
		return nil
	}
	json.Unmarshal(x, &data)*/
	data = fileToJSON(path)
	if data == nil {
		return nil, fmt.Errorf("Cannot read json file.")
	}
	return cmd.PostObj(cmd.EntityStrToInt(n.entity), n.entity, data)
}

type helpNode struct {
	entry string
}

func (n *helpNode) execute() (interface{}, error) {
	cmd.Help(n.entry)
	return nil, nil
}

type focusNode struct {
	path node
}

func (n *focusNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	cmd.FocusUI(path)
	return nil, nil
}

type cdNode struct {
	path node
}

func (n *cdNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	return cmd.CD(path), nil
}

type lsNode struct {
	path node
}

func (n *lsNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	return cmd.LS(path), nil
}

type lsAttrNode struct {
	path node
	attr string
}

func (n *lsAttrNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	cmd.LSATTR(path, n.attr)
	return nil, nil
}

type lsAttrGenericNode struct {
	path node
	attr string
}

func (n *lsAttrGenericNode) execute() (interface{}, error) {
	path, err := AssertString(&n.path, "Path")
	if err != nil {
		return nil, err
	}
	cmd.LSATTR(path, n.attr)
	return nil, nil
}

type getUNode struct {
	path node
	u    int
}

func (n *getUNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	cmd.GetByAttr(path, n.u)
	return nil, nil
}

type getSlotNode struct {
	path node
	slot node
}

func (n *getSlotNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	slot, err := n.slot.execute()
	if err != nil {
		return nil, err
	}
	cmd.GetByAttr(path, slot)
	return nil, nil
}

type loadNode struct {
	path node
}

func (n *loadNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	cmd.LoadFile(path)
	return nil, nil
}

type loadTemplateNode struct {
	path node
}

func (n *loadTemplateNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	data := fileToJSON(path)
	if data == nil {
		return nil, fmt.Errorf("Cannot read json file : %s", path)
	}
	cmd.LoadTemplate(data, path)
	return path, nil
}

type printNode struct {
	expr node
}

func (n *printNode) execute() (interface{}, error) {
	val, err := n.expr.execute()
	if err != nil {
		return nil, err
	}
	return cmd.Print([]interface{}{val}), nil
}

type deleteObjNode struct {
	path node
}

func (n *deleteObjNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	return cmd.DeleteObj(path), nil
}

type deleteSelectionNode struct{}

func (n *deleteSelectionNode) execute() (interface{}, error) {
	return cmd.DeleteSelection(), nil
}

type isEntityDrawableNode struct {
	path node
}

func (n *isEntityDrawableNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	return cmd.IsEntityDrawable(path), nil
}

type isAttrDrawableNode struct {
	path node
	attr string
}

func (n *isAttrDrawableNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Object path should be a string")
	}
	return cmd.IsAttrDrawable(path, n.attr, nil, false), nil
}

type getObjectNode struct {
	path node
}

func (n *getObjectNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Object path should be a string")
	}
	v, _ := cmd.GetObject(path, false)
	if v == nil {
		return nil, fmt.Errorf("Cannot find object at path %s", path)
	}
	return v, nil
}

type selectObjectNode struct {
	path node
}

func (n *selectObjectNode) execute() (interface{}, error) {
	var selection []string
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Object path should be a string")
	}
	if path != "" {
		selection = []string{path}
	}

	cmd.CD(path)
	return cmd.SetClipBoard(selection)
}

type searchObjectsNode struct {
	objType string
	nodeMap map[string]interface{}
}

func (n *searchObjectsNode) execute() (interface{}, error) {
	valMap, err := evalMapNodes(n.nodeMap)
	if err != nil {
		return nil, err
	}
	resMap, err := resMap(valMap, n.objType, false)
	if err != nil {
		return nil, err
	}
	v := cmd.SearchObjects(n.objType, resMap)
	return v, nil
}

// TODO: Need to restore recursive updates or to remove it
// entirely
type recursiveUpdateObjNode struct {
	arg0 interface{}
	arg1 interface{}
	arg2 interface{}
}

func (n *recursiveUpdateObjNode) execute() (interface{}, error) {
	//Old code was removed since
	//it broke the OCLI syntax easy update
	if _, ok := n.arg2.(bool); ok {
		//Weird edge case
		//to solve issue with:
		// for i in $(ls) do $i[attr]="string"

		//n.arg0 = referenceToNode
		//n.arg1 = attributeString, (used as an index)
		//n.arg2 = someValue (usually a string)
		nodeVal, err := n.arg0.(node).execute()
		if err != nil {
			return nil, err
		}
		objMap := nodeVal.(map[string]interface{})

		if checkIfObjectNode(objMap) == true {
			val, err := n.arg2.(node).execute()
			if err != nil {
				return nil, err
			}
			updateArgs := map[string]interface{}{n.arg1.(string): val}
			id := objMap["id"].(string)
			entity := objMap["category"].(string)
			cmd.RecursivePatch("", id, entity, updateArgs)
		}

	} else {
		if n.arg2.(string) == "recursive" {
			cmd.RecursivePatch(n.arg0.(string), "", "", n.arg1.(map[string]interface{}))
		}
	}
	return nil, nil
}

type updateObjNode struct {
	path       node
	attributes map[string]interface{}
	hasSharp   bool //Refers to indexing in 'description' array attributes
}

func (n *updateObjNode) execute() (interface{}, error) {
	path, err := AssertString(&n.path, "Object path")
	if err != nil {
		return nil, err
	}

	attributes, err := evalMapNodes(n.attributes)
	if err != nil {
		return nil, err
	}
	if path == "_" {
		return nil, cmd.UpdateSelection(attributes)
	}

	//Check if the syntax refers to update or an interact command
	//
	for i := range attributes {
		vals := []string{"label", "labelFont", "content",
			"alpha", "tilesName", "tilesColor", "U", "slots", "localCS"}

		invalidVals := []string{"separator", "areas"}
		if AssertInStringValues(i, invalidVals) {
			msg := "This is invalid syntax. You must specify" +
				" 2 arrays (and for separator commands, the type) separated by '@' "
			return nil, fmt.Errorf(msg)
		}

		if AssertInStringValues(i, vals) {
			//labelFont should be 'bold' or 'italic' here in this node
			if i != "labelFont" && i != "label" && !IsBool(attributes[i]) &&
				attributes[i] != "true" && attributes[i] != "false" {
				msg := "Only boolean values can be used for interact commands"
				return nil, fmt.Errorf(msg)
			}

			if i == "labelFont" && attributes[i] != "bold" && attributes[i] != "italic" {
				msg := "The font can only be bold or italic" +
					" or be in the form of color@[colorValue]." +
					"\n\nFor more information please refer to: " +
					"\nhttps://github.com/ditrit/OGrEE-3D/wiki/CLI-langage#interact-with-objects"
				return nil, fmt.Errorf(msg)
			}
			return nil, cmd.InteractObject(path, i, attributes[i], n.hasSharp)
		}
	}
	return cmd.UpdateObj(path, "", "", attributes, false)
}

type specialUpdateNode struct {
	path     node
	variable string
	first    node
	second   node
	sepType  string
}

func (n *specialUpdateNode) execute() (interface{}, error) {
	path, err := AssertString(&n.path, "Object path")
	if err != nil {
		return nil, err
	}

	first, err := n.first.execute()
	if err != nil {
		return nil, err
	}
	second, err := n.second.execute()
	if err != nil {
		return nil, err
	}
	if n.variable == "areas" {
		if n.sepType != "" {
			return nil, fmt.Errorf("Unrecognised argument. Only 2 arrays can be specified")
		}
		areas := map[string]interface{}{"reserved": first, "technical": second}
		attributes, e := parseAreas(areas)
		if e != nil {
			return nil, e
		}
		return cmd.UpdateObj(path, "", "", attributes, false)
	} else if n.variable == "separator" {

		errorResponder := func(attr string, multi bool) (interface{}, error) {
			var errorMsg string
			if multi {
				errorMsg = "Invalid " + attr + " attributes provided." +
					" They must be arrays/lists/vectors with 2 elements."
			} else {
				errorMsg = "Invalid " + attr + " attribute provided." +
					" It must be an array/list/vector with 2 elements."
			}

			segment := " Please refer to the wiki or manual reference" +
				" for more details on how to create objects " +
				"using this syntax"

			return nil, fmt.Errorf(errorMsg + segment)
		}

		sepType := strings.ToLower(n.sepType)
		if sepType != "wireframe" && sepType != "plain" {
			msg := "Separator type must be specified " +
				"and can only be 'wireframe' or 'plain'"
			return nil, fmt.Errorf(msg)
		}

		if !IsInfArr(first) {
			if !IsInfArr(second) {
				return errorResponder("Starting and ending", true)
			}
			return errorResponder("Starting", false)
		}

		if !IsInfArr(second) {
			return errorResponder("Ending", false)
		}

		startLen := len(first.([]interface{}))
		endLen := len(second.([]interface{}))

		if startLen != 2 && endLen == 2 {
			return errorResponder("starting position", false)
		}

		if endLen != 2 && startLen == 2 {
			return errorResponder("ending position", false)
		}

		if startLen != 2 && endLen != 2 {
			return errorResponder("starting and ending position", true)
		}

		obj, _ := cmd.GetObject(path, true)
		if obj == nil {
			return nil, fmt.Errorf("cannot find object")
		}
		attr := obj["attributes"].(map[string]interface{})
		var sepArray []interface{}
		separators, _ := attr["separators"]
		if IsInfArr(separators) {
			sepArray = separators.([]interface{})
			sepArray = append(sepArray, map[string]interface{}{
				"startPosXYm": first, "endPosXYm": second, "type": sepType})

			sepArrStr, _ := json.Marshal(&sepArray)
			attr["separators"] = string(sepArrStr)
		} else {
			var sepStr string
			nextSep := map[string]interface{}{
				"startPosXYm": first, "endPosXYm": second, "type": sepType}

			nextSepStr, _ := json.Marshal(nextSep)
			if IsString(separators) && separators != "" && separators != "[]" {
				sepStr = separators.(string)
				size := len(sepStr)
				sepStr = sepStr[:size-1] + "," + string(nextSepStr) + "]"
			} else {
				sepStr = "[" + string(nextSepStr) + "]"
			}

			attr["separators"] = sepStr
		}

		return cmd.UpdateObj(path, "", "", attr, false)

	} else if n.variable == "labelFont" {
		//This section will be expanded later on as
		//the language grows
		if !IsStringValue(first, "color") {
			msg := "'color' attribute can only specified via this syntax"
			return nil, fmt.Errorf(msg)
		}

		c, ok := AssertColor(second)
		if ok == false {
			msg := "Please provide a valid 6 length hex value for the color"
			return nil, fmt.Errorf(msg)
		}
		second = "color@" + c

		//attr := map[string]interface{}{}

		return nil,
			cmd.InteractObject(path, "labelFont", second, false)
	} else {
		return nil, fmt.Errorf("Invalid attribute specified for room update")
	}
	//Control should not reach here
	//code added to suppress compiler error
	return nil, fmt.Errorf("Invalid syntax")
}

type easyUpdateNode struct {
	nodePath     string
	jsonPath     string
	deleteAndPut bool
}

func (n *easyUpdateNode) execute() (interface{}, error) {
	data := make(map[string]interface{})
	data = fileToJSON(n.jsonPath)
	if data == nil {
		return nil, fmt.Errorf("Cannot open json file")
	}
	return cmd.UpdateObj(n.nodePath, "", "", data, n.deleteAndPut)
}

type lsObjNode struct {
	path      node
	entity    int
	recursive bool
	sort      string
	attrList  []string
	format    string
}

func (n *lsObjNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	var objects []any
	if n.recursive {
		objects = cmd.LSOBJECTRecursive(path, n.entity)
	} else {
		objects = cmd.LSOBJECT(path, n.entity)
	}
	if n.sort != "" {
		objects = cmd.SortObjects(&objects, n.sort).GetData()
	}
	if n.attrList != nil {
		if n.format != "" {
			cmd.DispfWithAttrs(n.format, &objects, &n.attrList)
		} else {
			cmd.DispWithAttrs(&objects, &n.attrList)
		}
	} else {
		if n.sort != "" {
			//We want to display the attribute used for sorting
			attrList := append(n.attrList, n.sort)
			cmd.DispWithAttrs(&objects, &attrList)
		} else {
			for i := range objects {
				object, ok := objects[i].(map[string]interface{})
				if ok && object != nil && object["name"] != nil {
					println(object["name"].(string))
				}
			}
		}
	}
	return objects, nil
}

type treeNode struct {
	path  node
	depth int
}

func (n *treeNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	cmd.Tree(path, n.depth)
	return nil, nil
}

type drawNode struct {
	path     node
	depth    int
	argument map[string]string
}

func (n *drawNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}

	if n.argument != nil {
		if len(n.argument) > 1 {
			msg := "Too many flags supplied, only -f acceptable"
			return nil, fmt.Errorf(msg)
		}
		if n.argument["f"] == "n" || n.argument["f"] == "y" {
			return nil, cmd.Draw(path, n.depth, n.argument["f"])
		}
		msg := "Unrecognised argument, only -f and 'y' or 'n' are acceptable"
		return nil, fmt.Errorf(msg)
	}
	return nil, cmd.Draw(path, n.depth, "")
}

type undrawNode struct {
	path node
}

func (n *undrawNode) execute() (interface{}, error) {
	if n.path == nil {
		return nil, cmd.Undraw("")
	}
	path, e := AssertString(&(n.path), "Path")
	if e != nil {
		return nil, e
	}
	return nil, cmd.Undraw(path)
}

type lsogNode struct{}

func (n *lsogNode) execute() (interface{}, error) {
	cmd.LSOG()
	return nil, nil
}

type lsenterpriseNode struct{}

func (n *lsenterpriseNode) execute() (interface{}, error) {
	cmd.LSEnterprise()
	return nil, nil
}

type exitNode struct{}

func (n *exitNode) execute() (interface{}, error) {
	cmd.Exit()
	return nil, nil
}

type clrNode struct{}

func (n *clrNode) execute() (interface{}, error) {
	cmd.Clear()
	return nil, nil
}

type envNode struct{}

func (n *envNode) execute() (interface{}, error) {
	cmd.Env(dynamicSymbolTable, funcTable)
	return nil, nil
}

type selectNode struct{}

func (n *selectNode) execute() (interface{}, error) {
	return cmd.ShowClipBoard(), nil
}

type pwdNode struct{}

func (n *pwdNode) execute() (interface{}, error) {
	return cmd.PWD(), nil
}

type grepNode struct{}

func (n *grepNode) execute() (interface{}, error) {
	return nil, nil
}

type selectChildrenNode struct {
	paths []node
}

func (n *selectChildrenNode) execute() (interface{}, error) {
	paths, err := evalNodeArr[string](&n.paths, []string{})
	if err != nil {
		return nil, err
	}
	v, err := cmd.SetClipBoard(paths)
	if err != nil {
		return nil, err
	}
	if cmd.State.DebugLvl > cmd.NONE {
		println("Selection made!")
	}

	return v, nil
}

type updateSelectNode struct {
	data map[string]interface{}
}

func (n *updateSelectNode) execute() (interface{}, error) {
	cmd.UpdateSelection(n.data)
	return nil, nil
}

type unsetFuncNode struct {
	funcName string
}

func (n *unsetFuncNode) execute() (interface{}, error) {
	delete(funcTable, n.funcName)
	return nil, nil
}

type unsetVarNode struct {
	varName string
}

func (n *unsetVarNode) execute() (interface{}, error) {
	delete(dynamicSymbolTable, n.varName)
	return nil, nil
}

type unsetAttrNode struct {
	path node
}

func (n *unsetAttrNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	arr := strings.Split(path, ":")
	if len(arr) != 2 {
		msg := "You must specify the attribute to delete with a colon!\n" +
			"(ie. $> unset path/to/object:attributex). \n" +
			"Please refer to the language reference help for more details" +
			"\n($> man unset)"
		return nil, fmt.Errorf(msg)
	}
	path = arr[0]
	data := map[string]interface{}{arr[1]: nil}

	return cmd.UpdateObj(path, "", "", data, true)

}

type setEnvNode struct {
	arg  string
	expr node
}

func (n *setEnvNode) execute() (interface{}, error) {
	val, err := n.expr.execute()
	if err != nil {
		return nil, err
	}
	cmd.SetEnv(n.arg, val)
	return nil, nil
}

type hierarchyNode struct {
	path  node
	depth int
}

func (n *hierarchyNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	return cmd.GetHierarchy(path, n.depth, false), nil

}

type createTenantNode struct {
	path  node
	color node
}

func (n *createTenantNode) execute() (interface{}, error) {
	pathVal, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	colorInf, err := n.color.execute()
	if err != nil {
		return nil, err
	}
	color, ok := AssertColor(colorInf)
	if !ok {
		msg := "Please provide a valid 6 length hex value for the color"
		return nil, fmt.Errorf(msg)
	}
	attributes := map[string]any{"attributes": map[string]any{"color": color}}
	err = cmd.GetOCLIAtrributes(path, cmd.TENANT, attributes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createSiteNode struct {
	path        node
	orientation node
}

func (n *createSiteNode) execute() (interface{}, error) {
	pathVal, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	orientationAny, err := n.orientation.execute()
	if err != nil {
		return nil, err
	}
	orientation, ok := orientationAny.(string)
	if !ok {
		return nil, fmt.Errorf("orientation should be a string")
	}
	if !checkIfOrientation(orientation) {
		return nil, fmt.Errorf("You must provide a valid orientation")
	}
	attributes := map[string]any{"attributes": map[string]any{"orientation": orientation}}
	err = cmd.GetOCLIAtrributes(path, cmd.SITE, attributes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createBuildingNode struct {
	path  node
	posXY node
	size  node
}

func (n *createBuildingNode) execute() (interface{}, error) {
	pathVal, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	posXY, err := n.posXY.execute()
	if err != nil {
		return nil, err
	}
	size, err := n.size.execute()
	if err != nil {
		return nil, err
	}
	attributes := map[string]any{"attributes": map[string]any{"posXY": posXY, "size": size}}
	err = cmd.GetOCLIAtrributes(path, cmd.BLDG, attributes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createRoomNode struct {
	path        node
	posXY       node
	size        node
	orientation node
	floorUnit   node
}

func (n *createRoomNode) execute() (interface{}, error) {
	pathVal, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	posXY, err := n.posXY.execute()
	if err != nil {
		return nil, err
	}
	size, err := n.size.execute()
	if err != nil {
		return nil, err
	}
	orientationAny, err := n.orientation.execute()
	if err != nil {
		return nil, err
	}
	orientation, ok := orientationAny.(string)
	if !ok {
		return nil, fmt.Errorf("orientation should be a string")
	}
	attributes := map[string]any{"attributes": map[string]any{"posXY": posXY, "size": size, "orientation": orientation}}
	if n.floorUnit != nil {
		floorUnitAny, err := n.floorUnit.execute()
		if err != nil {
			return nil, err
		}
		floorUnit, ok := floorUnitAny.(string)
		if !ok {
			return nil, fmt.Errorf("floorUnit should be a string")
		}
		attributes["attributes"].(map[string]any)["floorUnit"] = floorUnit
	}
	err = cmd.GetOCLIAtrributes(path, cmd.ROOM, attributes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createRoomFromTemplateNode struct {
	path     node
	posXY    node
	template node
}

func (n *createRoomFromTemplateNode) execute() (interface{}, error) {
	pathVal, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	posXY, err := n.posXY.execute()
	if err != nil {
		return nil, err
	}
	templateAny, err := n.template.execute()
	if err != nil {
		return nil, err
	}
	template, ok := templateAny.(string)
	if !ok {
		return nil, fmt.Errorf("template path should be a string")
	}
	attributes := map[string]any{"attributes": map[string]any{"posXY": posXY, "template": template}}
	err = cmd.GetOCLIAtrributes(path, cmd.ROOM, attributes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createRackNode struct {
	path           node
	posXY          node
	sizeOrTemplate node
	orientation    node
}

func (n *createRackNode) execute() (interface{}, error) {
	pathVal, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	posXY, err := n.posXY.execute()
	if err != nil {
		return nil, err
	}
	orientation, err := n.orientation.execute()
	if err != nil {
		return nil, err
	}
	sizeOrTemplate, err := n.sizeOrTemplate.execute()
	if err != nil {
		return nil, err
	}
	attr := make(map[string]interface{})
	if !checkIfTemplate(sizeOrTemplate) {
		attr["size"] = sizeOrTemplate
	} else {
		attr["template"] = sizeOrTemplate
	}
	attr["posXY"] = posXY
	attr["orientation"] = orientation
	attributes := map[string]interface{}{"attributes": attr}
	err = cmd.GetOCLIAtrributes(path, cmd.RACK, attributes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createDeviceNode struct {
	path            node
	posUOrSlot      node
	sizeUOrTemplate node
	side            node
}

func (n *createDeviceNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	posUOrSlot, err := n.posUOrSlot.execute()
	if err != nil {
		return nil, err
	}
	sizeUOrTemplate, err := n.sizeUOrTemplate.execute()
	if err != nil {
		return nil, err
	}
	attr := map[string]interface{}{"posU/slot": posUOrSlot}
	if !checkIfTemplate(sizeUOrTemplate) {
		attr["sizeU"] = sizeUOrTemplate
	} else {
		attr["template"] = sizeUOrTemplate
	}
	if n.side != nil {
		side, err := n.side.execute()
		if err != nil {
			return nil, err
		}
		attr["orientation"] = side
	}
	attributes := map[string]interface{}{"attributes": attr}
	err = cmd.GetOCLIAtrributes(path, cmd.DEVICE, attributes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createGroupNode struct {
	path  node
	paths []node
}

func (n *createGroupNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	var objs []string
	data := map[string]interface{}{}
	for i := range n.paths {
		v, err := n.paths[i].execute()
		if err != nil {
			return nil, err
		}
		obj, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("")
		}
		obj = filepath.Base(obj)
		objs = append(objs, obj)
	}

	data["attributes"] = map[string]interface{}{"content": objs}
	err = cmd.GetOCLIAtrributes(path, cmd.GROUP, data)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createCorridorNode struct {
	path      node
	leftRack  node
	rightRack node
	temp      node
}

func (n *createCorridorNode) execute() (interface{}, error) {
	path, err := AssertString(&n.path, "Path for corridor")
	if err != nil {
		return nil, err
	}

	leftRack, err2 := AssertString(&n.leftRack, "Path for left rack")
	if err2 != nil {
		return nil, err2
	}

	rightRack, err3 := AssertString(&n.rightRack, "Path for right rack")
	if err3 != nil {
		return nil, err3
	}

	temp, err4 := AssertString(&n.temp, "Temperature")
	if err4 != nil {
		return nil, err4
	}
	tempIsValid := AssertInStringValues(temp, []string{"warm", "cold"})
	if !tempIsValid {
		return nil,
			fmt.Errorf("Temperature should be either 'warm' or 'cold'")
	}
	leftRack = filepath.Base(leftRack)
	rightRack = filepath.Base(rightRack)

	attributes := map[string]interface{}{
		"content": leftRack + "," + rightRack, "temperature": temp}

	data := map[string]interface{}{"attributes": attributes}

	err = cmd.GetOCLIAtrributes(path, cmd.CORIDOR, data)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createOrphanNode struct {
	path     node
	template node
	sensor   bool
}

func (n *createOrphanNode) execute() (interface{}, error) {
	pathVal, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("path should be a string")
	}
	templateAny, err := n.template.execute()
	if err != nil {
		return nil, err
	}
	template, ok := templateAny.(string)
	if !ok {
		return nil, fmt.Errorf("template path should be a string")
	}
	attributes := map[string]interface{}{"attributes": map[string]any{"template": template}}

	var t int
	if n.sensor {
		t = cmd.STRAYSENSOR
	} else {
		t = cmd.STRAY_DEV
	}
	err = cmd.GetOCLIAtrributes(path, t, attributes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type uiDelayNode struct {
	time float64
}

func (n *uiDelayNode) execute() (interface{}, error) {
	cmd.UIDelay(n.time)
	return nil, nil
}

type uiToggleNode struct {
	feature string
	enable  bool
}

func (n *uiToggleNode) execute() (interface{}, error) {
	cmd.UIToggle(n.feature, n.enable)
	return nil, nil
}

type uiHighlightNode struct {
	path node
}

func (n *uiHighlightNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	return nil, cmd.UIHighlight(path)
}

type cameraMoveNode struct {
	command  string
	position []float64
	rotation []float64
}

func (n *cameraMoveNode) execute() (interface{}, error) {
	if len(n.position) != 3 {
		return nil, fmt.Errorf("Position (first argument) is invalid\nPlease provide a vector3")
	}
	if len(n.rotation) != 2 {
		return nil, fmt.Errorf("Rotation (second argument) is invalid\nPlease provide a vector2")
	}
	cmd.CameraMove(n.command, n.position, n.rotation)
	return nil, nil
}

type cameraWaitNode struct {
	time float64
}

func (n *cameraWaitNode) execute() (interface{}, error) {
	cmd.CameraWait(n.time)
	return nil, nil
}

type linkObjectNode struct {
	source      node
	destination node
	slot        node
}

func (n *linkObjectNode) execute() (interface{}, error) {
	var slot interface{}
	source, err := AssertString(&n.source, "Source Object Path")
	if err != nil {
		return nil, err
	}

	dest, err1 := AssertString(&n.destination, "Destination Object Path")
	if err1 != nil {
		return nil, err1
	}

	if n.slot != nil {
		s, e := n.slot.execute()
		if e != nil {
			return nil, e
		}
		slot = s
	} else {
		slot = nil
	}

	cmd.LinkObject(source, dest, slot)
	return nil, nil
}

type unlinkObjectNode struct {
	source      node
	destination node
}

func (n *unlinkObjectNode) execute() (interface{}, error) {
	destination := ""
	source, err := AssertString(&n.source, "Source Object Path")
	if err != nil {
		return nil, err
	}

	if n.destination != nil {
		var e error
		destination, e = AssertString(&n.destination, "Destination Object Path")
		if e != nil {
			return nil, e
		}
	}
	cmd.UnlinkObject(source, destination)
	return nil, nil
}

type symbolReferenceNode struct {
	va string
}

func (s *symbolReferenceNode) execute() (interface{}, error) {
	val, ok := dynamicSymbolTable[s.va]
	if !ok {
		return nil, fmt.Errorf("Undefined variable %s", s.va)
	}
	switch v := val.(type) {
	case string, int, bool, float64, float32, map[int]interface{}:
		if cmd.State.DebugLvl >= 3 {
			println("So You want the value: ", v)
		}
	}
	return val, nil
}

type objReferenceNode struct {
	va    string
	index node
}

func (o *objReferenceNode) execute() (interface{}, error) {
	val, ok := dynamicSymbolTable[o.va]
	if !ok {
		return nil, fmt.Errorf("Undefined variable %s", o.va)
	}
	if _, ok := val.(map[string]interface{}); !ok {
		return nil, fmt.Errorf(o.va + " Is not an indexable object")
	}
	object := val.(map[string]interface{})

	idx, e := o.index.execute()
	if e != nil {
		return nil, e
	}
	if _, ok := idx.(string); !ok {
		return nil, fmt.Errorf("The index must resolve to a string")
	}
	index := idx.(string)

	if mainAttr, ok := object[index]; ok {
		return mainAttr, nil
	} else {
		if attrInf, ok := object["attributes"]; ok {
			if attrDict, ok := attrInf.(map[string]interface{}); ok {
				if _, ok := attrDict[index]; ok {
					return attrDict[index], nil
				}
			}
		}
	}

	msg := "This object " + o.va + " cannot be indexed with " + index +
		". Please check the object you are referencing and try again"

	return nil, fmt.Errorf(msg)

}

type arrayReferenceNode struct {
	variable string
	idx      node
}

func (n *arrayReferenceNode) execute() (interface{}, error) {
	v, ok := dynamicSymbolTable[n.variable]
	if !ok {
		return nil, fmt.Errorf("Undefined variable %s", n.variable)
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil, fmt.Errorf("You can only index an array.")
	}
	idx, err := n.idx.execute()
	if err != nil {
		return nil, err
	}
	i, ok := idx.(int)
	if !ok {
		return nil, fmt.Errorf("Index should be an integer.")
	}
	if i < 0 || i >= len(arr) {
		return nil, fmt.Errorf(
			"Index out of range\n"+
				"Array length : %d"+
				"But desired index at : %d",
			len(arr), i,
		)
	}
	return arr[i], nil
}

type assignNode struct {
	variable string
	val      node
}

func (a *assignNode) execute() (interface{}, error) {
	val, err := a.val.execute()
	if err != nil {
		return nil, err
	}
	switch v := val.(type) {
	case bool, int, float64, string, []interface{}, map[string]interface{}:
		dynamicSymbolTable[a.variable] = v
		if cmd.State.DebugLvl >= 3 {
			println("You want to assign", a.variable, "with value of", v)
		}
		return nil, nil
	}
	return nil, fmt.Errorf("Invalid type to assign variable %s", a.variable)
}

// Checks the map and sees if it is an object type
func checkIfObjectNode(x map[string]interface{}) bool {
	if idInf, ok := x["id"]; ok {
		if id, ok := idInf.(string); ok {
			if len(id) == 24 {
				if catInf, ok := x["category"]; ok {
					if _, ok := catInf.(string); ok {
						return true
					}
				}

				if slugInf, ok := x["slug"]; ok {
					if _, ok := slugInf.(string); ok {
						return true
					}
				}
			}
		}
	}
	return false
}

// Hack function for the [room]:areas=[r1,r2,r3,r4]@[t1,t2,t3,t4]
// command
func parseAreas(areas map[string]interface{}) (map[string]interface{}, error) {
	var reservedStr string
	var techStr string

	errorResponder := func(attr string) (map[string]interface{}, error) {
		errorMsg := "Invalid " + attr + " attribute provided." +
			" It must be an array/list/vector with 4 elements." +
			" Please refer to the wiki or manual reference" +
			" for more details on how to create objects " +
			"using this syntax"
		return nil, fmt.Errorf(errorMsg)
	}

	if reserved, ok := areas["reserved"].([]interface{}); ok {
		if tech, ok := areas["technical"].([]interface{}); ok {
			if len(reserved) == 4 && len(tech) == 4 {
				var r [4]*bytes.Buffer
				var t [4]*bytes.Buffer
				for i := 3; i >= 0; i-- {
					r[i] = bytes.NewBufferString("")
					fmt.Fprintf(r[i], "%v", reserved[i])
					t[i] = bytes.NewBufferString("")
					fmt.Fprintf(t[i], "%v", tech[i])
				}
				reservedStr = "{\"left\":" + r[3].String() + ",\"right\":" + r[2].String() + ",\"top\":" + r[0].String() + ",\"bottom\":" + r[1].String() + "}"
				techStr = "{\"left\":" + t[3].String() + ",\"right\":" + t[2].String() + ",\"top\":" + t[0].String() + ",\"bottom\":" + t[1].String() + "}"
				areas["reserved"] = reservedStr
				areas["technical"] = techStr
			} else {
				if len(reserved) != 4 && len(tech) == 4 {
					return errorResponder("reserved")
				} else if len(tech) != 4 && len(reserved) == 4 {
					return errorResponder("technical")
				} else { //Both invalid
					return errorResponder("reserved and technical")
				}
			}
		} else {
			return errorResponder("technical")
		}
	} else {
		return errorResponder("reserved")
	}
	return areas, nil
}
