package main

import (
	"fmt"
	"log"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
)

func messages(msg *tg.Message) {
	token, err := getToken(msg.From.ID)
	if err != nil {
		log.Println(err.Error())
		start(msg)
		return
	}
	log.Println("MSG TOKEN:", token)
}

func start(msg *tg.Message) {
	_, off := msg.Time().Zone()
	switch {
	case msg.IsCommand():
		if msg.CommandArguments() != "" {
			usr, err := createUser(msg.From.ID, msg.CommandArguments(), off)
			if err != nil {
				reply := tg.NewMessage(msg.Chat.ID, "Invalid token. Try reset token in your profile page and send it again.")
				reply.ReplyToMessageID = msg.MessageID
				bot.Send(reply)
				return
			}
			reply := tg.NewMessage(msg.Chat.ID, fmt.Sprintf("It's all, %s! Just wait notifications! :3", usr.Firstname))
			reply.ReplyToMessageID = msg.MessageID
			reply.ReplyMarkup = tg.ForceReply{
				ForceReply: false,
				Selective:  false,
			}
			bot.Send(reply)
			return
		}
		reply := tg.NewMessage(msg.Chat.ID, "Hello, stranger!\nFor start you need send me your personal token. Go to your profile page, grab it from right sidebar and send it here. Easy!")
		reply.ReplyToMessageID = msg.MessageID
		reply.ReplyMarkup = tg.ForceReply{
			ForceReply: false,
			Selective:  false,
		}
		bot.Send(reply)
	default:
		usr, err := createUser(msg.From.ID, msg.Text, off)
		if err != nil {
			reply := tg.NewMessage(msg.Chat.ID, "Invalid token. Try reset token in your profile page and send it again.")
			reply.ReplyToMessageID = msg.MessageID
			bot.Send(reply)
			return
		}
		reply := tg.NewMessage(msg.Chat.ID, fmt.Sprintf("It's all, %s! Just wait notifications! :3", usr.Firstname))
		reply.ReplyToMessageID = msg.MessageID
		reply.ReplyMarkup = tg.ForceReply{
			ForceReply: false,
			Selective:  false,
		}
		bot.Send(reply)
	}
}
