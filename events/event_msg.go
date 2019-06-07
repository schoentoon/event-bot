package events

import (
	"database/sql"
	"log"

	"gitlab.schoentoon.com/schoentoon/event-bot/idhash"
	"gitlab.schoentoon.com/schoentoon/event-bot/templates"
	"gitlab.schoentoon.com/schoentoon/event-bot/utils"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type Event struct {
	Name        string
	Description string
	AnswerMode  string
	Yes         []tgbotapi.User
	No          []tgbotapi.User
	Maybe       []tgbotapi.User
}

func FormatEvent(tx *sql.Tx, eventID int64) (string, Event, error) {
	row := tx.QueryRow(`SELECT name, description, answers_options
		FROM public.events
		WHERE id = $1`,
		eventID)
	var event Event
	err := row.Scan(&event.Name, &event.Description, &event.AnswerMode)
	if err != nil {
		return "", event, err
	}

	_, err = tx.Exec(`DECLARE answers_cursor CURSOR FOR
		SELECT id, first_name, last_name, username, answer
		FROM public.users
		INNER JOIN public.answers
		ON users.id = answers.user_id
		WHERE answers.event_id = $1`,
		eventID)
	if err != nil {
		return "", event, err
	}
	defer tx.Exec(`CLOSE answers_cursor`)

	for {
		var user tgbotapi.User
		var answer string

		row := tx.QueryRow(`FETCH NEXT FROM answers_cursor`)
		err = row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.UserName, &answer)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return "", event, err
		}
		switch answer {
		case idhash.VoteYes.String():
			event.Yes = append(event.Yes, user)
		case idhash.VoteMaybe.String():
			event.Maybe = append(event.Maybe, user)
		case idhash.VoteNo.String():
			event.No = append(event.No, user)
		}
	}

	switch event.AnswerMode {
	case idhash.ChangeAnswerYesMaybe.String():
		event.No = []tgbotapi.User{}
	case idhash.ChangeAnswerYesNo.String():
		event.Maybe = []tgbotapi.User{}
	case idhash.ChangeAnswerYes.String():
		event.No = []tgbotapi.User{}
		event.Maybe = []tgbotapi.User{}
	}

	rendered, err := templates.Execute("event.tmpl", event)
	if err != nil {
		return "", event, err
	}

	return rendered, event, nil
}

func updateExistingMessages(tx *sql.Tx, bot *tgbotapi.BotAPI, eventID int64) error {
	msg, event, err := FormatEvent(tx, eventID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DECLARE inline_message_id_cursor CURSOR FOR
		SELECT inline_message_id
		FROM public.inline_messages
		WHERE event_id = $1`,
		eventID)
	if err != nil {
		return err
	}
	defer tx.Exec(`CLOSE inline_message_id_cursor`)

	edit := tgbotapi.EditMessageTextConfig{
		Text: msg,
	}
	edit.ReplyMarkup = utils.CreateInlineKeyboard(event.AnswerMode, eventID)
	edit.ParseMode = "HTML"

	for {
		var id string
		row := tx.QueryRow(`FETCH NEXT FROM inline_message_id_cursor`)
		err = row.Scan(&id)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return err
		}

		edit.InlineMessageID = id
		_, err = bot.Send(edit)
		if err != nil {
			log.Println(err)
		}

		_, err = tx.Exec(`UPDATE public.inline_messages
			SET needs_update = false
			WHERE inline_message_id = $1`,
			id)
		if err != nil {
			return err
		}
	}

	return nil
}

func NeedsUpdate(tx *sql.Tx, eventID int64) error {
	_, err := tx.Exec(`UPDATE public.inline_messages
		SET needs_update = true
		WHERE event_id = $1`,
		eventID)
	return err
}
