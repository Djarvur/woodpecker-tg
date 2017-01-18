package main

import (
	"fmt"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mattn/go-redmine"
)

var usr *redmine.User

func messages(msg *tg.Message) {
	if msg.IsCommand() {
		commands(msg)
		return
	}

	if msg.ReplyToMessage != nil {
		var err error
		usr, err = getCurrentUser(config["webhook_set"].(string), msg.Text)
		if err != nil {
			text := "Invalid API key. Be sure what you send actual API key from your Redmine account and try send it again."
			reply := tg.NewMessage(msg.Chat.ID, text)
			reply.ReplyMarkup = tg.ForceReply{true, false}
			bot.Send(reply)
			return
		}
		text := fmt.Sprintf("Authorized as %s %s.", usr.Firstname, usr.Lastname)
		reply := tg.NewMessage(msg.Chat.ID, text)
		bot.Send(reply)
		return
	}
}

func commands(msg *tg.Message) {
	switch strings.ToLower(msg.Command()) {
	case "start":
		startCommand(msg)
	}
}

func startCommand(msg *tg.Message) {
	var err error
	if msg.CommandArguments() != "" {
		usr, err = getCurrentUser(config["webhook_set"].(string), msg.CommandArguments())
		if err != nil {
			text := "Invalid API key. Be sure what you send actual API key from your Redmine account and try send it again."
			reply := tg.NewMessage(msg.Chat.ID, text)
			reply.ReplyMarkup = tg.ForceReply{true, false}
			bot.Send(reply)
			return
		}
		text := fmt.Sprintf("Authorized as %s %s.", usr.Firstname, usr.Lastname)
		reply := tg.NewMessage(msg.Chat.ID, text)
		bot.Send(reply)
		return
	}

	text := "Hello!\nFor beginning you need connect me to Redmine. Go to you profile page, find your personal API key and send me it."
	reply := tg.NewMessage(msg.Chat.ID, text)
	reply.ReplyMarkup = tg.ForceReply{true, false}
	bot.Send(reply)
}
