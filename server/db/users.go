package db

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func (manager *Manager) Register(table, email, password, username string) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	var id int
	if table == "hr" {
		query := fmt.Sprintf(`INSERT INTO %s (email, hash_password, username) VALUES ($1, $2, $3) RETURNING id`, table)
		err = manager.Conn.QueryRow(query, email, hashedPassword, username).Scan(&id)
		if err != nil {
			return 0, err
		}
	} else if table == "users" {
		query := fmt.Sprintf(`INSERT INTO %s (email, hash_password) VALUES ($1, $2) RETURNING id`, table)
		err = manager.Conn.QueryRow(query, email, hashedPassword).Scan(&id)
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}

func (manager *Manager) Authenticate(table, email, password string) (int, string, error) {
	var hash, username string
	var id int
	var err error

	if table == "hr" {
		err = manager.Conn.QueryRow(fmt.Sprintf(`SELECT hash_password, username, id FROM %s WHERE email=$1`, table), email).Scan(&hash, &username, &id)
	} else if table == "users" {
		err = manager.Conn.QueryRow(fmt.Sprintf(`SELECT hash_password, id FROM %s WHERE email=$1`, table), email).Scan(&hash, &id)
	}
	if err != nil {
		return 0, "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return 0, "", err
	}

	return id, username, nil
}

func (manager *Manager) UpdateUser(role string, email string, username string, userId int) error {
	if role == "hr" {
		query := `
			UPDATE hr
			SET email = $1, username = $2
			WHERE id = $3`
		_, err := manager.Conn.Exec(query, email, username, userId)
		if err != nil {
			return err
		}
	} else if role == "users" {
		query := `
			UPDATE users
			SET email = $1
			WHERE id = $2`
		_, err := manager.Conn.Exec(query, email, userId)
		if err != nil {
			return err
		}
	}
	return nil
}