package utils

import (
	"fmt"

	"gitlab.schoentoon.com/schoentoon/event-bot/templates"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func CreateInlineKeyboard(eventID int64) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_yes.tmpl", nil), fmt.Sprintf("event/yes/%d", eventID)),
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_maybe.tmpl", nil), fmt.Sprintf("event/maybe/%d", eventID)),
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_no.tmpl", nil), fmt.Sprintf("event/no/%d", eventID)),
		),
	)

	return &keyboard
}
