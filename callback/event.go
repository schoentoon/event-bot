package callback

import (
	"database/sql"

	"gitlab.schoentoon.com/schoentoon/event-bot/database"
	"gitlab.schoentoon.com/schoentoon/event-bot/events"
	"gitlab.schoentoon.com/schoentoon/event-bot/idhash"
	"gitlab.schoentoon.com/schoentoon/event-bot/templates"
	"gitlab.schoentoon.com/schoentoon/event-bot/utils"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func handleEvent(db *sql.DB, bot *tgbotapi.BotAPI, eventID int64, answer idhash.HashType, from *tgbotapi.User) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = utils.InsertUserTx(tx, from)
	if err != nil {
		return database.TxRollback(tx, err)
	}

	row := tx.QueryRow(`SELECT answer
		FROM public.answers
		WHERE user_id = $1
		AND event_id = $2`,
		from.ID, eventID)
	var oldAnswer string
	err = row.Scan(&oldAnswer)
	if err != nil {
		oldAnswer = ""
	}

	_, err = tx.Exec(`INSERT INTO public.answers
		(user_id, event_id, answer)
		VALUES
		($1, $2, $3)
		ON CONFLICT (user_id, event_id)
		DO UPDATE
		SET answer = EXCLUDED.answer`,
		from.ID, eventID, answer.String())
	if err != nil {
		return database.TxRollback(tx, err)
	}

	// if the previous answer is equal, we don't need to go and update all messages
	if answer.String() != oldAnswer {
		err = events.NeedsUpdate(tx, eventID)
		if err != nil {
			return database.TxRollback(tx, err)
		}
	}

	return tx.Commit()
}

func handleChangeEventName(db *sql.DB, bot *tgbotapi.BotAPI, eventID int64, callback *tgbotapi.CallbackQuery) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = database.EventWantsEdit(tx, eventID, callback.From.ID)
	if err != nil {
		return database.TxRollback(tx, err)
	}

	err = database.ChangeUserStateTx(tx, callback.From.ID, "waiting_for_event_name")
	if err != nil {
		return database.TxRollback(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	rendered, err := templates.Execute("change_event_name.tmpl", nil)
	if err != nil {
		return err
	}
	reply := tgbotapi.NewMessage(int64(callback.From.ID), rendered)

	_, err = bot.Send(reply)
	return err
}

func handleChangeEventDescription(db *sql.DB, bot *tgbotapi.BotAPI, eventID int64, callback *tgbotapi.CallbackQuery) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = database.EventWantsEdit(tx, eventID, callback.From.ID)
	if err != nil {
		return database.TxRollback(tx, err)
	}

	err = database.ChangeUserStateTx(tx, callback.From.ID, "waiting_for_description")
	if err != nil {
		return database.TxRollback(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	rendered, err := templates.Execute("change_event_description.tmpl", nil)
	if err != nil {
		return err
	}
	reply := tgbotapi.NewMessage(int64(callback.From.ID), rendered)

	_, err = bot.Send(reply)
	return err
}
