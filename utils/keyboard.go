package utils

import (
	"gitlab.schoentoon.com/schoentoon/event-bot/idhash"
	"gitlab.schoentoon.com/schoentoon/event-bot/templates"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func CreateInlineKeyboard(eventID int64) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_yes.tmpl", nil), idhash.Encode(idhash.VoteYes, eventID)),
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_maybe.tmpl", nil), idhash.Encode(idhash.VoteMaybe, eventID)),
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_no.tmpl", nil), idhash.Encode(idhash.VoteNo, eventID)),
		),
	)

	return &keyboard
}
