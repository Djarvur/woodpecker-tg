// Я тут понапишу комментариев
//
// Те, что я помечу FIXME - обязательны к исправлению
// Это требования заказчика
//
// Те, что я помечу TODO - требуют обсуждения
// и вырабатывания общей позиции
//
// Комментарии, по которым изменения внесены, следует удалять

package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"

	// FIXME: плохой выбор формата для конфига
	// мы предпочитаем yaml (toml лучше, но yaml чаще)
	hjson "github.com/hjson/hjson-go"

	// FIXME: сомнительный выбор
	// нам нужны логи и для release-сборки тоже
	// советую поглядеть на github.com/powerman/structlog
	log "github.com/kirillDanshin/dlog"

	// TODO: использование fasthttp вместо net/http требует обоснования
	// это твой проект, Максим, поэтому ты не обязан отчитываться
	// но мне любопытно
	f "github.com/valyala/fasthttp"
)

var (
	bot *tg.BotAPI

	// TODO: я весь функционал, связанный с конфигами, выношу в файл config.go
	// FIXME: структура конфига должна быть типизованной
	cfg map[string]interface{}
)

// FIXME: init() не место для чтения конфига и нициализации бота!
func init() {
	// Open configuration file and read content
	// FIXME: имя конфиг-файла должно передаваться параметром командной строки
	config, err := ioutil.ReadFile("config.hjson")
	if err != nil {
		panic(err.Error())
	}
	if err := hjson.Unmarshal(config, &cfg); err != nil {
		panic(err.Error())
	}

	// Initialize bot
	// FIXME: структура конфига должна быть типизованной
	bot, err = tg.NewBotAPI(cfg["token"].(string))
	if err != nil {
		panic(err.Error())
	}
	log.F("[BOT] Authorized as @%s\n", bot.Self.UserName)

}

// файл с функцией main() принято называть main.go
// это не правило, но так существенно удобнее при review
func main() {

	// TODO: я весь функционал, связанный с конфигами, выношу в файл config.go
	debug := flag.Bool("debug", false, "enable debug logs")
	// FIXME: описание флага
	webhook := flag.Bool("webhook", false, "enable debug logs")
	flag.Parse()
	/////////////////////

	bot.Debug = *debug // More logs

	updates := make(<-chan tg.Update)
	updates, err := SetUpdates(*webhook)
	if err != nil {
		panic(err.Error())
	}

	// Updater
	for update := range updates {
		if update.Message != nil {
			go Messages(update.Message)
		}
	}
}

func SetUpdates(isWebhook bool) (<-chan tg.Update, error) {
	bot.RemoveWebhook() // Just in case
	if isWebhook == true {
		// FIXME: этот if сильно ухудшает читабельность кода
		if _, err := bot.SetWebhook(
			tg.NewWebhook(
				// FIXME: структура конфига должна быть типизованной
				fmt.Sprint(cfg["webhook_set"].(string), cfg["token"].(string)),
			),
		); err != nil {
			return nil, err
		}
		// FIXME: структура конфига должна быть типизованной
		// TODO: я бы вынес создание горутины в main для очевидности
		go f.ListenAndServe(cfg["webhook_serve"].(string), nil)
		updates := bot.ListenForWebhook(
			// FIXME: структура конфига должна быть типизованной
			fmt.Sprint(cfg["webhook_listen"].(string), cfg["token"].(string)),
		)
		return updates, nil
	} else {

		// FIXME:  else не нужен - выход из if только по return
		// TODO: else блок короче, именно его я бы спрятал в if
		upd := tg.NewUpdate(0)
		upd.Timeout = 60
		updates, err := bot.GetUpdatesChan(upd)
		if err != nil {
			return nil, err
		}
		return updates, nil
	}
}
