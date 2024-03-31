package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type PostgresDB struct {
	db          *bun.DB
	Users       UserModel
	Hosts       HostModel
	Detectors   DetectorModel
	Recipients  RecipientModel
	Preferences PreferenceModel
	Sessions    SessionModel
}

func New(dsn string) (*PostgresDB, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(1000)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return &PostgresDB{
		db:          db,
		Users:       UserModel{db: db, log: logger.New("user_repo")},
		Hosts:       HostModel{db: db, log: logger.New("host_repo")},
		Detectors:   DetectorModel{db: db, log: logger.New("detector_repo")},
		Recipients:  RecipientModel{db: db, log: logger.New("recipient_repo")},
		Preferences: PreferenceModel{db: db, log: logger.New("preferences_repo")},
		Sessions:    SessionModel{db: db, log: logger.New("session_repo")},
	}, nil
}

func (db *PostgresDB) Close() error {
	return db.db.Close()
}

func (db *PostgresDB) Bun() *bun.DB {
	return db.db
}
