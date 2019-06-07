package commands

import (
	"database/sql"
	"log"

	"gitlab.schoentoon.com/schoentoon/event-bot/database"
	"gitlab.schoentoon.com/schoentoon/event-bot/idhash"
	"gitlab.schoentoon.com/schoentoon/event-bot/templates"

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
		rendered, err := templates.Execute("created_new_event.tmpl", nil)
		if err != nil {
			return err
		}
		reply = tgbotapi.NewMessage(msg.Chat.ID, rendered)
	} else {
		rendered, err := templates.Execute("something_went_wrong_try_later.tmpl", nil)
		if err != nil {
			return err
		}
		reply = tgbotapi.NewMessage(msg.Chat.ID, rendered)
		log.Printf("Error while creating new event %v", err)
	}

	reply.ReplyToMessageID = msg.MessageID
	_, err = bot.Send(reply)
	return err
}

func HandleNewEventName(db *sql.DB, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	if len(msg.Text) == 0 || len(msg.Text) > 128 {
		rendered, err := templates.Execute("name_too_long.tmpl", nil)
		if err != nil {
			return err
		}

		reply := tgbotapi.NewMessage(msg.Chat.ID, rendered)
		reply.ReplyToMessageID = msg.MessageID

		_, err = bot.Send(reply)
		return err
	}

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

	rendered, err := templates.Execute("name_set_enter_description.tmpl", nil)
	if err != nil {
		return err
	}
	reply := tgbotapi.NewMessage(msg.Chat.ID, rendered)
	reply.ReplyToMessageID = msg.MessageID

	_, err = bot.Send(reply)
	return err
}

func HandleNewEventDescription(db *sql.DB, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	if len(msg.Text) == 0 || len(msg.Text) > 512 {
		rendered, err := templates.Execute("description_too_long.tmpl", nil)
		if err != nil {
			return err
		}

		reply := tgbotapi.NewMessage(msg.Chat.ID, rendered)
		reply.ReplyToMessageID = msg.MessageID

		_, err = bot.Send(reply)
		return err
	}

	eventID, err := func(db *sql.DB, msg *tgbotapi.Message) (int64, error) {
		tx, err := db.Begin()
		if err != nil {
			return -1, err
		}

		text := msg.Text
		if msg.IsCommand() {
			if msg.Command() == "skip" {
				text = ""
			}
		}

		_, err = tx.Exec(`UPDATE public.drafts SET description = $1 WHERE user_id = $2`, text, msg.From.ID)
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

	rendered, err := templates.Execute("event_created.tmpl", nil)
	if err != nil {
		return err
	}
	reply := tgbotapi.NewMessage(msg.Chat.ID, rendered)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_settings.tmpl", nil), idhash.Encode(idhash.Settings, eventID)),
			tgbotapi.NewInlineKeyboardButtonSwitch(templates.Button("button_share.tmpl", nil), idhash.Encode(idhash.Event, eventID)),
		),
	)
	reply.ReplyMarkup = keyboard
	reply.ReplyToMessageID = msg.MessageID

	_, err = bot.Send(reply)
	return err
}
