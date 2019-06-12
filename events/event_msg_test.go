package events

import (
	"log"
	"testing"

	"gitlab.schoentoon.com/schoentoon/event-bot/templates"
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
	var user Vote
	user.ID = 1337
	user.FirstName = "First name"
	user.LastName = "Last name"
	user.UserName = "User name"
	user.Attendees = 1
	event.Yes = []Vote{user}
	event.Maybe = []Vote{user}
	event.No = []Vote{user}

	_, err := templates.Execute("event.tmpl", event)
	if err != nil {
		t.Fatal(err)
	}
}
