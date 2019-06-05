package utils

import (
	"fmt"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func CreateInlineKeyboard(eventID int64) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Yes", fmt.Sprintf("yes/%d", eventID)),
			tgbotapi.NewInlineKeyboardButtonData("Maybe", fmt.Sprintf("maybe/%d", eventID)),
			tgbotapi.NewInlineKeyboardButtonData("No", fmt.Sprintf("no/%d", eventID)),
		),
	)

	return &keyboard
}
