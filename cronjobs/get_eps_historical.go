package cronjobs

import (
	"linebot-gemini-pro/db"
	"log"
	"os"
	"time"
)

type EPS_RowData struct {
	date time.Time
	EPS  float64
}

func GetEPS_Historical(table string, stock string, companies []string) (map[string][]EPS_RowData, []string, error) {

	var rowData EPS_RowData
	reuslt_map := make(map[string][]EPS_RowData)

	file, _ := os.Create("/home/vagrant/go_project/ekils.github.io/logs/Write-to-PE.log")
	log.SetOutput(file)

	// companies := GetSubsCompanies(table, stock)
	var ok_companies []string
	for _, company := range companies {

		var resultList []EPS_RowData

		err := db.Controller_CheckTable(company + "_EPS")
		if err != nil {
			log.Printf("Table: %v 不存在", company+"_EPS")
		} else {
			// columns := strings.Join([]string{"date", "EPS"}, ", ")
			columns2 := "date"
			columns1 := "EPS"
			rows, _ := db.Controller_GetContnet(company+"_EPS", columns1, columns2)

			for rows.Next() {
				// log.Println("Now GET... ", company)
				if err := rows.Scan(&rowData.date, &rowData.EPS); err != nil {
					return reuslt_map, ok_companies, err
				}
				resultList = append(resultList, rowData)
			}
			if err := rows.Err(); err != nil {
				return reuslt_map, ok_companies, err
			}
			rows.Close()
			reuslt_map[company] = resultList
			ok_companies = append(ok_companies, company)
		}

	}

	// log.Println("reuslt_map:", reuslt_map)

	return reuslt_map, ok_companies, nil
}

// func InsertIntoEPS_Table(table string, stock string) {

// }

// func To_EP_Ratio(table string, stock string) {

// }
