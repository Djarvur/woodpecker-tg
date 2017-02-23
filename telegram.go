package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
)

func messages(msg *tg.Message) {
	usr, err := getUser(msg.From.ID)
	if err != nil {
		log.Println(err.Error())
		start(msg)
		return
	}

	if !msg.IsCommand() {
		return
	}

	switch strings.ToLower(msg.Command()) {
	case "issue":
		args := strings.SplitN(msg.CommandArguments(), " ", 2)

		if !strings.HasPrefix(args[0], "#") {
			text := "Please, use this command with issue id (`/issue #123`)."
			message(msg.From.ID, text, -1)
			return
		}
		iIDs := strings.TrimPrefix(args[0], "#")

		iID, err := strconv.Atoi(iIDs)
		if err != nil {
			text := "Issue ID must be as intenfer (`0-9999...`)."
			message(msg.From.ID, text, -1)
			return
		}

		if len(args) <= 1 {
			text := "Please, use this command with issue id *AND* text of comment (`/issue #123 Sample text`)."
			message(msg.From.ID, text, -1)
			return
		}

		note := args[1]
		log.Println("id:", iID)
		log.Println("note:", note)
		if err := updateIssue(usr, iID, note); err != nil {
			log.Println(err.Error())
			text := fmt.Sprintf("Commenting process interrupted by the following errors:\n_%s_\n\nTry repeat this action later, or contact to manager.", err.Error())
			message(msg.From.ID, text, -1)
			return
		}
		text := fmt.Sprintf("Your comment: `%s`\nTo issue #%d has been sent!\n\nYou are free from it for the next 24 hours.", note, iID)
		message(msg.From.ID, text, iID)
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

func message(to int, text string, issue int) {
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
