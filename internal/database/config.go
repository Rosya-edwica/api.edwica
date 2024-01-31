package database

import (
	"fmt"

	"github.com/Rosya-edwica/api.edwica/config"
	"github.com/Rosya-edwica/api.edwica/pkg/logger"
	"github.com/go-faster/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func New(cfg *config.DB) (*sqlx.DB, error) {
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.User, cfg.Password, cfg.Addr, cfg.Port, cfg.DB)
	conn, err := sqlx.Open("mysql", dataSource)
	if err != nil {
		logger.Log.Error("database.config.connection:" + err.Error())
		return nil, errors.Wrap(err, "mysql-connection")
	}
	db = conn
	err = db.Ping()
	if err != nil {
		logger.Log.Error("database.config.ping:" + err.Error())
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
		logger.Log.Error("database.config.close:" + err.Error())
		return errors.Wrap(err, "closing mysql connection")
	}
	return nil
}
