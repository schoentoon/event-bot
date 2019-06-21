package utils

import (
	"encoding/json"

	"gitlab.schoentoon.com/schoentoon/event-bot/templates"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"github.com/getsentry/sentry-go"
)

// ErrorWithChattable an error that also indicates that the attached Chattable should be send to indicate the error to the enduser
type ErrorWithChattable struct {
	ErrorMsg error
	Chattable *tgbotapi.Chattable
	CallbackAnswer *tgbotapi.CallbackConfig
}

func NewErrorWithChattable(err error, sendable tgbotapi.Chattable) *ErrorWithChattable {
	return &ErrorWithChattable{
		ErrorMsg: err,
		Chattable: &sendable,
	}
}

func NewErrorWithCallbackAnswer(err error, answer tgbotapi.CallbackConfig) *ErrorWithChattable {
	return &ErrorWithChattable{
		ErrorMsg: err,
		CallbackAnswer: &answer,
	}
}

func NewErrorWithChattableFromTemplate(err error, template_name string, chatID int64) *ErrorWithChattable {
	hub := sentry.CurrentHub().Clone()
	hub.CaptureException(err)

	rendered, terr := templates.Execute(template_name, nil)
	if terr != nil {
		panic(err)
	}
	msg := tgbotapi.NewMessage(chatID, rendered)
	return NewErrorWithChattable(err, msg)
}

func NewErrorWithCallbackAnswerTemplate(err error, template_name string, callback *tgbotapi.CallbackQuery) *ErrorWithChattable {
	hub := sentry.CurrentHub().Clone()
	data, err := json.Marshal(callback)
	if err != nil {
		panic(err)
	}
	hub.Scope().SetExtras(map[string]interface{}{"request":string(data)})

	hub.CaptureException(err)

	rendered, terr := templates.Execute(template_name, nil)
	if terr != nil {
		panic(err)
	}
	answer := tgbotapi.NewCallback(callback.ID, rendered)
	return NewErrorWithCallbackAnswer(err, answer)
}

func (e *ErrorWithChattable) Error() string {
	return e.ErrorMsg.Error()
}

func (e *ErrorWithChattable) Send(bot *tgbotapi.BotAPI) (error) {
	switch {
	case e.Chattable != nil:
		_, err := bot.Send(*e.Chattable)
		return err
	case e.CallbackAnswer != nil:
		_, err := bot.AnswerCallbackQuery(*e.CallbackAnswer)
		return err
	default:
		return nil
	}
}