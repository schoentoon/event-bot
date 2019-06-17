package utils

import (
	"gitlab.schoentoon.com/schoentoon/event-bot/idhash"
	"gitlab.schoentoon.com/schoentoon/event-bot/templates"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func CreateEventCreatedKeyboard(eventID int64) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_settings.tmpl", nil), idhash.Encode(idhash.Settings, eventID)),
			tgbotapi.NewInlineKeyboardButtonSwitch(templates.Button("button_share.tmpl", nil), idhash.Encode(idhash.Event, eventID)),
		),
	)

	return &keyboard
}
