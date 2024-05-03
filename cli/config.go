package cli

import (
	"fmt"

	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/nukiapi"
	"github.com/spf13/viper"
)

type Config struct {
	SmartlockID  int64          `mapstructure:"smartlock_id"`
	NukiAPIToken string         `mapstructure:"nuki_api_token"`
	AddressID    int64          `mapstructure:"address_id"`
	Senders      []SenderConfig `mapstructure:"senders"`
	TelegramBot  struct {
		Enabled    bool   `mapstructure:"enabled"`
		SenderName string `mapstructure:"sender_name"`
	} `mapstructure:"telegram_bot"`
	LogsReader          nukiapi.LogsReader          `mapstructure:"-"`
	SmartlockReader     nukiapi.SmartlockReader     `mapstructure:"-"`
	SmartlockAuthReader nukiapi.SmartlockAuthReader `mapstructure:"-"`
	ReservationsReader  nukiapi.ReservationsReader  `mapstructure:"-"`
}

func (c *Config) initReaders() {
	c.LogsReader = nukiapi.LogsReader{
		APICaller:   nukiapi.APICaller{Token: c.NukiAPIToken},
		SmartlockID: c.SmartlockID,
		Limit:       20,
	}
	c.SmartlockReader = nukiapi.SmartlockReader{
		APICaller:   nukiapi.APICaller{Token: c.NukiAPIToken},
		SmartlockID: c.SmartlockID,
	}
	c.ReservationsReader = nukiapi.ReservationsReader{
		APICaller: nukiapi.APICaller{Token: c.NukiAPIToken},
		AddressID: c.AddressID,
	}
	c.SmartlockAuthReader = nukiapi.SmartlockAuthReader{
		APICaller:   nukiapi.APICaller{Token: c.NukiAPIToken},
		SmartlockID: c.SmartlockID,
	}
}

type SenderConfig struct {
	Name     string                    `mapstructure:"name"`
	Telegram *messaging.TelegramSender `mapstructure:"telegram"`
	Console  *messaging.ConsoleSender  `mapstructure:"console"`
}

func (sc *SenderConfig) GetSender() (messaging.Sender, error) {
	if sc.Telegram != nil {
		if sc.Telegram.Token == "" || sc.Telegram.ChatID == 0 {
			return nil, fmt.Errorf("error creating telegram sender, token or chatid is not specified for %s", sc.Name)
		}

		sc.Telegram.Name = sc.Name
		return sc.Telegram, nil
	}

	if sc.Console != nil {
		sc.Console.Name = sc.Name
		return sc.Console, nil
	}

	return nil, fmt.Errorf("cannot find sender for %s", sc.Name)
}

func (c *Config) LoadConfig(vi *viper.Viper) error {
	if err := vi.ReadInConfig(); err != nil {
		return err
	}

	if err := vi.Unmarshal(c); err != nil {
		return err
	}
	c.initReaders()

	return nil
}

func (c *Config) GetSender(name string) (messaging.Sender, error) {
	for _, s := range c.Senders {
		if s.Name == name {
			return s.GetSender()
		}
	}
	return nil, fmt.Errorf("unable to find sender %s", name)
}
