package commands

import (
	"database/sql"
	"log"
	"time"

	"gitlab.com/schoentoon/event-bot/database"
	"gitlab.com/schoentoon/event-bot/events"
	"gitlab.com/schoentoon/event-bot/templates"
	"gitlab.com/schoentoon/event-bot/timestamp"
	"gitlab.com/schoentoon/event-bot/utils"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func HandleNewEventCommand(db *sql.DB, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	err := func(db *sql.DB, msg *tgbotapi.Message) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}

		_, err = tx.Exec(`INSERT INTO public.drafts (user_id)
			VALUES ($1)`,
			msg.From.ID)
		if err != nil {
			return database.TxRollback(tx, err)
		}

		err = database.ChangeUserStateTx(tx, msg.From.ID, "waiting_for_event_name")
		if err != nil {
			return database.TxRollback(tx, err)
		}

		return tx.Commit()
	}(db, msg)

	var reply tgbotapi.MessageConfig
	if err == nil {
		rendered, err := templates.Execute("created_new_event.tmpl", nil)
		if err != nil {
			return err
		}
		reply = tgbotapi.NewMessage(msg.Chat.ID, rendered)
	} else {
		return utils.NewErrorWithChattableFromTemplate(err, "something_went_wrong_try_later.tmpl", msg.Chat.ID)
	}

	reply.ReplyToMessageID = msg.MessageID
	_, err = utils.Send(bot, reply)
	return err
}

func HandleNewEventName(db *sql.DB, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	if len(msg.Text) == 0 || len(msg.Text) > 128 {
		rendered, err := templates.Execute("name_too_long.tmpl", nil)
		if err != nil {
			return err
		}

		reply := tgbotapi.NewMessage(msg.Chat.ID, rendered)
		reply.ReplyToMessageID = msg.MessageID

		_, err = utils.Send(bot, reply)
		return err
	}

	edited, err := func(db *sql.DB, msg *tgbotapi.Message) (bool, error) {
		tx, err := db.Begin()
		if err != nil {
			return false, err
		}

		row := tx.QueryRow(`SELECT id FROM public.events
			WHERE wants_edit
			AND owner = $1`, msg.From.ID)
		var eventID int64
		err = row.Scan(&eventID)
		if err != nil {
			// in case of a no rows error we're changing a name of an event
			if err == sql.ErrNoRows {
				_, err = tx.Exec(`UPDATE public.drafts SET name = $1 WHERE user_id = $2`, msg.Text, msg.From.ID)
				if err != nil {
					return false, database.TxRollback(tx, err)
				}

				err = database.ChangeUserStateTx(tx, msg.From.ID, "waiting_for_description")
				if err != nil {
					return false, database.TxRollback(tx, err)
				}

				return false, tx.Commit()
			}
			return false, database.TxRollback(tx, err)
		}

		_, err = tx.Exec(`UPDATE public.events
			SET name = $1,
			wants_edit = false
			WHERE id = $2
			AND wants_edit`, msg.Text, eventID)
		if err != nil {
			return false, database.TxRollback(tx, err)
		}

		err = database.ChangeUserStateTx(tx, msg.From.ID, "no_command")
		if err != nil {
			return false, database.TxRollback(tx, err)
		}

		err = events.NeedsUpdate(tx, eventID)
		if err != nil {
			return false, database.TxRollback(tx, err)
		}

		var messageID int
		row = tx.QueryRow(`SELECT settings_message_id
			FROM public.events
			WHERE id = $1`,
			eventID)
		err = row.Scan(&messageID)
		if err != nil {
			return false, database.TxRollback(tx, err)
		}

		rendered, _, err := events.FormatEventSettings(tx, eventID)
		if err != nil {
			return false, err
		}

		edit := tgbotapi.NewEditMessageText(msg.Chat.ID, messageID, rendered)
		edit.ReplyMarkup = utils.CreateEventCreatedKeyboard(eventID)
		edit.ParseMode = "HTML"

		_, err = utils.Send(bot, edit)
		if err != nil {
			log.Println(err)
		}

		return true, tx.Commit()
	}(db, msg)

	var rendered string
	if err != nil {
		return utils.NewErrorWithChattableFromTemplate(err, "something_went_wrong_try_later.tmpl", msg.Chat.ID)
	} else if edited {
		rendered, err = templates.Execute("name_edited.tmpl", nil)
	} else {
		rendered, err = templates.Execute("name_set_enter_description.tmpl", nil)
	}
	if err != nil {
		return err
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, rendered)
	reply.ReplyToMessageID = msg.MessageID

	_, err = utils.Send(bot, reply)
	return err
}

func HandleNewEventDescription(db *sql.DB, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	description := msg.Text
	if msg.IsCommand() && msg.Command() == "skip" {
		description = ""
	} else if len(msg.Text) == 0 || len(msg.Text) > 512 {
		rendered, err := templates.Execute("description_too_long.tmpl", nil)
		if err != nil {
			return err
		}

		reply := tgbotapi.NewMessage(msg.Chat.ID, rendered)
		reply.ReplyToMessageID = msg.MessageID

		_, err = utils.Send(bot, reply)
		return err
	}

	edited, err := func(db *sql.DB, msg *tgbotapi.Message) (bool, error) {
		tx, err := db.Begin()
		if err != nil {
			return false, err
		}

		row := tx.QueryRow(`SELECT id FROM public.events
			WHERE wants_edit
			AND owner = $1`, msg.From.ID)
		var eventID int64
		err = row.Scan(&eventID)
		if err != nil {
			// in case of a no rows error we're changing a description of an event
			if err == sql.ErrNoRows {
				_, err = tx.Exec(`UPDATE public.drafts SET description = $1 WHERE user_id = $2`, description, msg.From.ID)
				if err != nil {
					return false, database.TxRollback(tx, err)
				}

				err = database.ChangeUserStateTx(tx, msg.From.ID, "waiting_for_timestamp")
				if err != nil {
					return false, database.TxRollback(tx, err)
				}

				return false, tx.Commit()
			}
			return false, database.TxRollback(tx, err)
		}

		_, err = tx.Exec(`UPDATE public.events
			SET description = $1,
			wants_edit = false
			WHERE id = $2
			AND wants_edit`, description, eventID)
		if err != nil {
			return false, database.TxRollback(tx, err)
		}

		err = database.ChangeUserStateTx(tx, msg.From.ID, "no_command")
		if err != nil {
			return false, database.TxRollback(tx, err)
		}

		err = events.NeedsUpdate(tx, eventID)
		if err != nil {
			return false, database.TxRollback(tx, err)
		}

		var messageID int
		row = tx.QueryRow(`SELECT settings_message_id
			FROM public.events
			WHERE id = $1`,
			eventID)
		err = row.Scan(&messageID)
		if err != nil {
			return false, database.TxRollback(tx, err)
		}

		rendered, _, err := events.FormatEventSettings(tx, eventID)
		if err != nil {
			return false, err
		}

		edit := tgbotapi.NewEditMessageText(msg.Chat.ID, messageID, rendered)
		edit.ReplyMarkup = utils.CreateEventCreatedKeyboard(eventID)
		edit.ParseMode = "HTML"

		_, err = utils.Send(bot, edit)
		if err != nil {
			log.Println(err)
		}

		return true, tx.Commit()
	}(db, msg)

	var rendered string
	if err != nil {
		return utils.NewErrorWithChattableFromTemplate(err, "something_went_wrong_try_later.tmpl", msg.Chat.ID)
	} else if edited {
		rendered, err = templates.Execute("description_edited.tmpl", nil)
	} else {
		rendered, err = templates.Execute("description_set_enter_timestamp.tmpl", nil)
	}
	if err != nil {
		return err
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, rendered)
	reply.ReplyToMessageID = msg.MessageID

	_, err = utils.Send(bot, reply)
	return err
}

func HandleNewEventTimestamp(db *sql.DB, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	when, err := timestamp.ParseTimestampMessage(msg.Text)
	if err != nil {
		rendered, err := templates.Execute("invalid_timestamp.tmpl", nil)
		if err != nil {
			return err
		}

		reply := tgbotapi.NewMessage(msg.Chat.ID, rendered)
		reply.ReplyToMessageID = msg.MessageID

		_, err = utils.Send(bot, reply)
		return err
	}

	edited, err := func(db *sql.DB, msg *tgbotapi.Message, when time.Time) (bool, error) {
		tx, err := db.Begin()
		if err != nil {
			return false, err
		}

		row := tx.QueryRow(`SELECT id FROM public.events
			WHERE wants_edit
			AND owner = $1`, msg.From.ID)
		var eventID int64
		err = row.Scan(&eventID)
		if err != nil {
			// in case of a no rows error we're changing a when of an event
			if err == sql.ErrNoRows {
				_, err = tx.Exec(`UPDATE public.drafts SET "when" = $1 WHERE user_id = $2`, when, msg.From.ID)
				if err != nil {
					return false, database.TxRollback(tx, err)
				}

				err = database.ChangeUserStateTx(tx, msg.From.ID, "waiting_for_location")
				if err != nil {
					return false, database.TxRollback(tx, err)
				}

				return false, tx.Commit()
			}
			return false, database.TxRollback(tx, err)
		}

		_, err = tx.Exec(`UPDATE public.events
			SET "when" = $1,
			wants_edit = false
			WHERE id = $2
			AND wants_edit`, when, eventID)
		if err != nil {
			return false, database.TxRollback(tx, err)
		}

		err = database.ChangeUserStateTx(tx, msg.From.ID, "no_command")
		if err != nil {
			return false, database.TxRollback(tx, err)
		}

		err = events.NeedsUpdate(tx, eventID)
		if err != nil {
			return false, database.TxRollback(tx, err)
		}

		var messageID int
		row = tx.QueryRow(`SELECT settings_message_id
			FROM public.events
			WHERE id = $1`,
			eventID)
		err = row.Scan(&messageID)
		if err != nil {
			return false, database.TxRollback(tx, err)
		}

		rendered, _, err := events.FormatEventSettings(tx, eventID)
		if err != nil {
			return false, err
		}

		edit := tgbotapi.NewEditMessageText(msg.Chat.ID, messageID, rendered)
		edit.ReplyMarkup = utils.CreateEventCreatedKeyboard(eventID)
		edit.ParseMode = "HTML"

		_, err = utils.Send(bot, edit)
		if err != nil {
			log.Println(err)
		}

		return true, tx.Commit()
	}(db, msg, when)

	var rendered string
	if err != nil {
		return utils.NewErrorWithChattableFromTemplate(err, "something_went_wrong_try_later.tmpl", msg.Chat.ID)
	} else if edited {
		rendered, err = templates.Execute("timestamp_edited.tmpl", nil)
	} else {
		rendered, err = templates.Execute("timestamp_set_enter_location.tmpl", nil)
	}
	if err != nil {
		return err
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, rendered)
	reply.ReplyToMessageID = msg.MessageID

	_, err = utils.Send(bot, reply)
	return err
}

func HandleNewEventLocation(db *sql.DB, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	if len(msg.Text) == 0 || len(msg.Text) > 128 {
		rendered, err := templates.Execute("invalid_location.tmpl", nil)
		if err != nil {
			return err
		}

		reply := tgbotapi.NewMessage(msg.Chat.ID, rendered)
		reply.ReplyToMessageID = msg.MessageID

		_, err = utils.Send(bot, reply)
		return err
	}

	where := msg.Text

	err := func(db *sql.DB, msg *tgbotapi.Message, where string) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}

		row := tx.QueryRow(`SELECT id FROM public.events
			WHERE wants_edit
			AND owner = $1`, msg.From.ID)
		var eventID int64
		err = row.Scan(&eventID)
		if err != nil {
			// in case of a no rows error we're changing a name of an event
			if err == sql.ErrNoRows {
				_, err = tx.Exec(`UPDATE public.drafts SET location = $1 WHERE user_id = $2`, where, msg.From.ID)
				if err != nil {
					return database.TxRollback(tx, err)
				}

				row := tx.QueryRow(`INSERT INTO public.events ("owner", name, description, "when", location)
					SELECT user_id "owner", name, description, "when", location FROM public.drafts WHERE user_id = $1
					RETURNING id`,
					msg.From.ID)
				err = row.Scan(&eventID)
				if err != nil {
					return database.TxRollback(tx, err)
				}

				err = database.ChangeUserStateTx(tx, msg.From.ID, "no_command")
				if err != nil {
					return database.TxRollback(tx, err)
				}

				_, err = tx.Exec(`DELETE FROM public.drafts WHERE user_id = $1`, msg.From.ID)
				if err != nil {
					return database.TxRollback(tx, err)
				}

				rendered, _, err := events.FormatEventSettings(tx, eventID)
				if err != nil {
					return database.TxRollback(tx, err)
				}

				reply := tgbotapi.NewMessage(msg.Chat.ID, rendered)
				reply.ReplyMarkup = utils.CreateEventCreatedKeyboard(eventID)
				reply.ReplyToMessageID = msg.MessageID
				reply.ParseMode = "HTML"

				m, err := utils.Send(bot, reply)
				if err != nil {
					return database.TxRollback(tx, err)
				}
				err = events.SetSettingsMessageIDTx(tx, eventID, m.MessageID)
				if err != nil {
					return database.TxRollback(tx, err)
				}

				return tx.Commit()
			}
			return database.TxRollback(tx, err)
		}

		_, err = tx.Exec(`UPDATE public.events
			SET location = $1,
			wants_edit = false
			WHERE id = $2
			AND wants_edit`,
			where, eventID)
		if err != nil {
			return database.TxRollback(tx, err)
		}

		err = database.ChangeUserStateTx(tx, msg.From.ID, "no_command")
		if err != nil {
			return database.TxRollback(tx, err)
		}

		err = events.NeedsUpdate(tx, eventID)
		if err != nil {
			return database.TxRollback(tx, err)
		}
		var messageID int
		row = tx.QueryRow(`SELECT settings_message_id
			FROM public.events
			WHERE id = $1`,
			eventID)
		err = row.Scan(&messageID)
		if err != nil {
			return database.TxRollback(tx, err)
		}

		rendered, _, err := events.FormatEventSettings(tx, eventID)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		edit := tgbotapi.NewEditMessageText(msg.Chat.ID, messageID, rendered)
		edit.ReplyMarkup = utils.CreateEventCreatedKeyboard(eventID)
		edit.ParseMode = "HTML"

		_, err = utils.Send(bot, edit)

		return err
	}(db, msg, where)

	if err != nil {
		return utils.NewErrorWithChattableFromTemplate(err, "something_went_wrong_try_later.tmpl", msg.Chat.ID)
	}

	return nil
}
