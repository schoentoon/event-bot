package utils

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

var (
	tgSendDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "bot_send_duration",
			Help: "How long does it take to actually send a message",
		},
		[]string{"error"},
	)
	tgAnswerCallbackQuery = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "bot_answer_callback",
			Help: "How long does it take to send an answer to a callbackquery",
		},
		[]string{"error"},
	)
)

func init() {
	prometheus.MustRegister(
		tgSendDuration,
		tgAnswerCallbackQuery,
	)
}

func HandleSummary(summary *prometheus.SummaryVec, f func() error) error {
	start := time.Now()
	err := f()
	took := time.Since(start)

	if err != nil {
		summary.WithLabelValues(err.Error()).Observe(float64(took) / float64(time.Second))
	} else {
		summary.WithLabelValues("").Observe(float64(took) / float64(time.Second))
	}

	return err
}

func Send(bot *tgbotapi.BotAPI, msg tgbotapi.Chattable) (m tgbotapi.Message, err error) {
	start := time.Now()

	m, err = bot.Send(msg)

	took := time.Since(start)

	if err != nil {
		tgSendDuration.WithLabelValues(err.Error()).Observe(float64(took) / float64(time.Second))
	} else {
		tgSendDuration.WithLabelValues("").Observe(float64(took) / float64(time.Second))
	}

	return m, err
}

func AnswerCallbackQuery(bot *tgbotapi.BotAPI, msg tgbotapi.CallbackConfig) (r tgbotapi.APIResponse, err error) {
	start := time.Now()

	r, err = bot.AnswerCallbackQuery(msg)

	took := time.Now().Sub(start)

	if err != nil {
		tgAnswerCallbackQuery.WithLabelValues(err.Error()).Observe(float64(took) / float64(time.Second))
	} else {
		tgAnswerCallbackQuery.WithLabelValues("").Observe(float64(took) / float64(time.Second))
	}

	return r, err
}
