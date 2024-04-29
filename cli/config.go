package cli

import (
	"fmt"
	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/spf13/viper"
)

type Config struct {
	SmartlockID  int64          `mapstructure:"smartlock_id"`
	NukiAPIToken string         `mapstructure:"nuki_api_token"`
	AddressID    int64          `mapstructure:"address_id"`
	Senders      []SenderConfig `mapstructure:"senders"`
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

	return vi.Unmarshal(c)
}

func (c Config) GetSender(name string) (messaging.Sender, error) {
	for _, s := range c.Senders {
		if s.Name == name {
			return s.GetSender()
		}
	}
	return nil, fmt.Errorf("unable to find sender %s", name)
}
