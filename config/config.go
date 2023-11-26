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
	Addr     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	DB       string `yaml:"name" env-required:"true"`
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
		Addr:     viper.GetString("db_host"),
		Port:     viper.GetInt("db_port"),
		User:     viper.GetString("db_user"),
		Password: viper.GetString("db_password"),
		DB:       viper.GetString("db_name"),
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
