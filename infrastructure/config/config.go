package config

import (
	"fmt"
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

func Load(appName, environment string) (*Config, error) {
	path := fmt.Sprintf("./projects/%s/%s.env", appName, environment)
	viper.SetConfigFile(path)
	viper.SetConfigType("env")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf(".env cannot read: %w", err)
	}

	cfg := &Config{}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("config parse error: %w", err)
	}

	return cfg, nil
}
