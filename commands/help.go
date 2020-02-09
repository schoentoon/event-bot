package commands

import (
	"gitlab.com/schoentoon/event-bot/templates"
	"gitlab.com/schoentoon/event-bot/utils"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func SendHelp(bot *tgbotapi.BotAPI, chatID int64) error {
	rendered, err := templates.Execute("help.tmpl", nil)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(chatID, rendered)
	msg.ParseMode = "HTML"

	_, err = utils.Send(bot, msg)

	return err
}

func SendStart(bot *tgbotapi.BotAPI, chatID int64) error {
	rendered, err := templates.Execute("start.tmpl", nil)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(chatID, rendered)
	msg.ParseMode = "HTML"

	_, err = utils.Send(bot, msg)

	return err
}
