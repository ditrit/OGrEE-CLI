package main

import (
	cmd "cli/controllers"
	l "cli/logger"
	"fmt"
	"strconv"
)

type numNode interface {
	getNum() (float64, error)
	execute() (interface{}, error)
}

type floatNode interface {
	getFloat() (float64, error)
	execute() (interface{}, error)
}

type floatLeaf struct {
	val float64
}

func (l floatLeaf) getFloat() (float64, error) {
	return l.val, nil
}
func (l floatLeaf) execute() (interface{}, error) {
	return l.val, nil
}

func (l floatLeaf) getNum() (float64, error) {
	return l.val, nil
}

type intNode interface {
	getInt() (int, error)
}

type intLeaf struct {
	val int
}

func (l intLeaf) getInt() (int, error) {
	return l.val, nil
}
func (l intLeaf) execute() (interface{}, error) {
	return l.val, nil
}
func (l intLeaf) getNum() (float64, error) {
	return float64(l.val), nil
}

func numToString(num any) string {
	switch v := num.(type) {
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	panic("non numeric type to convert to string")
}

type arithNode struct {
	op    string
	left  node
	right node
}

func (a *arithNode) execute() (interface{}, error) {
	lv, err := a.left.execute()
	if err != nil {
		return nil, err
	}
	if cmd.State.DebugLvl >= 3 {
		println("Left:", lv)
	}
	rv, err := a.right.execute()
	if err != nil {
		return nil, err
	}
	if cmd.State.DebugLvl >= 3 {
		println("Right: ", rv)
	}

	leftIntVal, leftInt := lv.(int)
	rightIntVal, rightInt := rv.(int)
	leftFloatVal, leftFloat := lv.(float64)
	rightFloatVal, rightFloat := rv.(float64)
	leftStringVal, leftString := lv.(string)
	rightStringVal, rightString := rv.(string)

	if leftInt && rightInt {
		switch a.op {
		case "+":
			return leftIntVal + rightIntVal, nil
		case "-":
			return leftIntVal - rightIntVal, nil
		case "*":
			return leftIntVal * rightIntVal, nil
		case "/":
			return leftIntVal / rightIntVal, nil
		case "%":
			return leftIntVal % rightIntVal, nil
		default:
			return nil, fmt.Errorf("Invalid operator for integer operands")
		}
	}
	if (leftInt || leftFloat) && (rightInt || rightFloat) {
		if leftInt {
			leftFloatVal = float64(leftIntVal)
		}
		if rightInt {
			rightFloatVal = float64(rightIntVal)
		}
		switch a.op {
		case "+":
			return leftFloatVal + rightFloatVal, nil
		case "-":
			return leftFloatVal - rightFloatVal, nil
		case "*":
			return leftFloatVal * rightFloatVal, nil
		case "/":
			return leftFloatVal / rightFloatVal, nil
		default:
			return nil, fmt.Errorf("Invalid operator for float operands")
		}
	}
	if leftString || rightString {
		if !leftString && (leftFloat || leftInt) {
			leftStringVal = numToString(lv)
		}
		if !rightString && (rightFloat || rightInt) {
			rightStringVal = numToString(lv)
		}
		switch a.op {
		case "+":
			return leftStringVal + rightStringVal, nil
		default:
			return nil, fmt.Errorf("Invalid operator for string operands")
		}
	}
	l.GetWarningLogger().Println("Invalid arithmetic operation attempted")
	return nil, fmt.Errorf("Invalid arithmetic operation attempted")
}

type negateNode struct {
	val node
}

func (n *negateNode) execute() (interface{}, error) {
	v, err := n.val.execute()
	if err != nil {
		return nil, err
	}
	intVal, isInt := v.(int)
	if isInt {
		return -intVal, nil
	}
	floatVal, isFloat := v.(float64)
	if isFloat {
		return -floatVal, nil
	}
	return nil, fmt.Errorf("cannot negate non numeric value")
}
