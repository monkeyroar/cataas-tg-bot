package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const maxResultsSize = 50
const thumbWidth = 100
const thumbHeight = 100
const textSize = 50

type botConfig struct {
	token   string
	webhook string
	service Service
}

type Bot interface {
	Start()
}

func (config *botConfig) Start() {
	if config.token == "" {
		log.Panic("No Telegram token provided")
	}
	bot, err := tgbotapi.NewBotAPI(config.token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = config.webhook == ""
	log.Printf("Authorized on account %s", bot.Self.UserName)

	var updates tgbotapi.UpdatesChannel
	if bot.Debug {
		updates = startLongPolling(bot)
	} else {
		updates = startWebhook(bot, config.webhook)
	}

	for update := range updates {
		log.Printf("Got update: %+v", update)
		if update.Message != nil {
			msgText := "This bot works only in inline mode"
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
			bot.Send(msg)
		}
		if update.InlineQuery != nil {
			go config.handleInlineQuery(update.InlineQuery, bot)
		}
	}
}

func startLongPolling(bot *tgbotapi.BotAPI) tgbotapi.UpdatesChannel {
	log.Printf("Starting bot in long polling mode")
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	return bot.GetUpdatesChan(updateConfig)
}

func startWebhook(bot *tgbotapi.BotAPI, webhook string) tgbotapi.UpdatesChannel {
	log.Printf("Starting bot in webhook mode")
	log.Printf("Webhook URL: %s", webhook)

	wh, err := tgbotapi.NewWebhook(webhook)
	if err != nil {
		log.Panic(err)
	}

	_, err = bot.Request(wh)
	if err != nil {
		log.Panic(err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Panic(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/")
	go http.ListenAndServe("0.0.0.0:80", nil)
	return updates
}

func (config *botConfig) handleInlineQuery(query *tgbotapi.InlineQuery, bot *tgbotapi.BotAPI) {
	resultsSize := maxResultsSize
	results := make([]interface{}, 0, resultsSize)
	var says *string
	if trim := strings.TrimSpace(query.Query); trim != "" {
		says = &trim
	}
	cataasRequest := CataasRequest{resultsSize, says, textSize, thumbWidth, thumbHeight}
	cataasResponse, err := config.service.GetCats(&cataasRequest)
	if err == nil {
		for _, response := range cataasResponse {
			var pic interface{}
			switch response.photoType {
			case Image:
				img := tgbotapi.NewInlineQueryResultPhotoWithThumb(response.id, response.photoUrl.String(), response.thumbUrl.String())
				img.Width = thumbWidth
				img.Height = thumbHeight
				pic = img
			case Gif:
				gif := tgbotapi.NewInlineQueryResultGIF(response.id, response.photoUrl.String())
				gif.ThumbURL = response.thumbUrl.String()
				gif.Width = thumbWidth
				gif.Height = thumbHeight
				pic = gif
			}
			results = append(results, pic)
		}
	} else {
		errorMsg := tgbotapi.NewInlineQueryResultArticle("1", "Error", err.Error())
		results = append(results, errorMsg)
	}

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		IsPersonal:    true,
		CacheTime:     1,
		Results:       results,
	}

	if _, err := bot.Request(inlineConf); err != nil {
		log.Println(err)
	}
}

func CreateBot(service Service) Bot {
	return &botConfig{os.Getenv("BOT_TOKEN"), os.Getenv("WEBHOOK_URL"), service}
}
