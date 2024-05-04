package telegrambot

import (
	"fmt"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func (b *nukiBot) handlerCode(update telego.Update, msg *telego.SendMessageParams) {
	res, err := b.reservationsReader.Execute()
	if err != nil {
		msg.Text = fmt.Sprintf("Unable to get reservations from API, err=%v", err)
		return
	}

	var keyboardButtons []telego.InlineKeyboardButton
	for _, r := range res {
		keyboardButtons = append(keyboardButtons,
			tu.InlineKeyboardButton(fmt.Sprintf("%s (%s)", r.Name, r.Reference)).
				WithCallbackData(NewCallbackData("code", r.Reference)))
	}
	keyboard := tu.InlineKeyboard(keyboardButtons)

	msg.ReplyMarkup = keyboard
	msg.Text = "Select a reservation"
	msg.ProtectContent = true
}

func (b *nukiBot) callbackCode(update telego.Update, msg *telego.SendMessageParams) {
	data := GetDataFromCallbackData(update.CallbackQuery)
	if data == "" {
		msg.Text = "Unknown data"
		return
	}
	res, err := b.smartlockAuthReader.Execute()
	if err != nil {
		msg.Text = fmt.Sprintf("Unable to get smartlock auth from API, err=%v", err)
	}
	for _, v := range res {
		if v.Name == data {
			msg.Text = fmt.Sprintf("code for %s: %d", v.Name, v.Code)
			return
		}
	}
	msg.Text = fmt.Sprintf("Unable to find code for %s", data)
}
