package main

import (
	"awesomeProject/util/common"
	"awesomeProject/util/stack"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// state
const (
	Unknown    = 0
	ParseKey   = 1
	ParseValue = 2
)

type Handler struct {
	data         string
	f            func(int32, *Handler) error //下一步状态
	stack        *stack.Stack                // 存储{}[]
	currentState int                         // " 状态
	startIdx     int
	currentIdx   int
	root         *Node
	currentNode  *Node
}

type Node struct {
	parent  *Node
	kind    reflect.Kind
	keyName string

	mapVal map[string]any
	arrVal *[]any
}

func NewHandler() *Handler {
	return &Handler{
		f:     start,
		stack: stack.New(),
	}
}

func CheckValid(s string) error {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return errors.New("input is empty")
	}

	handler := NewHandler()
	handler.data = s
	for idx, c := range s {
		handler.currentIdx = idx
		err := handler.f(c, handler)
		if err != nil {
			fmt.Println("err happen after:", s[0:idx], "idx:"+strconv.Itoa(idx))
			return err
		}
	}

	// TODO test
	root := handler.root
	if root.kind == reflect.Map {
		res, _ := json.Marshal(root.mapVal)
		fmt.Println(string(res))
	} else {
		res, _ := json.Marshal(root.arrVal)
		fmt.Println(string(res))
	}

	return nil
}

/*
开始
*/
func start(c int32, handler *Handler) error {
	if common.IsBlank(c) {
		return nil
	}

	node := &Node{}
	handler.root = node
	handler.currentNode = node

	switch c {
	case common.BraceLeft:
		// 一定是放在这里
		handler.stack.Push(c)
		node.kind = reflect.Map
		node.mapVal = map[string]any{}
		handler.f = braceLeft
		return nil
	case common.BracketLeft:
		handler.stack.Push(c)
		node.kind = reflect.Array
		node.arrVal = &[]any{}
		handler.f = bracketLeft
		return nil
	default:
		return errors.New(string(c) + " should be { or [")
	}
}

/*
{
*/
func braceLeft(c int32, handler *Handler) error {
	if common.IsBlank(c) {
		return nil
	}

	switch c {
	case common.DoubleQuotes:
		handler.currentState = ParseKey
		handler.f = doubleQuotes
		handler.startIdx = handler.currentIdx
		return nil
	case common.BraceRight:
		pop := handler.stack.Pop()
		if pop == nil || *pop != common.BraceLeft {
			return errors.New("} can not match {")
		}

		// 跳出嵌套
		preNode := handler.currentNode.parent
		handler.currentNode = preNode

		// 是最外层的{}，最后应该只剩空白符
		if handler.stack.IsEmpty() {
			handler.f = blankOnly
			return nil
		}

		return braceRight(c, handler)
	default:
		return errors.New(string(c) + " is invalid")
	}
}

/*
}
*/
func braceRight(c int32, handler *Handler) error {
	handler.f = lookForEndToken
	return nil
}

/*
[
*/
func bracketLeft(c int32, handler *Handler) error {
	if common.IsBlank(c) {
		return nil
	}

	if c == common.BracketRight {
		pop := handler.stack.Pop()
		if pop == nil || *pop != common.BracketLeft {
			return errors.New("] can not match [")
		}

		// 跳出嵌套
		preNode := handler.currentNode.parent
		handler.currentNode = preNode

		// 是最外层的[]，最后应该只剩空白符
		if handler.stack.IsEmpty() {
			handler.f = blankOnly
			return nil
		}
		return bracketRight(c, handler)
	}

	return parseValue(c, handler)
}

/*
]
*/
func bracketRight(c int32, handler *Handler) error {
	handler.f = lookForEndToken
	return nil
}

/*
,
*/
func comma(c int32, handler *Handler) error {
	if common.IsBlank(c) {
		return nil
	}

	peek := handler.stack.Peek()
	if peek == nil {
		return errors.New(", should within obj or arr")
	}

	// 在obj中。寻找 "
	if *peek == common.BraceLeft {
		handler.f = doubleQuotes
		handler.currentState = ParseKey
		handler.startIdx = handler.currentIdx
		return nil
	}

	// 在arr中。寻找*val*
	return parseValue(c, handler)
}

/*
" 允许key是空的
*/
func doubleQuotes(c int32, handler *Handler) error {
	switch handler.currentState {
	case ParseKey:
		if c == common.DoubleQuotes {
			handler.f = lookForColon
			handler.currentState = ParseValue

			key := handler.data[handler.startIdx+1 : handler.currentIdx]
			handler.currentNode.keyName = key
			return nil
		}

		// continue
		return nil
	case ParseValue:
		if c == common.DoubleQuotes {
			handler.f = lookForEndToken
			val := handler.data[handler.startIdx+1 : handler.currentIdx]
			setValue(handler, val)
			return nil
		}

		// continue
		return nil
	default:
		return errors.New(string(c) + " is invalid")
	}
}

/*
:
*/
func lookForColon(c int32, handler *Handler) error {
	if common.IsBlank(c) {
		return nil
	}

	if c == common.Colon {
		handler.f = parseValue
		return nil
	}

	return errors.New(string(c) + " is invalid, ':' instead")
}

/*
解析value
*/
func parseValue(c int32, handler *Handler) error {
	if common.IsBlank(c) {
		return nil
	}

	handler.currentState = ParseValue

	if common.IsNumBeginning(c) {
		handler.startIdx = handler.currentIdx
		return parseNumSymbol(c, handler)
	}

	switch c {
	// obj
	case common.BraceLeft:
		handler.stack.Push(c)
		handler.f = braceLeft

		m := map[string]any{}
		setValue(handler, m)
		preNode := handler.currentNode
		handler.currentNode = &Node{
			parent: preNode,
			kind:   reflect.Map,
			mapVal: m,
		}
		return nil

	// arr
	case common.BracketLeft:
		handler.stack.Push(c)
		handler.f = bracketLeft

		arr := &[]any{}
		setValue(handler, arr)
		preNode := handler.currentNode
		handler.currentNode = &Node{
			parent: preNode,
			kind:   reflect.Array,
			arrVal: arr,
		}

		return nil

	//string
	case common.DoubleQuotes:
		handler.f = doubleQuotes
		handler.startIdx = handler.currentIdx
		return nil

	// true
	case common.AlpT:
		handler.f = parseTrue
		handler.startIdx = handler.currentIdx
		return nil

	// false
	case common.AlpF:
		handler.f = parseFalse
		handler.startIdx = handler.currentIdx
		return nil

	// null
	case common.AlpN:
		handler.f = parseNull
		handler.startIdx = handler.currentIdx
		return nil
	default:
		return errors.New(string(c) + " is invalid for the beginning of value")
	}
}

/*
合法的num：没有加号。负号与数字之间没有空白符。小数点的前后一定要有数字
*/
func parseNumSymbol(c int32, handler *Handler) error {
	if c == common.Sub {
		handler.f = parseNumSymbol
		return nil
	}

	if c == common.Zero {
		handler.f = parseNum0
		return nil
	}

	if c >= common.One && c <= common.Nine {
		handler.f = parseNum19
		return nil
	}

	return errors.New(string(c) + " is invalid for a number")
}

func parseNum0(c int32, handler *Handler) error {
	if c == common.Point {
		handler.f = point
		return nil
	}

	i := handler.data[handler.currentIdx-1 : handler.currentIdx]
	if common.IsNum(i) {
		val := handler.data[handler.startIdx:handler.currentIdx]
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		setValue(handler, f)
	}

	return lookForEndToken(c, handler)
}

func parseNum19(c int32, handler *Handler) error {
	if c >= common.Zero && c <= common.Nine {
		handler.f = parseNum19
		return nil
	}

	return parseNum0(c, handler)
}

func point(c int32, handler *Handler) error {
	if c >= common.Zero && c <= common.Nine {
		handler.f = numAfterPoint
		return nil
	}

	return errors.New(string(c) + " is invalid for a number")
}

func numAfterPoint(c int32, handler *Handler) error {
	if c >= common.Zero && c <= common.Nine {
		return nil
	}

	i := handler.data[handler.currentIdx-1 : handler.currentIdx]
	if common.IsNum(i) {
		val := handler.data[handler.startIdx:handler.currentIdx]
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		setValue(handler, f)
	}

	return lookForEndToken(c, handler)
}

func parseTrue(c int32, handler *Handler) error {
	if handler.currentIdx-handler.startIdx == 1 && c == common.AlpR {
		return nil
	}
	if handler.currentIdx-handler.startIdx == 2 && c == common.AlpU {
		return nil
	}
	if handler.currentIdx-handler.startIdx == 3 && c == common.AlpE {
		setValue(handler, true)
		handler.f = lookForEndToken
		return nil
	}
	return errors.New(string(c) + " parse 'true' error")
}

func parseFalse(c int32, handler *Handler) error {
	if handler.currentIdx-handler.startIdx == 1 && c == common.AlpA {
		return nil
	}
	if handler.currentIdx-handler.startIdx == 2 && c == common.AlpL {
		return nil
	}
	if handler.currentIdx-handler.startIdx == 3 && c == common.AlpS {
		return nil
	}
	if handler.currentIdx-handler.startIdx == 4 && c == common.AlpE {
		setValue(handler, false)
		handler.f = lookForEndToken
		return nil
	}
	return errors.New(string(c) + " parse 'false' error")
}

func parseNull(c int32, handler *Handler) error {
	if handler.currentIdx-handler.startIdx == 1 && c == common.AlpU {
		return nil
	}
	if handler.currentIdx-handler.startIdx == 2 && c == common.AlpL {
		return nil
	}
	if handler.currentIdx-handler.startIdx == 3 && c == common.AlpL {
		handler.f = lookForEndToken
		setValue(handler, nil)
		return nil
	}
	return errors.New(string(c) + " parse 'null' error")
}

/*
, or } or ]
*/
func lookForEndToken(c int32, handler *Handler) error {
	if common.IsBlank(c) {
		return nil
	}

	switch c {
	case common.BraceRight:
		pop := handler.stack.Pop()
		if pop == nil || *pop != common.BraceLeft {
			return errors.New("} can not match {")
		}
		preNode := handler.currentNode.parent
		handler.currentNode = preNode

		return braceRight(c, handler)
	case common.BracketRight:
		pop := handler.stack.Pop()
		if pop == nil || *pop != common.BracketLeft {
			return errors.New("] can not match [")
		}
		preNode := handler.currentNode.parent
		handler.currentNode = preNode
		return bracketRight(c, handler)
	case common.Comma:
		handler.f = comma
		return nil
	default:
		return errors.New(string(c) + " is invalid")
	}
}

/*
只剩空白符，防止如 {}xxx
*/
func blankOnly(c int32, handler *Handler) error {
	if !common.IsBlank(c) {
		return errors.New(string(c) + " should be empty")
	}
	return nil
}

func setValue(handler *Handler, val any) {
	if handler.currentNode.kind == reflect.Map {
		handler.currentNode.mapVal[handler.currentNode.keyName] = val
	} else if handler.currentNode.kind == reflect.Array {
		*handler.currentNode.arrVal = append(*handler.currentNode.arrVal, val)
	}
}
