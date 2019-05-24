package main

import (
	"log"
	"database/sql"

	"gopkg.in/telegram-bot-api.v4"
)

func handleNewEvent(db *sql.DB, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	err := func(db *sql.DB, msg *tgbotapi.Message) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}

		row := tx.QueryRow(`INSERT INTO public.events ("owner")
			VALUES ($1)
			RETURNING id`,
				msg.From.ID)
		var id int64
		err = row.Scan(&id)
		if err != nil {
			return TxRollback(tx, err)
		}
		log.Printf("created new event with id %d", id)

		return tx.Commit()
	}(db, msg)

	var reply tgbotapi.MessageConfig
	if err == nil {
		reply = tgbotapi.NewMessage(msg.Chat.ID, "Created new event, please enter the name for the event. Use /skip to have no name.")
	} else {
		reply = tgbotapi.NewMessage(msg.Chat.ID, "Something went wrong while creating the event, please try again later.")
		log.Printf("Error while creating new event %v", err)
	}

	reply.ReplyToMessageID = msg.MessageID
	_, err = bot.Send(reply)
	return err
}


func handlePrivateMessage(db *sql.DB, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	row := db.QueryRow(`SELECT id, insert_state
		FROM public.events
		WHERE "owner" = $1
		AND insert_state != 'done'`,
			msg.Chat.ID)
	var state string
	var id int64

	err := row.Scan(&id, &state)
	if err != nil {
		return err
	}

	err, processedState := func(db *sql.DB, msg *tgbotapi.Message, id int64, state string) (error, int) {
		tx, err := db.Begin()
		if err != nil {
			return err, -1
		}

		text := msg.Text
		if text == "/skip" {
			text = ""
		}

		processedState := -1
		if state == "waiting_for_name" {
			processedState = 1
			_, err = tx.Exec(`UPDATE public.events
				SET name = $1,
				insert_state = 'waiting_for_description'
				WHERE id = $2`,
					text, id)
		} else if state == "waiting_for_description" {
			processedState = 2
			_, err = tx.Exec(`UPDATE public.events
				SET description = $1,
				insert_state = 'done'
				WHERE id = $2`,
					text, id)
		}

		if err != nil {
			return TxRollback(tx, err), -1
		}

		return tx.Commit(), processedState
	}(db, msg, id, state)

	var reply tgbotapi.MessageConfig
	if err == nil {
		switch processedState {
		case 1:
			reply = tgbotapi.NewMessage(msg.Chat.ID, "Set name accordingly, please enter description for the event. Use /skip to have no description.")
		case 2:
			reply = tgbotapi.NewMessage(msg.Chat.ID, "Congratz! Event created succesfully!")
		}
	} else {
		reply = tgbotapi.NewMessage(msg.Chat.ID, "Something went wrong while creating the event, please try again later.")
		log.Printf("Error while creating new event %v", err)
	}

	reply.ReplyToMessageID = msg.MessageID
	_, err = bot.Send(reply)
	return err
}