package telegrambot

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *nukiBot) handlerResa(update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
	res, err := b.reservationsReader.Execute()
	if err != nil {
		msg.Text = fmt.Sprintf("Unable to get reservations from API, err=%v", err)
		return
	}

	now := time.Now()
	var lines []string
	for _, r := range res {
		isBold := now.After(r.StartDate) && now.Before(r.EndDate)
		loc, err := time.LoadLocation(b.sender.Timezone)
		if err != nil {
			loc = time.UTC
		}
		startDate := r.StartDate.In(loc).Format("02/01 15:04")
		endDate := r.EndDate.In(loc).Format("02/01 15:04")
		line := fmt.Sprintf("%s (%s) - %s -> %s", r.Name, r.Reference, startDate, endDate)
		if isBold {
			line = "*" + line + "*"
		}
		lines = append(lines, line)
	}
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.Text = fmt.Sprintf("Reservations:\n%s", strings.Join(lines, "\n"))
}
