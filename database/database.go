package database

import (
	"database/sql"
	"log"
)

// TxRollback helper funtion to automatically rollback and log issues with rollbacks
func TxRollback(tx *sql.Tx, err error) error {
	rollbackerr := tx.Rollback()
	if rollbackerr != nil {
		log.Printf("Error while rolling back transaction %v", rollbackerr)
	}

	return err
}

// UpgradeDatabase fills in the database schema accordingly
func UpgradeDatabase(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS public.events (
		id serial NOT NULL,
		"owner" bigint NOT NULL,
		name varchar NULL,
		description varchar NULL,
		CONSTRAINT events_pk PRIMARY KEY (id)
	);`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TYPE user_state AS ENUM ('no_command', 'waiting_for_event_name', 'waiting_for_description')`)
	if err != nil {
		log.Printf("%v, continueing anyway..", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS public.user_states (
		user_id bigint NOT NULL,
		state user_state DEFAULT 'no_command',
		CONSTRAINT user_state_pk PRIMARY KEY (user_id)
	);`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS public.drafts (
		user_id bigint NOT NULL,
		name varchar NULL,
		description varchar NULL,
		CONSTRAINT draft_pk PRIMARY KEY (user_id)
	);`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS public.inline_messages (
		event_id serial REFERENCES events(id),
		inline_message_id varchar NOT NULL,
		needs_update boolean DEFAULT false,
		PRIMARY KEY (event_id, inline_message_id)
	);`)
	if err != nil {
		return err
	}

	// this has to match with the fields in idhash/types.go
	_, err = db.Exec(`CREATE TYPE answers_enum AS ENUM ('VoteYes', 'VoteNo', 'VoteMaybe')`)
	if err != nil {
		log.Printf("%v, continueing anyway..", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS public.users (
		id bigint NOT NULL,
		first_name varchar,
		last_name varchar,
		username varchar,
		PRIMARY KEY (id)
	)`)
	if err != nil {
		return nil
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS public.answers (
		user_id bigint REFERENCES users(id),
		event_id serial REFERENCES events(id),
		answer answers_enum NOT NULL,
		PRIMARY KEY (user_id, event_id)
	);`)
	if err != nil {
		return err
	}

	return nil
}
