package commands

import (
	"database/sql"
	"fmt"
	"log"

	"gitlab.schoentoon.com/schoentoon/event-bot/database"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func HandleNewEventCommand(db *sql.DB, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	err := func(db *sql.DB, msg *tgbotapi.Message) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}

		_, err = tx.Exec(`INSERT INTO public.drafts (user_id)
			VALUES ($1)`,
			msg.From.ID)
		if err != nil {
			return database.TxRollback(tx, err)
		}

		err = database.ChangeUserStateTx(tx, msg.From.ID, "waiting_for_event_name")
		if err != nil {
			return database.TxRollback(tx, err)
		}

		return tx.Commit()
	}(db, msg)

	var reply tgbotapi.MessageConfig
	if err == nil {
		reply = tgbotapi.NewMessage(msg.Chat.ID, "Created new event, please enter the name for the event. Use /skip to have no name.")
	} else {
		reply = tgbotapi.NewMessage(msg.Chat.ID, "Something went wrong while creating the event, please try again later.")
		log.Printf("Error while creating new event %v", err)
	}

	reply.ReplyToMessageID = msg.MessageID
	_, err = bot.Send(reply)
	return err
}

func HandleNewEventName(db *sql.DB, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	err := func(db *sql.DB, msg *tgbotapi.Message) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}

		_, err = tx.Exec(`UPDATE public.drafts SET name = $1 WHERE user_id = $2`, msg.Text, msg.From.ID)
		if err != nil {
			return database.TxRollback(tx, err)
		}

		err = database.ChangeUserStateTx(tx, msg.From.ID, "waiting_for_description")
		if err != nil {
			return database.TxRollback(tx, err)
		}

		return tx.Commit()
	}(db, msg)

	reply := tgbotapi.NewMessage(msg.Chat.ID, "Set name accordingly, please enter description for the event. Use /skip to have no description.")
	reply.ReplyToMessageID = msg.MessageID

	_, err = bot.Send(reply)
	return err
}

func HandleNewEventDescription(db *sql.DB, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	eventID, err := func(db *sql.DB, msg *tgbotapi.Message) (int64, error) {
		tx, err := db.Begin()
		if err != nil {
			return -1, err
		}

		_, err = tx.Exec(`UPDATE public.drafts SET description = $1 WHERE user_id = $2`, msg.Text, msg.From.ID)
		if err != nil {
			return -1, database.TxRollback(tx, err)
		}

		row := tx.QueryRow(`INSERT INTO public.events ("owner", name, description)
			SELECT user_id "owner", name, description FROM public.drafts WHERE user_id = $1
			RETURNING id`,
			msg.From.ID)
		var id int64
		err = row.Scan(&id)
		if err != nil {
			return -1, database.TxRollback(tx, err)
		}

		err = database.ChangeUserStateTx(tx, msg.From.ID, "no_command")
		if err != nil {
			return -1, database.TxRollback(tx, err)
		}

		_, err = tx.Exec(`DELETE FROM public.drafts WHERE user_id = $1`, msg.From.ID)
		if err != nil {
			return -1, database.TxRollback(tx, err)
		}

		return id, tx.Commit()
	}(db, msg)

	reply := tgbotapi.NewMessage(msg.Chat.ID, "Congratz! Event created succesfully!")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonSwitch("Share", fmt.Sprintf("event/%d", eventID)),
		),
	)
	reply.ReplyMarkup = keyboard
	reply.ReplyToMessageID = msg.MessageID

	_, err = bot.Send(reply)
	return err
}
