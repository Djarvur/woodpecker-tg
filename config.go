package main

import (
	"flag"
	"log"

	config "github.com/olebedev/config"
)

var (
	cfg      *config.Config
	endpoint string
	token    string
	set      string
	listen   string
	serve    string

	configFlag  = flag.String("config", "config.yml", "load custom config file")
	dbFlag      = flag.String("db", "woodpecker.db", "select custom db file")
	debugFlag   = flag.Bool("debug", false, "enable debug logs")
	webhookFlag = flag.Bool("webhook", false, "enable webhook mode")
)

func init() {
	flag.Parse()

	var err error
	cfg, err = config.ParseYamlFile(*configFlag)
	if err != nil {
		log.Fatalln(err.Error())
	}

	endpoint, err = cfg.String("redmine.endpoint")
	if err != nil {
		log.Fatalln(err.Error())
	}

	token, err = cfg.String("telegram.token")
	if err != nil {
		log.Fatalln(err.Error())
	}

	set, err = cfg.String("telegram.webhook.set")
	if err != nil {
		log.Fatalln(err.Error())
	}

	listen, err = cfg.String("telegram.webhook.listen")
	if err != nil {
		log.Fatalln(err.Error())
	}

	serve, err = cfg.String("telegram.webhook.serve")
	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Printf("config %s loaded", *configFlag)
}
