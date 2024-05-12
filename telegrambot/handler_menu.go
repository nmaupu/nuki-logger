package telegrambot

import (
	"fmt"

	"github.com/enescakir/emoji"
	"github.com/rs/zerolog/log"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

var (
	menuBattery    = fmt.Sprintf("%s %s", emoji.Battery.String(), "Battery")
	menuCode       = fmt.Sprintf("%s %s", emoji.InputNumbers.String(), "Code")
	menuHelp       = fmt.Sprintf("%s %s", emoji.QuestionMark.String(), "Help")
	menuLogs       = fmt.Sprintf("%s %s", emoji.FileFolder.String(), "Logs")
	menuResas      = fmt.Sprintf("%s %s", emoji.OpenBook.String(), "Resas")
	menuListModify = fmt.Sprintf("%s %s", emoji.Pencil.String(), "List modifs")
	menuModify     = fmt.Sprintf("%s %s", emoji.Gear.String(), "Modify")
)

func (b *nukiBot) handlerMenu(update telego.Update, msg *telego.SendMessageParams) {
	log.Debug().Msg("menuHandler called")
	keyboard := tu.Keyboard(
		tu.KeyboardRow(tu.KeyboardButton(menuBattery), tu.KeyboardButton(menuLogs), tu.KeyboardButton(menuCode)),
		tu.KeyboardRow(tu.KeyboardButton(menuResas), tu.KeyboardButton(menuModify), tu.KeyboardButton(menuListModify)),
		tu.KeyboardRow(tu.KeyboardButton(menuHelp)),
	).WithResizeKeyboard().WithInputFieldPlaceholder("Menu")

	msg.Text = "Menu"
	msg.ReplyMarkup = keyboard
	msg.ProtectContent = true
}
