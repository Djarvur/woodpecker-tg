package main

import (
	"fmt"
	"log"
	// "strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	// "github.com/mattn/go-redmine"
)

func messages(msg *tg.Message) {
	token, err := getToken(msg.From.ID)
	if err != nil {
		log.Println(err.Error())
		Start(msg)
		return
	}
	log.Println("MSG TOKEN:", token)

	/*
		if msg.IsCommand() {
			switch strings.ToLower(msg.Command()) {
			case "offset":
				if msg.CommandArguments() != "" {
					if err := ChangeOffset(msg.From.ID, msg.CommandArguments()); err != nil {
						reply := tg.NewMessage(msg.Chat.ID, "I don't understand you offset. Please, use only number.")
						reply.ReplyToMessageID = msg.MessageID
						bot.Send(reply)
						return
					}
				}
				reply := tg.NewMessage(msg.Chat.ID, "Please, use this command with number.")
				reply.ReplyToMessageID = msg.MessageID
				bot.Send(reply)
			case "issues":
				log.Println("====== GET ME ======")
				issues, _ := redmine.NewClient(cfg["endpoint"].(string), token).IssuesByFilter(&redmine.IssueFilter{AssignedToId: "me"})
				for _, issue := range issues {
					log.Println(issue.GetTitle())
				}
			}
		}
	*/
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
