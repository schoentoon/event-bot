package commands

import (
	"gitlab.schoentoon.com/schoentoon/event-bot/templates"
	"gitlab.schoentoon.com/schoentoon/event-bot/utils"
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
