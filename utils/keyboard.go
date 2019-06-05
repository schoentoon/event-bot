package utils

import (
	"fmt"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func CreateInlineKeyboard(eventID int64) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Yes", fmt.Sprintf("event/yes/%d", eventID)),
			tgbotapi.NewInlineKeyboardButtonData("Maybe", fmt.Sprintf("event/maybe/%d", eventID)),
			tgbotapi.NewInlineKeyboardButtonData("No", fmt.Sprintf("event/no/%d", eventID)),
		),
	)

	return &keyboard
}
