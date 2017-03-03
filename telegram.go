package main

import (
	"fmt"
	_ "log" // just to safisfy Sublime Go plugin
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
)

func messages(msg *tg.Message) {
	if msg.IsCommand() && strings.ToLower(msg.Command()) == "ping" {
		ping(msg)
		return
	}

	usr, err := getUser(msg.From.ID)
	if err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		if msg.IsCommand() && strings.ToLower(msg.Command()) == "start" {
			start(msg)
			return
		}
		message(msg.From.ID, "It seems with your Redmine token something wrong. Please, send a new token through a `/start [token]` command.")
		return
	}

	if !msg.IsCommand() {
		update(usr, msg)
		return
	}

	switch strings.ToLower(msg.Command()) {
	case "start":
		start(msg)
	case "help":
		text := "`/start [token]` - reconnect account by new token;\n`/help` - show this message;\n`[any text]` - send note with `any text` to last task and go to next;\n`/skip` - skip last task and go to next;\n`/last` - get info about your last task;\n`/ping [telegram|redmine|db]` - get current status of your connection;"
		message(usr.Telegram, text)
	case "skip":
		skip(usr, msg)
	case "last":
		checkIssues(usr)
	}
}

func ping(msg *tg.Message) {
	log.Println("====== PING COMMAND ======")
	switch strings.ToLower(msg.CommandArguments()) {
	case "telegram":
		message(msg.From.ID, "If you see this message, the connection with Telegram *is stable*. ☺️")
	case "redmine":
		if err := pingRedmine(); err != nil {
			log.Println("!!!!!! ERROR !!!!!!")
			log.Println(err.Error())
			message(msg.From.ID, "Connection to Redmine *is not okay*. 😕")
			return
		}
		message(msg.From.ID, "Connection to Redmine *is okay*. ☺️")
	case "db":
		writable, err := pingDB()
		if err != nil {
			log.Println("!!!!!! ERROR !!!!!!")
			log.Println(err.Error())
			message(msg.From.ID, "Connection to BoltDB *is not okay*. 😕")
			return
		}

		if writable {
			message(msg.From.ID, "Connection to BoltDB *is okay*, but *is not writable*. 😕")
			return
		}

		message(msg.From.ID, "Connection to BoltDB *is okay* and *is writable*. ☺️")
	}
}

func start(msg *tg.Message) {
	log.Println("====== START COMMAND ======")
	if msg.CommandArguments() != "" {
		if _, err := createUser(msg.From.ID, msg.CommandArguments()); err != nil {
			log.Println("!!!!!! ERROR !!!!!!")
			log.Println(err.Error())
			message(msg.From.ID, "Invalid token. Try reset token in your profile page and send it again.")
			return
		}
		message(msg.From.ID, "It's all! Just wait notifications! :3")
		return
	}
	message(msg.From.ID, "Hello, stranger!\nFor start you need send me your personal token. Go to your profile page, grab it from right sidebar and send it here. Easy!")
}

func update(usr *dbUser, msg *tg.Message) {
	log.Println("====== UPDATE COMMAND ======")
	if err := updateIssue(usr, msg.Text); err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		text := fmt.Sprintf(
			"Commenting process interrupted by the following errors:\n_%s_\nTry repeat this action later, or contact to manager.",
			err.Error(),
		)
		message(msg.From.ID, text)
		return
	}
	text := fmt.Sprintf(
		"Your comment: `%s`\nTo issue %s has been sent!\nYou are free from it for the next 24 hours.",
		msg.CommandArguments(),
		makeIssueUrl(usr.Task),
	)
	message(msg.From.ID, text)
	go changeIssue(usr, 0)

	go checkIssues(usr)
}

func skip(usr *dbUser, msg *tg.Message) {
	log.Println("====== SKIP COMMAND ======")
	if err := updateIssue(usr, "Skipped"); err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		text := fmt.Sprintf(
			"Commenting process interrupted by the following errors:\n_%s_\nTry repeat this action later, or contact to manager.",
			err.Error(),
		)
		message(msg.From.ID, text)
		return
	}
	text := fmt.Sprintf(
		"Issue %s has been skipped.\nYou are free from it for the next 24 hours.",
		makeIssueUrl(usr.Task),
	)
	message(msg.From.ID, text)
	go changeIssue(usr, 0)

	go checkIssues(usr)
}

func message(to int, text string) {
	log.Println("====== MESSAGE ======")
	notify := tg.NewMessage(int64(to), text)
	notify.ParseMode = "markdown"
	if _, err := bot.Send(notify); err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
	}
}
