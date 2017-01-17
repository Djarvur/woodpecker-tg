package main

import (
	"flag"
	"log"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
)

const token = ""

var bot *tg.BotAPI

func init() {
	debug := flag.Bool("debug", false, "enable debug logs")
	flag.Parse()

	log.Println("[INIT] I'm alive!")

	var err error
	bot, err = tg.NewBotAPI(token)
	if err != nil {
		log.Fatalf("[BOT] Initialize error: %+v", err)
	}
	log.Printf("[BOT] Authorized as @%s", bot.Self.UserName)
	bot.Debug = *debug
}

func main() {
	updates := make(<-chan tg.Update)

	upd := tg.NewUpdate(0)
	upd.Timeout = 60
	updates, err := bot.GetUpdatesChan(upd)
	if err != nil {
		log.Printf("[ERROR][UPDATES] %s", err.Error())
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
		log.Printf("[ERROR][SEND] %s", err.Error())
	}
}
