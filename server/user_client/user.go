package userclient

import (
	"gaspr/cookies"
	"gaspr/db"
	"net/http"
)

func LoginUser(email string, password string, db *db.DBManager, store *cookies.CookieManager, w http.ResponseWriter, r *http.Request) error {

	id, username, err := db.CheckUser(email, password)
	if err != nil {
		return err
	}

	store.Session.Values["username"] = username
	store.Session.Values["user_id"] = id

	return store.Session.Save(r, w)
}

func Logout(username string, store *cookies.CookieManager, w http.ResponseWriter, r *http.Request) error {
	store.Session.Values["username"] = ""
	store.Session.Values["user_id"] = ""

	return store.Session.Save(r, w)
}

func RegisterUser(password string, email string, db *db.DBManager,  store *cookies.CookieManager, w http.ResponseWriter, r *http.Request) error {
	
	id, username, err := db.RegisterUser(email, password)
	if err != nil {
		return err
	}

	store.Session.Values["username"] = username
	store.Session.Values["user_id"] = id

	return store.Session.Save(r, w)
}