package cronjobs

import (
	"linebot-gemini-pro/db"
	"log"
	"os"
)

func GetSubsCompanies(table string, col string) []string {

	file, _ := os.Create("logs/Write-to-PE.log")
	log.SetOutput(file)

	var companies []string
	rows, _ := db.Controller_GetContnet(table, col)
	for rows.Next() {
		var rowData string
		if err := rows.Scan(&rowData); err != nil {
			log.Println("Process 遇到問題...", err)
		}
		log.Println("Company:", rowData)
		companies = append(companies, rowData)
	}
	rows.Close()
	return companies
}
