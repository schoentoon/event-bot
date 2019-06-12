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

func handleEvent(db *sql.DB, bot *tgbotapi.BotAPI, eventID int64, answer idhash.HashType, callback *tgbotapi.CallbackQuery) error {
	plusone, err := func() (bool, error) {
		plusone := false
		tx, err := db.Begin()
		if err != nil {
			return false, err
		}

		err = utils.InsertUserTx(tx, callback.From)
		if err != nil {
			return false, database.TxRollback(tx, err)
		}

		row := tx.QueryRow(`SELECT answer
			FROM public.answers
			WHERE user_id = $1
			AND event_id = $2`,
			callback.From.ID, eventID)
		var oldAnswer string
		err = row.Scan(&oldAnswer)
		if err != nil {
			oldAnswer = ""
		}
		if answer == idhash.VoteYes && answer.String() == oldAnswer {
			_, err = tx.Exec(`UPDATE public.answers
				SET attendees = attendees + 1
				WHERE event_id = $1`,
				eventID)
			if err != nil {
				return false, database.TxRollback(tx, err)
			}
			plusone = true
		} else if answer.String() != oldAnswer {
			_, err = tx.Exec(`INSERT INTO public.answers
				(user_id, event_id, answer)
				VALUES
				($1, $2, $3)
				ON CONFLICT (user_id, event_id)
				DO UPDATE
				SET answer = EXCLUDED.answer,
				attendees = 0`,
				callback.From.ID, eventID, answer.String())
			if err != nil {
				return false, database.TxRollback(tx, err)
			}
		} else {
			return false, tx.Commit()
		}

		err = events.NeedsUpdate(tx, eventID)
		if err != nil {
			return false, database.TxRollback(tx, err)
		}

		return plusone, tx.Commit()
	}()

	var rendered string
	if err != nil {
		rendered, err = templates.Execute("something_went_wrong_try_later.tmpl", nil)
	} else {
		switch answer {
		case idhash.VoteYes:
			rendered, err = templates.Execute("ack_yes_vote.tmpl", !plusone)
		case idhash.VoteMaybe:
			rendered, err = templates.Execute("ack_maybe_vote.tmpl", nil)
		case idhash.VoteNo:
			rendered, err = templates.Execute("ack_no_vote.tmpl", nil)
		default:
			return nil
		}
	}
	if err != nil {
		return err
	}

	reply := tgbotapi.NewCallback(callback.ID, rendered)
	_, err = utils.AnswerCallbackQuery(bot, reply)
	return err
}

func handleChangeEventProperty(db *sql.DB, bot *tgbotapi.BotAPI, eventID int64, callback *tgbotapi.CallbackQuery, newState, tmpl string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = database.EventWantsEdit(tx, eventID, callback.From.ID)
	if err != nil {
		return database.TxRollback(tx, err)
	}

	err = database.ChangeUserStateTx(tx, callback.From.ID, newState)
	if err != nil {
		return database.TxRollback(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	rendered, err := templates.Execute(tmpl, nil)
	if err != nil {
		return err
	}
	reply := tgbotapi.NewMessage(int64(callback.From.ID), rendered)

	_, err = utils.Send(bot, reply)
	return err
}
