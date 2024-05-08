package telegrambot

import (
	"context"
	"fmt"

	"github.com/enescakir/emoji"
	"github.com/looplab/fsm"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/rs/zerolog/log"
)

var chatSessions = map[int64]Command{}

type CommandHandler func(update telego.Update, msgResponse *telego.SendMessageParams)

type Command struct {
	StateMachine *fsm.FSM
	NextFSMEvent string
	Handler      CommandHandler
}

func resetChatSession(chatID int64, cmd *Command) {
	log.Trace().
		Int64("chatID", chatID).
		Msg("Resetting chat session")
	if cmd != nil && cmd.StateMachine != nil {
		if err := cmd.StateMachine.Event(context.Background(), FSMEventReset); err != nil {
			log.Error().Err(err).Msg("An error occurred resetting chat session")
		}
	}
	delete(chatSessions, chatID)
}

type Commands map[string]Command

func (c Command) GetNextFSMEvent() string {
	if c.NextFSMEvent == "" {
		return FSMEventDefault
	}
	return c.NextFSMEvent
}

func (c Commands) start(b *nukiBot) error {
	bot, err := telego.NewBot(b.Sender.Token)
	if err != nil {
		return err
	}

	updates, err := bot.UpdatesViaLongPolling(nil)
	if err != nil {
		return err
	}

	go func() {
		defer bot.StopLongPolling()
	POLL:
		for update := range updates {
			// Execute all filters before proceeding
			for _, filterFunc := range b.filters {
				if !filterFunc(update) {
					logger := log.With().Logger()
					if update.Message != nil {
						logger = logger.With().
							Int64("from_id", update.Message.From.ID).
							Str("from_username", update.Message.From.Username).
							Str("from_firstname", update.Message.From.FirstName).
							Str("from_lastname", update.Message.From.LastName).
							Str("from_lang", update.Message.From.LanguageCode).
							Str("message", update.Message.Text).Logger()
					}
					logger.Warn().Msg("Message filtered.")
					continue POLL
				}
			}

			if update.CallbackQuery == nil && update.Message == nil {
				continue
			}

			// Init destinationChatID
			var destinationChatID int64
			if update.Message != nil {
				destinationChatID = update.Message.Chat.ID
			} else {
				destinationChatID = update.CallbackQuery.Message.GetChat().ID
			}

			if update.Message != nil && !IsPrivateMessage(update) {
				// Command are only executed through private messages, deleting message.
				log.Debug().Msg("Ignoring commands sent to group")
				err := bot.DeleteMessage(tu.Delete(update.Message.Chat.ChatID(), update.Message.MessageID))
				if err != nil {
					log.Error().Err(err).
						Int64("chat_id", update.Message.Chat.ID).
						Int("message_id", update.Message.MessageID).
						Str("message", update.Message.Text).
						Msg("Unable to delete unwanted message")
				}
				// if it's a command: answer response to the member
				destinationChatID = update.Message.From.ID
				log.Debug().
					Int64("from_id", destinationChatID).
					Msg("Deleted unwanted group message, sending a private response.")
			}

			var msg *telego.SendMessageParams

			if isCallback(update) { // Callback keyboard's button
				if err := bot.AnswerCallbackQuery(tu.CallbackQuery(update.CallbackQuery.ID)); err != nil {
					log.Error().Err(err).Msg("Unable to answer callback.")
				}

				msg, err = c.handleCallback(update, destinationChatID)
				if err != nil {
					msg = &telego.SendMessageParams{Text: err.Error()}
				}
			} else { // Direct message
				msg, err = c.handleMessage(bot, update, destinationChatID)
				if err != nil {
					msg = &telego.SendMessageParams{Text: err.Error()}
				}
			}

			// Sending message to client
			msg.ChatID = tu.ID(destinationChatID)
			_, err := bot.SendMessage(msg)
			if err != nil {
				log.Error().Err(err).
					Str("msg", msg.Text).
					Msg("An error occurred while sending message")
			}
		}
	}()

	return nil
}

func IsPrivateMessage(update telego.Update) bool {
	return update.Message != nil && update.Message.Chat.Type == telego.ChatTypePrivate
}

func isCallback(update telego.Update) bool {
	return update.Message == nil && update.CallbackQuery != nil
}

func (c Commands) handleCallback(update telego.Update, destinationChatID int64) (*telego.SendMessageParams, error) {
	// Should already have a FSM registered
	command, ok := chatSessions[destinationChatID]
	if !ok {
		return nil, fmt.Errorf("This button is expired, use menu or initiate a new command!")
	}

	log.Debug().
		Str("msg", update.CallbackQuery.Data).
		Str("fsm_state", command.StateMachine.Current()).
		Msg("Received callback")

	cmd := GetCommandFromCallbackData(update.CallbackQuery)
	data := GetDataFromCallbackData(update.CallbackQuery)
	err := command.StateMachine.Event(context.Background(), cmd, data)
	if err != nil {
		return nil, err
	}
	return getMetadataSendMessageParams(FSMMetadataMessage, command.StateMachine)
}

func (c Commands) handleMessage(bot *telego.Bot, update telego.Update, destinationChatID int64) (*telego.SendMessageParams, error) {
	command, ok := c[update.Message.Text]
	if !ok {
		// 2 possibilities here:
		//   - unknown command
		//   - a response to a previous command as part of a conversation with the bot
		command, ok = chatSessions[destinationChatID]
		if !ok || command.NextFSMEvent == "" { // Unknown command
			return tu.Message(tu.ID(destinationChatID), fmt.Sprintf("I don't understand %s", emoji.ManShrugging.String())), nil
		}
	} else { // reinit for a new command to be processed
		cmd, ok := chatSessions[destinationChatID]
		if ok {
			resetChatSession(destinationChatID, &cmd)
		}
	}

	// If we have a basic handler, execute that instead
	if command.Handler != nil {
		msg := &telego.SendMessageParams{}
		command.Handler(update, msg)
		return msg, nil
	}

	if command.StateMachine == nil {
		return tu.Message(tu.ID(destinationChatID), "internal error, fsm is nil"), nil
	}

	log.Debug().
		Str("msg", update.Message.Text).
		Str("fsm_state", command.StateMachine.Current()).
		Str("next_fsm_event", command.GetNextFSMEvent()).
		Msgf("Telegram message received.")

	err := command.StateMachine.Event(context.Background(), command.GetNextFSMEvent(), update.Message.Text)
	if err != nil {
		if errRecoverEvent, _ := getMetadataString(FSMMetadataErrRecoverEvent, command.StateMachine); errRecoverEvent != "" {
			// Send error message to the client
			if _, err := bot.SendMessage(tu.Message(tu.ID(destinationChatID), err.Error())); err != nil {
				log.Error().Err(err).Send()
			}
			// Transition to the recover event
			if err := command.StateMachine.Event(context.Background(), errRecoverEvent); err != nil {
				log.Error().Err(err).Msg("An error occurred calling error callback")
			}
		} else {
			return nil, err
		}
	}
	// Get next fsm event from metadata if any
	command.NextFSMEvent, err = getMetadataString(FSMMetadataNextEvent, command.StateMachine)
	if err != nil {
		command.NextFSMEvent = ""
	}
	chatSessions[destinationChatID] = command

	return getMetadataSendMessageParams(FSMMetadataMessage, command.StateMachine)
}
