package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

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
	code, body, err := f.Get(nil, fmt.Sprint(endpoint, "/users/current.json?key=", apikey))
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
