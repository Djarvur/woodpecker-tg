package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	bolt "github.com/boltdb/bolt"
)

type dbUser struct {
	Redmine  int
	Telegram int64
	Token    string
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

					// TODO: целесообразно ли запускать по горутине на сообщение?
					// вернее - не надо ли придумать им лимит, например, на атомиках
					go checkIssues(usr)
					return nil
				})
			}); err != nil {
				log.Println(err.Error())
			}
		}
	}()
}

// TODO: сомнительна полезность vault.db для такого малого количества данных.
// возможно, надо сделать соответствующую структуру в памяти,
// всасывать ее при старте
// и дампить на диск при изменениях

func createUser(id int, tkn string) (*dbUser, error) {
	r, err := getCurrentUser(fmt.Sprint(scheme, "://", endpoint), tkn)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists([]byte(strconv.Itoa(id)))
		if err != nil {
			return err
		}

		bkt.Put([]byte("redmine"), []byte(string(r.Id)))
		bkt.Put([]byte("telegram"), []byte(string(id)))
		bkt.Put([]byte("token"), []byte(tkn))
		return nil
	})

	return &dbUser{r.Id, int64(id), tkn}, err
}

func getUser(id int) (*dbUser, error) {
	var usr dbUser
	err := db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(strconv.Itoa(id)))
		if bkt == nil {
			return fmt.Errorf("user %v doesn't exist", id)
		}

		r, err := strconv.Atoi(string(bkt.Get([]byte("redmine"))))
		if err != nil {
			return err
		}

		t, err := strconv.Atoi(string(bkt.Get([]byte("telegram"))))
		if err != nil {
			return err
		}

		usr.Redmine = r
		usr.Telegram = int64(t)
		usr.Token = string(bkt.Get([]byte("token")))
		return nil
	})

	if _, err := getCurrentUser(fmt.Sprint(scheme, "://", endpoint), usr.Token); err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	return &usr, err
}
