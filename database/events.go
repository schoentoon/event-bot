package database

import "database/sql"

func EventWantsEdit(tx *sql.Tx, eventID int64, userID int) error {
	_, err := tx.Exec(`UPDATE public.events
		SET wants_edit = false
		WHERE owner = $1`, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE public.events
		SET wants_edit = true
		WHERE id = $1`, eventID)
	return err
}
