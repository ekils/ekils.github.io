package cronjobs

import (
	"fmt"
	"linebot-gemini-pro/db"
	"log"
	"os"
	"strconv"
	"time"
)

var eps_current_date time.Time
var eps_last_date time.Time
var eps_current_EPS float64
var PE float64
var dict = make(map[string]float64)

func GenPE_Ratio(reuslt_map_price map[string][]Price_RowData, reuslt_map_eps map[string][]EPS_RowData) (string, error) {

	file, _ := os.Create("logs/Write-to-PE.log")
	log.SetOutput(file)

	for company, eps_struct := range reuslt_map_eps {
		dict = Handle_EPS_Logic(company, eps_struct, reuslt_map_price)

		_, err := Handle_EPS_Column(company, dict)

		if err != nil {
			return err.Error(), err
		}
	}
	return "ok", nil
}

func Handle_EPS_Logic(company string, eps_struct []EPS_RowData, reuslt_map_price map[string][]Price_RowData) map[string]float64 {

	log.Println("In Logic Company:", company)

	for index := len(eps_struct) - 1; index >= 0; index-- {

		if index == len(eps_struct)-1 {
			eps_current_date = eps_struct[index].date //最新一期 eps 日期
			eps_current_EPS = eps_struct[index].EPS
			log.Printf("-----EPS區間 (最新) %s ; %f ------:", eps_current_date.Format("2006-01-02 15:04:05"), eps_current_EPS)

			for _, rowData := range reuslt_map_price[company] {
				if rowData.date.After(eps_current_date) {
					PE = To_PE_Ratio(eps_current_EPS, rowData.StockPrice)
					// log.Println(rowData.date.Format("2006-01-02 15:04:05"), rowData.StockPrice, PE)
					dict[rowData.date.Format("2006-01-02 15:04:05")] = PE
				}
			}
		} else {
			eps_current_date = eps_struct[index].date //最新一期 eps 日期
			eps_last_date = eps_struct[index+1].date  //上一期 eps 日期
			eps_current_EPS = eps_struct[index].EPS
			log.Printf("-----EPS區間 %s ;  %f ------:", eps_current_date.Format("2006-01-02 15:04:05"), eps_current_EPS)
			log.Println("前期 EPS 區間:", eps_last_date)
			for _, rowData := range reuslt_map_price[company] {
				if rowData.date.After(eps_current_date) && rowData.date.Before(eps_last_date) {
					PE = To_PE_Ratio(eps_current_EPS, rowData.StockPrice)
					// log.Println(rowData.date.Format("2006-01-02 15:04:05"), rowData.StockPrice, PE)
					dict[rowData.date.Format("2006-01-02 15:04:05")] = PE
				}
			}
		}
	}
	return dict
}

func To_PE_Ratio(eps_current_EPS float64, StockPrice float64) float64 {

	result := StockPrice / eps_current_EPS
	PE, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", result), 64)
	return PE

}

func Handle_EPS_Column(company string, dict map[string]float64) (string, error) {

	table := company + "_PE"
	// 確認欄位是否存在
	err := db.Controller_CheckColumnExist(table, company)
	if err != nil { //不存在
		log.Printf(" =======不存在 Table %s, 創建Table中 =======", table)
		err := db.Controller_CreateTable(table)
		if err != nil {
			return err.Error(), err
		}
		err = db.Controller_AddCloumnWithTable(table, company)
		if err != nil {
			return err.Error(), err
		}

		log.Printf(" =======Table %s 與欄位已建立完成, 寫入中 =======", table)

		err = db.Controller_AddPE(table, company, dict)
		if err != nil {
			return err.Error(), err
		}

	} else { // 存在
		log.Printf(" =======已存在 Table %s,寫入資料中 =======", table)
		err := db.Controller_AddPE(table, company, dict)
		if err != nil {
			return err.Error(), err
		}

	}
	return "ok", nil
}
