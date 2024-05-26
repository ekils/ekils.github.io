package cronjobs

import (
	// "crypto/rand"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"linebot-gemini-pro/db"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	grob "github.com/MetalBlueberry/go-plotly/graph_objects"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/joho/godotenv"
	"gonum.org/v1/gonum/stat"
)

var formats = []string{"_PE", "_EPS"}
var reuslt_EPS = make(map[string][]EPS_Plot)
var reuslt_PE = make(map[string][]PE_Plot)
var reuslt_Price = make(map[string][]Price_Plot)
var group_setting = 20 // 過去五年 = 20期
var pyplotip string
var html_location string
var logplot string

type PE_Plot struct {
	Date string
	PE   float64
}

type EPS_Plot struct {
	Date string
	EPS  float64
}

type Price_Plot struct {
	Date  string
	Price float64
}

type Color interface{}

type LineShape struct {
	Type string           `json:"type,omitempty"`
	X0   time.Time        `json:"x0,omitempty"`
	Y0   float64          `json:"y0,omitempty"`
	X1   time.Time        `json:"x1,omitempty"`
	Y1   float64          `json:"y1,omitempty"`
	Line grob.ScatterLine `json:"line,omitempty"`
}

func init() {
	errors := godotenv.Load(".env")
	if errors != nil {
		log.Fatalf("Error loading .env file")
	}
	pyplotip = os.Getenv("PyPlotIP")
	html_location = os.Getenv("HtmlLocation")
	logplot = os.Getenv("Log_Plot")

}

func Plot(companies []string) {

	file, _ := os.Create(logplot)
	log.SetOutput(file)

	for _, company := range companies {
		company_map_plot_info := PrePlot(company)
		GenPlot(company, company_map_plot_info)
	}

}

func PrePlot(company string) map[string][]interface{} {

	/*============================================================
		[邏輯]
		0. 計算整體的 P/E_LOG ,並取得整體 P/E_LOG 斜率
		1. 先找出前20期的 eps 作為 group 區間
		2. 然後依照各區間找出最高點的價格
		3. 區間斜率
	=============================================================*/

	var (
		interval_data_y0             []float64
		interval_data_y1             []float64
		slope                        float64
		stdev                        float64
		eps_date_group               []string
		eps_date_group_with_add_next []time.Time
		eps_data_group               []float64
		max_pelog10_list             []float64
		filtered_partial             dataframe.DataFrame
		filtered                     dataframe.DataFrame
		plot_info_list               []interface{}
		company_map_plot_info        = make(map[string][]interface{})
	)

	// 0-1 取得訂閱單裡所有的 eps pe
	log.Printf("// 0-1 取得訂閱單裡所有的 eps pe")
	fmt.Printf("// 0-1 取得訂閱單裡所有的 eps pe")
	reuslt_EPS, reuslt_PE, err := Get_EPS_and_PE(company)

	if err != nil {
		fmt.Println(err)
		fmt.Println(reuslt_PE, reuslt_EPS)
		log.Println(err)
		log.Println(reuslt_PE, reuslt_EPS)
	}

	// 0-2 取得訂閱單裡所有的 price
	log.Printf("// 0-2 取得訂閱單裡所有的 price")
	fmt.Printf("// 0-2 取得訂閱單裡所有的 price")
	reuslt_Price, err := Get_PriceData(company)
	if err != nil {
		fmt.Println(err)
		fmt.Println(reuslt_Price)
		log.Println(err)
		log.Println(reuslt_Price)
	}

	// 0-3 收集整體的 P/E_LOG (從2015-01-02開始: 735600)
	log.Printf("// 0-3 收集整體的 P/E_LOG (從2015-01-02開始: 735600)")
	fmt.Printf("// 0-3 收集整體的 P/E_LOG (從2015-01-02開始: 735600)")
	var dfs = make(map[string]dataframe.DataFrame)

	dfs[company] = dataframe.LoadStructs(reuslt_PE[company]) //建立 dataframe
	fmt.Printf("[company]: %v", company)
	fmt.Printf("dfs[company]: %v", dfs[company])
	log.Printf("[company]: %v", company)
	log.Printf("dfs[company]: %v", dfs[company])
	ordinal_data := ToOrdinal(dfs[company].Col("Date")) //新增 col: Ordinal
	dfs[company] = dfs[company].Mutate(ordinal_data).Rename("Ordinal", "X0")
	fmt.Printf("dfs[company]2: %v", dfs[company])
	log.Printf("dfs[company]2: %v", dfs[company])
	//調整資料區間:
	dfs[company] = dfs[company].Filter(
		dataframe.F{Colname: "Ordinal", Comparator: ">=", Comparando: int(735600)})
	fmt.Printf("dfs[company]3: %v", dfs[company])
	log.Printf("dfs[company]3: %v", dfs[company])
	// 準備加入 price:
	t_PriceDataframe := dataframe.LoadStructs(reuslt_Price[company])

	dateMap := make(map[string]bool)
	for _, date := range dfs[company].Col("Date").Records() {
		dateMap[date] = true
	}
	// 以 pe ratio 裡有的date為主,取對應price 的date再合併:
	PriceDataframe := t_PriceDataframe.Filter(
		dataframe.F{
			Colname:    "Date",
			Comparator: series.In,
			Comparando: dfs[company].Col("Date"),
		},
	)
	ordinal_data_p := ToOrdinal(PriceDataframe.Col("Date")) //新增 col: Ordinal
	PriceDataframe = PriceDataframe.Mutate(ordinal_data_p).Rename("Ordinal", "X0")
	fmt.Printf("PriceDataframe: %v", PriceDataframe)
	log.Printf("PriceDataframe: %v", PriceDataframe)
	// 準備加入 price 調整資料區間:
	PriceDataframe = PriceDataframe.Filter(
		dataframe.F{Colname: "Ordinal", Comparator: ">=", Comparando: int(735600)})
	fmt.Printf("PriceDataframe2: %v", PriceDataframe)
	fmt.Println("PriceDataframe3:", PriceDataframe.Nrow())
	fmt.Println("dfs:", dfs[company].Nrow())
	log.Printf("PriceDataframe2: %v", PriceDataframe)
	log.Println("PriceDataframe3:", PriceDataframe.Nrow())
	log.Println("dfs:", dfs[company].Nrow())
	PriceSeries := PriceDataframe.Col("Price") // 加入 Price Col:
	fmt.Printf("PriceSeries: %v", PriceSeries)
	log.Printf("PriceSeries: %v", PriceSeries)
	dfs[company] = dfs[company].Mutate(PriceSeries).Rename("Price", "Price")
	fmt.Printf("dfs[company]4: %v", dfs[company])
	fmt.Printf("dfs[company].Col(PE): %v", dfs[company].Col("PE"))
	log.Printf("dfs[company]4: %v", dfs[company])
	log.Printf("dfs[company].Col(PE): %v", dfs[company].Col("PE"))
	logSeries := DataToLog10(dfs[company].Col("PE")) //新增 col: PE
	dfs[company] = dfs[company].Mutate(logSeries).Rename("PE_LOG10", "X0")

	// fmt.Println(dfs)
	// 1. 找出前20期的 eps 作為 group 區間
	// 0-4. 斜率
	// 0-5. std
	// 2. 依照各區間找 p😈ice 高點:
	// 3. 區間斜率

	// 1-1. 依照 company 取得 前20期的 eps date, data

	eps_date_group, eps_data_group = GroupedEPS(company, reuslt_EPS)
	fmt.Println(eps_data_group)
	fmt.Println(eps_date_group)
	log.Println(eps_data_group)
	log.Println(eps_date_group)
	// 1-2. 依照 company 增加下一期的預估時間
	eps_date_group_with_add_next = AddNextGroupEPS_Date(eps_date_group)
	fmt.Println(eps_date_group_with_add_next)
	log.Println(eps_date_group_with_add_next)

	// 0-4 看最新 ~前20期的
	eps_data_xxth := eps_date_group[len(eps_date_group)-1]
	// fmt.Println("eps_data_xxth:", eps_data_xxth)
	eps_data_now := time.Now()
	eps_data_xxth_times := String2Time(eps_data_xxth)
	eps_data_xxth_time2unix := eps_data_xxth_times.Unix()
	eps_data_xxth_ordinal := UnixToProlepticGregorianOrdinal(eps_data_xxth_time2unix)
	eps_data_now_time2unix := eps_data_now.Unix()
	eps_data_now_ordinal := UnixToProlepticGregorianOrdinal(eps_data_now_time2unix)
	filtered = dfs[company].FilterAggregation(
		dataframe.And,
		dataframe.F{Colname: "Ordinal", Comparator: "<=", Comparando: int(eps_data_now_ordinal)},
		dataframe.F{Colname: "Ordinal", Comparator: ">=", Comparando: int(eps_data_xxth_ordinal)})
	//之後可刪-----:
	// fmt.Println("\n")
	fmt.Printf(" 🦖🦖🦖 [看 %s DataFrame 裡最近的資料]... 🦖🦖🦖", company)
	log.Printf(" 🦖🦖🦖 [看 %s DataFrame 裡最近的資料]... 🦖🦖🦖", company)
	test := watch_tail(15, filtered)
	fmt.Println(test)
	log.Println(test)
	//------------
	// 0-4 取得過去20期斜率
	xs := filtered.Col("Ordinal")
	ys := filtered.Col("PE_LOG10")
	// rowCount := filtered.Nrow()
	// fmt.Println("Number of rows in DataFrame:", rowCount)
	slope = GetSlope(xs, ys)
	fmt.Printf("-----> Slope of %s <-----: %.6f\n", company, slope)
	log.Printf("-----> Slope of %s <-----: %.6f\n", company, slope)
	// 0-5 std:
	stdev = stat.StdDev(filtered.Col("PE_LOG10").Float(), nil)
	fmt.Printf("-----> Std   of %s <-----: %.6f\n", company, stdev)
	log.Printf("-----> Std   of %s <-----: %.6f\n", company, stdev)

	// 2. 依照各區間找 price 高點:
	fmt.Printf("eps_data_group:%v", eps_data_group)
	log.Printf("eps_data_group:%v", eps_data_group)
	for index := 0; index <= (len(eps_data_group) - 1); index++ {
		if index == 0 {
			defaults := eps_date_group[index]
			fmt.Printf("index =0 :%v", defaults)
			log.Printf("index =0 :%v", defaults)
			s2t := String2Time(defaults)
			t2int64 := s2t.Unix()
			t2int64_with_ordinal := UnixToProlepticGregorianOrdinal(t2int64)
			fmt.Printf("區間範圍  %s ~\n", defaults)
			log.Printf("區間範圍  %s ~\n", defaults)
			filtered_partial = filtered.Filter(
				dataframe.F{Colname: "Ordinal", Comparator: ">=", Comparando: int(t2int64_with_ordinal)})

		} else {
			defaults := eps_date_group[index-1]
			fmt.Printf("index before :%v", defaults)
			log.Printf("index before :%v", defaults)
			s2t := String2Time(defaults)
			t2int64 := s2t.Unix()
			t2int64_with_ordinal := UnixToProlepticGregorianOrdinal(t2int64)

			defaults_1 := eps_date_group[index]
			fmt.Printf("index now :%v", defaults_1)
			log.Printf("index now :%v", defaults_1)
			s2t_1 := String2Time(defaults_1)
			t2int64_1 := s2t_1.Unix()
			t2int64_with_ordinal_1 := UnixToProlepticGregorianOrdinal(t2int64_1)

			fmt.Printf("區間範圍  %s ~ %s\n", defaults_1, defaults)
			log.Printf("區間範圍  %s ~ %s\n", defaults_1, defaults)

			filtered_partial = filtered.FilterAggregation(
				dataframe.And,
				dataframe.F{Colname: "Ordinal", Comparator: "<=", Comparando: int(t2int64_with_ordinal)},
				dataframe.F{Colname: "Ordinal", Comparator: ">", Comparando: int(t2int64_with_ordinal_1)})
		}
		max_price := filtered_partial.Col("Price").Max()
		fmt.Println("max_price:", max_price)
		log.Println("max_price:", max_price)
		filter := dataframe.F{Colname: "Price", Comparator: series.Eq, Comparando: max_price}
		filteredDF := filtered_partial.Filter(filter)

		fmt.Printf("filteredDF1:%v", filteredDF)
		log.Printf("filteredDF1:%v", filteredDF)

		// 透過 filter 找 col A 的row 對應col B的值
		pelog10_string := filteredDF.Col("PE_LOG10").Records()[0]
		pelog10_float, _ := strconv.ParseFloat(pelog10_string, 64)
		// fmt.Println("pelog10_float: ", pelog10_float)
		max_pelog10_list = append(max_pelog10_list, pelog10_float)
		fmt.Println("PE:", filteredDF.Col("PE").Records()[0])
		fmt.Println("Price:", filteredDF.Col("Price").Records()[0])
		fmt.Println("max_pelog10_list:", max_pelog10_list)
		log.Println("PE:", filteredDF.Col("PE").Records()[0])
		log.Println("Price:", filteredDF.Col("Price").Records()[0])
		log.Println("max_pelog10_list:", max_pelog10_list)
	}
	// 3. 區間斜率
	for i := 0; i <= len(max_pelog10_list)-1; i++ {
		y1 := (max_pelog10_list[i] + slope*45)
		y0 := (max_pelog10_list[i] - slope*45)
		interval_data_y0 = append(interval_data_y0, y0)
		interval_data_y1 = append(interval_data_y1, y1)
		fmt.Println("最高本益比區間:", y0, y1)
		log.Println("最高本益比區間:", y0, y1)
	}

	// // 時間轉換sample:
	// defaults := "2015-01-02T00:00:00Z"
	// s2t := String2Time(defaults)
	// t2int64 := s2t.Unix()
	// t2int64_with_ordinal := UnixToProlepticGregorianOrdinal(t2int64)
	// fmt.Println(t2int64_with_ordinal)
	// int2time := RevertUnixToTimeStamp(int(t2int64_with_ordinal))
	// fmt.Println(int2time)
	plot_info_list = append(plot_info_list, eps_date_group_with_add_next) //[0]
	plot_info_list = append(plot_info_list, eps_data_group)               //[1]
	plot_info_list = append(plot_info_list, filtered)                     //[2]
	plot_info_list = append(plot_info_list, slope)                        //[3]
	plot_info_list = append(plot_info_list, stdev)                        //[4]
	plot_info_list = append(plot_info_list, interval_data_y0)             //[5]
	plot_info_list = append(plot_info_list, interval_data_y1)             //[6]
	company_map_plot_info[company] = plot_info_list
	fmt.Println("PrePlot Done -----")
	log.Println("PrePlot Done -----")
	return company_map_plot_info
}

func GenPlot(company string, company_map_plot_info map[string][]interface{}) {

	// var x_timetime_list []time.Time
	// Prepare Data:
	plot_info_list := company_map_plot_info[company]
	// fmt.Println(plot_info_list)
	filtered := plot_info_list[2].(dataframe.DataFrame)
	x_timestring := filtered.Col("Date")
	y_data_PE_Log := filtered.Col("PE_LOG10")

	// X data準備:
	x_timestring_list := x_timestring.Records() // recode() --> 轉 []string
	// for _, x := range x_timestring_list {
	// 	x_timetime := String2Time(x)
	// 	// x_timetime_list = append(x_timetime_list, x_timetime)
	// }
	// Y 資料準備:
	src := y_data_PE_Log.Float()
	yCopied := make([]float64, len(src))
	copy(yCopied, src)
	sort.Slice(yCopied, func(i, j int) bool {
		return yCopied[i] < yCopied[j]
	})
	yminValue := yCopied[0]
	ymaxValue := yCopied[len(yCopied)-1]

	// Prepare Layout:
	eps_date_group_with_add_next := plot_info_list[0].([]time.Time)
	interval_data_y0 := plot_info_list[5].([]float64)
	interval_data_y1 := plot_info_list[6].([]float64)
	slope := plot_info_list[3].(float64)
	stdev := plot_info_list[4].(float64)

	jsonData := Data2Json(company, eps_date_group_with_add_next, x_timestring_list, y_data_PE_Log.Float(), yminValue, ymaxValue, interval_data_y0, interval_data_y1, slope, stdev)

	fmt.Println("準備進入 PythonPlot")
	log.Println("準備進入 PythonPlot")
	response := PythonPlot(jsonData, company)
	fmt.Println(response)
	log.Println(response)

}

func Data2Json(company string, eps_date_group_with_add_next []time.Time, x_timestring_list []string, y_data_list []float64, yminValue float64, ymaxValue float64, interval_data_y0 []float64, interval_data_y1 []float64, slope float64, stdev float64) []byte {

	data := map[string]interface{}{
		"company":                      company,
		"x_timestring_list":            x_timestring_list,
		"eps_date_group_with_add_next": eps_date_group_with_add_next,
		"y_data_list":                  y_data_list,
		"yminValue":                    yminValue,
		"ymaxValue":                    ymaxValue,
		"interval_data_y0":             interval_data_y0,
		"interval_data_y1":             interval_data_y1,
		"slope":                        slope,
		"stdev":                        stdev,
	}
	jsonData, _ := json.Marshal(data)
	return jsonData
}

func PythonPlot(jsonData []byte, company string) string {

	resp, err := http.Post(pyplotip, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error:", err)
		log.Println("Error:", err)
		return err.Error()
	}
	defer resp.Body.Close()

	err = Data2Html(resp, company)
	if err != nil {
		fmt.Println("Error:", err)
		log.Println("Error:", err)
		return err.Error()
	}
	fmt.Printf("PythonPlot: %v", resp.Status)
	log.Printf("PythonPlot: %v", resp.Status)
	return resp.Status
}

func Data2Html(resp *http.Response, company string) error {

	htmlContent, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		log.Println("Error:", err)
		return err
	}

	filr_path := fmt.Sprintf(html_location+"PE_Trend_%s.html", company)
	err = os.WriteFile(filr_path, htmlContent, os.ModePerm)
	if err != nil {
		fmt.Println("Error:", err)
		log.Println("Error:", err)
		return err
	}
	return nil
}

func GetSlope(xs series.Series, ys series.Series) float64 {

	/*==============计算协方差===================
	ssxm = mean( (x-mean(x))^2 )
	ssxym = mean( (x-mean(x)) * (y-mean(y)) )
	==========================================*/

	// 将 series 转换为 []float64
	xFloats := xs.Float()
	yFloats := ys.Float()
	xmean := stat.Mean(xFloats, nil)
	ymean := stat.Mean(yFloats, nil)
	// fmt.Println("xmean:", xmean)
	// fmt.Println("ymean:", ymean)
	var deltax_list []float64
	var deltay_list []float64
	for i := 0; i < len(xFloats); i++ {
		x := xFloats[i]
		y := yFloats[i]
		deltax := (x - xmean) * (x - xmean)
		deltax_list = append(deltax_list, deltax)
		deltay := (x - xmean) * (y - ymean)
		deltay_list = append(deltay_list, deltay)
	}

	ssxm := stat.Mean(deltax_list, nil)
	ssxym := stat.Mean(deltay_list, nil)
	slope := ssxym / ssxm
	return slope
}

func ToOrdinal(s series.Series) series.Series {

	var tempTime []int
	for i := 0; i < s.Len(); i++ {
		times := String2Time(s.Elem(i).String())
		time2unix := times.Unix()
		unix2 := UnixToProlepticGregorianOrdinal(time2unix)
		tempTime = append(tempTime, int(unix2))
	}
	return series.Ints(tempTime)
}

func UnixToProlepticGregorianOrdinal(unixTimestamp int64) int64 { // 將時間轉為unix
	// 1970年1月1日是 proleptic Gregorian 日历的第一天
	// 因此 proleptic Gregorian ordinal = Unix 时间戳 / 86400 + 719163
	return (unixTimestamp / 86400) + 719163
}

func RevertUnixToTimeStamp(timeunix int) time.Time { // 將unit轉為時間
	unixTimestamp := int64((timeunix - 719163) * 86400)
	// 将 Unix 时间戳转换为时间对象
	t := time.Unix(unixTimestamp, 0)
	// 打印转换后的时间对象
	fmt.Println("Unix 时间戳对应的日期是：", t.Format("2006-01-02"))
	log.Println("Unix 时间戳对应的日期是：", t.Format("2006-01-02"))
	return t
}

func DataToLog10(s series.Series) series.Series {

	var log10 []float64
	floats := s.Float()
	for _, f := range floats {
		log10 = append(log10, math.Log10(f))
	}
	return series.Floats(log10)
}

func String2Time(date_string string) time.Time {

	datetime, _ := time.Parse(time.RFC3339, date_string)
	return datetime
}

func Time2String(datetime time.Time) string {

	date_string := datetime.Format("2006-01-02 15:04:05")
	return date_string
}

func Get_EPS_and_PE(company string) (map[string][]EPS_Plot, map[string][]PE_Plot, error) {

	for _, format := range formats {

		com_string := company + format

		if format == "_PE" {
			var rowData PE_Plot
			var resultList []PE_Plot
			var dateString string
			log.Printf("Get_EPS_and_PE-----if")
			fmt.Printf("Get_EPS_and_PE-----if")
			rows, _ := db.Controller_GetContnet(com_string, company, "date")

			for rows.Next() {
				if err := rows.Scan(&dateString, &rowData.PE); err != nil {
					fmt.Println(".....:", err)
					log.Println(".....:", err)
					return reuslt_EPS, reuslt_PE, err
				}
				// fmt.Printf("append: %s", format)
				rowData.Date = dateString
				resultList = append(resultList, rowData)
			}
			if err := rows.Err(); err != nil {
				return reuslt_EPS, reuslt_PE, err
			}
			rows.Close()
			reuslt_PE[company] = resultList

		} else {
			var rowData EPS_Plot
			var resultList []EPS_Plot
			log.Printf("Get_EPS_and_PE-----else")
			fmt.Printf("Get_EPS_and_PE-----else")
			rows, _ := db.Controller_GetContnet(com_string, "EPS", "date")
			for rows.Next() {
				if err := rows.Scan(&rowData.Date, &rowData.EPS); err != nil {
					fmt.Println(".....:", err)
					log.Println(".....:", err)
					return reuslt_EPS, reuslt_PE, err
				}
				// fmt.Printf("append: %s", format)
				resultList = append(resultList, rowData)
			}
			if err := rows.Err(); err != nil {
				return reuslt_EPS, reuslt_PE, err
			}
			rows.Close()
			reuslt_EPS[company] = resultList
		}

	}

	return reuslt_EPS, reuslt_PE, nil
}

func Get_PriceData(company string) (map[string][]Price_Plot, error) {

	var rowData Price_Plot
	var resultList []Price_Plot
	log.Printf("Get_PriceData-----")
	fmt.Printf("Get_PriceData-----")
	rows, _ := db.Controller_GetContnet("STOCKHISTORY", company, "date")
	for rows.Next() {
		if err := rows.Scan(&rowData.Date, &rowData.Price); err != nil {
			fmt.Println(".....:", err)
			log.Println(".....:", err)
			return reuslt_Price, err
		}
		resultList = append(resultList, rowData)

		if err := rows.Err(); err != nil {
			return reuslt_Price, err
		}
	}
	rows.Close()
	reuslt_Price[company] = resultList

	return reuslt_Price, nil
}

func GroupedEPS(com string, reuslt_EPS map[string][]EPS_Plot) ([]string, []float64) { //依照每三個月 eps 分期, 抓x期

	var eps_date_group []string
	var eps_data_group []float64
	var count int
	company := com //+ "_EPS"
	count = 1
	for index := len(reuslt_EPS[company]) - 1; index >= 0; index-- {
		if count <= group_setting+1 { // ＋補 區間頭
			eps_date_group = append(eps_date_group, reuslt_EPS[company][index].Date)
			eps_data_group = append(eps_data_group, reuslt_EPS[company][index].EPS)
			count += 1
		}
	}
	fmt.Printf("******* %s 前20期 EPS ******* : ", company)
	fmt.Println(eps_date_group)
	log.Printf("******* %s 前20期 EPS ******* : ", company)
	log.Println(eps_date_group)
	return eps_date_group, eps_data_group
}

func AddNextGroupEPS_Date(eps_date_group []string) []time.Time { // 推估下一個 eps日 (繪圖需要)

	var new_eps_date_group []time.Time

	for _, eps_datestring := range eps_date_group {
		dateString := eps_datestring
		parsedTime := String2Time(dateString)
		new_eps_date_group = append(new_eps_date_group, parsedTime)
	}

	// 計算時間差
	dayDistance := new_eps_date_group[0].Sub(new_eps_date_group[1])
	// 使用 Add 函數添加時間差
	new_date := new_eps_date_group[0].Add(dayDistance)

	new_eps_date_group = append([]time.Time{new_date}, new_eps_date_group...)
	return new_eps_date_group
}

func watch_tail(quantity int, df dataframe.DataFrame) dataframe.DataFrame {

	var how_many []int
	for index := 0; index <= quantity; index++ {
		how_many = append(how_many, index)
	}
	sorted := df.Arrange(
		dataframe.RevSort("Ordinal"),
	)
	last := sorted.Subset(how_many)
	return last
}

func GenRandomString(howmany_byte int) string {

	b := make([]byte, howmany_byte)
	var str string

	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Failed to generate random bytes:", err)
		log.Println("Failed to generate random bytes:", err)
		return err.Error()
	}
	for _, v := range b {
		str += fmt.Sprintf("%d", v)
	}
	return str
}
