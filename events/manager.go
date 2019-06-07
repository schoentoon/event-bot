package events

import (
	"database/sql"
	"log"
	"time"

	"gitlab.schoentoon.com/schoentoon/event-bot/database"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func UpdateLoop(db *sql.DB, bot *tgbotapi.BotAPI) {
	tick := time.Tick(time.Millisecond * 500)
	for range tick {
		err := run(db, bot)
		if err != nil {
			log.Println(err)
		}
	}
}

func run(db *sql.DB, bot *tgbotapi.BotAPI) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DECLARE event_messages_cursor CURSOR FOR
		SELECT DISTINCT event_id
		FROM public.inline_messages
		WHERE needs_update = true`)
	if err != nil {
		return err
	}
	defer tx.Exec(`CLOSE event_messages_cursor`)

	for {
		var id int64
		row := tx.QueryRow(`FETCH NEXT FROM event_messages_cursor`)
		err = row.Scan(&id)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return database.TxRollback(tx, err)
		}
		err = updateExistingMessages(tx, bot, id)
		if err != nil {
			return database.TxRollback(tx, err)
		}
	}

	return tx.Commit()
}
