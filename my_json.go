package main

import (
	"awesomeProject/util/common"
	"awesomeProject/util/stack"
	"errors"
	"fmt"
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
	f              func(int32, *Handler) error //下一步状态
	stack          *stack.Stack                // 存储{}[]
	currentState   int                         // " 状态
	nextKeyWordIdx int                         // 表示解析到关键字的第几个下标
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
	for idx, c := range s {
		err := handler.f(c, handler)
		if err != nil {
			fmt.Println("err happen after:", s[0:idx], "idx:"+strconv.Itoa(idx))
			return err
		}
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

	switch c {
	case common.BraceLeft:
		// 一定是放在这里
		handler.stack.Push(c)
		handler.f = braceLeft
		return nil
	case common.BracketLeft:
		handler.stack.Push(c)
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
		return nil
	case common.BraceRight:
		pop := handler.stack.Pop()
		if pop == nil || *pop != common.BraceLeft {
			return errors.New("} can not match {")
		}

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

		// 是最外层的{}，最后应该只剩空白符
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
			return nil
		}

		// continue
		return nil
	case ParseValue:
		if c == common.DoubleQuotes {
			handler.f = lookForEndToken
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
		return parseNumSymbol(c, handler)
	}

	switch c {
	// obj
	case common.BraceLeft:
		handler.stack.Push(c)
		handler.f = braceLeft
		return nil

	// arr
	case common.BracketLeft:
		handler.stack.Push(c)
		handler.f = bracketLeft
		return nil

	//string
	case common.DoubleQuotes:
		handler.f = doubleQuotes
		return nil

	// true
	case common.AlpT:
		handler.f = parseTrue
		handler.nextKeyWordIdx = 1
		return nil

	// false
	case common.AlpF:
		handler.f = parseFalse
		handler.nextKeyWordIdx = 1
		return nil

	// null
	case common.AlpN:
		handler.f = parseNull
		handler.nextKeyWordIdx = 1
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
		handler.f = numAfterPoint
		return nil
	}
	return lookForEndToken(c, handler)
}

func parseTrue(c int32, handler *Handler) error {
	if handler.nextKeyWordIdx == 1 && c == common.AlpR {
		handler.nextKeyWordIdx = 2
		return nil
	}
	if handler.nextKeyWordIdx == 2 && c == common.AlpU {
		handler.nextKeyWordIdx = 3
		return nil
	}
	if handler.nextKeyWordIdx == 3 && c == common.AlpE {
		handler.f = lookForEndToken
		return nil
	}
	return errors.New(string(c) + " parse 'true' error")
}

func parseFalse(c int32, handler *Handler) error {
	if handler.nextKeyWordIdx == 1 && c == common.AlpA {
		handler.nextKeyWordIdx = 2
		return nil
	}
	if handler.nextKeyWordIdx == 2 && c == common.AlpL {
		handler.nextKeyWordIdx = 3
		return nil
	}
	if handler.nextKeyWordIdx == 3 && c == common.AlpS {
		handler.nextKeyWordIdx = 4
		return nil
	}
	if handler.nextKeyWordIdx == 4 && c == common.AlpE {
		handler.f = lookForEndToken
		return nil
	}
	return errors.New(string(c) + " parse 'false' error")
}

func parseNull(c int32, handler *Handler) error {
	if handler.nextKeyWordIdx == 1 && c == common.AlpU {
		handler.nextKeyWordIdx = 2
		return nil
	}
	if handler.nextKeyWordIdx == 2 && c == common.AlpL {
		handler.nextKeyWordIdx = 3
		return nil
	}
	if handler.nextKeyWordIdx == 3 && c == common.AlpL {
		handler.f = lookForEndToken
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
		return braceRight(c, handler)
	case common.BracketRight:
		pop := handler.stack.Pop()
		if pop == nil || *pop != common.BracketLeft {
			return errors.New("] can not match [")
		}
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
