package events

import (
	"database/sql"
	"log"
	"time"

	"gitlab.com/schoentoon/event-bot/database"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func UpdateLoop(db *sql.DB, bot *tgbotapi.BotAPI, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
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

	_, err = tx.Exec(`CLOSE event_messages_cursor`)
	if err != nil {
		return database.TxRollback(tx, err)
	}

	return tx.Commit()
}
