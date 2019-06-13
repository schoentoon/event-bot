package utils

import (
	"database/sql"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func InsertUser(db *sql.DB, user *tgbotapi.User) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}

	update, err := InsertUserTx(tx, user)
	if err != nil {
		return false, err
	}

	return update, tx.Commit()
}

func InsertUserTx(tx *sql.Tx, user *tgbotapi.User) (bool, error) {
	var i int
	row := tx.QueryRow(`SELECT 1
		FROM public.users
		WHERE id = $1
		AND first_name = $2
		AND last_name = $3
		AND username = $4`,
		user.ID, user.FirstName, user.LastName, user.UserName)
	err := row.Scan(&i)

	// if nothing has changed there is no point in doing the insert
	if err == nil && i == 1 {
		return false, nil
	}

	_, err = tx.Exec(`INSERT INTO public.users
		(id, first_name, last_name, username)
		VALUES
		($1, $2, $3, $4)
		ON CONFLICT (id)
		DO UPDATE
		SET first_name = EXCLUDED.first_name,
		last_name = EXCLUDED.last_name,
		username = EXCLUDED.username`,
		user.ID, user.FirstName, user.LastName, user.UserName)

	_, err = tx.Exec(`UPDATE inline_messages
		SET needs_update = true
		WHERE event_id = (
			SELECT event_id
			FROM answers
			WHERE user_id = $1
		)`,
		user.ID)

	return i != 1, err
}
