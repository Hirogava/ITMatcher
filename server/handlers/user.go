package handlers

import (
	"encoding/json"
	"fmt"
	"gaspr/db"
	"gaspr/services/cookies"
	"log"
	"net/http"
)

func Login(manager *db.Manager, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
        http.Error(w, "Content-Type должен быть application/json", http.StatusBadRequest)
        return
    }

    var requestData struct {
        Role     string `json:"role"`
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    err := json.NewDecoder(r.Body).Decode(&requestData)
    if err != nil {
        http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
        return
    }

    log.Println("role:", requestData.Role, "email:", requestData.Email, "password:", requestData.Password)

	id, username, err := manager.Authenticate(requestData.Role, requestData.Email, requestData.Password)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	store := cookies.NewCookieManager(r)
	store.Session.Values["role"] = requestData.Role
	store.Session.Values["username"] = username
	store.Session.Values["user_id"] = id

	w.Header().Set("Content-Type", "application/json")
	store.Session.Save(r, w)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "успешный вход"})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	store := cookies.NewCookieManager(r)
    store.Session.Options.MaxAge = -1
    if err := store.Session.Save(r, w); err != nil {
        http.Error(w, "Ошибка выхода", http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Register(manager *db.Manager, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
        http.Error(w, "Content-Type должен быть application/json", http.StatusBadRequest)
        return
    }

    var requestData struct {
        Role     string `json:"role"`
        Email    string `json:"email"`
        Username string `json:"username"`
        Password string `json:"password"`
    }

    err := json.NewDecoder(r.Body).Decode(&requestData)
    if err != nil {
        http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
        return
    }

    log.Println("role:", requestData.Role, "email:", requestData.Email, "username:", requestData.Username, "password:", requestData.Password)

	id, err := manager.Register(requestData.Role, requestData.Email, requestData.Username, requestData.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed: %v", err), http.StatusUnauthorized)
		return
	}

	store := cookies.NewCookieManager(r)
	store.Session.Values["role"] = requestData.Role
	store.Session.Values["username"] = requestData.Username
	store.Session.Values["user_id"] = id

	w.Header().Set("Content-Type", "application/json")
	store.Session.Save(r, w)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "успешная регистрация"})
}
