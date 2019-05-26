package main

import (
	"database/sql"
	"flag"
	"log"

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
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "newevent":
					handleNewEvent(db, bot, update.Message)
				}
			} else if update.Message.Chat.IsPrivate() {
				handlePrivateMessage(db, bot, update.Message)
			} else {
				log.Printf("Unhandled message %#v", update.Message)
				edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, update.Message.MessageID, "penis")
				bot.Send(edit)
			}
		} else if update.InlineQuery != nil {
			handleInlineQuery(db, bot, update.InlineQuery)
		} else if update.ChosenInlineResult != nil {
			log.Printf("%#v", update.ChosenInlineResult)
		}
		log.Printf("%#v", update)
	}
}
