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
					id, err := strconv.Atoi(string(name))
					if err != nil {
						return err
					}

					usr, err := getUser(id)
					if err != nil {
						return err
					}

					go checkIssues(usr)
					return nil
				})
			}); err != nil {
				log.Println(err.Error())
			}
		}
	}()
}

func pingDB() (bool, error) {
	var write bool
	err := db.View(func(tx *bolt.Tx) error {
		write = tx.Writable()
		return nil
	})
	return write, err
}

func createUser(id int, tkn string) (*dbUser, error) {
	log.Println("====== CREATE USER ======")
	r, err := getCurrentUser(tkn)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	err = db.Batch(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists([]byte(strconv.Itoa(id)))
		if err != nil {
			return err
		}

		bkt.Put([]byte("redmine"), []byte(strconv.Itoa(r.Id)))
		bkt.Put([]byte("telegram"), []byte(strconv.Itoa(id)))
		bkt.Put([]byte("task"), []byte(strconv.Itoa(r.Id)))
		bkt.Put([]byte("token"), []byte(tkn))
		return nil
	})

	usr := &dbUser{
		Redmine:  r.Id,
		Telegram: id,
		Token:    tkn,
	}

	return usr, err
}

func getUser(id int) (*dbUser, error) {
	log.Println("====== GET USER ======")
	var usr dbUser
	if err := db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(strconv.Itoa(id)))
		if bkt == nil {
			return fmt.Errorf("user %v doesn't exist", id)
		}

		usr.Redmine, _ = strconv.Atoi(string(bkt.Get([]byte("redmine"))))
		usr.Telegram, _ = strconv.Atoi(string(bkt.Get([]byte("telegram"))))
		usr.Task, _ = strconv.Atoi(string(bkt.Get([]byte("task"))))
		usr.Token = string(bkt.Get([]byte("token")))
		return nil
	}); err != nil {
		return nil, err
	}

	if _, err := getCurrentUser(usr.Token); err != nil {
		text := "Your token is broken. Please, send me new valid token."
		go message(id, text, -1)
		go removeUser(id)
		return nil, fmt.Errorf("invalid token")
	}

	return &usr, err
}

func removeUser(id int) error {
	log.Println("====== REMOVE USER ======")
	return db.Batch(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(strconv.Itoa(id)))
	})
}

func changeIssue(usr *dbUser, id int) error {
	return db.Batch(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(strconv.Itoa(usr.Telegram)))
		return bkt.Put([]byte("task"), []byte(strconv.Itoa(id)))
	})
}
