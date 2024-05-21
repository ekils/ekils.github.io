package cronjobs

import (
	"linebot-gemini-pro/db"
	"log"
	"os"
	"time"
)

type Price_RowData struct {
	date       time.Time
	StockPrice float64
}

func GetPrice_Historical(sub_table string, parse_table string, col string, companies []string) (map[string][]Price_RowData, error) {

	file, _ := os.Create("/logs/Write-to-PE.log")
	log.SetOutput(file)

	var rowData Price_RowData
	reuslt_map := make(map[string][]Price_RowData)

	for _, company := range companies {

		var resultList []Price_RowData

		rows, _ := db.Controller_GetContnet(parse_table, company, "date")

		for rows.Next() {
			// log.Println("Now GET... ", company)
			if err := rows.Scan(&rowData.date, &rowData.StockPrice); err != nil {
				log.Println(".....:", err)
				return reuslt_map, err
			}
			resultList = append(resultList, rowData)
		}
		if err := rows.Err(); err != nil {
			return reuslt_map, err
		}
		rows.Close()
		reuslt_map[company] = resultList
	}
	// log.Println("reuslt_map:", reuslt_map)

	return reuslt_map, nil

}
