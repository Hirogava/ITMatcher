package handlers

import (
	"fmt"
	"gaspr/db"
	"gaspr/services/cookies"
	"net/http"
)

func Login(manager *db.Manager, w http.ResponseWriter, r *http.Request) {
	role := r.FormValue("role")
	email := r.FormValue("email")
	password := r.FormValue("password")

	id, username, err := manager.Authenticate(role, email, password)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	store := cookies.NewCookieManager(r)
	store.Session.Values["role"] = role
	store.Session.Values["username"] = username
	store.Session.Values["user_id"] = id

	store.Session.Save(r, w)
	w.WriteHeader(http.StatusOK)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	store := cookies.NewCookieManager(r)

	store.Session.Values["role"] = ""
	store.Session.Values["username"] = ""
	store.Session.Values["user_id"] = ""

	store.Session.Save(r, w)
}

func Register(manager *db.Manager, w http.ResponseWriter, r *http.Request) {
	role := r.FormValue("role")
	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")

	id, err := manager.Register(role, email, password, username)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed: %v", err), http.StatusUnauthorized)
		return
	}

	store := cookies.NewCookieManager(r)
	store.Session.Values["role"] = role
	store.Session.Values["username"] = username
	store.Session.Values["user_id"] = id

	store.Session.Save(r, w)
	w.WriteHeader(http.StatusOK)
}
