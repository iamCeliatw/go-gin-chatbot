package colly

import (
	// "fmt"
	// "encoding/json"
	// "fmt"
	"strconv"

	"github.com/gocolly/colly"
)

type Restaurant struct {
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Star    float64 `json:"star"`
}

func GetRestaurantInfo(url string) []Restaurant {
	var restaurants []Restaurant
	var currentRestaurant Restaurant
	count := 0

	c := colly.NewCollector()

	c.OnHTML(".jsx-1309326380 .info-rows", func(e *colly.HTMLElement) {
		if len(restaurants) >= 5 {
			return // 只處理前五家
		}

		currentRestaurant.Name = e.ChildText(".jsx-1309326380 .title-text")
		currentRestaurant.Star, _ = strconv.ParseFloat(e.ChildText(".jsx-2373119553 .text"), 64)
		currentRestaurant.Address = e.ChildText(".jsx-1309326380 .address-row")

		restaurants = append(restaurants, currentRestaurant)

	})

	c.Visit(url)
	c.Wait() // 等待所有请求完成

	if count > 5 {
		restaurants = restaurants[:5] // 只返回前五家餐厅
	}

	return restaurants
}
