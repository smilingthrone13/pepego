package database

import (
	"apubot/internal/config"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pkg/errors"
	"log"
)

type DB struct {
	conn *sql.DB
}

func New(cfg *config.Config) *DB {
	conn, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	err = conn.Ping()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	migrationsDir := "./migrations"
	err = migrationUp(cfg.DBPath, migrationsDir)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	return &DB{conn: conn}
}

func (db *DB) Conn() *sql.DB {
	return db.conn
}

func migrationUp(connString, migrationsDir string) error {
	m, err := migrate.New(
		"file://"+migrationsDir,
		"sqlite3://"+connString,
	)
	if err != nil {
		return errors.Wrap(err, "can not create migrate instance")
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return errors.Wrap(err, "error applying migrations")
	}

	log.Println("Successfully applied migrations")

	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}
