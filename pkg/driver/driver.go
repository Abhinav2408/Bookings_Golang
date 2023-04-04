package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// DB holds the connection pool
type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

const maxOpenDBConn = 10
const maxIdleDBConn = 5
const maxLifetimeDB = 5 * time.Minute

//creates database pool for postgres

func ConnectSQL(dsn string) (*DB, error) {

	d, err := NewDatabase(dsn)
	if err != nil {
		panic(err)
	}

	d.SetMaxOpenConns(maxOpenDBConn)
	d.SetConnMaxIdleTime(maxIdleDBConn)
	d.SetConnMaxLifetime(maxLifetimeDB)

	dbConn.SQL = d

	err = testDb(d)
	if err != nil {
		return nil, err
	}

	return dbConn, nil

}

//tries to ping db

func testDb(d *sql.DB) error {
	err := d.Ping()
	if err != nil {
		return err
	}
	return nil
}

//creates new database

func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil

}
