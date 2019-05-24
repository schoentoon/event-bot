package main

import (
	"log"
	"database/sql"
	"flag"

	_ "github.com/lib/pq"
	"gopkg.in/telegram-bot-api.v4"
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
	err = UpgradeDatabase(db)
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
	
	for update := range updates {
		if update.Message != nil {
			if update.Message.Text == "/newevent" {
				handleNewEvent(db, bot, update.Message)
			} else {
				handlePrivateMessage(db, bot, update.Message)
			}
		}
		log.Println(update)
	}
}