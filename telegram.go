package main

import (
	"fmt"
	_ "log" // just to safisfy Sublime Go plugin
	"net/url"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
)

func messages(msg *tg.Message) {
	if msg.IsCommand() && strings.ToLower(msg.Command()) == "ping" {
		reply := tg.NewMessage(msg.Chat.ID, "pong")
		reply.ReplyToMessageID = msg.MessageID
		bot.Send(reply)
		return
	}

	usr, err := getUser(msg.From.ID)
	if err != nil {
		log.Println(err.Error())
		start(msg)
		return
	}

	if !msg.IsCommand() {
		task := usr.Task
		text := "Your connection to Redmine is correctly, all right. 👌🏻"
		if task != 0 {
			text += "\n*P.S.:* You last task pinned to this message."
		} else {
			task = -1
		}

		message(usr.Telegram, text, usr.Task)
		return
	}

	switch strings.ToLower(msg.Command()) {
	case "start":
		start(msg)
	case "update":
		update(usr, msg)
	case "skip":
		skip(usr, msg)
	case "last":
		checkIssues(usr)
	}
}

func start(msg *tg.Message) {
	log.Println("====== START COMMAND ======")
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

func update(usr *dbUser, msg *tg.Message) {
	log.Println("====== UPDATE COMMAND ======")
	if msg.CommandArguments() == "" {
		text := fmt.Sprintf("Please, use this command with some text", err.Error())
		message(msg.From.ID, text, -1)
		return
	}

	if err := updateIssue(usr, msg.CommandArguments()); err != nil {
		log.Println(err.Error())
		text := fmt.Sprintf("Commenting process interrupted by the following errors:\n_%s_\nTry repeat this action later, or contact to manager.", err.Error())
		message(msg.From.ID, text, -1)
		return
	}
	text := fmt.Sprintf("Your comment: `%s`\nTo issue #%d has been sent!\nYou are free from it for the next 24 hours.", msg.CommandArguments(), usr.Task)
	message(msg.From.ID, text, usr.Task)
	go changeIssue(usr, 0)

	go checkIssues(usr)
}

func skip(usr *dbUser, msg *tg.Message) {
	log.Println("====== SKIP COMMAND ======")
	if err := updateIssue(usr, "Skipped"); err != nil {
		log.Println(err.Error())
		text := fmt.Sprintf("Commenting process interrupted by the following errors:\n_%s_\nTry repeat this action later, or contact to manager.", err.Error())
		message(msg.From.ID, text, -1)
		return
	}
	text := fmt.Sprintf("Issue #%d has been skipped.\nYou are free from it for the next 24 hours.", usr.Task)
	message(msg.From.ID, text, usr.Task)
	go changeIssue(usr, 0)

	go checkIssues(usr)
}

func message(to int, text string, issue int) {
	log.Println("====== MESSAGE ======")

	issueURL := url.URL{
		Scheme: scheme,
		Host:   endpoint,
		Path:   fmt.Sprint("issues/", issue),
	}

	notify := tg.NewMessage(int64(to), text)
	notify.ParseMode = "markdown"
	if issue != -1 {
		notify.ReplyMarkup = tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonURL(
					fmt.Sprintf("Open issue #%d", issue),
					issueURL.String(),
				),
			),
		)
	}
	if _, err := bot.Send(notify); err != nil {
		log.Println(err.Error())
	}
}
