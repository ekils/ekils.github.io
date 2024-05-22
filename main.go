package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"linebot-gemini-pro/cronjobs"
	"linebot-gemini-pro/operations"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

var (
	bot                   *messaging_api.MessagingApiAPI
	api_adress            string
	channelsecret         string
	channelToken          string
	err                   error
	filePath              string
	filePath_DDM          string
	invoke                string
	response              Response
	cronjob_hoursetting   string
	cronjob_minutesetting string
)

type Response struct {
	Message string `json:"message"`
	Price   string `json:"price"`
}

func init() {
	errors := godotenv.Load(".env")
	if errors != nil {
		log.Fatalf("Error loading .env file")
	}
	api_adress = os.Getenv("API_ADRESS")
	channelsecret = os.Getenv("ChannelSecret")
	channelToken = os.Getenv("ChannelAccessToken")
	filePath = os.Getenv("FlexMessagePath")
	filePath_DDM = os.Getenv("FlexMessagePathDDM")
	cronjob_hoursetting = os.Getenv("CronjobHourSetting")
	cronjob_minutesetting = os.Getenv("CronjobMinutesSetting")
	invoke = ""

}

func main() {

	// 這裡放置主程式的其他邏輯
	// 因為上面的 Goroutine 已經開始執行了，所以這裡可以繼續執行其他操作
	// 如果有需要，你可以在這裡等待 Goroutine 的結束
	// 等待程式結束
	bot, err = messaging_api.NewMessagingApiAPI(channelToken)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/get/inq", func(w http.ResponseWriter, r *http.Request) {
		operations.Inguiry(w, r)
	})
	http.HandleFunc("/post/subs", func(w http.ResponseWriter, r *http.Request) {
		operations.Subscription(w, r)
	})
	http.HandleFunc("/del/unsubs", func(w http.ResponseWriter, r *http.Request) {
		operations.UnSubscription(w, r)
	})
	http.HandleFunc("/post/addeps", func(w http.ResponseWriter, r *http.Request) {
		operations.CollectEPS(w, r)
	})
	http.HandleFunc("/post/deleps", func(w http.ResponseWriter, r *http.Request) {
		operations.DeleteEPS(w, r)
	})
	http.HandleFunc("/post/adddivend", func(w http.ResponseWriter, r *http.Request) {
		operations.AddDividend(w, r)
	})
	http.HandleFunc("/post/deldivend", func(w http.ResponseWriter, r *http.Request) {
		operations.DelDividend(w, r)
	})

	http.HandleFunc("/get/queryDDM", func(w http.ResponseWriter, r *http.Request) {
		operations.QueryDDM(w, r)
	})

	http.HandleFunc("/post/question", func(w http.ResponseWriter, r *http.Request) {
		operations.Question(w, r)
	})
	http.HandleFunc("/post/ddmquestion", func(w http.ResponseWriter, r *http.Request) {
		operations.DDMQuestion(w, r)
	})

	port := os.Getenv("PORT")
	addr := ":" + port
	log.Println("Listen and Serve on " + addr)

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	// 設定時間
	now := time.Now()

	// hour_setting, _ := strconv.Atoi(cronjob_hoursetting)
	// minutes_setting, _ := strconv.Atoi(cronjob_minutesetting)
	hour_setting := now.Hour()      //測試用:  time.Minute * 2
	minutes_setting := now.Minute() //測試用:  time.Minute * 2

	// 計算下一次執行時間
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), hour_setting, minutes_setting, 0, 0, now.Location()).Add(time.Minute)
	if now.After(nextRun) {
		// nextRun = nextRun.Add(time.Hour * 24)
		nextRun = nextRun.Add(time.Minute * 1) //測試用
	}

	// 循環執行
	for {
		waitDuration := time.Until(nextRun)
		fmt.Println("Waiting for", waitDuration)
		time.Sleep(waitDuration)
		// 執行 cron job
		fmt.Println("Executing cron job...")
		cronjobs.CronJobs()
		// 計算下一次執行時間
		// nextRun = nextRun.Add(time.Hour * 24)
		nextRun = nextRun.Add(time.Minute * 60) //測試用
	}

}

func callbackHandler(w http.ResponseWriter, r *http.Request) {

	jsonString, _ := os.ReadFile(filePath)
	jsonString_DDM, _ := os.ReadFile(filePath_DDM)
	events, err := webhook.ParseRequest(channelsecret, r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	for _, event := range events.Events {
		switch e := event.(type) {
		case webhook.MessageEvent:
			switch message := e.Message.(type) {
			case webhook.TextMessageContent:
				demo := message.Text + invoke
				if strings.Contains(demo, "[點選功能]") || strings.Contains(demo, "請輸入查詢股票代號:") || strings.Contains(demo, "更新EPS選項:") || strings.Contains(demo, "更新Subs選項:") || strings.Contains(demo, "更新BVPS選項:") {
					demo = message.Text
					log.Println("demo:", demo)
				} else {
					log.Println("demo:(message.Text + invoke)", demo)
				}
				switch demo {
				case "[點選功能]":
					// Unmarshal JSON
					flexContainer, err := messaging_api.UnmarshalFlexContainer([]byte(jsonString))
					if err != nil {
						log.Println(err)
					}
					if _, err := bot.ReplyMessage(
						&messaging_api.ReplyMessageRequest{
							ReplyToken: e.ReplyToken,
							Messages: []messaging_api.MessageInterface{
								&messaging_api.FlexMessage{
									AltText:  "Flex message alt text",
									Contents: flexContainer,
								},
							},
						},
					); err != nil {
						log.Println(err)
					}
				case "更新EPS選項:", "更新Subs選項:", "更新BVPS選項:":
					if message.Text == "更新EPS選項:" {
						invoke = "$"
					} else if message.Text == "更新Subs選項:" {
						invoke = "#"
					} else if message.Text == "更新BVPS選項:" {
						invoke = "&"
					}
					msg := templete_msg()
					if _, err := bot.ReplyMessage(
						&messaging_api.ReplyMessageRequest{
							ReplyToken: e.ReplyToken,
							Messages:   []messaging_api.MessageInterface{msg},
						},
					); err != nil {
						log.Println(err)
					}
				case "請輸入查詢股票代號:":
					invoke = demo
				case "[股息投資]":
					flexContainer, err := messaging_api.UnmarshalFlexContainer([]byte(jsonString_DDM))
					if err != nil {
						log.Println(err)
					}
					if _, err := bot.ReplyMessage(
						&messaging_api.ReplyMessageRequest{
							ReplyToken: e.ReplyToken,
							Messages: []messaging_api.MessageInterface{
								&messaging_api.FlexMessage{
									AltText:  "Flex message alt text",
									Contents: flexContainer,
								},
							},
						},
					); err != nil {
						log.Println(err)
					}

				case "更新Dividend":
					invoke = "Yahoo"
					msg := templete_msg()
					if _, err := bot.ReplyMessage(
						&messaging_api.ReplyMessageRequest{
							ReplyToken: e.ReplyToken,
							Messages:   []messaging_api.MessageInterface{msg},
						},
					); err != nil {
						log.Println(err)
					}
				case "+Yahoo":
					SimpleText(e.ReplyToken, demo+": 請輸入股票代碼後換行再貼上 Dividend資料 ⬇")
					invoke = "+Yahoo_with_data"

				case "-Yahoo":
					SimpleText(e.ReplyToken, demo+": 請輸要刪除的 Dividend 股票代碼 ⬇")
					invoke = "-Yahoo_with_data"

				case "查詢 Dividend":
					SimpleText(e.ReplyToken, "https://finance.yahoo.com/lookup")
					invoke = ""

				case "PB 歷史":
					SimpleText(e.ReplyToken, "https://www.macrotrends.net/stocks/charts/JPM/jpmorgan-chase/price-book")
					invoke = ""

				case "查詢配息公司股價":
					SimpleText(e.ReplyToken, "請輸入查詢配息公司的股票代號:")
					invoke = "請輸入查詢股票代號:"

				case "合理估值":
					SimpleText(e.ReplyToken, "請輸入公司股票代號以利估值:")
					invoke = "估值"

				case "*$":
					var temp_slice []string
					resp := CommunicateWithAPI(message.Text, api_adress, "post", "question")
					log.Println("resp:", resp)

					json.Unmarshal([]byte(resp), &response)
					ret := response.Message + "\n"
					temp_slice = append(temp_slice, ret)
					ret = strings.Join(temp_slice, "")
					if err := SimpleText(e.ReplyToken, ret); err != nil {
						log.Print(err)
					}

				case "*#":
					var temp_slice []string
					resp := CommunicateWithAPI(message.Text, api_adress, "post", "question")
					log.Println("resp:", resp)

					json.Unmarshal([]byte(resp), &response)
					ret := response.Message + "\n"
					temp_slice = append(temp_slice, ret)
					ret = strings.Join(temp_slice, "")
					if err := SimpleText(e.ReplyToken, ret); err != nil {
						log.Print(err)
					}
				case "*Yahoo":
					var temp_slice []string
					resp := CommunicateWithAPI(message.Text, api_adress, "post", "ddmquestion")
					log.Println("resp:", resp)

					json.Unmarshal([]byte(resp), &response)
					ret := response.Message + "\n"
					temp_slice = append(temp_slice, ret)
					ret = strings.Join(temp_slice, "")
					if err := SimpleText(e.ReplyToken, ret); err != nil {
						log.Print(err)
					}
					invoke = ""

				default:
					switch invoke {
					case "請輸入查詢股票代號:":
						var temp_slice []string
						resp := CommunicateWithAPI(message.Text, api_adress, "get", "inq")
						log.Println(resp)

						json.Unmarshal([]byte(resp), &response)
						ret := response.Message + "  股價：" + response.Price + "\n"
						temp_slice = append(temp_slice, ret)
						ret = strings.Join(temp_slice, "")
						if err := SimpleText(e.ReplyToken, ret); err != nil {
							log.Print(err)
						}
						invoke = ""

					case "$":
						if message.Text == "+" {
							invoke = "$+"
							SimpleText(e.ReplyToken, "請輸入股票代碼後換行再貼上EPS資料(2015第一季開始)⬇︎")
						} else if message.Text == "-" {
							invoke = "$-"
							SimpleText(e.ReplyToken, "請輸入要移除EPS資料的股票代碼 ⬇︎")
						} else if message.Text == "[股息投資]" {
							invoke = ""
							flexContainer, err := messaging_api.UnmarshalFlexContainer([]byte(jsonString_DDM))
							if err != nil {
								log.Println(err)
							}
							if _, err := bot.ReplyMessage(
								&messaging_api.ReplyMessageRequest{
									ReplyToken: e.ReplyToken,
									Messages: []messaging_api.MessageInterface{
										&messaging_api.FlexMessage{
											AltText:  "Flex message alt text",
											Contents: flexContainer,
										},
									},
								},
							); err != nil {
								log.Println(err)
							}
						} else {
							invoke = ""
							SimpleText(e.ReplyToken, "輸入錯誤,輪詢結束")
						}

					case "#":
						if message.Text == "+" {
							invoke = "#+"
							SimpleText(e.ReplyToken, "請輸入要訂閱的股票 ⬇︎")
						} else if message.Text == "-" {
							invoke = "#-"
							SimpleText(e.ReplyToken, "請輸入要取消訂閱的股票 ⬇︎")
						} else if message.Text == "[股息投資]" {
							invoke = ""
							flexContainer, err := messaging_api.UnmarshalFlexContainer([]byte(jsonString_DDM))
							if err != nil {
								log.Println(err)
							}
							if _, err := bot.ReplyMessage(
								&messaging_api.ReplyMessageRequest{
									ReplyToken: e.ReplyToken,
									Messages: []messaging_api.MessageInterface{
										&messaging_api.FlexMessage{
											AltText:  "Flex message alt text",
											Contents: flexContainer,
										},
									},
								},
							); err != nil {
								log.Println(err)
							}
						} else {
							invoke = ""
							SimpleText(e.ReplyToken, "輸入錯誤,輪詢結束")
						}

					case "&":
						if message.Text == "+" {
							SimpleText(e.ReplyToken, "&+")
						} else if message.Text == "-" {
							SimpleText(e.ReplyToken, "&-")
						}

					case "#+":
						var temp_slice []string
						resp := CommunicateWithAPI(message.Text, api_adress, "post", "subs")
						log.Println("resp:", resp)

						json.Unmarshal([]byte(resp), &response)
						ret := response.Message + "\n"
						temp_slice = append(temp_slice, ret)
						ret = strings.Join(temp_slice, "")
						if err := SimpleText(e.ReplyToken, ret); err != nil {
							log.Print(err)
						}
						invoke = ""

					case "#-":
						var temp_slice []string
						resp := CommunicateWithAPI(message.Text, api_adress, "del", "unsubs")
						log.Println("resp:", resp)

						json.Unmarshal([]byte(resp), &response)
						ret := response.Message + "\n"
						temp_slice = append(temp_slice, ret)
						ret = strings.Join(temp_slice, "")
						if err := SimpleText(e.ReplyToken, ret); err != nil {
							log.Print(err)
						}
						invoke = ""
					case "$+":
						var temp_slice []string
						resp := CommunicateWithAPI(message.Text, api_adress, "post", "addeps")
						log.Println("resp:", resp)

						json.Unmarshal([]byte(resp), &response)
						ret := response.Message + "\n"
						temp_slice = append(temp_slice, ret)
						ret = strings.Join(temp_slice, "")
						if err := SimpleText(e.ReplyToken, ret); err != nil {
							log.Print(err)
						}
						invoke = ""

					case "$-":
						var temp_slice []string
						resp := CommunicateWithAPI(message.Text, api_adress, "post", "deleps")
						log.Println("resp:", resp)

						json.Unmarshal([]byte(resp), &response)
						ret := response.Message + "\n"
						temp_slice = append(temp_slice, ret)
						ret = strings.Join(temp_slice, "")
						if err := SimpleText(e.ReplyToken, ret); err != nil {
							log.Print(err)
						}

					case "+Yahoo_with_data":
						var temp_slice []string
						invoke = ""
						resp := CommunicateWithAPI(message.Text, api_adress, "post", "adddivend")
						log.Println("resp:", resp)

						json.Unmarshal([]byte(resp), &response)
						ret := response.Message + "\n"
						temp_slice = append(temp_slice, ret)
						ret = strings.Join(temp_slice, "")
						if err := SimpleText(e.ReplyToken, ret); err != nil {
							log.Print(err)
						}

					case "-Yahoo_with_data":
						var temp_slice []string
						invoke = ""
						resp := CommunicateWithAPI(message.Text, api_adress, "post", "deldivend")
						log.Println("resp:", resp)

						json.Unmarshal([]byte(resp), &response)
						ret := response.Message + "\n"
						temp_slice = append(temp_slice, ret)
						ret = strings.Join(temp_slice, "")
						if err := SimpleText(e.ReplyToken, ret); err != nil {
							log.Print(err)
						}

					case "估值":
						invoke = ""
						resp := CommunicateWithAPI(message.Text, api_adress, "get", "queryDDM")
						var response Response
						err := json.Unmarshal([]byte(resp), &response)
						if err != nil {
							log.Println("DDM解码 JSON 失败:", err)
							return
						}
						SimpleText(e.ReplyToken, response.Message)

					default:
						SimpleText(e.ReplyToken, demo+": 前面輪詢已結束或並無包含查詢關鍵字")
						invoke = ""
					}
				}

			}
		}
	}
}

func CommunicateWithAPI(demo string, api_adress string, crud string, apipath string) string {
	url := api_adress + "/" + crud + "/" + apipath
	var jsonStr = []byte(demo)
	crud = strings.ToUpper(crud)
	req, _ := http.NewRequest(crud, url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("Response Status: %s\n", resp.Status)
	log.Printf("Response Body: %s\n", string(body))
	return string(body)
}

func SimpleText(replyToken, text string) error {
	if _, err := bot.ReplyMessage(
		&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				&messaging_api.TextMessage{
					Text: text,
				},
			},
		},
	); err != nil {
		return err
	}
	return nil
}

func templete_msg() messaging_api.TextMessage {
	msg := &messaging_api.TextMessage{
		Text: "請點選下列按鈕:",
		QuickReply: &messaging_api.QuickReply{
			Items: []messaging_api.QuickReplyItem{
				{
					ImageUrl: "https://github.com/ekils/JUSTupload/blob/main/images/plus.png?raw=true",
					Action: &messaging_api.MessageAction{
						Label: "增加",
						Text:  "+",
					},
				},
				{
					ImageUrl: "https://github.com/ekils/JUSTupload/blob/main/images/minus.png?raw=true",
					Action: &messaging_api.MessageAction{
						Label: "移除",
						Text:  "-",
					},
				},
				{
					ImageUrl: "https://github.com/ekils/JUSTupload/blob/main/images/question.png?raw=true",
					Action: &messaging_api.MessageAction{
						Label: "查詢",
						Text:  "*",
					},
				},
			},
		},
	}
	return *msg
}
