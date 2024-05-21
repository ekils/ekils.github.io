package operations

import (
	"encoding/json"
	"fmt"
	"linebot-gemini-pro/db"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	response string
)

type CollectEPSResponse struct {
	Message string `json:"message"`
}

type CollectBVPSResponse struct {
	Message string `json:"message"`
}

type Dividend struct {
	date     time.Time
	Dividend float64
}

func CollectEPS(w http.ResponseWriter, r *http.Request) {

	layout := "1/2/2006"

	match_slice := SharedArchitecture2(r)
	stock := match_slice[0]
	stock = strings.ToUpper(stock)

	_, _, content := db.Controller_CheckStocks([]string{stock})
	str := strings.Join(match_slice[1:], " ")

	if content != nil {
		response = "股票代碼不存在"
	} else {
		// 移除不必要雜字
		content_split := strings.FieldsFunc(str, func(r rune) bool {
			return unicode.IsSpace(r) || r == '	' || r == '\n'
		})
		// 检查拆分後的數據長度是否為偶數
		if len(content_split)%2 != 0 {
			response = "輸入 EPS 格式錯誤"
			log.Println(response)
		} else {
			// 拆2個list, 然後轉 map
			list_key := make([]interface{}, 0)
			list_value := make([]interface{}, 0)

			for i, v := range content_split {
				if i%2 == 0 {
					list_key = append(list_key, v)
				} else {
					list_value = append(list_value, v)
				}
			}
			// 檢查 list是否相等
			if len(list_key) != len(list_value) {
				response = "輸入 EPS 格式錯誤"
				log.Println(response)
			} else {
				content_map := make(map[interface{}]interface{})
				for i := 0; i < len(list_key); i++ {
					content_map[list_key[i]] = list_value[i]
				}
				// 再次移除雜訊 $
				for key, value := range content_map {
					strValue, ok := value.(string)
					if !ok {
						continue
					}
					content_map[key] = strings.Replace(strValue, "$", "", -1)
					log.Println("content_map:", content_map)
				}
				for date, price := range content_map {
					dateStr, _ := date.(string)
					priceStr, _ := price.(string)
					// 將日期字符串解析為 timestamp
					t, err := time.Parse(layout, dateStr)
					if err != nil {
						response = fmt.Sprint("日期解析失败:", err)
					} else {
						// 將價格字符串轉換為 float
						price, _ := strconv.ParseFloat(priceStr, 64)

						//確認是否有該table：
						table := stock + "_EPS"
						table = strings.ToUpper(table)
						err := db.Controller_CheckTable(table)
						if err != nil { //代表沒有該table, 要create, 再寫入
							fmt.Println("😍.....")
							err := db.Controller_CreateTable(table)
							if err != nil {
								response = err.Error()
							}
						}
						err = db.Controller_AddEPS(table, stock, t, price)
						if err != nil {
							response = err.Error()
						}
						response = fmt.Sprintf("已增加 %s EPS 資料", stock)

					}
				}
			}
		}
		response_eps := SubscriptionResponse{Message: response}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response_eps)
	}
}

func DeleteEPS(w http.ResponseWriter, r *http.Request) {

	match_slice := SharedArchitecture2(r)

	stock := match_slice[0]
	stock = strings.ToUpper(stock)
	_, _, content := db.Controller_CheckStocks([]string{stock})
	if content != nil {
		response = "股票代碼不存在"
	} else {
		table := stock + "_EPS"
		table = strings.ToUpper(table)
		err := db.Controller_DelEPS(table)
		if err != nil {
			response = err.Error()
		} else {
			response = fmt.Sprintf("已移除 %s EPS 資料", stock)
		}
	}
	response_eps := SubscriptionResponse{Message: response}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response_eps)
}

func AddDividend(w http.ResponseWriter, r *http.Request) {

	layout := "2006-1-2"

	match_slice := SharedArchitecture2(r)
	stock := match_slice[0]
	stock = strings.ToUpper(stock)

	_, _, content := db.Controller_CheckStocks([]string{stock})
	str := strings.Join(match_slice[1:], " ")

	if content != nil {
		response = "股票代碼不存在"
	} else {
		// 移除不必要雜字
		content_split := strings.FieldsFunc(str, func(r rune) bool {
			return unicode.IsSpace(r) || r == '	' || r == '\n'
		})
		// 检查拆分後的數據長度是否為偶數
		if len(content_split)%2 != 0 {
			response = "輸入 Dividend 格式錯誤"
			log.Println(response)
		} else {
			// 拆2個list, 然後轉 map
			list_key := make([]interface{}, 0)
			list_value := make([]interface{}, 0)

			for i, v := range content_split {
				if i%2 == 0 {
					list_key = append(list_key, v)
				} else {
					list_value = append(list_value, v)
				}
			}
			// 檢查 list是否相等
			if len(list_key) != len(list_value) {
				response = "輸入 Dividend 格式錯誤"
				log.Println(response)
			} else {
				content_map := make(map[interface{}]interface{})
				for i := 0; i < len(list_key); i++ {
					content_map[list_key[i]] = list_value[i]
				}
				// 再次移除雜訊 $
				for key, value := range content_map {
					strValue, ok := value.(string)
					if !ok {
						continue
					}
					content_map[key] = strings.Replace(strValue, "$", "", -1)
					log.Println("content_map:", content_map)
				}
				for date, price := range content_map {
					dateStr, _ := date.(string)
					priceStr, _ := price.(string)
					// 將日期字符串解析為 timestamp
					t, err := time.Parse(layout, dateStr)
					if err != nil {
						response = fmt.Sprint("日期解析失败:", err)
					} else {
						// 將價格字符串轉換為 float
						price, _ := strconv.ParseFloat(priceStr, 64)
						log.Println(price)
						//確認是否有該table：
						table := stock + "_Dividend"
						table = strings.ToUpper(table)
						err := db.Controller_CheckTable(table)
						if err != nil { //代表沒有該table, 要create, 再寫入
							err := db.Controller_CreateTable(table)
							if err != nil {
								response = err.Error()
							}
						}

						db.Controller_AddSubs("NYSEDDM", "Company", stock)
						err = db.Controller_AddDividend(table, stock, t, price)
						if err != nil {
							response = err.Error()
						}
						response = fmt.Sprintf("已增加 %s 的 Dividend 資料", stock)

					}
				}
			}
		}
		response_dividend := SubscriptionResponse{Message: response}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response_dividend)
	}
}

func DelDividend(w http.ResponseWriter, r *http.Request) {
	match_slice := SharedArchitecture2(r)

	stock := match_slice[0]
	stock = strings.ToUpper(stock)
	_, _, content := db.Controller_CheckStocks([]string{stock})
	if content != nil {
		response = "股票代碼不存在"
	} else {
		table := stock + "_DIVIDEND"
		table = strings.ToUpper(table)
		db.Controller_UnSubs("NYSEDDM", "Company", stock)
		err := db.Controller_DelDividend(table)
		if err != nil {
			response = err.Error()
		} else {
			response = fmt.Sprintf("已移除 %s 的 Dividend 資料", stock)
		}
	}
	response_dividend := SubscriptionResponse{Message: response}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response_dividend)
}

func QueryDDM(w http.ResponseWriter, r *http.Request) { //估值用

	match_slice := SharedArchitecture2(r)
	stock := match_slice[0]
	stock = strings.ToUpper(stock)
	table_name := stock + "_DIVIDEND"

	rows, content := db.Controller_QueryDataOrderByDate(`date`, `Dividend`, table_name, `date`, 24)

	if content != nil {
		qr := "QueryDataOrderByDate 有問題"
		response := SubscriptionResponse{Message: qr}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		// 創建一個用於存儲 Dividend 的切片
		dividends := make([]Dividend, 0)

		// 遍歷查詢結果，並將每一行轉換為 Dividend 類型的結構，然後添加到切片中
		for rows.Next() {
			var dividend Dividend
			if err := rows.Scan(&dividend.date, &dividend.Dividend); err != nil {
				log.Fatal(err)
			}
			dividends = append(dividends, dividend) // lists
		}

		var ddm_data []float64
		for _, dividend := range dividends {
			// log.Printf("Timestamp: %s, Price: %f\n", dividend.date, dividend.Dividend)
			ddm_data = append(ddm_data, dividend.Dividend)
		}

		FV_str, FV := GetReasonablePrice(stock, ddm_data)

		FV_data := strconv.FormatFloat(FV, 'f', 2, 64) //小數點後2位
		response := SubscriptionResponse{Message: stock + " " + FV_str + ": " + FV_data}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

}

func DDMQuestion(w http.ResponseWriter, r *http.Request) { //查詢資料庫裡有哪些股息公司
	rows, _ := db.Controller_GetContnet("NYSEDDM", "Company")
	eturn_string := ReformatResponseSubs(rows)
	fianl_response := eturn_string

	response := SubscriptionResponse{Message: fianl_response}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CollectBVPS(match_slice []string) {
}
