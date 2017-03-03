package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "log" // just to safisfy Sublime Go plugin
	"net/url"
	"strconv"
	"strings"
	"time"

	redmine "github.com/mattn/go-redmine"
	http "github.com/valyala/fasthttp"
)

type redmineUser struct {
	User redmine.User `json:"user"`
}

type redmineIssues struct {
	Issues []redmine.Issue `json:"issues"`
}

type redmineErrors struct {
	Errors []string `json:"errors"`
}

func makeIssueUrl(id int) string {
	uri := url.URL{
		Scheme: scheme,
		Host:   endpoint,
		Path:   fmt.Sprint("issues/", id),
	}
	return fmt.Sprintf("[#%d](%s)", id, uri.String())
}

func pingRedmine() error {
	log.Println("====== PING REDMINE ======")
	req := url.URL{
		Scheme: scheme,
		Host:   endpoint,
		Path:   "news.json",
	}

	code, _, err := http.Get(nil, req.String())
	if err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		return err
	}
	if code != http.StatusOK {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println("not 200")
		return fmt.Errorf("not 200")
	}
	return nil
}

func getCurrentUser(apikey string) (*redmine.User, error) {
	log.Println("====== GET CURRENT USER ======")
	req := &url.URL{
		Scheme: scheme,
		Host:   endpoint,
		Path:   "users/current.json",
	}

	q := req.Query()
	q.Set("key", apikey)
	req.RawQuery = q.Encode()

	code, body, err := http.Get(nil, req.String())
	if err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	var rUsr redmineUser
	if code != 200 {
		var rErr redmineErrors
		err = decoder.Decode(&rErr)
		if err == nil {
			log.Println("!!!!!! ERROR !!!!!!")
			log.Printf(strings.Join(rErr.Errors, "\n"))
			err = fmt.Errorf(strings.Join(rErr.Errors, "\n"))
		}
	} else {
		err = decoder.Decode(&rUsr)
	}
	if err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		return nil, err
	}
	return &rUsr.User, nil
}

func getIssues(apikey string, assignedTo string, offset, limit int, timestamp *time.Time) ([]redmine.Issue, error) {
	log.Println("====== GET ISSUES ======")

	req := &url.URL{
		Scheme: scheme,
		Host:   endpoint,
		Path:   "issues.json",
	}

	q := req.Query()
	q.Set("key", apikey)
	if assignedTo != "" {
		q.Set("assigned_to_id", assignedTo)
	}
	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if timestamp != nil {
		q.Set("updated_on", fmt.Sprint("<=", timestamp.Format(time.RFC3339)))
	}
	req.RawQuery = q.Encode()

	code, body, err := http.Get(nil, req.String())
	if err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	var rIssues redmineIssues
	if code != 200 {
		var rErr redmineErrors
		err = decoder.Decode(&rErr)
		if err == nil {
			log.Println("!!!!!! ERROR !!!!!!")
			log.Printf(strings.Join(rErr.Errors, "\n"))
			err = fmt.Errorf(strings.Join(rErr.Errors, "\n"))
		}
	} else {
		err = decoder.Decode(&rIssues)
	}
	if err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		return nil, err
	}
	return rIssues.Issues, nil
}

func checkIssues(usr *dbUser) {
	log.Println("====== CHECK ISSUES ======")

	ts := time.Now().UTC().AddDate(0, 0, -1)
	issues, err := getIssues(usr.Token, strconv.Itoa(usr.Redmine), 0, 1, &ts)
	if err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		return
	}

	if len(issues) > 0 {
		checkIssue(usr, issues[0])
	} else {
		message(usr.Telegram, "No one issue for you right now. 🏖")
	}
}

func checkIssue(usr *dbUser, issue redmine.Issue) {
	log.Println("====== CHECK SINGLE ISSUE ======")
	updTime, err := time.Parse(time.RFC3339, issue.UpdatedOn)
	if err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		return
	}

	/*
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
							message(usr.Telegram, text)
						}
					}
				}
				return
			}

		if issue.AssignedTo.Id != usr.Redmine {
			return
		}
	*/

	log.Println("====== MORE THAN 24 HOURS ======")
	text := fmt.Sprintf(
		"Issue %s *%s*\n🗄 Project: %s\nLast updated: %s",
		makeIssueUrl(issue.Id),
		issue.Subject,
		issue.Project.Name,
		updTime.String(),
	)
	message(usr.Telegram, text)
	go changeIssue(usr, issue.Id)
}

func updateIssue(usr *dbUser, note string) error {
	log.Println("====== UPDATE ISSUE ======")

	if usr.Task == 0 {
		return fmt.Errorf("not selected task")
	}

	c := redmine.NewClient(fmt.Sprint(scheme, "://", endpoint), usr.Token)
	issue, err := c.Issue(usr.Task)
	if err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		return err
	}

	issue.Notes = fmt.Sprintf("%s via @%s", note, bot.Self.UserName)
	issue.PriorityId = issue.Priority.Id

	return c.UpdateIssue(*issue)
}
