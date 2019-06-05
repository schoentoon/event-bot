package main

import (
	"sync"
	"log"
	"database/sql"

	"gitlab.schoentoon.com/schoentoon/event-bot/commands"
	"gitlab.schoentoon.com/schoentoon/event-bot/callback"
	"gitlab.schoentoon.com/schoentoon/event-bot/database"
	"gitlab.schoentoon.com/schoentoon/event-bot/inline"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func worker(wg *sync.WaitGroup, db *sql.DB, bot *tgbotapi.BotAPI, ch tgbotapi.UpdatesChannel) {
	defer wg.Done()
	for update := range ch {
		err := job(update, db, bot)
		if err != nil {
			log.Printf("%#v on %#v", err, update)
		}
	}
}

func job(update tgbotapi.Update, db *sql.DB, bot *tgbotapi.BotAPI) error {
	if update.Message != nil {
		if update.Message.Chat.IsPrivate() == false {
			return nil
		}
		state, err := database.GetUserState(db, update.Message.From.ID)
		if err != nil {
			return err
		}
		switch state {
		case "no_command":
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "newevent":
					return commands.HandleNewEventCommand(db, bot, update.Message)
				}
			}
		case "waiting_for_event_name":
			return commands.HandleNewEventName(db, bot, update.Message)
		case "waiting_for_description":
			return commands.HandleNewEventDescription(db, bot, update.Message)
		}
	} else if update.InlineQuery != nil {
		return inline.HandleInlineQuery(db, bot, update.InlineQuery)
	} else if update.ChosenInlineResult != nil {
		return inline.HandleChoseInlineResult(db, update.ChosenInlineResult)
	} else if update.CallbackQuery != nil {
		log.Printf("CALLBACK %#v", update.CallbackQuery)
		return callback.HandleCallback(db, bot, update.CallbackQuery)
	}

	return nil
}