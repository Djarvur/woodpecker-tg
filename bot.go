package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/hjson/hjson-go"
	"github.com/kirillDanshin/dlog"
)

var (
	bot    *tg.BotAPI
	config map[string]interface{}
	locale map[string]string
)

func init() {
	// Open configuration file and read content
	cfgFile, err := ioutil.ReadFile("config.hjson")
	if err != nil {
		panic(err.Error())
	}
	if err = hjson.Unmarshal(cfgFile, &config); err != nil {
		panic(err.Error())
	}

	// Initialize bot
	bot, err = tg.NewBotAPI(config["token"].(string))
	if err != nil {
		panic(err.Error())
	}
	dlog.F("[BOT] Authorized as @%s", bot.Self.UserName)

}

func main() {
	db, err := sql.Open(
		"mysql",
		fmt.Sprintf("%s:%s@/%s", config["mysql_user"].(string), config["mysql_pass"].(string), config["mysql_db"].(string)),
	)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	debug := flag.Bool("debug", false, "enable debug logs")
	webhook := flag.Bool("webhook", false, "enable debug logs")
	flag.Parse()

	bot.Debug = *debug // More logs

	updates := make(<-chan tg.Update)
	updates, err = setUpdates(*webhook)
	if err != nil {
		dlog.F("[ERROR] Ошибка получения обновлений: %s", err.Error())
	}

	// Updater
	for update := range updates {
		if update.Message != nil {
			go messages(update.Message, db)
		}
	}
}

func setUpdates(isWebhook bool) (<-chan tg.Update, error) {
	bot.RemoveWebhook() // Just in case
	if isWebhook == true {
		if _, err := bot.SetWebhook(
			tg.NewWebhook(
				fmt.Sprint(config["webhook_set"].(string), config["token"].(string)),
			),
		); err != nil {
			return nil, err
		}
		go http.ListenAndServe(config["webhook_serve"].(string), nil)
		updates := bot.ListenForWebhook(
			fmt.Sprint(config["webhook_listen"].(string), config["token"].(string)),
		)
		return updates, nil
	} else {
		upd := tg.NewUpdate(0)
		upd.Timeout = 60
		updates, err := bot.GetUpdatesChan(upd)
		if err != nil {
			return nil, err
		}
		return updates, nil
	}
}
