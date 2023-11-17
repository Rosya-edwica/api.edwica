package database

import (
	"fmt"

	"github.com/go-faster/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	Addr     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	DB       string `yaml:"name" env-required:"true"`
}

var db *sqlx.DB

func New(cfg *Config) (*sqlx.DB, error) {
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.User, cfg.Password, cfg.Addr, cfg.Port, cfg.DB)
	conn, err := sqlx.Open("mysql", dataSource)
	if err != nil {
		return nil, errors.Wrap(err, "mysql-connection")
	}
	db = conn
	err = db.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "mysql-ping-failed")
	}
	return db, nil
}

func GetDB() *sqlx.DB {
	return db
}

func Close(conn *sqlx.DB) error {
	err := conn.Close()
	if err != nil {
		return errors.Wrap(err, "closing mysql connection")
	}
	return nil
}
