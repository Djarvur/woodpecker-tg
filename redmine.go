package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	redmine "github.com/mattn/go-redmine"
	// log "github.com/kirillDanshin/dlog"
)

func GetCurrentUser(endpoint, apikey string) (*redmine.User, error) {
	c := redmine.NewClient(endpoint, apikey)
	resp, err := c.Get(fmt.Sprint(endpoint, "/users/current.json?key=", apikey))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var r = struct {
		User redmine.User `json:"user"`
	}{}
	if resp.StatusCode != 200 {
		var er = struct {
			Errors []string `json:"errors"`
		}{}
		err = decoder.Decode(&er)
		if err == nil {
			err = errors.New(strings.Join(er.Errors, "\n"))
		}
	} else {
		err = decoder.Decode(&r)
	}
	if err != nil {
		return nil, err
	}
	return &r.User, nil
}
