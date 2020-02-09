package main

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"gitlab.com/schoentoon/event-bot/callback"
	"gitlab.com/schoentoon/event-bot/commands"
	"gitlab.com/schoentoon/event-bot/database"
	"gitlab.com/schoentoon/event-bot/inline"
	"gitlab.com/schoentoon/event-bot/utils"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

var (
	inlineQueryDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "bot_inline_query_seconds",
			Help: "How long does it take to handle inline query requests",
		},
		[]string{"error"},
	)
	callbackDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "bot_callback_seconds",
			Help: "How long does it take to handle a callback",
		},
		[]string{"error"},
	)
	messageDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "bot_message_seconds",
			Help: "How long does it take to process a message",
		},
		[]string{"error", "user_state"},
	)
	commandDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "bot_command_seconds",
			Help: "How long does it take to process a command",
		},
		[]string{"error", "command"},
	)
)

func init() {
	prometheus.MustRegister(inlineQueryDuration,
		callbackDuration,
		messageDuration,
		commandDuration,
	)
}

func worker(wg *sync.WaitGroup, db *sql.DB, bot *tgbotapi.BotAPI, ch tgbotapi.UpdatesChannel) {
	defer wg.Done()
	for update := range ch {
		err := job(update, db, bot)
		if err != nil {
			if cerr, ok := err.(*utils.ErrorWithChattable); ok {
				err := cerr.Send(bot)
				if err != nil {
					log.Println(err)
				}
			}
			log.Printf("%#v %s on %#v", err, err, update)
		}
	}
}

func job(update tgbotapi.Update, db *sql.DB, bot *tgbotapi.BotAPI) error {
	defer utils.Recover(update)

	if update.Message != nil {
		if !update.Message.Chat.IsPrivate() {
			return nil
		}
		start := time.Now()

		_, err := utils.InsertUser(db, update.Message.From)
		if err != nil {
			return err
		}

		state, err := database.GetUserState(db, update.Message.From.ID)
		if err != nil {
			return err
		}

		switch state {
		case "no_command":
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "newevent":
					err = commands.HandleNewEventCommand(db, bot, update.Message)
				case "help":
					err = commands.SendHelp(bot, update.Message.Chat.ID)
				case "start":
					err = commands.SendStart(bot, update.Message.Chat.ID)
				}
			}
		case "waiting_for_event_name":
			err = commands.HandleNewEventName(db, bot, update.Message)
		case "waiting_for_description":
			err = commands.HandleNewEventDescription(db, bot, update.Message)
		case "waiting_for_timestamp":
			err = commands.HandleNewEventTimestamp(db, bot, update.Message)
		case "waiting_for_location":
			err = commands.HandleNewEventLocation(db, bot, update.Message)
		}

		took := time.Since(start)
		if state == "no_command" {
			if err != nil {
				commandDuration.WithLabelValues(err.Error(), update.Message.Command()).Observe(float64(took) / float64(time.Second))
			} else {
				commandDuration.WithLabelValues("", update.Message.Command()).Observe(float64(took) / float64(time.Second))
			}
		} else {
			if err != nil {
				messageDuration.WithLabelValues(err.Error(), state).Observe(float64(took) / float64(time.Second))
			} else {
				messageDuration.WithLabelValues("", state).Observe(float64(took) / float64(time.Second))
			}
		}
		return err

	} else if update.InlineQuery != nil {
		return utils.HandleSummary(inlineQueryDuration, func() error {
			return inline.HandleInlineQuery(db, bot, update.InlineQuery)
		})
	} else if update.ChosenInlineResult != nil {
		return inline.HandleChoseInlineResult(db, update.ChosenInlineResult)
	} else if update.CallbackQuery != nil {
		return utils.HandleSummary(callbackDuration, func() error {
			return callback.HandleCallback(db, bot, update.CallbackQuery)
		})
	}

	return nil
}
