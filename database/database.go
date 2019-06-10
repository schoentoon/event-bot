package database

import (
	"database/sql"
	"log"
	"time"
)

// TxRollback helper funtion to automatically rollback and log issues with rollbacks
func TxRollback(tx *sql.Tx, err error) error {
	rollbackerr := tx.Rollback()
	if rollbackerr != nil {
		log.Printf("Error while rolling back transaction %v", rollbackerr)
	}

	return err
}

func waitForStartup(db *sql.DB) {
	err := db.Ping()
	for err != nil {
		err = db.Ping()
		time.Sleep(time.Millisecond * 100)
	}
}

// UpgradeDatabase fills in the database schema accordingly
func UpgradeDatabase(db *sql.DB) error {
	waitForStartup(db)

	// this has to match with the fields in idhash/types.go
	_, err := db.Exec(`CREATE TYPE answers_setting AS ENUM ('ChangeAnswerYesNoMaybe',
		'ChangeAnswerYesMaybe',
		'ChangeAnswerYesNo',
		'ChangeAnswerYes')`)
	if err != nil {
		log.Printf("%v, continueing anyway..", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS public.events (
		id serial NOT NULL,
		"owner" bigint NOT NULL,
		name varchar NOT NULL,
		description varchar NOT NULL,
		"when" timestamp with time zone NOT NULL,
		answers_options answers_setting DEFAULT 'ChangeAnswerYesNoMaybe',
		wants_edit boolean DEFAULT false,
		settings_message_id integer DEFAULT NULL,
		PRIMARY KEY (id)
	);`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TYPE user_state AS ENUM ('no_command',
		'waiting_for_event_name',
		'waiting_for_description',
		'waiting_for_timestamp')`)
	if err != nil {
		log.Printf("%v, continueing anyway..", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS public.user_states (
		user_id bigint NOT NULL,
		state user_state DEFAULT 'no_command',
		PRIMARY KEY (user_id)
	);`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS public.drafts (
		user_id bigint NOT NULL,
		name varchar NULL,
		description varchar NULL,
		"when" timestamp with time zone NULL,
		PRIMARY KEY (user_id)
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
	_, err = db.Exec(`CREATE TYPE answers_enum AS ENUM ('VoteYes',
		'VoteNo',
		'VoteMaybe')`)
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
