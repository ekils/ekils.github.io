package cronjobs

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

/*
! 與db互動都要在 db_controller 操作

! Cron jobs裡：

	只要有被訂閱的, 每天都會執行:
	1. 有訂閱的股票 price 更新
	2. 有訂閱的股票 pe-ratio 欄位更新 / pb-ratio 欄位更新
	3. 圖表推播

	[點擊圖檔]:
	產生圖表
*/
func CronJobs() {

	file, _ := os.Create("/logs/Write-to-PE.log")
	log.SetOutput(file)

	// 測試用:
	// var AAPL []string
	// AAPL = append(AAPL, "AAPL")
	// Script(AAPL)

	// Step0 : 清空 html 檔案
	cmd := `
	rm -rf ./html/*.html;
	`
	combinedCmd := exec.Command("sh", "-c", cmd)
	if err := combinedCmd.Run(); err != nil {
		fmt.Printf("Cron Step 清空 Html檔案出錯......: %v", err)
	}

	// Step1: 更新股價資訊 parse price history:
	fmt.Printf("/ Step1: 更新股價資訊 parse price history: ")
	log.Printf("/ Step1: 更新股價資訊 parse price history: ")
	reuslt, companies, err := ParsePrice("NYSE", "STOCKHISTORY", "Company")
	if err != nil {
		log.Println(err)
		fmt.Println(err)
	}
	log.Println(reuslt)
	fmt.Println(reuslt)

	// Step2: 取得股價 parse price history:
	fmt.Printf("// Step2: 取得股價 parse price history:")
	log.Printf("// Step2: 取得股價 parse price history:")
	reuslt_map_price, err := GetPrice_Historical("NYSE", "STOCKHISTORY", "Company", companies)
	// log.Printf("reuslt_map_price: %v", reuslt_map_price)
	if err != nil {
		log.Println(err)
		fmt.Println(err)
	}

	// // Step3: 取得eps parse eps history:(不用parse eps 是因為這是手動增加)
	fmt.Printf("// Step3: 取得eps parse eps history:(不用parse eps 是因為這是手動增加)")
	log.Printf("// Step3: 取得eps parse eps history:(不用parse eps 是因為這是手動增加)")
	reuslt_map_eps, ok_companies, err := GetEPS_Historical("NYSE", "Company", companies)
	log.Printf("reuslt_map_eps: %v", reuslt_map_eps)
	fmt.Printf("reuslt_map_eps: %v", reuslt_map_eps)

	if err != nil {
		log.Println(err)
		fmt.Println(err)
	}

	// Step3: gen pe-ratio report to db :
	fmt.Printf("Step3: gen pe-ratio report to db :")
	log.Printf("Step3: gen pe-ratio report to db :")
	response, err = GenPE_Ratio(reuslt_map_price, reuslt_map_eps)
	if err != nil {
		log.Println(err)
		fmt.Println(err)
	} else {
		log.Println(response)
		fmt.Println(response)
	}

	// Step4: Gen plot link
	fmt.Printf("/ Step4: Gen plot link")
	log.Printf("/ Step4: Gen plot link")
	// var ok_companies = []string{"MSFT"}

	if len(ok_companies) > 0 {
		Plot(ok_companies)
		//Step5: Plot through Liff
		Script(ok_companies)
	} else {
		log.Println("EPS Table都沒有資料, 請先Update")
		fmt.Println("EPS Table都沒有資料, 請先Update")
	}

}
