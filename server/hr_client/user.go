package hrclient

import (
	"fmt"
	"gaspr/db"

	"golang.org/x/crypto/bcrypt"
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

func Register(username string, password string, email string) error {
	hash_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Ошибка генерации хеша: %w", err)
	}
	return nil
}