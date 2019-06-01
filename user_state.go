package main

import (
	"database/sql"
)

func GetUserState(db *sql.DB, userID int64) (string, error) {
	row := db.QueryRow("SELECT state FROM user_states WHERE user_id = $1", userID)
	var out string
	err := row.Scan(&out)
	if err == sql.ErrNoRows {
		row = db.QueryRow("INSERT INTO user_states (user_id) VALUES ($1) RETURNING state", userID)
		err = row.Scan(&out)
		if err != nil {
			return "", err
		}
	}
	return out, nil
}

func ChangeUserState(db *sql.DB, userID int64, newState string) (error) {
	_, err := db.Exec(`INSERT INTO user_states
		(user_id, state)
		VALUES
		($1, $2)
		ON CONFLICT (user_id)
		DO UPDATE SET state = EXCLUDED.state`)
	
	return err
}