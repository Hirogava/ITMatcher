package hrclient

import (
	"gaspr/db"
)

func Login(email string, password string, db *db.DBManager) error {

	err := db.CheckHr(email, password)
	if err != nil {
		return err
	}

	return nil
}

func Logout(username string) string {
	return ""
}

func Register(password string, email string, db *db.DBManager) error {
	
	err := db.RegisterHr(email, password)
	if err != nil {
		return err
	}

	return nil
}