package common

import "strconv"

const (
	BraceLeft    = 123 // {		=> " or }		入栈
	BraceRight   = 125 // }		=> *end_token*	出栈
	BracketLeft  = 91  // [		=> *val* or ]	入栈
	BracketRight = 93  // ]		=> *end_token*	出栈
	DoubleQuotes = 34  // "		=> " in string or : or *end_token*
	Comma        = 44  // ,		=> " in obj or *val* in arr
	Colon        = 58  // :		=> *val*

	//  num 48~57
	Sub   = 45 // - 	=> num
	Point = 46 // . 	=> num

	Zero = 48
	One  = 49
	Nine = 57

	// 空白符
	HT    = 9
	LF    = 10
	VT    = 11
	FF    = 12
	CR    = 13
	Space = 32

	// 关键字
	AlpA = 97  // l
	AlpE = 101 // [end_token]
	AlpF = 102 // a
	AlpL = 108 // l or s or [end_token]
	AlpN = 110 // u
	AlpR = 114 // u
	AlpS = 115 // e
	AlpT = 116 // r
	AlpU = 117 // e or l
)

func IsBlank(c int32) bool {
	return c >= HT && c <= CR || c == Space
}

func IsNumBeginning(c int32) bool {
	return c >= Zero && c <= Nine || c == Sub
}

func IsNum(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
