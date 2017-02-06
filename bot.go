package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/hjson/hjson-go"
	log "github.com/kirillDanshin/dlog"
	f "github.com/valyala/fasthttp"
)

var (
	bot *tg.BotAPI
	cfg map[string]interface{}
)

func init() {
	// Open configuration file and read content
	config, err := ioutil.ReadFile("config.hjson")
	if err != nil {
		panic(err.Error())
	}
	if err := hjson.Unmarshal(config, &cfg); err != nil {
		panic(err.Error())
	}

	// Initialize bot
	bot, err = tg.NewBotAPI(cfg["token"].(string))
	if err != nil {
		panic(err.Error())
	}
	log.F("[BOT] Authorized as @%s\n", bot.Self.UserName)

}

func main() {
	debug := flag.Bool("debug", false, "enable debug logs")
	webhook := flag.Bool("webhook", false, "enable debug logs")
	flag.Parse()

	bot.Debug = *debug // More logs

	updates := make(<-chan tg.Update)
	updates, err := SetUpdates(*webhook)
	if err != nil {
		panic(err.Error())
	}

	// Updater
	for update := range updates {
		if update.Message != nil {
			go Messages(update.Message)
		}
	}
}

func SetUpdates(isWebhook bool) (<-chan tg.Update, error) {
	bot.RemoveWebhook() // Just in case
	if isWebhook == true {
		if _, err := bot.SetWebhook(
			tg.NewWebhook(
				fmt.Sprint(cfg["webhook_set"].(string), cfg["token"].(string)),
			),
		); err != nil {
			return nil, err
		}
		go f.ListenAndServe(cfg["webhook_serve"].(string), nil)
		updates := bot.ListenForWebhook(
			fmt.Sprint(cfg["webhook_listen"].(string), cfg["token"].(string)),
		)
		return updates, nil
	} else {
		upd := tg.NewUpdate(0)
		upd.Timeout = 60
		updates, err := bot.GetUpdatesChan(upd)
		if err != nil {
			return nil, err
		}
		return updates, nil
	}
}
