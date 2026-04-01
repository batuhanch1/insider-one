package config

import (
	"context"
	"fmt"
	"insider-one/infrastructure/logging"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App           AppConfig           `mapstructure:"APP"`
	DB            DBConfig            `mapstructure:"DB"`
	Rabbit        RabbitConfig        `mapstructure:"RABBIT"`
	SmsProvider   SmsProviderConfig   `mapstructure:"SMSPROVIDER"`
	EmailProvider EmailProviderConfig `mapstructure:"EMAILPROVIDER"`
	PushProvider  PushProviderConfig  `mapstructure:"PUSHPROVIDER"`
}

type AppConfig struct {
	Port  int    `mapstructure:"PORT"`
	Name  string `mapstructure:"NAME"`
	Debug bool   `mapstructure:"DEBUG"`
}

type DBConfig struct {
	Port     int    `mapstructure:"PORT"`
	Host     string `mapstructure:"HOST"`
	Name     string `mapstructure:"NAME"`
	User     string `mapstructure:"USER"`
	Password string `mapstructure:"PASSWORD"`
}

type RabbitConfig struct {
	Port     int    `mapstructure:"PORT"`
	Host     string `mapstructure:"HOST"`
	User     string `mapstructure:"USER"`
	Password string `mapstructure:"PASSWORD"`
}

type SmsProviderConfig struct {
	Host     string `mapstructure:"HOST"`
	User     string `mapstructure:"USER"`
	Password string `mapstructure:"PASSWORD"`
}
type EmailProviderConfig struct {
	Host     string `mapstructure:"HOST"`
	User     string `mapstructure:"USER"`
	Password string `mapstructure:"PASSWORD"`
}
type PushProviderConfig struct {
	Host     string `mapstructure:"HOST"`
	User     string `mapstructure:"USER"`
	Password string `mapstructure:"PASSWORD"`
}

func Load(ctx context.Context, appName, environment string) (*Config, error) {
	path := fmt.Sprintf("./projects/%s/%s.env", appName, environment)
	viper.SetConfigFile(path)
	viper.SetConfigType("env")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		err = fmt.Errorf(".env cannot read: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	cfg := &Config{}

	if err := viper.Unmarshal(cfg); err != nil {
		err = fmt.Errorf("config parse error: %w", err)
		logging.Error(ctx, err)
		return nil, err
	}

	return cfg, nil
}
