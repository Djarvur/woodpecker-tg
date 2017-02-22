package main

import (
	"fmt"
	"log"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
)

func messages(msg *tg.Message) {
	if _, err := getUser(msg.From.ID); err != nil {
		log.Println(err.Error())
		start(msg)
		return
	}
}

func start(msg *tg.Message) {
	switch {
	case msg.IsCommand():
		if msg.CommandArguments() != "" {
			if _, err := createUser(msg.From.ID, msg.CommandArguments()); err != nil {
				reply := tg.NewMessage(msg.Chat.ID, "Invalid token. Try reset token in your profile page and send it again.")
				reply.ReplyToMessageID = msg.MessageID
				bot.Send(reply)
				return
			}
			reply := tg.NewMessage(msg.Chat.ID, "It's all! Just wait notifications! :3")
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
		if _, err := createUser(msg.From.ID, msg.Text); err != nil {
			reply := tg.NewMessage(msg.Chat.ID, "Invalid token. Try reset token in your profile page and send it again.")
			reply.ReplyToMessageID = msg.MessageID
			bot.Send(reply)
			return
		}
		reply := tg.NewMessage(msg.Chat.ID, "It's all! Just wait notifications! :3")
		reply.ReplyToMessageID = msg.MessageID
		reply.ReplyMarkup = tg.ForceReply{
			ForceReply: false,
			Selective:  false,
		}
		bot.Send(reply)
	}
}

func notification(to int, text string, issue int) {
	notify := tg.NewMessage(int64(to), text)
	notify.ParseMode = "markdown"
	if issue != -1 {
		notify.ReplyMarkup = tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonURL(
					fmt.Sprintf("Open issue #%d", issue),
					fmt.Sprintf("%s/issues/%d", fmt.Sprint(scheme, "://", endpoint), issue),
				),
			),
		)
	}
	if _, err := bot.Send(notify); err != nil {
		log.Println(err.Error())
	}
}
