package db

import (
	"database/sql"
)

type DBManager struct {
	DB *sql.DB
}

func NewDBManager(driver string, connStr string) (*DBManager, error) {
	db, err := sql.Open(driver, connStr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &DBManager{DB: db}, nil
}

func (d *DBManager) Close() {
	d.DB.Close()
	d.DB = nil
}