package operations

import (
	"encoding/json"
	"fmt"
	"io"
	"linebot-gemini-pro/db"
	"net/http"
	"regexp"
	"strings"
	"unicode"
)

type InquiryResponsse struct {
	Message string `json:"message"`
	Price   string `json:"price"`
}

func Inguiry(w http.ResponseWriter, r *http.Request) {
	s, _ := io.ReadAll(r.Body)
	var byte2string = string(s)
	var priceString string

	byte2string_temp := strings.FieldsFunc(byte2string, func(r rune) bool {
		return unicode.IsSpace(r) || r == '{' || r == '}' || r == ':'
	})

	byte2string = strings.Join(byte2string_temp, "")

	re := regexp.MustCompile("[a-zA-Z]+")
	match_slice := re.FindAllString(byte2string, -1)

	stock, chart, return_str := db.Controller_CheckStocks(match_slice)
	if return_str != nil {
		response = "股票代碼不存在"
	} else {
		priceString = fmt.Sprintf("%.2f", chart)
	}

	response := InquiryResponsse{Message: fmt.Sprintf("查詢: %s", stock)}
	response.Price = priceString

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
