package database

import (
	"github.com/Rosya-edwica/api.edwica/config"
	"github.com/Rosya-edwica/api.edwica/pkg/logger"
	"github.com/go-faster/errors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB

func New(cfg *config.DB) (*sqlx.DB, error) {
	conn, err := sqlx.Open("postgres", cfg.URL)
	if err != nil {
		logger.Log.Error("database.config.connection:" + err.Error())
		return nil, errors.Wrap(err, "postgres-connection")
	}
	db = conn
	err = db.Ping()
	if err != nil {
		logger.Log.Error("database.config.ping:" + err.Error())
		return nil, errors.Wrap(err, "postgres-ping-failed")
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
		return errors.Wrap(err, "closing postgres connection")
	}
	return nil
}
