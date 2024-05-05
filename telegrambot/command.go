package telegrambot

import (
	"context"
	"fmt"

	"github.com/looplab/fsm"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/rs/zerolog/log"
)

var (
	stateMachines = map[int64]*fsm.FSM{}
)

type CommandHandler func(update telego.Update, msgResponse *telego.SendMessageParams)

type Command struct {
	FSM      *fsm.FSM
	Handler  CommandHandler
	Callback CommandHandler
}
type Commands map[string]Command

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

	POLLING:
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

			// Init an empty message to be sent
			msgToSend := tu.Message(tu.ID(destinationChatID), "")

			var stateM *fsm.FSM
			if update.Message != nil { // direct message from user
				command, ok := c[update.Message.Text]
				if !ok {
					msgToSend.Text = "Unknown command."
					sendMessage(bot, msgToSend)
					continue POLLING
				}
				stateM = command.FSM
				stateM.SetMetadata("msg", msgToSend)
				stateMachines[destinationChatID] = stateM

				log.Debug().
					Str("msg", update.Message.Text).
					Str("fsm_state", stateM.Current()).
					Msgf("Telegram message received.")

				if err := stateM.Event(context.Background(), "run"); err != nil {
					msgToSend.Text = fmt.Sprintf("An error occurred, err=%v", err)
					sendMessage(bot, msgToSend)
					continue POLLING
				}
				sendMessage(bot, msgToSend)
			} else { // Telegram button's callback
				// Should already have a FSM registered
				var ok bool
				stateM, ok = stateMachines[destinationChatID]
				if !ok {
					msgToSend.Text = "Button is expired, use menu."
					sendMessage(bot, msgToSend)
					continue POLLING
				}
				stateM.SetMetadata("msg", msgToSend)

				log.Debug().
					Str("msg", update.CallbackQuery.Data).
					Str("fsm_state", stateM.Current()).
					Msg("Received callback")
				if err := bot.AnswerCallbackQuery(tu.CallbackQuery(update.CallbackQuery.ID)); err != nil {
					log.Error().Err(err).Msg("Unable to answer callback.")
				}
				if err := stateM.Event(context.Background(),
					GetCommandFromCallbackData(update.CallbackQuery),
					GetDataFromCallbackData(update.CallbackQuery)); err != nil {
					msgToSend.Text = fmt.Sprintf("An error occurred, err=%v", err)
				}
				sendMessage(bot, msgToSend)
			}
		}
	}()

	return nil
}

func isPrivateMessage(update telego.Update) bool {
	return update.Message != nil && update.Message.Chat.Type == telego.ChatTypePrivate
}

func sendMessage(bot *telego.Bot, msg *telego.SendMessageParams) {
	if bot == nil || msg == nil {
		log.Error().Msg("Cannot send a null message to telegram")
	}

	_, err := bot.SendMessage(msg)
	if err != nil {
		log.Error().Err(err).
			Str("msg", msg.Text).
			Msg("An error occurred while sending message")
	}
}
