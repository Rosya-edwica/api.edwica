package config

import (
	"github.com/go-faster/errors"
	"github.com/spf13/viper"
)

var (
	DefaultPath = "config/"
	DefaultName = "config"
	DefaultType = "yaml"
)

func setViperSettings() {
	viper.AddConfigPath(DefaultPath)
	viper.SetConfigName(DefaultName)
	viper.SetConfigType(DefaultType)
}

type DB struct {
	URL string `yaml:"db_url"`
}

type Server struct {
	Port string `yaml:"port" env-required:"true"`
}

type Telegram struct {
	Token string
	Chats []string
}

func LoadDBConfig(path string) (*DB, error) {
	if path != "" {
		DefaultPath = path
	}
	setViperSettings()
	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "reading db config")
	}
	return &DB{
		URL: viper.GetString("db_url"),
	}, nil
}

func LoadHTTPConfig() (*Server, error) {
	setViperSettings()
	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "reading http config")
	}
	return &Server{
		Port: viper.GetString("http_port"),
	}, nil
}

func LoadTelegramConfig() (*Telegram, error) {
	setViperSettings()
	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "reading telegram config")
	}
	return &Telegram{
		Token: viper.GetString("telegram_token"),
		Chats: viper.GetStringSlice("telegram_chats"),
	}, nil
}
