package util

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// 移除字符串左侧所有空白字符，如果有空格' '与字符相邻，则保留这些空格
// 如:
//     (1) "   \r\n       \n       \n   中国"    处理后为   "   中国"
//     (2) "   \r\n       \n       \n中国"       处理后为   "中国"
func TrimLeftSpaceKeep(s string) string {
	start := 0
	spaceNum := 0
	for start < len(s) {
		wid := 1
		r := rune(s[start])
		if r >= utf8.RuneSelf {
			r, wid = utf8.DecodeRuneInString(s[start:])
		}

		if unicode.IsSpace(r) == false {
			start -= spaceNum
			return s[start:]
		}

		if r == ' ' {
			spaceNum += 1
		} else {
			spaceNum = 0
		}

		start += wid
	}

	return s
}

// 移除字符串右侧所有空白字符
func TrimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}
