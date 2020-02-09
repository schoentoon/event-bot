package events

import (
	"gitlab.com/schoentoon/event-bot/idhash"
	"gitlab.com/schoentoon/event-bot/templates"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func CreateInlineKeyboard(event Event, eventID int64) *tgbotapi.InlineKeyboardMarkup {
	var keyboard tgbotapi.InlineKeyboardMarkup
	// 'ChangeAnswerYesNoMaybe', 'ChangeAnswerYesMaybe', 'ChangeAnswerYesNo', 'ChangeAnswerYes'
	switch event.AnswerMode {
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

	if event.PubliclyShareable {
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonSwitch(templates.Button("button_share.tmpl", nil), idhash.Encode(idhash.Event, eventID)),
			),
		)
	}

	return &keyboard
}
