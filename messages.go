package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kirillDanshin/dlog"
	"github.com/mattn/go-redmine"
)

var (
	user     *redmine.User
	verified = false
)

func messages(msg *tg.Message, db *sql.DB) {
	// Check userID=chatID in DB
	reg, err := db.Prepare(fmt.Sprintf("INSERT IGNORE INTO `users` SET `chat_id` = '%d'", msg.Chat.ID))
	if err != nil {
		panic(err.Error())
	}
	defer reg.Close()
	dlog.F("reg: %#v", reg)

	if msg.IsCommand() {
		commands(msg)
		return
	}

	// TODO: check ReplyToMessage itself
	if msg.ReplyToMessage != nil {
		var err error
		user, err = getCurrentUser(config["webhook_set"].(string), msg.Text)
		if err != nil {
			text := "Invalid API key. Be sure what you send actual API key from your Redmine account and try send it again."
			reply := tg.NewMessage(msg.Chat.ID, text)
			reply.ReplyMarkup = tg.ForceReply{true, false}
			bot.Send(reply)
			return
		}

		verified = true
		text := fmt.Sprintf("Authorized as %s %s.", user.Firstname, user.Lastname)
		reply := tg.NewMessage(msg.Chat.ID, text)
		bot.Send(reply)
		return
	}
}

func commands(msg *tg.Message) {
	// More commands soon
	switch strings.ToLower(msg.Command()) {
	case "start":
		startCommand(msg)
	}
}

func startCommand(msg *tg.Message) {
	var err error
	if msg.CommandArguments() == "" {
		text := "Hello!\nFor beginning you need connect me to Redmine. Go to you profile page, find your personal API key and send me it."
		reply := tg.NewMessage(msg.Chat.ID, text)
		reply.ReplyMarkup = tg.ForceReply{true, false}
		bot.Send(reply)
		return
	}
	user, err = getCurrentUser(config["webhook_set"].(string), msg.CommandArguments())
	if err != nil {
		text := "Invalid API key. Be sure what you send actual API key from your Redmine account and try send it again."
		reply := tg.NewMessage(msg.Chat.ID, text)
		reply.ReplyMarkup = tg.ForceReply{true, false}
		bot.Send(reply)
		return
	}
	verified = true
	text := fmt.Sprintf("Authorized as %s %s.", user.Firstname, user.Lastname)
	reply := tg.NewMessage(msg.Chat.ID, text)
	bot.Send(reply)
}
