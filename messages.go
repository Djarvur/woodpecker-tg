package main

import (
	"fmt"
	"log"
	// "strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	// "github.com/mattn/go-redmine"
)

func Messages(msg *tg.Message) {
	token, err := GetToken(msg.From.ID)
	if err != nil {
		log.Println(err.Error())
		Start(msg)
		return
	}
	log.Println("MSG TOKEN:", token)
}

func Start(msg *tg.Message) {
	_, off := msg.Time().Zone()
	switch {
	case msg.IsCommand():
		if msg.CommandArguments() != "" {
			usr, err := CreateUser(msg.From.ID, msg.CommandArguments(), off)
			if err != nil {
				reply := tg.NewMessage(msg.Chat.ID, "Invalid token. Try reset token in your profile page and send it again.")
				reply.ReplyToMessageID = msg.MessageID
				bot.Send(reply)
				return
			}
			reply := tg.NewMessage(msg.Chat.ID, fmt.Sprintf("It's all, %s! Just wait notifications! :3", usr.Firstname))
			reply.ReplyToMessageID = msg.MessageID
			reply.ReplyMarkup = tg.ForceReply{false, false}
			bot.Send(reply)
			return
		}
		reply := tg.NewMessage(msg.Chat.ID, "Hello, stranger!\nFor start you need send me your personal token. Go to your profile page, grab it from right sidebar and send it here. Easy!")
		reply.ReplyToMessageID = msg.MessageID
		reply.ReplyMarkup = tg.ForceReply{false, false}
		bot.Send(reply)
	default:
		usr, err := CreateUser(msg.From.ID, msg.Text, off)
		if err != nil {
			reply := tg.NewMessage(msg.Chat.ID, "Invalid token. Try reset token in your profile page and send it again.")
			reply.ReplyToMessageID = msg.MessageID
			bot.Send(reply)
			return
		}
		reply := tg.NewMessage(msg.Chat.ID, fmt.Sprintf("It's all, %s! Just wait notifications! :3", usr.Firstname))
		reply.ReplyToMessageID = msg.MessageID
		reply.ReplyMarkup = tg.ForceReply{false, false}
		bot.Send(reply)
	}
}
