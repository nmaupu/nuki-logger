package telegrambot

import (
	"fmt"

	"github.com/mymmrac/telego"
)

func (b *nukiBot) handlerBattery(update telego.Update, msg *telego.SendMessageParams) {
	res, err := b.SmartlockReader.Execute()
	if err != nil {
		msg.Text = fmt.Sprintf("Unable to read smartlock status from API, err=%v", err)
	} else {
		msg.Text = res.PrettyFormat()
	}
}
