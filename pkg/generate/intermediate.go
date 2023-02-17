package generate

import (
	"encoding/json"
	"fmt"
)

type Node any

type propNode struct {
	Name  string
	Value any
}

type objNode struct {
	Properties []*propNode
}

type listNode struct {
	Items []any
}

func toString(node Node) string {
	b, _ := json.Marshal(node)
	return string(b)
}

type TokenParser struct {
	stack 	*Stack[Node]
	current Node
}

func MakeTokenParser() *TokenParser {
	return &TokenParser{stack: MakeStack[Node]()}
}

// add a new object to the current object
// push the current object down
// add the new object to the head
func (this *TokenParser) push(node Node) error {
	switch current := this.current.(type) {
	case *listNode:
		current.Items = append(current.Items, node)
	case *propNode:
		current.Value = node
	case *objNode:
		switch node := node.(type) {
		case *propNode:
			current.Properties = append(current.Properties, node)
		default:
			return fmt.Errorf("cannot add non-property to an object: trying to add %#v to %#v", node, current)
		}
	}
	this.stack.Push(this.current)
	this.current = node
	return nil
}

// pop the current node off the stack and return it
// if the node is a Property, pop to its Object
func (this *TokenParser) pop() Node {
	node := this.current
	parent, _ := this.stack.Pop()
	switch parent := parent.(type) {
	case *propNode:
		return this.pop()
	default:
		this.current = parent
	}
	return node
}

func (this *TokenParser) addValue(value any) error {
	switch current := this.current.(type) {
	case *propNode:
		current.Value = value
		this.pop()
	case *listNode:
		current.Items = append(current.Items, value)
	case *objNode:
		return fmt.Errorf("cannot add non-property to an object: trying to add %#v to %#v", value, current)
	}
	return nil
}

func (this *TokenParser) Parse(tokens []token) (Node, error) {
	var lastClosed Node
	for i := range tokens {
		switch token := tokens[i].(type) {
		case jsonArrOpen:
			node := &listNode{Items: []any{}}
			if err := this.push(node); err != nil {
				return nil, err
			}

		case jsonObjOpen:
			node := &objNode{Properties: []*propNode{}}
			if err := this.push(node); err != nil {
				return nil, err
			}

		case jsonArrClose, jsonObjClose:
			lastClosed = this.pop()

		case jsonString:
			switch this.current.(type) {
			case *objNode:
				node := &propNode{Name: token.value}
				if err := this.push(node); err != nil {
					return nil, err
				}
			default:
				if err := this.addValue(token.value); err != nil {
					return nil, err
				}
			}

		case jsonValue:
			if err := this.addValue(token.Value()); err != nil {
				return nil, err
			}
		}
	}
	return lastClosed, nil
}



