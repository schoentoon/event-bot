package main

import (
	"log"
	"database/sql"
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
	_, err := db.Exec(`CREATE TYPE insert_state AS ENUM ('waiting_for_name', 'waiting_for_description', 'done')`)
	if err != nil {
		log.Printf("%v, continueing anyway..", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS public.events (
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
		CONSTRAINT user_id_pk PRIMARY KEY (user_id)
	);`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS public.drafts (
		user_id bigint NOT NUL,
		name varchar NULL,
		description varchar NULL,
		CONSTRAINT user_id_pk PRIMARY KEY (user_id)
	);`)
	if err != nil {
		return err
	}

	return nil
}