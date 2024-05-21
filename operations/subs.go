package operations

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"linebot-gemini-pro/db"
	"log"

	"net/http"
)

type SubscriptionResponse struct {
	Message string `json:"message"`
}

func Question(w http.ResponseWriter, r *http.Request) {
	rows, _ := db.Controller_GetContnet("NYSE", "Company")
	eturn_string := ReformatResponseSubs(rows)
	fianl_response := response + "\n" + eturn_string

	response := SubscriptionResponse{Message: fianl_response}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func Subscription(w http.ResponseWriter, r *http.Request) {

	match_slice := SharedArchitecture(r)
	stock, _, return_str := db.Controller_CheckStocks(match_slice)
	log.Println("STOCK!!!!:", stock)
	if return_str != nil {
		response = "股票代碼不存在"
	} else {

		boolin := db.Controller_AddSubs("NYSE", "Company", stock)
		if boolin {
			response = fmt.Sprintf("已訂閱 %s,並請記得更新該股票的EPS", stock)
		} else {
			response = "資料庫裡已存在"
		}
	}
	//回傳目前有哪些訂閱內容:
	rows, _ := db.Controller_GetContnet("NYSE", "Company")
	eturn_string := ReformatResponseSubs(rows)
	fianl_response := response + "\n" + eturn_string

	response := SubscriptionResponse{Message: fianl_response}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UnSubscription(w http.ResponseWriter, r *http.Request) {

	match_slice := SharedArchitecture(r)
	stock, _, return_str := db.Controller_CheckStocks(match_slice)
	if return_str != nil {
		response = "股票代碼不存在"
	} else {
		err := db.Controller_UnSubs("NYSE", "Company", stock)
		if err != nil {
			response = err.Error()
		}
		response = fmt.Sprintf("已取消訂閱 %s", stock)
	}

	//回傳目前有哪些訂閱內容:
	rows, _ := db.Controller_GetContnet("NYSE", "Company")
	eturn_string := ReformatResponseSubs(rows)
	fianl_response := response + "\n" + eturn_string

	response := SubscriptionResponse{Message: fianl_response}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func ReformatResponseSubs(rows *sql.Rows) string {

	type RowData struct {
		SubStock string
	}
	var rowData RowData
	var resultList []RowData
	eturn_string := "[目前資料庫裡已有訂閱]:\n"

	for rows.Next() {
		if err := rows.Scan(&rowData.SubStock); err != nil {
			response = err.Error()
		}
		resultList = append(resultList, rowData)
	}
	rows.Close()

	for _, row := range resultList {
		eturn_string += fmt.Sprintf("%s ,", row.SubStock)
	}
	eturn_string = eturn_string[:len(eturn_string)-1]
	return eturn_string
}
