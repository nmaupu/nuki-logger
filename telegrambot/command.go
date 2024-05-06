package telegrambot

import (
	"context"
	"fmt"

	"github.com/enescakir/emoji"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/rs/zerolog/log"
)

var chatSessions = map[int64]Command{}

type CommandHandler func(update telego.Update, msgResponse *telego.SendMessageParams)

type Command struct {
	StateMachine FSM
	NextFSMEvent string
	Handler      CommandHandler
	Callback     CommandHandler
}

func resetChatSession(chatID int64, cmd *Command) {
	log.Trace().
		Int64("chatID", chatID).
		Msg("Resetting chat session")
	if cmd != nil && cmd.StateMachine.FSM != nil {
		if err := cmd.StateMachine.FSM.Event(context.Background(), FSMEventReset); err != nil {
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

		for update := range updates {
			if update.CallbackQuery == nil && update.Message == nil {
				continue
			}

			var destinationChatID int64
			if update.Message != nil && !isPrivateMessage(update) {
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
				destinationChatID = int64(update.Message.From.ID)
			}

			if destinationChatID == 0 {
				if update.Message != nil {
					destinationChatID = update.Message.Chat.ID
				} else {
					destinationChatID = update.CallbackQuery.Message.GetChat().ID
				}
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
				msg, err = c.handleMessage(update, destinationChatID)
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

func isPrivateMessage(update telego.Update) bool {
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
	stateM := command.StateMachine

	log.Debug().
		Str("msg", update.CallbackQuery.Data).
		Str("fsm_state", stateM.Current()).
		Msg("Received callback")

	cmd := GetCommandFromCallbackData(update.CallbackQuery)
	data := GetDataFromCallbackData(update.CallbackQuery)
	err := stateM.Event(context.Background(), cmd, data)
	if err != nil {
		return nil, err
	}
	return stateM.getMetadataSendMessageParams(FSMMetadataMessage)
}

func (c Commands) handleMessage(update telego.Update, destinationChatID int64) (*telego.SendMessageParams, error) {
	command, ok := c[update.Message.Text]
	if !ok {
		// 2 possibilities here:
		//   - unknown command
		//   - a response to a previous command as part of a conversation with the bot
		var ok bool
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

	if command.StateMachine.FSM == nil {
		return tu.Message(tu.ID(destinationChatID), "internal error, fsm is nil"), nil
	}

	log.Debug().
		Str("msg", update.Message.Text).
		Str("fsm_state", command.StateMachine.Current()).
		Str("next_fsm_event", command.GetNextFSMEvent()).
		Msgf("Telegram message received.")

	err := command.StateMachine.Event(context.Background(), command.GetNextFSMEvent(), update.Message.Text)
	if err != nil {
		return nil, err
	}
	// Get next fsm event from metadata if any
	command.NextFSMEvent, err = command.StateMachine.getMetadataString(FSMMetadataNextEvent)
	if err != nil {
		command.NextFSMEvent = ""
	}
	chatSessions[destinationChatID] = command

	return command.StateMachine.getMetadataSendMessageParams(FSMMetadataMessage)
}
