// package main

package flexmessage

import (
	"github.com/line/line-bot-sdk-go/linebot"
)

func CreateWeatherBubble(cityName, imageURL, formattedDate string, formatInfo string) *linebot.BubbleContainer {
	flexValueForWeather := 2
	flexValueForAddress := 5

	return &linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,
		Hero: &linebot.ImageComponent{
			Type:        linebot.FlexComponentTypeImage,
			URL:         imageURL,
			Size:        linebot.FlexImageSizeTypeFull,
			AspectRatio: linebot.FlexImageAspectRatioType20to13,
			AspectMode:  linebot.FlexImageAspectModeTypeCover,
		},
		Body: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Text:   cityName + "的36小時天氣資訊",
					Weight: linebot.FlexTextWeightTypeRegular,
					Size:   linebot.FlexTextSizeTypeXl,
				},
				&linebot.TextComponent{
					Type:  linebot.FlexComponentTypeText,
					Text:  formattedDate,
					Wrap:  true,
					Color: "#666666",
					Size:  linebot.FlexTextSizeTypeSm,
					Flex:  &flexValueForAddress,
				},
				&linebot.SeparatorComponent{},
				&linebot.TextComponent{
					Type:  linebot.FlexComponentTypeText,
					Text:  "天氣狀況",
					Color: "#aaaaaa",
					Size:  linebot.FlexTextSizeTypeSm,
					Flex:  &flexValueForWeather,
				},
				&linebot.TextComponent{
					Type:  linebot.FlexComponentTypeText,
					Text:  formatInfo,
					Wrap:  true,
					Color: "#666666",
					Size:  linebot.FlexTextSizeTypeSm,
					Flex:  &flexValueForWeather,
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Action: linebot.NewURIAction("詳細內容", "https://www.cwa.gov.tw/V8/C/"),
					Height: "sm",
				},
			},
		},
	}
}
