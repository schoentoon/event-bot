package callback

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"gitlab.schoentoon.com/schoentoon/event-bot/database"
	"gitlab.schoentoon.com/schoentoon/event-bot/utils"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func HandleCallback(db *sql.DB, bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) error {
	split := strings.Split(callback.Data, "/")
	if len(split) < 3 {
		return errors.New("Split is less than 3, ignoring")
	}

	if split[0] == "event" {
		eventID, err := strconv.ParseInt(split[2], 10, 64)
		if err != nil {
			return err
		}
		return handleEvent(db, bot, eventID, split[1], callback.From)
	}

	return nil
}

func handleEvent(db *sql.DB, bot *tgbotapi.BotAPI, eventID int64, answer string, from *tgbotapi.User) error {
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
		from.ID, eventID, answer)
	if err != nil {
		return database.TxRollback(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	// if the previous answer is equal, we don't need to go and update all messages
	if answer == oldAnswer {
		return nil
	}

	return utils.UpdateExistingMessages(db, bot, eventID)
}
