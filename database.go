package main

import (
	"fmt"
	_ "log" // just to safisfy Sublime Go plugin
	"strconv"
	"time"

	bolt "github.com/boltdb/bolt"
)

type dbUser struct {
	Redmine  int
	Telegram int
	Token    string
	Task     int
}

var db *bolt.DB

// FIXME: init() не место для инициализаци базы
func initDB() {
	var err error
	log.Println("###### DB OPEN ######")
	db, err = bolt.Open(*dbFlag, 0600, nil)
	if err != nil {
		log.Fatalln(err.Error())
	}
	// defer db.Close()
	log.Printf("DB %s opened", *dbFlag)

	go func() {
		ticker := time.NewTicker(time.Minute * 15)
		for t := range ticker.C {
			now := time.Now()
			todayStart := time.Date(now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, time.UTC)
			todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 16, 0, 0, 0, time.UTC)

			log.Println("ticker:", t.String())
			log.Println("###### VIEW ######")
			if err := db.View(func(tx *bolt.Tx) error {
				log.Println("###### FOR EACH ######")
				return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
					id, err := strconv.Atoi(string(name))
					if err != nil {
						log.Println("!!!!!! ERROR !!!!!!")
						log.Println(err.Error())
						return err
					}

					usr, err := getUser(id)
					if err != nil {
						log.Println("!!!!!! ERROR !!!!!!")
						log.Println(err.Error())
						return err
					}

					log.Println("Now:", now.UTC().String())
					log.Println("StartTime:", todayStart.String())
					log.Println("EndTime:", todayEnd.String())

					if now.UTC().Before(todayStart) {
						if now.UTC().After(todayEnd) {
							if usr.Task != 0 {
								message(usr.Telegram, "The work day is over, good night. 😴")
								changeIssue(usr, 0)
							}
							return nil
						}
						go checkIssues(usr, false)
					}
					return nil
				})
			}); err != nil {
				log.Println("!!!!!! ERROR !!!!!!")
				log.Println(err.Error())
			}
		}
	}()
}

func pingDB() (bool, error) {
	log.Println("====== PING DB ======")
	var write bool
	log.Println("###### VIEW ######")
	err := db.View(func(tx *bolt.Tx) error {
		log.Println("###### WRITABLE ######")
		write = tx.Writable()
		log.Println("###### RETURN NIL ######")
		return nil
	})
	if err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
	}
	log.Println("###### VIEW DONE ######")
	return write, err
}

func createUser(id int, tkn string) (*dbUser, error) {
	log.Println("====== CREATE USER ======")
	r, err := getCurrentUser(tkn)
	if err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		return nil, err
	}

	log.Println("###### BATCH ######")
	if err := db.Batch(func(tx *bolt.Tx) error {
		log.Println("###### CREATE BUCKET IF NOT EXISTS ######")
		bkt, err := tx.CreateBucketIfNotExists([]byte(strconv.Itoa(id)))
		if err != nil {
			log.Println("!!!!!! ERROR !!!!!!")
			log.Println(err.Error())
			return err
		}
		log.Println("###### BUCKET IS OK ######")
		log.Println("###### PUT REDMINE ######")
		bkt.Put([]byte("redmine"), []byte(strconv.Itoa(r.Id)))
		log.Println("###### PUT TELEGRAM ######")
		bkt.Put([]byte("telegram"), []byte(strconv.Itoa(id)))
		log.Println("###### PUT TASK ######")
		bkt.Put([]byte("task"), []byte(strconv.Itoa(0)))
		log.Println("###### PUT TOKEN ######")
		bkt.Put([]byte("token"), []byte(tkn))

		log.Println("###### RETURN NIL ######")
		return nil
	}); err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		return nil, err
	}

	usr := &dbUser{
		Redmine:  r.Id,
		Telegram: id,
		Token:    tkn,
	}

	log.Println("###### BATCH DONE ######")

	return usr, err
}

func getUser(id int) (*dbUser, error) {
	log.Println("====== GET USER ======")
	var usr dbUser

	log.Println("###### VIEW ######")
	if err := db.View(func(tx *bolt.Tx) error {
		log.Println("###### SELECT BUCKET ######")
		bkt := tx.Bucket([]byte(strconv.Itoa(id)))
		if bkt == nil {
			log.Println("!!!!!! ERROR !!!!!!")
			log.Printf("user %v doesn't exist", id)
			return fmt.Errorf("user %v doesn't exist", id)
		}

		log.Println("###### BUCKET IS OK ######")
		log.Println("###### GET REDMINE ######")
		usr.Redmine, _ = strconv.Atoi(string(bkt.Get([]byte("redmine"))))
		log.Println("###### GET TELEGRAM ######")
		usr.Telegram, _ = strconv.Atoi(string(bkt.Get([]byte("telegram"))))
		log.Println("###### GET TASK ######")
		usr.Task, _ = strconv.Atoi(string(bkt.Get([]byte("task"))))
		log.Println("###### GET TOKEN ######")
		usr.Token = string(bkt.Get([]byte("token")))
		log.Println("###### RETURN NIL ######")
		return nil
	}); err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		return nil, err
	}

	if _, err := getCurrentUser(usr.Token); err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		changeIssue(&usr, 0)
		removeUser(id)
		text := "Your token is broken. Please, send me new valid token."
		message(id, text)
		return nil, fmt.Errorf("invalid token")
	}

	log.Println("###### VIEW DONE ######")

	log.Println("USER TASK:", usr.Task)

	return &usr, err
}

func removeUser(id int) error {
	log.Println("====== REMOVE USER ======")
	log.Println("###### BATCH ######")
	err := db.Batch(func(tx *bolt.Tx) error {
		log.Println("###### DELETE BUCKET ######")
		return tx.DeleteBucket([]byte(strconv.Itoa(id)))
	})
	if err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
	}

	log.Println("###### BATCH DONE ######")
	return err
}

func changeIssue(usr *dbUser, id int) error {
	log.Println("###### BATCH ######")
	err := db.Batch(func(tx *bolt.Tx) error {
		log.Println("###### SELECT BUCKET ######")
		bkt := tx.Bucket([]byte(strconv.Itoa(usr.Telegram)))
		log.Println("###### PUT TASK ######")
		return bkt.Put([]byte("task"), []byte(strconv.Itoa(id)))
	})
	if err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
	}

	log.Println("###### BATCH DONE ######")
	return err
}
