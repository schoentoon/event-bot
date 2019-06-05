package main

import (
	"database/sql"
	"flag"
	"log"

	"gitlab.schoentoon.com/schoentoon/event-bot/commands"
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
	log.Println("Accepting commands")

	for update := range updates {
		if update.Message != nil {
			if update.Message.Chat.IsPrivate() == false {
				continue
			}
			state, err := database.GetUserState(db, update.Message.From.ID)
			if err != nil {
				log.Println(err)
				continue
			}
			switch state {
			case "no_command":
				if update.Message.IsCommand() {
					switch update.Message.Command() {
					case "newevent":
						commands.HandleNewEventCommand(db, bot, update.Message)
					}
				}
			case "waiting_for_event_name":
				commands.HandleNewEventName(db, bot, update.Message)
			case "waiting_for_description":
				commands.HandleNewEventDescription(db, bot, update.Message)
			}
		} else if update.InlineQuery != nil {
			handleInlineQuery(db, bot, update.InlineQuery)
		} else if update.ChosenInlineResult != nil {
			log.Printf("%#v", update.ChosenInlineResult)
		} else if update.CallbackQuery != nil {
			log.Printf("CALLBACK %#v", update.CallbackQuery)
		}
		log.Printf("%#v", update)
	}
}
