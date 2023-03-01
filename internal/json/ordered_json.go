package json

import (
	"encoding/json"
	"fmt"

	"github.com/erik-overdahl/tabs_server/internal/util"
)

type JSON any

type KeyValue struct {
	Key   string
	Value any
}

type Object struct {
	Items []*KeyValue
}

type List struct {
	Items []any
}

func toString(node JSON) string {
	b, _ := json.Marshal(node)
	return string(b)
}

type TokenParser struct {
	stack   *util.Stack[JSON]
	current JSON
}

func MakeTokenParser() *TokenParser {
	return &TokenParser{stack: util.MakeStack[JSON]()}
}

// add a new object to the current object
// push the current object down
// add the new object to the head
func (this *TokenParser) push(node JSON) error {
	switch current := this.current.(type) {
	case *List:
		current.Items = append(current.Items, node)
	case *KeyValue:
		current.Value = node
	case *Object:
		switch node := node.(type) {
		case *KeyValue:
			current.Items = append(current.Items, node)
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
func (this *TokenParser) pop() JSON {
	node := this.current
	parent, _ := this.stack.Pop()
	switch parent := parent.(type) {
	case *KeyValue:
		return this.pop()
	default:
		this.current = parent
	}
	return node
}

func (this *TokenParser) addValue(value any) error {
	switch current := this.current.(type) {
	case *KeyValue:
		current.Value = value
		this.pop()
	case *List:
		current.Items = append(current.Items, value)
	case *Object:
		return fmt.Errorf("cannot add non-property to an object: trying to add %#v to %#v", value, current)
	}
	return nil
}

func (this *TokenParser) Parse(tokens []token) (JSON, error) {
	var lastClosed JSON
	for i := range tokens {
		switch token := tokens[i].(type) {
		case jsonArrOpen:
			node := &List{Items: []any{}}
			if err := this.push(node); err != nil {
				return nil, err
			}

		case jsonObjOpen:
			node := &Object{Items: []*KeyValue{}}
			if err := this.push(node); err != nil {
				return nil, err
			}

		case jsonArrClose, jsonObjClose:
			lastClosed = this.pop()

		case jsonString:
			switch this.current.(type) {
			case *Object:
				node := &KeyValue{Key: token.value}
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
