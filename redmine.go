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

func getCurrentUser(apikey string) (*redmine.User, error) {
	req := &url.URL{
		Scheme: scheme,
		Host:   endpoint,
		Path:   "users/current.json",
	}

	q := req.Query()
	q.Set("key", apikey)
	req.RawQuery = q.Encode()

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

	if _, err := getCurrentUser(usr.Token); err != nil {
		text := "Invalid token. Try reset token in your profile page and send it again.\nP.S.: But now I going to sleep. Zzz..."
		notification(usr.Telegram, text, -1)
		go removeUser(usr.Telegram)
		return
	}

	c := redmine.NewClient(fmt.Sprint(scheme, "://", endpoint), usr.Token)
	issues, err := c.Issues()
	if err != nil {
		log.Println(err.Error())
		return
	}

	for _, issue := range issues {
		go checkIssue(usr, issue)
	}
}

func checkIssue(usr *dbUser, issue redmine.Issue) {
	updTime, err := time.Parse(time.RFC3339, issue.UpdatedOn)
	if err != nil {
		log.Println(err.Error())
		return
	}

	if issue.AssignedTo == nil {
		log.Printf("issue #%d is not assigned to anyone!", issue.Id)
		c := redmine.NewClient(fmt.Sprint(scheme, "://", endpoint), usr.Token)
		mships, err := c.Memberships(issue.Project.Id)
		if err != nil {
			log.Println(err.Error())
		}

		for _, mship := range mships {
			for _, role := range mship.Roles {
				if role.Id == 3 {
					text := fmt.Sprintf("⚠️ *This task is not assigned to anyone!*\n%s\nLast updated: %s", issue.GetTitle(), updTime.String())
					notification(usr.Telegram, text, issue.Id)
				}
			}
		}
		return
	}

	if issue.AssignedTo.Id != usr.Redmine {
		log.Printf("issue #%d is not assigned to user %d", issue.Id, usr.Redmine)
		return
	}

	log.Printf("issue #%d is assigned to user %d...", issue.Id, usr.Redmine)

	if time.Now().UTC().After(updTime.Add(time.Hour * 24)) {
		log.Println("====== MORE THAN 24 HOURS ======")
		text := fmt.Sprintf("_Use_ `/issue #%d sample text` _for comment issue and reset timer._\n%s\nLast updated: %s", issue.Id, issue.GetTitle(), updTime.String())
		notification(usr.Telegram, text, issue.Id)
	}

	if time.Now().UTC().After(updTime.Add(time.Hour * 48)) {
		log.Println("====== WARNING! ======")
		// TODO: Send notify for all managers who have access to current issue assigned to current token 9_6
		/*
			text := fmt.Sprintf("⚠️ *THIS TASK HAS BEEN FUCKED UP!*\n%s\nLast updated: %s", issue.GetTitle(), updTime.String())
			notification(usr.Telegram, text, issue.Id)
		*/
	}
}
