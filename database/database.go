package database

import (
	"database/sql"
	"log"
	"time"

	"github.com/lopezator/migrator"
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

func initMigrator() (*migrator.Migrator) {
	return migrator.New(
		&migrator.Migration{
			Name: "Create answers_settings ENUM",
			Func: func(tx *sql.Tx) error {
				// this has to match with the fields in idhash/types.go
				_, err := tx.Exec(
					`CREATE TYPE answers_setting AS ENUM ('ChangeAnswerYesNoMaybe',
						'ChangeAnswerYesMaybe',
						'ChangeAnswerYesNo',
						'ChangeAnswerYes')`)
				return err
			},
		},
		&migrator.Migration{
			Name: "Create events table",
			Func: func(tx *sql.Tx) error {
				_, err := tx.Exec(
					`CREATE TABLE public.events (
						id serial NOT NULL,
						"owner" bigint NOT NULL,
						name varchar NOT NULL,
						description varchar NOT NULL,
						"when" timestamp with time zone NOT NULL,
						location varchar NOT NULL,
						answers_options answers_setting DEFAULT 'ChangeAnswerYesNoMaybe',
						wants_edit boolean DEFAULT false,
						settings_message_id integer DEFAULT NULL,
						PRIMARY KEY (id)
					)`)
				return err
			},
		},
		&migrator.Migration{
			Name: "Create user_state ENUM",
			Func: func(tx *sql.Tx) error {
				_, err := tx.Exec(
					`CREATE TYPE user_state AS ENUM ('no_command',
						'waiting_for_event_name',
						'waiting_for_description',
						'waiting_for_timestamp',
						'waiting_for_location')`)
				return err
			},
		},
		&migrator.Migration{
			Name: "Create user_states table",
			Func: func(tx *sql.Tx) error {
				_, err := tx.Exec(
					`CREATE TABLE public.user_states (
						user_id bigint NOT NULL,
						state user_state DEFAULT 'no_command',
						PRIMARY KEY (user_id)
					)`)
				return err
			},
		},
		&migrator.Migration{
			Name: "Create drafts table",
			Func: func(tx *sql.Tx) error {
				_, err := tx.Exec(
					`CREATE TABLE public.drafts (
						user_id bigint NOT NULL,
						name varchar NULL,
						description varchar NULL,
						"when" timestamp with time zone NULL,
						location varchar NULL,
						PRIMARY KEY (user_id)
					)`)
				return err
			},
		},
		&migrator.Migration{
			Name: "Create inline_messages table",
			Func: func(tx *sql.Tx) error {
				_, err := tx.Exec(
					`CREATE TABLE public.inline_messages (
						event_id serial REFERENCES events(id),
						inline_message_id varchar NOT NULL,
						needs_update boolean DEFAULT false,
						PRIMARY KEY (event_id, inline_message_id)
					)`)
				return err
			},
		},
		&migrator.Migration{
			Name: "Create answers ENUM",
			Func: func(tx *sql.Tx) error {
				_, err := tx.Exec(
					`CREATE TYPE answers_enum AS ENUM
					('VoteYes',
					'VoteNo',
					'VoteMaybe')`)
				return err
			},
		},
		&migrator.Migration{
			Name: "Create users table",
			Func: func(tx *sql.Tx) error {
				_, err := tx.Exec(
					`CREATE TABLE public.users (
						id bigint NOT NULL,
						first_name varchar,
						last_name varchar,
						username varchar,
						PRIMARY KEY (id)
					)`)
				return err
			},
		},
		&migrator.Migration{
			Name: "Create answers table",
			Func: func(tx *sql.Tx) error {
				_, err := tx.Exec(
					`CREATE TABLE public.answers (
						user_id bigint REFERENCES users(id),
						event_id serial REFERENCES events(id),
						answer answers_enum NOT NULL,
						attendees smallint DEFAULT 0,
						PRIMARY KEY (user_id, event_id)
					)`)
				return err
			},
		},
	)
}

// UpgradeDatabase fills in the database schema accordingly
func UpgradeDatabase(db *sql.DB) error {
	waitForStartup(db)

	migrator := initMigrator()

	return migrator.Migrate(db)
}
