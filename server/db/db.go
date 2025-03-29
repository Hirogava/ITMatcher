package db

import (
	"database/sql"
	"fmt"
	"sync"

	"golang.org/x/crypto/bcrypt"
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

func (d *DBManager) CheckHr(email, password string) error{
	query := `SELECT hash_password FROM hr WHERE email=$1`
	var hash string
	err := d.DB.QueryRow(query, email).Scan(&hash)
	if err != nil{
		return err
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil{
		return err
	}
	
	return nil
}

func (db *DBManager) RegisterHr(email, password string) error{
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil{
		return fmt.Errorf("ошибка генерации хеша: %w", err)
	}
	_, err = db.DB.Exec("INSERT INTO hr (email, hash_password) VALUES ($1, $2)", email, hashedPassword)
	if err != nil{
    	return err
	}
	return nil
}