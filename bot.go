package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/hjson/hjson-go"
)

var (
	bot    *tg.BotAPI
	config map[string]interface{}
)

func init() {
	log.Println("[INIT] I'm alive!")

	// Open configuration file and read content
	cfgFile, err := ioutil.ReadFile("config.hjson")
	if err != nil {
		log.Fatalf("[ERROR] Ошибка чтения конфигурации: %s", err)
	}
	if err = hjson.Unmarshal(cfgFile, &config); err != nil {
		log.Fatalf("[ERROR] Ошибка декодирования конфигурации: %s", err.Error())
	}
	log.Println("[CONFIG] Успешно сконфигуровано!")

	// Initialize bot
	bot, err = tg.NewBotAPI(config["token"].(string))
	if err != nil {
		log.Fatalf("[ERROR] Ошибка инициализации бота: %+v", err)
	}
	log.Printf("[BOT] Авторизован как @%s", bot.Self.UserName)

}

func main() {
	debug := flag.Bool("debug", false, "enable debug logs")
	webhook := flag.Bool("webhook", false, "enable debug logs")
	flag.Parse()

	bot.Debug = *debug // More logs

	updates := make(<-chan tg.Update)
	updates, err := setUpdates(*webhook)
	if err != nil {
		log.Fatalf("[ERROR] Ошибка получения обновлений: %s", err.Error())
	}

	// Updater
	for update := range updates {
		if update.Message != nil {
			go sendEcho(update.Message)
		}
	}
}

func setUpdates(isWebhook bool) (<-chan tg.Update, error) {
	bot.RemoveWebhook() // Just in case
	if isWebhook == true {
		if _, err := bot.SetWebhook(
			tg.NewWebhook(
				fmt.Sprint(config["webhook_set"].(string), config["token"].(string)),
			),
		); err != nil {
			return nil, err
		}
		go http.ListenAndServe(config["webhook_serve"].(string), nil)
		updates := bot.ListenForWebhook(
			fmt.Sprint(config["webhook_listen"].(string), config["token"].(string)),
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

func sendEcho(msg *tg.Message) {
	log.Printf("[LOG] %s: %s", msg.From.UserName, msg.Text)

	echo := tg.NewMessage(msg.Chat.ID, msg.Text)
	echo.ReplyToMessageID = msg.MessageID

	if _, err := bot.Send(echo); err != nil {
		log.Printf("[ERROR] Ошибка отправки: %s", err.Error())
	}
}
