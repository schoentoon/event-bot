package callback

import (
	"database/sql"
	"log"

	"gitlab.schoentoon.com/schoentoon/event-bot/idhash"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func HandleCallback(db *sql.DB, bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) error {
	log.Printf("%#v %#v", callback, callback.Message)
	typ, id, err := idhash.Decode(callback.Data)
	if err != nil {
		return err
	}

	switch typ {
	case idhash.VoteYes:
		fallthrough
	case idhash.VoteMaybe:
		fallthrough
	case idhash.VoteNo:
		return handleEvent(db, bot, id, typ, callback.From)
	case idhash.MainMenu:
		return handleMainMenu(db, bot, id, callback)
	case idhash.Settings:
		return handleSettings(db, bot, id, callback)
	case idhash.SettingChangeAnswers:
		return handleChangeAnswers(db, bot, id, callback)
	case idhash.ChangeAnswerYesNoMaybe:
		fallthrough
	case idhash.ChangeAnswerYesMaybe:
		fallthrough
	case idhash.ChangeAnswerYesNo:
		fallthrough
	case idhash.ChangeAnswerYes:
		return handleChangeAnswerPicked(db, bot, id, typ, callback)
	case idhash.SettingChangeName:
		return handleChangeEventName(db, bot, id, callback)
	case idhash.SettingChangeDescription:
		return handleChangeEventDescription(db, bot, id, callback)
	}

	return nil
}
