package stock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ApiResponse struct {
	MsgArray []StockInfo `json:"msgArray"`
	// 其他字段...
}

type StockInfo struct {
	A  string `json:"a"`  //最佳五檔賣出價格
	B  string `json:"b"`  //最佳五檔買入價格
	C  string `json:"c"`  //股票代號/
	D  string `json:"d"`  //最近交易日期
	N  string `json:"n"`  //公司簡稱
	Nf string `json:"nf"` //公司全名
	Z  string `json:"z"`  //最近成交價格
	Tv string `json:"tv"` //當盤成交量
	V  string `json:"v"`  //累積成交量
	// 其他字段...
}

func GetStockPrice(stockCode string, dateTime string) StockInfo {
	url := "https://mis.twse.com.tw/stock/api/getStockInfo.jsp?ex_ch=tse_" + stockCode + ".tw_" + dateTime
	resp, err := http.Get(url)
	if err != nil {
		return StockInfo{}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	jsonData := string(body)

	// Parse JSON data into ApiResponse struct
	var apiResponse ApiResponse
	err = json.Unmarshal([]byte(jsonData), &apiResponse)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return StockInfo{}
	}
	byteResult, _ := json.MarshalIndent(apiResponse, "", "  ")
	fmt.Println(string(byteResult), "byteResult")
	// return apiResponse.MsgArray[0]
	if len(apiResponse.MsgArray) > 0 {
		return apiResponse.MsgArray[0] // 或其他您需要的字段
	} else {
		// 如果 msgArray 是空的，返回適當的錯誤信息或空字符串
		return StockInfo{} // 或者返回空字符串，根據您的需求
	}
}

// IsStockCode 函數用於檢查輸入的文本是否是股票代號
func IsStockCode(input string) bool {
	// 檢查輸入是否全為數字並且長度為 4 或 6
	fmt.Println("input", input)
	return (len(input) == 4 || len(input) == 5) && IsAllDigits(input)
}

// isAllDigits 檢查字符串是否全為數字
func IsAllDigits(str string) bool {
	for _, c := range str {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
