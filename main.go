package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
)

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

func main() {
	//rate
	fetchRates()
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
					if strings.ToLower(message.Text) == "美金" {
						reply = fmt.Sprintf("美元匯率現在是 %s", usdRate)
					} else if strings.ToLower(message.Text) == "日圓" {
						reply = fmt.Sprintf("日圓匯率現在是 %s", jpyRate)
					} else {
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
