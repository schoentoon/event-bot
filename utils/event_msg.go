package utils

import (
	"database/sql"
	"log"

	"gitlab.schoentoon.com/schoentoon/event-bot/templates"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type Event struct {
	Name        string
	Description string
	Yes         []tgbotapi.User
	No          []tgbotapi.User
	Maybe       []tgbotapi.User
}

func FormatEvent(db *sql.DB, eventID int64) (string, error) {
	row := db.QueryRow(`SELECT name, description
		FROM public.events
		WHERE id = $1`,
		eventID)
	var event Event
	err := row.Scan(&event.Name, &event.Description)
	if err != nil {
		return "", err
	}

	rows, err := db.Query(`SELECT id, first_name, last_name, username, answer
		FROM public.users
		INNER JOIN public.answers
		ON users.id = answers.user_id
		WHERE answers.event_id = $1`,
		eventID)
	if err != nil {
		return "", err
	}

	for rows.Next() {
		var user tgbotapi.User
		var answer string
		err = rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.UserName, &answer)
		if err != nil {
			return "", err
		}
		switch answer {
		case "yes":
			event.Yes = append(event.Yes, user)
		case "maybe":
			event.Maybe = append(event.Maybe, user)
		case "no":
			event.No = append(event.No, user)
		}
	}

	rendered, err := templates.Execute("event.tmpl", event)
	if err != nil {
		return "", err
	}

	return rendered, nil
}

func UpdateExistingMessages(db *sql.DB, bot *tgbotapi.BotAPI, eventID int64) error {
	msg, err := FormatEvent(db, eventID)
	if err != nil {
		return err
	}

	rows, err := db.Query(`SELECT inline_message_id
		FROM public.inline_messages
		WHERE event_id = $1`,
		eventID)
	if err != nil {
		return err
	}

	edit := tgbotapi.EditMessageTextConfig{
		Text: msg,
	}
	edit.ReplyMarkup = CreateInlineKeyboard(eventID)
	edit.ParseMode = "html"

	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return err
		}

		edit.InlineMessageID = id
		_, err = bot.Send(edit)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}
