package main

import (
	"github.com/kirillDanshin/dlog"
	"github.com/mattn/go-redmine"
)

func checkUser(id int64) error {
	reg, err := db.Prepare("INSERT IGNORE INTO `users` VALUES( ?, ?, ? )")
	if err != nil {
		return err
	}
	defer reg.Close()

	if _, err = reg.Exec(id, "", "connect"); err != nil {
		return err
	}

	return nil
}

func checkStatus(id int64) (string, error) {
	if err := checkUser(id); err != nil {
		return "", err
	}

	rows, err := db.Query("SELECT `status` FROM `users` WHERE `chat_id` = ?", id)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var status string
	for rows.Next() {
		if err := rows.Scan(&status); err != nil {
			return "", err
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	return status, nil
}

func setToken(id int64, apikey string) (*redmine.User, error) {
	if err := checkUser(id); err != nil {
		return nil, err
	}

	user, err := getCurrentUser(config["endpoint"].(string), apikey)
	if err != nil {
		dlog.Ln("====== REDMINE FAIL ======")
		return nil, err
	}

	dlog.D(*user)

	reg, err := db.Prepare("UPDATE `users` SET `token` = ?, `status` = ? WHERE `chat_id` = ?")
	if err != nil {
		dlog.Ln("====== OOPS ======")
		return nil, err
	}
	defer reg.Close()

	if _, err := reg.Exec(apikey, "main", id); err != nil {
		return nil, err
	}

	dlog.Ln("====== ITS OKAY ======")

	return user, nil
}

func getToken(id int64) (string, error) {
	rows, err := db.Query("SELECT `token` FROM `users` WHERE `chat_id` = ?", id)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var token string
	for rows.Next() {
		if err := rows.Scan(&token); err != nil {
			return "", err
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	return token, nil
}

func checkToken(id int64, apikey string) error {
	if err := checkUser(id); err != nil {
		return err
	}

	token, err := getToken(id)
	if err != nil {
		return err
	}

	if _, err := getCurrentUser(config["endpoint"].(string), token); err != nil {
		return err
	}

	return nil
}
