/*
 * Copyright © 2019 Hedzr Yeh.
 */

package vxconf

import (
	"regexp"
	"strconv"
	"unicode/utf8"
)

// var reFind = regexp.MustCompile(`^\s*[^\s\:]+\:\s*["']?.*\\u.*["']?\s*$`)
var reFind = regexp.MustCompile(`[^\s\:]+\:\s*["']?.*\\u.*["']?`)
var reFindU = regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)

func expandUnicodeInYamlLine(line []byte) []byte {
	// TODO: restrict this to the quoted string value
	return reFindU.ReplaceAllFunc(line, expandUnicodeRune)
}

func expandUnicodeRune(esc []byte) []byte {
	ri, _ := strconv.ParseInt(string(esc[2:]), 16, 32)
	r := rune(ri)
	repr := make([]byte, utf8.RuneLen(r))
	utf8.EncodeRune(repr, r)
	return repr
}

// UnescapeUnicode 解码 \uxxxx 为 unicode 字符; 但是输入的 b 应该是 yaml 格式
func UnescapeUnicode(b []byte) string {
	b = reFind.ReplaceAllFunc(b, expandUnicodeInYamlLine)
	return string(b)
}
