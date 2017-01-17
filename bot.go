package main

import (
	"flag"
	"io/ioutil"
	"log"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/hjson/hjson-go"
)

var (
	bot    *tg.BotAPI
	config map[string]interface{}
)

func init() {
	log.Println("[INIT] I'm alive!")

	debug := flag.Bool("debug", false, "enable debug logs")
	flag.Parse()

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

	bot.Debug = *debug // More logs
}

func main() {
	updates := make(<-chan tg.Update)

	upd := tg.NewUpdate(0)
	upd.Timeout = 60
	updates, err := bot.GetUpdatesChan(upd)
	if err != nil {
		log.Printf("[ERROR] Ошибка получения обновлений: %s", err.Error())
	}

	// Updater
	for update := range updates {
		if update.Message != nil {
			go sendEcho(update.Message)
		}
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
