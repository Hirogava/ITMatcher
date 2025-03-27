package db

import (
	"database/sql"
	"sync"
)

type DBManager struct {
	DB *sql.DB
	WG *sync.WaitGroup
	MU *sync.RWMutex
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
	wg := &sync.WaitGroup{}
	mu := &sync.RWMutex{}
	return &DBManager{DB: db, WG: wg, MU: mu}, nil
}

func (d *DBManager) Close() {
	d.DB.Close()
	d.DB = nil
}