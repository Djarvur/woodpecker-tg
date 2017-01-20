package main

import (
	"database/sql"
	"fmt"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kirillDanshin/dlog"
)

func messages(msg *tg.Message, db *sql.DB) {
	status, err := checkStatus(msg.Chat.ID)
	if err != nil {
		panic(err.Error())
	}
	dlog.Ln("status:", status)

	if status != "connect" {
		token, err := getToken(msg.Chat.ID)
		if err != nil {
			startCommand(msg)
			return
		}
		if err := checkToken(msg.Chat.ID, token); err != nil {
			startCommand(msg)
			return
		}
	}

	switch {
	case status == "connect":
		startCommand(msg)
	case status == "main":
		getMe(msg)
	}
}

func getMe(msg *tg.Message) {
	token, err := getToken(msg.Chat.ID)
	if err != nil {
		text := "Who am I? D:"
		reply := tg.NewMessage(msg.Chat.ID, text)
		bot.Send(reply)
		return
	}

	user, err := getCurrentUser(config["endpoint"].(string), token)
	if err != nil {
		text := "Who are you? D:"
		reply := tg.NewMessage(msg.Chat.ID, text)
		bot.Send(reply)
		return
	}
	text := fmt.Sprintf("Hello again, %s %s! \\(>^<)/\n\nYou birthday in Redmine: %s;\nYou last visit: %s;", user.Firstname, user.Lastname, user.CreatedOn, user.LatLoginOn)
	reply := tg.NewMessage(msg.Chat.ID, text)
	bot.Send(reply)
	return
}

func startCommand(msg *tg.Message) {
	if msg.IsCommand() == false && msg.Text != "" {
		connectProccess(msg, msg.Text)
		return
	}

	if msg.IsCommand() == true && strings.ToLower(msg.Command()) == "start" {
		if msg.CommandArguments() == "" {
			text := "Hello, stranger!\nFor beginning you need connect me to Redmine. Go to you profile page, find your personal API key and send me it."
			reply := tg.NewMessage(msg.Chat.ID, text)
			reply.ReplyMarkup = tg.ForceReply{true, false}
			bot.Send(reply)
			return
		}
		connectProccess(msg, msg.CommandArguments())
		return
	}
}

func connectProccess(msg *tg.Message, token string) {
	user, err := setToken(msg.Chat.ID, token)
	if err != nil {
		text := "Invalid API key. Be sure what you send actual API key from your Redmine account and try send it again."
		reply := tg.NewMessage(msg.Chat.ID, text)
		reply.ReplyMarkup = tg.ForceReply{true, false}
		bot.Send(reply)
		return
	}
	text := fmt.Sprintf("Authorized as %s %s.", user.Firstname, user.Lastname)
	reply := tg.NewMessage(msg.Chat.ID, text)
	bot.Send(reply)
	return
}
