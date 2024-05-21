package operations

import (
	"log"
	"math"
)

func GetReasonablePrice(stock string, ddm_data []float64) (string, float64) {

	sums := make([]float64, 0)
	// 遍歷原始切片，每4個數字相加一次
	for i := 0; i < len(ddm_data); i += 4 {
		sum := 0.0
		for j := 0; j < 4 && i+j < len(ddm_data); j++ {
			sum += ddm_data[i+j]
		}
		sums = append(sums, sum)
	}
	log.Printf(" %s 過去24季度配息: %v", stock, ddm_data)
	log.Printf(" %s 過去每年配息[4 Season]: %v", stock, sums)

	// 計算 1y ,3y, 5y 成長率:
	D0 := sums[0]
	year_1 := float64(sums[0])/float64(sums[1]) - 1
	year_3 := math.Pow(float64(sums[0])/float64(sums[3]), 1.0/3) - 1
	year_5 := math.Pow(float64(sums[0])/float64(sums[5]), 1.0/5) - 1

	// 計算平均成長率:
	year_avg := (year_1 + year_3 + year_5) / 3.0

	log.Printf("year_1: %f", year_1)
	log.Printf("year_3: %f", year_3)
	log.Printf("year_5: %f", year_5)
	log.Printf("year_avg: %f", year_avg)

	// 計算 合理值:
	FV_str, FV := DDMAlgorithm(D0, year_1, year_3, year_5, year_avg)

	return FV_str, FV
}

func DDMAlgorithm(D0 float64, year_1 float64, year_3 float64, year_5 float64, year_avg float64) (string, float64) {

	var FV_str string

	// Fair Value1 :
	ddm_r := 0.12 //12%
	FV := D0 * (1 + year_avg) / (ddm_r - year_avg)

	if FV > 0 {
		FV_str = "Fair Value(合理價)"

	} else {
		// Fair Value2:
		ddm_r := 0.12 //12%
		ddm_h := 7.0
		ddm_gL := 0.05 // 5%
		FV = D0*(1+ddm_gL)/(ddm_r-ddm_gL) + D0*ddm_h*(year_avg-ddm_gL)/(ddm_r-ddm_gL)
		// log.Printf("Fair Value2(合理價): %f", FV)
		FV_str = "Fair Value(合理價)2"
	}
	return FV_str, FV
}
