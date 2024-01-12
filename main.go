package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"my-go-app/colly"
	flexmessage "my-go-app/flexMessage"
	"my-go-app/format"
	"my-go-app/stock"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
)

// 定義 JSON 結構體以匹配您的數據

// 定義一個結構來存儲每個時間段的天氣資訊
type WeatherPeriod struct {
	StartTime string
	EndTime   string
	MaxT      string
	MinT      string
	PoP       string
}
type WeatherPeriodInfo struct {
	StartTime string
	EndTime   string
	PoP       string // 降雨機率
	MinT      string // 最低溫度
	MaxT      string // 最高溫度
}

type WeatherData struct {
	Records struct {
		Location []struct {
			LocationName   string           `json:"locationName"`
			WeatherElement []WeatherElement `json:"weatherElement"`
		} `json:"location"`
	} `json:"records"`
}

type WeatherElement struct {
	ElementName string `json:"elementName"`
	Time        []struct {
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
		Parameter struct {
			ParameterName string `json:"parameterName"`
			ParameterUnit string `json:"parameterUnit"`
		} `json:"parameter"`
	} `json:"time"`
}

var usdRate, jpyRate string // 全局變量存儲匯率
func fetchRates() {
	url := "https://rate.bot.com.tw/xrt/flcsv/0/day" // 牌告匯率 CSV 網址

	// 發送 HTTP GET 請求
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// 讀取響應體
	reader := csv.NewReader(bufio.NewReader(resp.Body))

	for line := 1; ; line++ {
		var record []string
		record, err = reader.Read()
		if err == io.EOF {
			break
		}
		if strings.Contains(record[0], "USD") {
			usdRate = record[12] // 假設匯率在第三列
		} else if strings.Contains(record[0], "JPY") {
			jpyRate = record[12] // 同上
		}
	}
	fmt.Printf("美元匯率現在是 %s\n", usdRate)
	fmt.Printf("日圓匯率現在是 %s\n", jpyRate)
}

type ElementHandler func(element WeatherElement, periods map[string]*WeatherPeriod)

var elementHandlers = map[string]ElementHandler{
	"PoP":  handlePoP,
	"MinT": handleMinT,
	"MaxT": handleMaxT,
}

// 修改處理函數來更新 WeatherPeriod 結構的對應字段
func handleMaxT(element WeatherElement, periods map[string]*WeatherPeriod) {
	for _, timePeriod := range element.Time {
		startTime := format.FormatTime(timePeriod.StartTime)
		if period, ok := periods[startTime]; ok {
			period.MaxT = timePeriod.Parameter.ParameterName
		}
	}
}

func handleMinT(element WeatherElement, periods map[string]*WeatherPeriod) {
	for _, timePeriod := range element.Time {
		startTime := format.FormatTime(timePeriod.StartTime)
		if period, ok := periods[startTime]; ok {
			period.MinT = timePeriod.Parameter.ParameterName
		}
	}
}

func handlePoP(element WeatherElement, periods map[string]*WeatherPeriod) {
	for _, timePeriod := range element.Time {
		startTime := format.FormatTime(timePeriod.StartTime)
		if period, ok := periods[startTime]; ok {
			period.PoP = timePeriod.Parameter.ParameterName
		}
	}

}

func fetchWeatherInfo(cityName string) ([]WeatherPeriodInfo, error) {
	//讀取.env
	err := godotenv.Load()
	apiKey := os.Getenv("WEATHER_API_KEY")
	url := "https://opendata.cwa.gov.tw/api/v1/rest/datastore/F-C0032-001?Authorization=" + apiKey + "&locationName=" + cityName
	// 發送 HTTP GET 請求
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	jsonData := string(body)

	// 解析 JSON 數據 把json轉成struct
	var data WeatherData
	err = json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}
	periods := make(map[string]*WeatherPeriod)
	for _, location := range data.Records.Location {
		if location.LocationName == cityName {
			for _, element := range location.WeatherElement {
				for _, timePeriod := range element.Time {
					startTime := format.FormatTime(timePeriod.StartTime)
					if _, ok := periods[startTime]; !ok {
						periods[startTime] = &WeatherPeriod{StartTime: startTime, EndTime: format.FormatTime(timePeriod.EndTime)}
					}
				}

				if handler, ok := elementHandlers[element.ElementName]; ok {
					handler(element, periods)
				}
			}
		}
	}

	var weatherInfos []WeatherPeriodInfo
	for _, period := range periods {
		info := WeatherPeriodInfo{
			StartTime: format.FormatToCh(period.StartTime),
			EndTime:   format.FormatToCh(period.EndTime),
			PoP:       period.PoP,
			MinT:      period.MinT,
			MaxT:      period.MaxT,
		}
		weatherInfos = append(weatherInfos, info)
	}
	fmt.Printf("%+v\n", weatherInfos)
	return weatherInfos, nil

	// return result.String(), nil
}

func main() {
	// stock.GetStockPrice("2330", "20240112")

	// 加載 .env 文件
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// 設置Line Bot
	bot, err := linebot.New(
		os.Getenv("LINE_BOT_CHANNEL_SECRET"),
		os.Getenv("LINE_BOT_CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// 創建一個Gin路由器
	router := gin.Default()

	// 設置Webhook路由
	router.POST("/callback", func(c *gin.Context) {
		events, err := bot.ParseRequest(c.Request)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			}
			return
		}

		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					var reply string
					inputText := strings.ToLower(message.Text)
					switch {
					case inputText == "美金":
						fetchRates()
						reply = fmt.Sprintf("美元匯率現在是 %s", usdRate)
					case inputText == "日圓":
						fetchRates()
						reply = fmt.Sprintf("日圓匯率現在是 %s", jpyRate)
					case inputText == "天氣":
						// 使用 fetchWeatherInfo 函數從 API 獲取天氣資訊
						cityName := "臺北市"
						weatherInfo, err := fetchWeatherInfo(cityName)
						// weatherInfo, err := fetchWeatherInfo("臺北市")
						if err != nil {
							log.Printf("Error fetching weather info: %v", err)
							reply = "抱歉，無法獲取天氣資訊。"
						} else {

							flexMessage := createWeatherFlexMessage(weatherInfo, cityName)
							// 發送 Flex Message 回覆給用戶
							if _, err = bot.ReplyMessage(event.ReplyToken, flexMessage).Do(); err != nil {
								log.Print(err)
							}
						}
					case stock.IsStockCode(inputText):
						// 獲取當前日期和時間
						currentTime := time.Now()
						// 格式化日期為 YYYYMMDD
						formattedDate := currentTime.Format("20060102")

						fmt.Println(inputText, formattedDate)
						// 調用獲取股票資訊的函數
						stockInfo := stock.GetStockPrice(inputText, formattedDate)
						fmt.Println(stockInfo, "回傳的股票訊息")
						if stockInfo == (stock.StockInfo{}) {
							log.Printf("Error fetching stock info: no data returned")
							reply = "抱歉，無法獲取股票資訊。"
						} else {
							reply = fmt.Sprintf("%s 的當日股價📈: %s\n成交量🤝:%s", stockInfo.N, stockInfo.Z, stockInfo.V)
						}
					case strings.Contains(inputText, "餐廳"):
						// 提取關鍵字
						keyword := ""
						splitText := strings.Split(inputText, "餐廳")
						fmt.Println(splitText, "splitText")
						if len(splitText) > 1 {
							keyword = strings.TrimSpace(splitText[0])
						}
						// result := colly.GetRestaurantInfo("https://ifoodie.tw/explore/%E5%8F%B0%E4%B8%AD%E5%B8%82/list?sortby=popular&opening=true&sortby=rating")
						url := "https://ifoodie.tw/explore/" + keyword + "/list?sortby=popular&opening=true&sortby=rating"
						fmt.Printf("url: %s\n", url, keyword)
						restaurants := colly.GetRestaurantInfo(url)
						// fmt.Println(restaurants, "restaurants")
						// 處理餐廳資訊，例如格式化為一條消息

						for _, restaurant := range restaurants {
							// 格式化消息並發送給 Line Bot
							// ...
							fmt.Println(restaurant, "哈囉")
							// 将每家餐厅的信息添加到回复字符串中
							reply += fmt.Sprintf("%s\n地址📍: %s\n評分⭐️: %.1f\n\n", restaurant.Name, restaurant.Address, restaurant.Star)
						}

						if reply == "" {
							reply = "抱歉，無法獲取餐廳資訊。"
						}

						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
							log.Print(err)
						}

					default:
						reply = message.Text // 或其他預設回應
					}
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
						log.Print(err)
					}
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	})

	// 運行伺服器
	router.Run(":80")
}

func createWeatherFlexMessage(weatherInfo []WeatherPeriodInfo, cityName string) *linebot.FlexMessage {
	// 創建一個 BubbleContainer 作為 Flex Message 的主要內容
	// 檢查切片是否至少有一個元素
	var formattedDate string
	var formattedDate2 string
	var formattedDate3 string
	var formattedInfo string
	var formattedInfo2 string
	var formattedInfo3 string

	if len(weatherInfo) > 0 {
		firstInfo := weatherInfo[0]
		secondInfo := weatherInfo[1]
		thirdInfo := weatherInfo[2]
		// 格式化第一個物件的資訊
		formattedDate = fmt.Sprintf("%s~%s", firstInfo.StartTime, firstInfo.EndTime)
		formattedDate2 = fmt.Sprintf("%s~%s", secondInfo.StartTime, secondInfo.EndTime)
		formattedDate3 = fmt.Sprintf("%s~%s", thirdInfo.StartTime, thirdInfo.EndTime)
		formattedInfo = fmt.Sprintf("降雨機率%s 溫度 %s~%s°C", firstInfo.PoP, firstInfo.MinT, firstInfo.MaxT)
		formattedInfo2 = fmt.Sprintf("降雨機率%s 溫度 %s~%s°C", secondInfo.PoP, secondInfo.MinT, secondInfo.MaxT)
		formattedInfo3 = fmt.Sprintf("降雨機率%s 溫度 %s~%s°C", thirdInfo.PoP, thirdInfo.MinT, thirdInfo.MaxT)
	} else {
		// 若切片是空的，可以設置一個預設值或錯誤信息
		formattedDate = "無可用天氣資訊"
		formattedDate2 = "無可用天氣資訊"
		formattedDate3 = "無可用天氣資訊"
	}

	var bubbles []*linebot.BubbleContainer

	imageURL := "https://images.unsplash.com/photo-1463947628408-f8581a2f4aca?q=80&w=2070&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D"
	bubble1 := flexmessage.CreateWeatherBubble(cityName, imageURL, formattedDate, formattedInfo)
	bubble2 := flexmessage.CreateWeatherBubble(cityName, imageURL, formattedDate2, formattedInfo2)
	bubble3 := flexmessage.CreateWeatherBubble(cityName, imageURL, formattedDate3, formattedInfo3)
	// ...
	bubbles = []*linebot.BubbleContainer{bubble1, bubble2, bubble3}
	carousel := linebot.CarouselContainer{
		Type:     linebot.FlexContainerTypeCarousel,
		Contents: bubbles,
	}
	// 創建一個 Flex Message 並返回
	flexMessage := linebot.NewFlexMessage("天氣資訊", &carousel)
	return flexMessage
}
