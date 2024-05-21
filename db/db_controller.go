package db

import (
	"database/sql"
	"log"
	"strings"
	"time"

	yahoofinance "github.com/shoenig/yahoo-finance"
)

/*
############## Operations  ##############
*/
func Controller_AddCloumnWithTable(table string, stock string) error {

	err := AddCloumnWithTable(table, stock)
	if err != nil {
		return err
	}
	return nil
}

func Controller_AddPrice(table string, stock string, dict map[string]string) error {

	err := AddPrice(table, stock, dict)
	if err != nil {
		return err
	}
	return nil
}

func Controller_AddPE(table string, stock string, dict map[string]float64) error {
	err := AddPE(table, stock, dict)
	if err != nil {
		return err
	}
	return nil
}

func Controller_AddSubs(table string, column string, stock string) bool {

	err := CheckIsInTable(table, column, stock)
	if err != nil { // error 代表沒找到, 要進行訂閱
		err := AddSubs(table, stock)
		if err != nil {
			return false
		} else {
			return true
		}
	} else { // nil 代表找到,通知已有資料
		return false
	}
}

func Controller_UnSubs(table string, column string, stock string) error {

	err := CheckIsInTable(table, column, stock)
	if err != nil {
		return err
	} else {
		err := DeleteSubs(table, stock)
		if err != nil {
			return err
		}
	}
	return nil
}

func Controller_AddEPS(table string, stock string, date time.Time, price float64) error {

	err := CheckIsInTable2(table, stock)
	if err != nil {
		AddCloumnWithTable(table, "EPS")
	}
	err = AddNewEPS(table, stock, date, price)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func Controller_DelEPS(table string) error {

	err := Controller_CheckTable(table)
	if err != nil {
		return err
	} else {
		enil := DeleteEPS(table)
		if enil != nil {
			return err
		}
		return nil
	}
}

func Controller_AddDividend(table string, stock string, date time.Time, price float64) error {

	err := CheckIsInTable2(table, stock)
	if err != nil {
		AddCloumnWithTable(table, "Dividend")
	}
	err = AddNewDividend(table, stock, date, price)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func Controller_DelDividend(table string) error {

	err := Controller_CheckTable(table)
	if err != nil {
		return err
	} else {
		enil := DeleteDividend(table)
		if enil != nil {
			return err
		}
		return nil
	}
}

/*
############## 共用操作  ##############
*/
func Controller_CheckStocks(match_slice []string) (string, float64, error) { // call yahoo api

	var chart_float float64
	stock := strings.ToUpper(match_slice[0])
	client := yahoofinance.New(nil)
	for _, stock := range match_slice {
		stock = strings.ToUpper(stock)
		log.Println("Controller_CheckStocks:", stock)
		chart, err := client.Lookup(stock)
		if err != nil {
			return stock, 0.0, err
		} else {
			chart_float = chart.Price()
		}
	}
	return stock, chart_float, nil
}

func Controller_CheckTable(table string) error {

	err := CheckTableExist(table)
	if err != nil {
		return err
	}
	return nil
}

func Controller_CreateTable(table string) error {

	err := AddNewTable(table)
	if err != nil {
		return err
	}
	return nil
}

func Controller_GetContnet(params ...interface{}) (*sql.Rows, error) {

	rows, err := GetContent(params...)
	if err != nil {
		return rows, err
	}
	return rows, nil
}

func Controller_QueryDataOrderByDate(col1 string, col2 string, table string, order_col string, number int) (*sql.Rows, error) {

	rows, err := QueryDataOrderByDate(col1, col2, table, order_col, number)
	if err != nil {
		return rows, err
	}
	return rows, nil
}

func Controller_CheckColumnExist(table string, col string) error {
	err := CheckColumnExist(table, col)
	if err != nil {
		return err //不存在
	}
	return nil //存在
}
