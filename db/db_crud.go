package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"
)

/*
############## Operations  ##############
*/
func AddNewTable(table string) error {
	// SqlScript := fmt.Sprintf("CREATE TABLE %s (date TIMESTAMP,PRIMARY KEY (date));", table)
	SqlScript := fmt.Sprintf("CREATE TABLE \"%s\" (\"date\" TIMESTAMP PRIMARY KEY);", table)

	s, err := dbConn.Prepare(SqlScript)
	if err != nil {
		return err
	}
	_, err = s.Exec()
	if err != nil {
		return err
	}
	s.Close()
	return nil
}

func AddCloumnWithTable(table string, column_name string) error {

	// var SqlScript = fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` FLOAT;", table, column_name)
	var SqlScript = fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN \"%s\" FLOAT;", table, column_name)
	s, err := dbConn.Prepare(SqlScript)
	if err != nil {
		return err
	}
	_, err = s.Exec()
	if err != nil {
		return err
	}
	s.Close()
	return nil
}

func AddPrice(table string, stock string, dictionary map[string]string) error {

	// 获取并排序键（日期）
	var dates []string
	for date := range dictionary {
		dates = append(dates, date)
	}

	sort.Slice(dates, func(i, j int) bool {
		timeI, _ := time.Parse("2006-01-02 15:04", dates[i])
		timeJ, _ := time.Parse("2006-01-02 15:04", dates[j])
		return timeI.After(timeJ)
	})

	for _, dateStr := range dates {
		priceStr := dictionary[dateStr]
		// 將日期字符串解析為 timestamp
		date, _ := time.Parse("2006-01-02 15:04:05", priceStr+":00")
		// 將價格字符串轉換為 float
		price, _ := strconv.ParseFloat(priceStr, 64)

		// SqlScript := fmt.Sprintf("INSERT INTO \"%s\" (\"date\", \"%s\") VALUES ($1, $2) ON CONFLICT (\"date\") DO UPDATE SET \"%s\" = excluded.\"%s\";", table, stock, stock, stock)

		// 05-25 改為 DO NOTHING:
		SqlScript := fmt.Sprintf("INSERT INTO \"%s\" (\"date\", \"%s\") VALUES ($1, $2) ON CONFLICT (\"date\") DO NOTHING;", table, stock)

		result, err := dbConn.Exec(SqlScript, date, price)
		if err != nil {
			log.Printf("SQL 執行錯誤：%v\n", err)
			continue
		}

		// 检查是否插入成功
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("获取影响的行数时出错：%v\n", err)
			fmt.Printf("获取影响的行数时出错：%v\n", err)
			continue
		}

		// 如果没有行受影响，说明记录已存在，跳出循环
		if rowsAffected == 0 {
			log.Printf("AddPrice %v %v 日期紀錄已存在，停止插入操作。", stock, date)
			fmt.Printf("AddPrice %v %v 日期紀錄已存在，停止插入操作。", stock, date)
			break
		}
	}
	return nil
}

func AddPE(table string, stock string, dict map[string]float64) error {

	// 创建一个切片来保存 map 的键
	var keys []string
	for key := range dict {
		keys = append(keys, key)
	}

	// 对键进行排序
	sort.Slice(keys, func(i, j int) bool {
		timeI, _ := time.Parse("2006-01-02 15:04:05", keys[i])
		timeJ, _ := time.Parse("2006-01-02 15:04:05", keys[j])
		return timeI.After(timeJ)
	})

	// 按排序后的键顺序访问 map 的值
	for _, key := range keys {
		date, _ := time.Parse("2006-01-02 15:04:05", key)
		pe := dict[key]
		// fmt.Printf("日期: %s, PE: %f\n", date, pe)

		// SqlScript := fmt.Sprintf("INSERT INTO \"%s\" (\"date\", \"%s\") VALUES ($1, $2) ON CONFLICT (\"date\") DO UPDATE SET \"%s\" = excluded.\"%s\";", table, stock, stock, stock)

		// 05-25 :改 DO NOTHING
		SqlScript := fmt.Sprintf("INSERT INTO \"%s\" (\"date\", \"%s\") VALUES ($1, $2) ON CONFLICT (\"date\") DO NOTHING;", table, stock)

		result, err := dbConn.Exec(SqlScript, date, pe)
		if err != nil {
			log.Printf("SQL 執行錯誤：%v\n", err)
			continue
		}
		// 检查是否插入成功
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("获取影响的行数时出错：%v\n", err)
			fmt.Printf("获取影响的行数时出错：%v\n", err)
			continue
		}

		// 如果没有行受影响，说明记录已存在，跳出循环
		if rowsAffected == 0 {
			log.Printf("AddPE %v %v 日期紀錄已存在，停止插入操作。", stock, date)
			fmt.Printf("AddPE %v %v 日期紀錄已存在，停止插入操作。", stock, date)
			break
		}

	}

	// for date, pe := range dict {
	// 	// 將日期字符串解析為 timestamp
	// 	date, _ := time.Parse("2006-01-02 15:04:05", date)
	// 	// fmt.Printf("AddPE date: %v", date)
	// 	// SqlScript := fmt.Sprintf("INSERT INTO %s (date, %s) VALUES (?, ?) ON DUPLICATE KEY UPDATE %s = values(%s);", table, stock, stock, stock)
	// 	SqlScript := fmt.Sprintf("INSERT INTO \"%s\" (\"date\", \"%s\") VALUES ($1, $2) ON CONFLICT (\"date\") DO UPDATE SET \"%s\" = excluded.\"%s\";", table, stock, stock, stock)
	// 	_, err = dbConn.Exec(SqlScript, date, pe)
	// 	if err != nil {
	// 		log.Printf("SQL 執行錯誤：%v\n", err)
	// 		continue
	// 	}
	// }
	return nil

}

func AddSubs(table string, stock string) error {

	// SqlScript := fmt.Sprintf("INSERT INTO NYSE (Company) VALUES ('%s') ON DUPLICATE KEY UPDATE Company = Company;", stock)
	SqlScript := fmt.Sprintf("INSERT INTO \"%s\" (\"Company\") VALUES ('%s') ON CONFLICT (\"Company\") DO NOTHING;", table, stock)

	log.Println("AddSubs SQL:", SqlScript)

	_, err = dbConn.Exec(SqlScript)
	if err != nil {
		log.Printf("SQL 執行錯誤：%v\n", err)
	}
	return nil
}

func DeleteSubs(table string, stock string) error {

	// SqlScript := fmt.Sprintf("DELETE FROM NYSE WHERE Company = '%s';", stock)
	SqlScript := fmt.Sprintf("DELETE FROM\"%s\" WHERE \"Company\" = $1;", table)
	fmt.Println("DELSubs SQL :  ", SqlScript)
	// _, err = dbConn.Exec(SqlScript)
	_, err = dbConn.Exec(SqlScript, stock)
	if err != nil {
		log.Printf("SQL 執行錯誤：%v\n", err)
	}

	return nil
}

func AddNewEPS(table string, stock string, date time.Time, price float64) error {

	// SqlScript := fmt.Sprintf("INSERT INTO %s (date, EPS) VALUES (?, ?) ON DUPLICATE KEY UPDATE EPS = values(EPS);", table)
	SqlScript := fmt.Sprintf("INSERT INTO \"%s\" (\"date\", \"EPS\") VALUES ($1, $2) ON CONFLICT (\"date\") DO UPDATE SET \"EPS\" = EXCLUDED.\"EPS\";", table)

	_, err = dbConn.Exec(SqlScript, date, price)
	if err != nil {
		return fmt.Errorf("SQL 執行錯誤：%v", err)
	}
	return nil
}

func DeleteEPS(table string) error {
	SqlScript := fmt.Sprintf("DROP TABLE \"%s\";", table)
	_, err = dbConn.Exec(SqlScript)
	if err != nil {
		return fmt.Errorf("SQL 執行錯誤：%v", err)
	}
	return nil
}

func AddNewDividend(table string, stock string, date time.Time, price float64) error {

	SqlScript := fmt.Sprintf("INSERT INTO \"%s\" (\"date\", \"Dividend\") VALUES ($1, $2) ON CONFLICT (\"date\") DO UPDATE SET \"Dividend\" = EXCLUDED.\"Dividend\";", table)
	_, err = dbConn.Exec(SqlScript, date, price)
	if err != nil {
		return fmt.Errorf("SQL 執行錯誤：%v", err)
	}
	return nil
}

func DeleteDividend(table string) error {
	SqlScript := fmt.Sprintf("DROP TABLE \"%s\";", table)
	_, err = dbConn.Exec(SqlScript)
	if err != nil {
		return fmt.Errorf("SQL 執行錯誤：%v", err)
	}
	return nil
}

func AddPERatio(stock string) error {
	return nil
}

/*
############## 共用操作  ##############
*/
func CheckTableExist(table string) error { //確認Table存在

	SqlScript := fmt.Sprintf("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE '%s';", table)
	var tableName string
	err = dbConn.QueryRow(SqlScript).Scan(&tableName)
	if err == sql.ErrNoRows {
		return errors.New("table 不存在")
	} else if err != nil {
		return err
	}
	return nil
}

func CheckColumnExist(table string, col string) error { //確認欄位存在

	SqlScript := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='public' AND table_name ='%s' AND column_name ='%s';", table, col)
	log.Println("CheckColumnExist SQL :  ", SqlScript)
	var count int
	nill := dbConn.QueryRow(SqlScript).Scan(&count)
	log.Printf("count:%d", count)
	if count > 0 {
		return nill //存在
	}
	return errors.New("不存在")
}

func CheckIsInTable(table string, column string, stock string) error { // table 查詢, 有確認 rows

	SqlScript := fmt.Sprintf("SELECT COUNT(*) FROM \"%s\" WHERE \"%s\" = '%s';", table, column, stock)
	log.Println("CheckIsInTable SQL :  ", SqlScript)
	var count int
	err = dbConn.QueryRow(SqlScript).Scan(&count)
	if err != nil {
		return fmt.Errorf("SQL 執行錯誤：%v", err)
	}
	log.Printf("查詢 %s 裡是否有%s股票: %d ", table, stock, count)

	if count > 0 {
		return nil
	} else {
		return fmt.Errorf("在 %s 表中未找到股票 %s", table, stock)
	}
}

func CheckIsInTable2(table string, stock string) error { //table 查詢 不用確認 rows 用到

	SqlScript := fmt.Sprintf("SELECT column_name FROM information_schema.columns WHERE table_schema = 'public' AND table_name = '%s' AND column_name LIKE '%s';", table, stock)

	rows, err := dbConn.Query(SqlScript)
	if err != nil {
		return fmt.Errorf("SQL 執行錯誤：%v", err)
	} else {
		log.Printf("查詢 %s 裡是否有股票: %s ", table, stock)
	}
	if rows.Next() {
		return nil
	} else {
		return fmt.Errorf("在 %s 表中未找到股票 %s", table, stock)
	}
}

func GetContent(params ...interface{}) (*sql.Rows, error) { // 取得欄位裡的資料

	var SqlScript string
	if len(params) > 2 {
		table, _ := params[0].(string)
		col, _ := params[1].(string)
		date, _ := params[2].(string)
		SqlScript = fmt.Sprintf("SELECT \"%s\" ,\"%s\" FROM \"%s\" ORDER BY \"%s\" ASC;", date, col, table, date)

	} else {
		table, _ := params[0].(string)
		col, _ := params[1].(string)
		SqlScript = fmt.Sprintf("SELECT \"%s\" FROM \"%s\";", col, table)
	}

	log.Println("SqlScript:", SqlScript)
	rows, err := dbConn.Query(SqlScript)
	if err != nil {
		return rows, err
	}
	return rows, nil
}

func QueryDataOrderByDate(col1 string, col2 string, table string, order_col string, number int) (*sql.Rows, error) {

	SqlScript := fmt.Sprintf("SELECT \"%s\" ,\"%s\" FROM \"%s\" ORDER BY \"%s\" DESC LIMIT %d;", col1, col2, table, order_col, number)

	log.Println("SqlScript:", SqlScript)
	rows, err := dbConn.Query(SqlScript)

	if err != nil {
		return rows, err
	}
	return rows, nil
}
