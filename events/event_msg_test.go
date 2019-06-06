package events

import (
	"log"
	"testing"

	"gitlab.schoentoon.com/schoentoon/event-bot/templates"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func init() {
	err := templates.Load("../tmpl")
	if err != nil {
		log.Fatal(err)
	}
}

func TestEvenTemplate(t *testing.T) {
	var event Event
	event.Name = "Event name"
	event.Description = "Event description"
	var user tgbotapi.User
	user.ID = 1337
	user.FirstName = "First name"
	user.LastName = "Last name"
	user.UserName = "User name"
	event.Yes = []tgbotapi.User{user}
	event.Maybe = []tgbotapi.User{user}
	event.No = []tgbotapi.User{user}

	_, err := templates.Execute("event.tmpl", event)
	if err != nil {
		t.Fatal(err)
	}
}
