package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	redmine "github.com/mattn/go-redmine"
)

var db *bolt.DB

func init() {
	var err error
	go func() {
		db, err = bolt.Open("woodpecker.db", 0600, nil)
		if err != nil {
			log.Fatalln(err.Error())
		}
		defer db.Close()
		log.Println("BD opened")
		select {}
	}()

	go func() {
		ticker := time.NewTicker(time.Minute * 15)
		for t := range ticker.C {
			log.Println(t.String())
			if err := db.View(func(tx *bolt.Tx) error {
				return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
					id, _ := strconv.Atoi(string(name))
					token := string(b.Get([]byte("token")))

					go SendIssue(int64(id), token)
					return nil
				})
			}); err != nil {
				log.Println(err.Error())
			}
		}
	}()
}

func SendIssue(id int64, token string) {
	log.Println("====== SEND ISSUE ======")
	log.Println("to", id)
	log.Println("token", token)
	c := redmine.NewClient(cfg["endpoint"].(string), token)
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
								fmt.Sprintf("%s/issues/%d", cfg["endpoint"].(string), issue.Id),
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
						fmt.Sprintf("%s/issues/%d", cfg["endpoint"].(string), issue.Id),
					),
				),
			)
			bot.Send(notify)
		}
	}
}

func CreateUser(id int, token string, offset int) (*redmine.User, error) {
	usr, err := GetCurrentUser(cfg["endpoint"].(string), token)
	if err != nil {
		return usr, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists([]byte(strconv.Itoa(id)))
		if err != nil {
			return err
		}

		bkt.Put([]byte("token"), []byte(token))
		bkt.Put([]byte("offset"), []byte(string(offset)))
		return nil
	})
	return usr, err
}

func GetToken(id int) (string, error) {
	var token string
	err := db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(strconv.Itoa(id)))
		if bkt == nil {
			return fmt.Errorf("user don't exist")
		}

		token = string(bkt.Get([]byte("token")))
		if token == "" {
			return fmt.Errorf("user don't exist")
		}
		return nil
	})

	if _, err := GetCurrentUser(cfg["endpoint"].(string), token); err != nil {
		return token, fmt.Errorf("invalid token")
	}

	log.Println("TOKEN:", token)
	return token, err
}
