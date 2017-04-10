package main

import (
	"fmt"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/powerman/structlog"
	http "github.com/valyala/fasthttp"
)

var (
	err error
	bot *tg.BotAPI
	log = structlog.New()
)

func main() {
	structlog.DefaultLogger.
		AppendPrefixKeys(structlog.KeyTime, structlog.KeySource).
		SetSuffixKeys(structlog.KeyStack).
		SetTimeValFormat("2006-01-02_15:04:05.999999")

	time.Local = time.UTC

	initConfig()

	initDB()
	defer db.Close()

	bot, err = tg.NewBotAPI(token)
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Print("Authorized as @", bot.Self.UserName)

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
	go http.ListenAndServe(serve, nil)
	updates := bot.ListenForWebhook(fmt.Sprint(listen, token))
	return updates, nil
}
