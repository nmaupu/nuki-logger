package cli

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/nukiapi"
	"github.com/spf13/viper"
)

// TimeHourMinute is a time.Time representing only hours and minutes
type TimeHourMinute time.Time

type Config struct {
	SmartlockID  int64          `mapstructure:"smartlock_id"`
	NukiAPIToken string         `mapstructure:"nuki_api_token"`
	AddressID    int64          `mapstructure:"address_id"`
	Senders      []SenderConfig `mapstructure:"senders"`
	TelegramBot  struct {
		Enabled           bool           `mapstructure:"enabled"`
		SenderName        string         `mapstructure:"sender_name"`
		DefaultCheckIn    TimeHourMinute `mapstructure:"default_check_in"`
		DefaultCheckOut   TimeHourMinute `mapstructure:"default_check_out"`
		RestrictToChatIDs []int64        `mapstructure:"restrict_private_chat_ids"`
	} `mapstructure:"telegram_bot"`
	HealthCheckPort     int                         `mapstructure:"health_check_port"`
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

	decoderConfigOpt := viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			StringToTimeHourMinuteHookFunc(),
		),
	)
	if err := vi.Unmarshal(c, decoderConfigOpt); err != nil {
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

func StringToTimeHourMinuteHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(TimeHourMinute(time.Now())) {
			return data, nil
		}

		// Convert it by parsing
		ti, err := time.Parse("15:04", data.(string))
		return TimeHourMinute(ti), err
	}
}
