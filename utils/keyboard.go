package utils

import (
	"gitlab.schoentoon.com/schoentoon/event-bot/idhash"
	"gitlab.schoentoon.com/schoentoon/event-bot/templates"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func CreateInlineKeyboard(answersOptions string, eventID int64) *tgbotapi.InlineKeyboardMarkup {
	var keyboard tgbotapi.InlineKeyboardMarkup
	// 'ChangeAnswerYesNoMaybe', 'ChangeAnswerYesMaybe', 'ChangeAnswerYesNo', 'ChangeAnswerYes'
	switch answersOptions {
	case "ChangeAnswerYesNoMaybe":
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_yes.tmpl", nil), idhash.Encode(idhash.VoteYes, eventID)),
				tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_maybe.tmpl", nil), idhash.Encode(idhash.VoteMaybe, eventID)),
				tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_no.tmpl", nil), idhash.Encode(idhash.VoteNo, eventID)),
			),
		)
	case "ChangeAnswerYesMaybe":
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_yes.tmpl", nil), idhash.Encode(idhash.VoteYes, eventID)),
				tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_maybe.tmpl", nil), idhash.Encode(idhash.VoteMaybe, eventID)),
			),
		)
	case "ChangeAnswerYesNo":
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_yes.tmpl", nil), idhash.Encode(idhash.VoteYes, eventID)),
				tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_no.tmpl", nil), idhash.Encode(idhash.VoteNo, eventID)),
			),
		)
	case "ChangeAnswerYes":
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_yes.tmpl", nil), idhash.Encode(idhash.VoteYes, eventID)),
			),
		)
	}

	return &keyboard
}

func CreateEventCreatedKeyboard(eventID int64) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_settings.tmpl", nil), idhash.Encode(idhash.Settings, eventID)),
			tgbotapi.NewInlineKeyboardButtonSwitch(templates.Button("button_share.tmpl", nil), idhash.Encode(idhash.Event, eventID)),
		),
	)

	return &keyboard
}
