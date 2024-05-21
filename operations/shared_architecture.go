package operations

import (
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode"
)

func SharedArchitecture(r *http.Request) []string { //僅拆股票出來
	s, _ := io.ReadAll(r.Body)
	var byte2string = string(s)
	byte2string_temp := strings.FieldsFunc(byte2string, func(r rune) bool {
		return unicode.IsSpace(r) || r == '{' || r == '}' || r == ':'
	})

	byte2string = strings.Join(byte2string_temp, "")

	re := regexp.MustCompile("[a-zA-Z]+")
	match_slice := re.FindAllString(byte2string, -1)

	return match_slice

}

func SharedArchitecture2(r *http.Request) []string { //多拆eps bvps資料
	s, _ := io.ReadAll(r.Body)
	var byte2string = string(s)
	byte2string_temp := strings.FieldsFunc(byte2string, func(r rune) bool {
		return unicode.IsSpace(r) || r == '{' || r == '}' || r == ':'
	})
	return byte2string_temp
}
