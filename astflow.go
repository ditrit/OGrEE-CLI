package main

import "fmt"

type ifNode struct {
	condition  node
	ifBranch   node
	elseBranch node
}

func (n *ifNode) execute() (interface{}, error) {
	val, err := n.condition.execute()
	if err != nil {
		return nil, err
	}
	condition, ok := val.(bool)
	if !ok {
		return nil, fmt.Errorf("Condition should be a boolean")
	}
	if condition == true {
		_, err := n.ifBranch.execute()
		if err != nil {
			return nil, err
		}
	} else {
		if n.elseBranch != nil {
			_, err := n.elseBranch.execute()
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

type whileNode struct {
	condition node
	body      node
}

func (n *whileNode) execute() (interface{}, error) {
	for true {
		val, err := n.condition.execute()
		if err != nil {
			return nil, err
		}
		condition, ok := val.(bool)
		if !ok {
			return nil, fmt.Errorf("Condition should be a boolean")
		}
		if !condition {
			break
		}
		_, err = n.body.execute()
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

type forNode struct {
	init        node
	condition   node
	incrementor node
	body        node
}

func (n *forNode) execute() (interface{}, error) {
	_, err := n.init.execute()
	if err != nil {
		return nil, err
	}
	for true {
		val, err := n.condition.execute()
		if err != nil {
			return nil, err
		}
		condition, ok := val.(bool)
		if !ok {
			return nil, fmt.Errorf("Condition should be a boolean")
		}
		if !condition {
			break
		}
		_, err = n.body.execute()
		if err != nil {
			return nil, err
		}
		_, err = n.incrementor.execute()
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

type forArrayNode struct {
	variable string
	arr      node
	body     node
}

func (n *forArrayNode) execute() (interface{}, error) {
	val, err := n.arr.execute()
	if err != nil {
		return nil, err
	}
	arr, ok := val.([]interface{})
	if !ok {
		return nil, fmt.Errorf("only an array can be iterated")
	}
	for _, v := range arr {
		_, err := (&assignNode{n.variable, &valueNode{v}}).execute()
		if err != nil {
			return nil, err
		}
		_, err = n.body.execute()
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

type forRangeNode struct {
	variable string
	start    int
	end      int
	body     node
}

func (n *forRangeNode) execute() (interface{}, error) {
	for i := n.start; i <= n.end; i++ {
		_, err := (&assignNode{n.variable, &intLeaf{i}}).execute()
		if err != nil {
			return nil, err
		}
		_, err = n.body.execute()
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}