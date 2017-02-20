package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	bolt "github.com/boltdb/bolt"

	// FIXME: не надо работать с telegram из файла database.go
	tg "github.com/go-telegram-bot-api/telegram-bot-api"

	// FIXME: не надо работать с redmine из файла database.go
	redmine "github.com/mattn/go-redmine"
)

var db *bolt.DB

// FIXME: init() не место для инициализаци базы
func init() {
	var err error
	go func() {
		db, err = bolt.Open(*dbFlag, 0600, nil)
		if err != nil {
			log.Fatalln(err.Error())
		}
		defer db.Close()
		log.Printf("DB %s opened", *dbFlag)
		select {}
	}()

	go func() {
		ticker := time.NewTicker(time.Minute * 15)
		for t := range ticker.C {
			log.Println(t.String())
			if err := db.View(func(tx *bolt.Tx) error {
				return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
					id, _ := strconv.Atoi(string(name))
					tkn := string(b.Get([]byte("token")))

					// TODO: целесообразно ли запускать по горутине на сообщение?
					// вернее - не надо ли придумать им лимит, например, на атомиках
					go sendIssue(int64(id), tkn)
					return nil
				})
			}); err != nil {
				log.Println(err.Error())
			}
		}
	}()
}

// FIXME: не надо работать с redmine из файла database.go
func sendIssue(id int64, token string) {
	log.Println("====== SEND ISSUE ======")
	log.Println("to", id)
	log.Println("token", token)

	c := redmine.NewClient(endpoint, token)
	issues, err := c.IssuesByFilter(&redmine.IssueFilter{AssignedToId: "me"})
	if err != nil {
		log.Println(err.Error())
		return
	}

	for _, issue := range issues {
		updTime, _ := time.Parse(time.RFC3339, issue.UpdatedOn)

		if time.Now().UTC().After(updTime.Add(time.Hour * 48)) {
			log.Println("====== WARNING! ======")

			roles, _ := c.Memberships(issue.Project.Id)
			for _, role := range roles {
				if role.Id == 3 {
					text := fmt.Sprintf("*This task has been fucked up!*\n%s\nLast updated: %s", issue.GetTitle(), updTime.String())
					notify := tg.NewMessage(id, text)
					notify.ParseMode = "markdown"
					notify.ReplyMarkup = tg.NewInlineKeyboardMarkup(
						tg.NewInlineKeyboardRow(
							tg.NewInlineKeyboardButtonURL(
								fmt.Sprintf("Open issue #%d", issue.Id),
								fmt.Sprintf("%s/issues/%d", endpoint, issue.Id),
							),
						),
					)
					bot.Send(notify)
				}
			}
		}

		if time.Now().UTC().After(updTime.Add(time.Hour * 24)) {
			log.Println("====== MORE THAN 24 HOURS ======")
			text := fmt.Sprintf("%s\nLast updated: %s", issue.GetTitle(), updTime.String())
			notify := tg.NewMessage(id, text)
			notify.ReplyMarkup = tg.NewInlineKeyboardMarkup(
				tg.NewInlineKeyboardRow(
					tg.NewInlineKeyboardButtonURL(
						fmt.Sprintf("Open issue #%d", issue.Id),
						fmt.Sprintf("%s/issues/%d", endpoint, issue.Id),
					),
				),
			)
			bot.Send(notify)
		}
	}

	// TODO: вот тут, похоже, нужен еще один цикл, с поиском ни на кого не повешенных задач
	// для пользователей, которые админы в своем проекте

}

// TODO: сомнительна полезность vault.db для такого малого количества данных.
// возможно, надо сделать соответствующую структуру в памяти,
// всасывать ее при старте
// и дампить на диск при изменениях

func createUser(id int, tkn string, offset int) (*redmine.User, error) {
	usr, err := getCurrentUser(endpoint, tkn)
	if err != nil {
		return usr, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists([]byte(strconv.Itoa(id)))
		if err != nil {
			return err
		}

		bkt.Put([]byte("token"), []byte(tkn))
		bkt.Put([]byte("offset"), []byte(string(offset)))
		return nil
	})
	return usr, err
}

func getToken(id int) (string, error) {
	var tkn string
	err := db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(strconv.Itoa(id)))
		if bkt == nil {
			return fmt.Errorf("user don't exist")
		}

		tkn = string(bkt.Get([]byte("token")))
		if tkn == "" {
			return fmt.Errorf("user '%v' doesn't exist", id)
		}
		return nil
	})

	if _, err := getCurrentUser(endpoint, tkn); err != nil {
		return tkn, fmt.Errorf("invalid token")
	}

	return tkn, err
}
