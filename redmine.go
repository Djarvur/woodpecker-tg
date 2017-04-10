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

func getIssues(apikey, assignedTo string, offset, limit int, timestamp *time.Time) ([]redmine.Issue, error) {
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
	req.RawQuery = q.Encode()

	reqStr := req.String()
	if timestamp != nil {
		reqStr += url.QueryEscape(fmt.Sprint("updated_on", "<=", timestamp.Format(time.RFC3339)))
	}

	code, body, err := http.Get(nil, reqStr)
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

func checkIssues(usr *dbUser, fromUser bool) {
	log.Println("====== CHECK ISSUES ======")

	ts := time.Now().UTC().AddDate(0, 0, -1)
	issues, err := getIssues(usr.Token, strconv.Itoa(usr.Redmine), 0, 1, &ts)
	if err != nil {
		log.Println("!!!!!! ERROR !!!!!!")
		log.Println(err.Error())
		message(usr.Telegram, err.Error())
		return
	}

	if len(issues) > 0 {
		checkIssue(usr, issues[0])
	} else if fromUser {
		changeIssue(usr, 0)
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
		"*%s*\n🏷 %s\n🗄 %s\n\n_%s_\n\n⏰ Last updated: %s",
		issue.Subject,
		makeIssueUrl(issue.Id),
		issue.Project.Name,
		issue.Description,
		updTime.String(),
	)
	message(usr.Telegram, text)
	changeIssue(usr, issue.Id)
}

func closeIssue(usr *dbUser, id int) error {
	if usr.Task == 0 {
		return fmt.Errorf("not selected task")
	}

	issue := redmine.Issue{
		Id: id,
	}
	issue.Status = &redmine.IdName{
		Id:   statusClosed,
		Name: "closed",
	}
	body, err := json.Marshal(issue)
	if err != nil {
		log.Fatalln(err.Error())
		return err
	}

	uri := &url.URL{
		Scheme: scheme,
		Host:   endpoint,
		Path:   fmt.Sprintf("issues/%d.json", id),
	}

	log.Println(uri.String())

	var req http.Request
	req.Header.SetMethod("PUT")
	req.Header.SetContentType("application/json; charset=utf-8")
	req.Header.Set("X-Redmine-API-Key", usr.Token)
	req.SetRequestURI(uri.String())
	req.SetBody(body)

	var resp http.Response
	err = http.Do(&req, &resp)
	if err != nil {
		log.Fatalln(err.Error())
		return err
	}

	log.Println(string(resp.Body()))

	text := fmt.Sprintf(
		"Issue %s has been closed.\nI will not remind you of this again, unless it is reopened.",
		makeIssueUrl(usr.Task),
	)
	message(usr.Telegram, text)
	changeIssue(usr, 0)
	checkIssues(usr, true)

	return nil
}

func updateIssue(usr *dbUser, note string, closed bool) error {
	log.Println("====== UPDATE ISSUE ======")

	if closed {
		closeIssue(usr, usr.Task)
		return nil
	}

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

	issue.CategoryId = 0
	issue.ProjectId = issue.Project.Id
	issue.PriorityId = issue.Priority.Id
	issue.TrackerId = issue.Tracker.Id
	if closed {
		issue.StatusId = statusClosed
	} else {
		issue.Notes = fmt.Sprintf("%s via @%s", note, bot.Self.UserName)
	}

	log.Println("TrackerID:", issue.Tracker.Id, "\nTrackerName:", issue.Tracker.Name)

	return c.UpdateIssue(*issue)
}
