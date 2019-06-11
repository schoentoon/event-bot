package database

import (
	"database/sql"
)

func GetUserState(db *sql.DB, userID int) (string, error) {
	row := db.QueryRow("SELECT state FROM user_states WHERE user_id = $1", userID)
	var out string
	err := row.Scan(&out)
	if err == sql.ErrNoRows {
		tx, err := db.Begin()
		if err != nil {
			return "", err
		}
		row = tx.QueryRow("INSERT INTO user_states (user_id) VALUES ($1) RETURNING state", userID)
		err = row.Scan(&out)
		if err != nil {
			return "", err
		}
		err = tx.Commit()
		if err != nil {
			return "", err
		}
	}
	return out, nil
}

func ChangeUserStateTx(tx *sql.Tx, userID int, newState string) error {
	_, err := tx.Exec(`INSERT INTO user_states
		(user_id, state)
		VALUES
		($1, $2)
		ON CONFLICT (user_id)
		DO UPDATE
		SET state = EXCLUDED.state`, userID, newState)

	return err
}

func ChangeUserState(db *sql.DB, userID int, newState string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO user_states
		(user_id, state)
		VALUES
		($1, $2)
		ON CONFLICT (user_id)
		DO UPDATE
		SET state = EXCLUDED.state`, userID, newState)
	if err != nil {
		return err
	}

	return tx.Commit()
}
