package cronjobs

import (
	"errors"
	"fmt"
	"linebot-gemini-pro/db"
	"log"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/markcheno/go-quote"
)

var (
	response      string
	price_history []string
	date_history  []string
)

func ParsePrice(sub_table string, parse_table string, col string) (string, []string, error) {

	companies := GetSubsCompanies(sub_table, col)

	today := time.Now()
	end := today.Format("2006-01-02")
	start := os.Getenv("TimeForStart") // 時間區間不夠就會從該股票起始抓起

	for _, company := range companies {
		stock := strings.ToUpper(company)
		fmt.Printf("===== 收集 %s 歷史股價中 =====\n", stock)
		log.Printf("===== 收集 %s 歷史股價中 =====\n", stock)
		spy, _ := quote.NewQuoteFromYahoo(stock, start, end, quote.Daily, false) // format: 開盤 , 最高, 最低, 收盤, 交易量
		data := spy.CSV()
		var ss *rune
		slice_data := strings.Split(data, "\n")
		for _, s := range slice_data {
			for _, char := range s {
				c := rune(char) // 宣告 unicode裡的類型為 rune
				ss = &c
			}
			if s != "" && !unicode.Is(unicode.Latin, *ss) { // 去除英文字母
				p := strings.Split(s, ",")
				date := p[0]
				price := p[len(p)-2]
				price_history = append(price_history, price)
				date_history = append(date_history, date)
			}
		}
		dict := make(map[string]string)
		for i := 0; i < len(date_history); i++ {
			dict[date_history[i]] = price_history[i]
		}

		if len(spy.Date) == 0 && len(spy.Open) == 0 && len(spy.High) == 0 && len(spy.Low) == 0 && len(spy.Close) == 0 && len(spy.Volume) == 0 {
			return response, companies, errors.New("股票代碼不存在")
		} else {
			// 確認欄位是否存在
			err := db.Controller_CheckColumnExist(parse_table, company)
			if err != nil { //不存在
				log.Printf(" =======不存在欄位 %s, 創建欄位中 =======", company)
				err := db.Controller_AddCloumnWithTable(parse_table, stock)
				if err != nil {
					return err.Error(), companies, err
				}
				log.Printf(" =======欄位 %s已建立完成, 寫入中 =======", company)
				err = db.Controller_AddPrice(parse_table, stock, dict)
				if err != nil {
					return err.Error(), companies, err
				}

			} else { // 存在
				log.Printf(" =======已存在欄位 %s,寫入資料中 =======", company)
				err := db.Controller_AddPrice(parse_table, stock, dict)
				if err != nil {
					return err.Error(), companies, err
				}

			}
		}
		response = fmt.Sprintf("已增加 %s 歷史股價", stock)
	}
	return response, companies, nil
}
