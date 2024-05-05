package telegrambot

import (
	"fmt"

	"github.com/enescakir/emoji"
	"github.com/rs/zerolog/log"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

var (
	menuBattery = fmt.Sprintf("%s %s", emoji.Battery.String(), "battery")
	menuCode    = fmt.Sprintf("%s %s", emoji.InputNumbers.String(), "code")
	menuHelp    = fmt.Sprintf("%s %s", emoji.QuestionMark.String(), "help")
	menuLogs    = fmt.Sprintf("%s %s", emoji.FileFolder.String(), "logs")
	menuResa    = fmt.Sprintf("%s %s", emoji.OpenBook.String(), "resas")
	menuModify  = fmt.Sprintf("%s %s", emoji.Gear.String(), "modify")
)

func (b *nukiBot) handlerMenu(update telego.Update, msg *telego.SendMessageParams) {
	log.Debug().Msg("menuHandler called")
	keyboard := tu.Keyboard(
		tu.KeyboardRow(tu.KeyboardButton(menuBattery), tu.KeyboardButton(menuLogs), tu.KeyboardButton(menuCode)),
		tu.KeyboardRow(tu.KeyboardButton(menuResa), tu.KeyboardButton(menuModify)),
		tu.KeyboardRow(tu.KeyboardButton(menuHelp)),
	).WithResizeKeyboard().WithInputFieldPlaceholder("Menu")

	msg.Text = "Menu"
	msg.ReplyMarkup = keyboard
	msg.ProtectContent = true
}
