package main

import (
	"fmt"
	"log"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	f "github.com/valyala/fasthttp"
)

var bot *tg.BotAPI

func init() {
	var err error
	bot, err = tg.NewBotAPI(token)
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Print("Authorized as @", bot.Self.UserName)
}

func main() {
	bot.Debug = *debugFlag

	var updates <-chan tg.Update
	updates, err := setUpdates(*webhookFlag)
	if err != nil {
		log.Fatalln(err.Error())
	}

	for update := range updates {
		if update.Message != nil {
			go messages(update.Message)
		}
	}
}

func setUpdates(isWebhook bool) (<-chan tg.Update, error) {
	bot.RemoveWebhook()
	if !isWebhook {
		upd := tg.NewUpdate(0)
		upd.Timeout = 60
		updates, err := bot.GetUpdatesChan(upd)
		if err != nil {
			return nil, err
		}
		return updates, nil
	}

	if _, err := bot.SetWebhook(tg.NewWebhook(fmt.Sprint(set, token))); err != nil {
		return nil, err
	}
	go f.ListenAndServe(serve, nil)
	updates := bot.ListenForWebhook(fmt.Sprint(listen, token))
	return updates, nil
}
