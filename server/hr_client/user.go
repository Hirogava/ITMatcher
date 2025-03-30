package hrclient

import (
	"gaspr/db"
	"gaspr/cookies"
)

func Login(email string, password string, db *db.DBManager, store *cookies.CookieManager) error {

	id, username, err := db.CheckHr(email, password)
	if err != nil {
		return err
	}

	store.Session.Values["username"] = username
	store.Session.Values["user_id"] = id

	return nil
}

func Logout(username string, store *cookies.CookieManager) error {
	store.Session.Values["username"] = ""
	store.Session.Values["user_id"] = ""

	return nil
}

func Register(password string, email string, db *db.DBManager,  store *cookies.CookieManager) error {
	
	id, username, err := db.RegisterHr(email, password)
	if err != nil {
		return err
	}

	store.Session.Values["username"] = username
	store.Session.Values["user_id"] = id

	return nil
}