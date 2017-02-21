package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	redmine "github.com/mattn/go-redmine"
	f "github.com/valyala/fasthttp"
)

type redmineUser struct {
	User redmine.User `json:"user"`
}

type redmineErrors struct {
	Errors []string `json:"errors"`
}

func getCurrentUser(endpoint, apikey string) (*redmine.User, error) {
	req := &url.URL{
		Scheme:   scheme,
		Host:     endpoint,
		Path:     "users/current.json",
		RawQuery: fmt.Sprint("key=", apikey),
	}

	code, body, err := f.Get(nil, req.String())
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	var rUsr redmineUser
	if code != 200 {
		var rErr redmineErrors
		err = decoder.Decode(&rErr)
		if err == nil {
			err = fmt.Errorf(strings.Join(rErr.Errors, "\n"))
		}
	} else {
		err = decoder.Decode(&rUsr)
	}
	if err != nil {
		return nil, err
	}
	return &rUsr.User, nil
}

func checkIssues(usr *dbUser) {
	log.Println("====== SEND ISSUE ======")
	log.Println("to", usr.Telegram)
	log.Println("token", usr.Token)

	c := redmine.NewClient(fmt.Sprint(scheme, "://", endpoint), usr.Token)
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
					notification(usr.Telegram, text, issue.Id)
				}
			}
		}

		if time.Now().UTC().After(updTime.Add(time.Hour * 24)) {
			log.Println("====== MORE THAN 24 HOURS ======")
			text := fmt.Sprintf("%s\nLast updated: %s", issue.GetTitle(), updTime.String())
			notification(usr.Telegram, text, issue.Id)
		}
	}

	// TODO: вот тут, похоже, нужен еще один цикл, с поиском ни на кого не повешенных задач
	// для пользователей, которые админы в своем проекте

}
