package callback

import (
	"database/sql"

	"gitlab.schoentoon.com/schoentoon/event-bot/events"
	"gitlab.schoentoon.com/schoentoon/event-bot/utils"

	"gitlab.schoentoon.com/schoentoon/event-bot/database"

	"gitlab.schoentoon.com/schoentoon/event-bot/idhash"
	"gitlab.schoentoon.com/schoentoon/event-bot/templates"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func handleMainMenu(db *sql.DB, bot *tgbotapi.BotAPI, eventID int64, callback *tgbotapi.CallbackQuery) error {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_settings.tmpl", nil), idhash.Encode(idhash.Settings, eventID)),
			tgbotapi.NewInlineKeyboardButtonSwitch(templates.Button("button_share.tmpl", nil), idhash.Encode(idhash.Event, eventID)),
		),
	)

	edit := tgbotapi.NewEditMessageReplyMarkup(callback.Message.Chat.ID, callback.Message.MessageID, keyboard)

	_, err := utils.Send(bot, edit)
	return err
}

func handleSettings(db *sql.DB, bot *tgbotapi.BotAPI, eventID int64, callback *tgbotapi.CallbackQuery) error {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_change_name.tmpl", nil), idhash.Encode(idhash.SettingChangeName, eventID)),
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_change_description.tmpl", nil), idhash.Encode(idhash.SettingChangeDescription, eventID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_change_timestamp.tmpl", nil), idhash.Encode(idhash.SettingChangeTime, eventID)),
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_change_location.tmpl", nil), idhash.Encode(idhash.SettingChangeLocation, eventID)),
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_change_answers.tmpl", nil), idhash.Encode(idhash.SettingChangeAnswers, eventID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_back.tmpl", nil), idhash.Encode(idhash.MainMenu, eventID)),
		),
	)

	edit := tgbotapi.NewEditMessageReplyMarkup(callback.Message.Chat.ID, callback.Message.MessageID, keyboard)

	_, err := utils.Send(bot, edit)
	return err
}

func handleChangeAnswers(db *sql.DB, bot *tgbotapi.BotAPI, eventID int64, callback *tgbotapi.CallbackQuery) error {
	//'yes_no_maybe', 'yes_maybe', 'yes_no', 'yes'
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_change_yes_no_maybe.tmpl", nil), idhash.Encode(idhash.ChangeAnswerYesNoMaybe, eventID)),
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_change_yes_maybe.tmpl", nil), idhash.Encode(idhash.ChangeAnswerYesMaybe, eventID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_change_yes_no.tmpl", nil), idhash.Encode(idhash.ChangeAnswerYesNo, eventID)),
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_change_yes.tmpl", nil), idhash.Encode(idhash.ChangeAnswerYes, eventID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(templates.Button("button_back.tmpl", nil), idhash.Encode(idhash.Settings, eventID)),
		),
	)

	edit := tgbotapi.NewEditMessageReplyMarkup(callback.Message.Chat.ID, callback.Message.MessageID, keyboard)

	_, err := utils.Send(bot, edit)
	return err
}

func handleChangeAnswerPicked(db *sql.DB, bot *tgbotapi.BotAPI, eventID int64, typ idhash.HashType, callback *tgbotapi.CallbackQuery) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE public.events
		SET answers_options = $1
		WHERE id = $2`,
		typ.String(), eventID)
	if err != nil {
		return database.TxRollback(tx, err)
	}

	// now we mark this event as needing an update
	err = events.NeedsUpdate(tx, eventID)
	if err != nil {
		return database.TxRollback(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	rendered, err := templates.Execute("answers_changed.tmpl", nil)
	if err != nil {
		return err
	}

	answer := tgbotapi.NewCallback(callback.ID, rendered)
	_, err = utils.AnswerCallbackQuery(bot, answer)
	if err != nil {
		return err
	}

	return handleSettings(db, bot, eventID, callback)
}
