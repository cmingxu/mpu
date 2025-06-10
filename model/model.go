package model

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"

	"github.com/rs/zerolog/log"
)

var (
	db *sqlx.DB
)

func Init(dbFile string) error {
	_db, err := sqlx.Connect("sqlite3", dbFile)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to database %s", dbFile)
	}

	db = _db

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)

	return err
}

func InitDB() error {
	tx := db.MustBegin()
	if _, err := tx.Exec(TemplateCreationSchema); err != nil {
		return errors.Wrapf(err, "failed to create templates table %s", TemplateCreationSchema)
	}

	if _, err := tx.Exec(AudioCreationSchema); err != nil {
		return errors.Wrapf(err, "failed to create audios table %s", AudioCreationSchema)
	}

	if _, err := tx.Exec(MovieCreationSchema); err != nil {
		return errors.Wrapf(err, "failed to create movies table %s", MovieCreationSchema)
	}

	if _, err := tx.Exec(TemplateInitializationStat); err != nil {
		log.Warn().Err(err).Msgf("template initialization stat failed, maybe already initialized: %s", TemplateInitializationStat)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func Close() error {
	db = nil
	return nil
}
