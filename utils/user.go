package utils

import (
	"database/sql"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func InsertUser(db *sql.DB, user *tgbotapi.User) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = InsertUserTx(tx, user)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func InsertUserTx(tx *sql.Tx, user *tgbotapi.User) error {
	_, err := tx.Exec(`INSERT INTO public.users
		(id, first_name, last_name, username)
		VALUES
		($1, $2, $3, $4)
		ON CONFLICT (id)
		DO UPDATE
		SET first_name = EXCLUDED.first_name,
		last_name = EXCLUDED.last_name,
		username = EXCLUDED.username`,
		user.ID, user.FirstName, user.LastName, user.UserName)
	return err
}
