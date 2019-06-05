package main

import (
	"database/sql"
	"flag"
	"log"
	"sync"

	"gitlab.schoentoon.com/schoentoon/event-bot/database"

	_ "github.com/lib/pq"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func main() {
	var cfgfile = flag.String("config", "config.yml", "Config file location")
	flag.Parse()

	cfg, err := ReadConfig(*cfgfile)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("postgres", cfg.Postgres.Addr)
	if err != nil {
		panic(err)
	}
	err = database.UpgradeDatabase(db)
	if err != nil {
		panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		panic(err)
	}
	if cfg.Telegram.Debug {
		log.Println("Enabling Telegram debug mode")
		bot.Debug = true
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}

	log.Printf("Starting %d workers", cfg.Workers)

	wg := &sync.WaitGroup{}
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go worker(wg, db, bot, updates)
	}

	wg.Wait()
}