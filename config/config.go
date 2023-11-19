package config

import (
	"github.com/Rosya-edwica/api.edwica/internal/database"
	"github.com/Rosya-edwica/api.edwica/internal/server"
	"github.com/go-faster/errors"
	"github.com/spf13/viper"
)

func init() {
	viper.AddConfigPath("config/")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
}

func LoadDBConfig() (*database.Config, error) {
	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "reading http config")
	}
	return &database.Config{
		Addr:     viper.GetString("db_host"),
		Port:     viper.GetInt("db_port"),
		User:     viper.GetString("db_user"),
		Password: viper.GetString("db_password"),
		DB:       viper.GetString("db_name"),
	}, nil
}

func LoadHTTPConfig() (*server.Config, error) {
	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "reading http config")
	}
	return &server.Config{
		Port: viper.GetString("http_port"),
	}, nil
}
