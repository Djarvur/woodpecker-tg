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
		message(usr.Telegram, "Please, use valid commands from /help commands list.")
		return
	}

	switch strings.ToLower(msg.Command()) {
	case "start":
		start(msg)
	case "help":
		text := "`/start [token]` - reconnect account by new token;\n`/help` - show this message;\n`/update [hours] [any text]` - send note with `any text` to last task and go to next;\n`/skip` - skip last task and go to next;\n`/last` - get info about your last task;\n`/ping [telegram|redmine|db]` - get current status of your connection;"
		message(usr.Telegram, text)
	case "skip":
		skip(usr, msg)
	case "last":
		checkIssues(usr, true)
	case "close":
		closeIssue(usr, usr.Task)
	case "update":
		update(usr, msg)
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
		usr, err := createUser(msg.From.ID, msg.CommandArguments())
		if err != nil {
			log.Println("!!!!!! ERROR !!!!!!")
			log.Println(err.Error())
			message(msg.From.ID, "Invalid token. Try reset token in your profile page and send it again.")
			return
		}
		message(msg.From.ID, "It's all! Just wait notifications! :3")
		checkIssues(usr, true)
		return
	}
	message(msg.From.ID, "Hello, stranger!\nFor start you need send me your personal token. Go to your profile page, grab it from right sidebar and send it by `/start [token]` command. Easy!")
}

func update(usr *dbUser, msg *tg.Message) {
	if usr.Task == 0 {
		message(usr.Telegram, "No one issue for you right now. 🏖")
		return
	}
	if msg.CommandArguments() == "" {
		message(usr.Telegram, "Please, use `/update` command with `[hours]` and/or `[any text]` arguments.")
		return
	}
	log.Println("====== UPDATE COMMAND ======")
	if err := updateIssue(usr, msg.CommandArguments(), false); err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		message(msg.From.ID, err.Error())
		return
	}
	text := fmt.Sprintf(
		"Your comment: `%s`\nTo issue %s has been sent!\nYou are free from it for the next 24 hours.",
		msg.Text,
		makeIssueUrl(usr.Task),
	)
	message(msg.From.ID, text)
	changeIssue(usr, 0)
	checkIssues(usr, true)
}

func skip(usr *dbUser, msg *tg.Message) {
	log.Println("====== SKIP COMMAND ======")
	if err := updateIssue(usr, "Skipped", false); err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		message(msg.From.ID, err.Error())
		return
	}
	text := fmt.Sprintf(
		"Issue %s has been skipped.\nYou are free from it for the next 24 hours.",
		makeIssueUrl(usr.Task),
	)
	message(msg.From.ID, text)
	changeIssue(usr, 0)
	checkIssues(usr, true)
}

/*
func closeIssue(usr *dbUser, msg *tg.Message) {
	log.Println("====== SKIP COMMAND ======")
	if err := updateIssue(usr, "Closed", true); err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		message(msg.From.ID, err.Error())
		return
	}
	text := fmt.Sprintf(
		"Issue %s has been closed.\nI will not remind you of this again, unless it is reopened.",
		makeIssueUrl(usr.Task),
	)
	message(msg.From.ID, text)
	changeIssue(usr, 0)
	checkIssues(usr, true)
}
*/

func message(to int, text string) {
	log.Println("====== MESSAGE ======")
	notify := tg.NewMessage(int64(to), text)
	notify.ParseMode = "markdown"
	if _, err := bot.Send(notify); err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
	}
}
