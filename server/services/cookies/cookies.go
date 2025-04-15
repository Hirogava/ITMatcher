package cookies

import (
	"gaspr/models"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

type Manager struct {
	Session *sessions.Session
}

var store *sessions.CookieStore

func Init(key string) {
	store = sessions.NewCookieStore([]byte(key))
	store.Options.HttpOnly = true
	store.Options.Secure = false
	store.Options.SameSite = http.SameSiteStrictMode
}

func NewCookieManager(r *http.Request) *Manager {
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("Ошибка при получении сессии: %v", session)
	}
	return &Manager{
		Session: session,
	}
}

func GetUsername(r *http.Request) *string {
	store := NewCookieManager(r)
	username, ok := store.Session.Values["username"].(string)
	if !ok {
		return nil
	}
	return &username
}

func GetId(r *http.Request) *int {
	store := NewCookieManager(r)
	id, ok := store.Session.Values["user_id"].(int)
	if !ok {
		return nil
	}
	return &id
}

func GetAccount(r *http.Request) *models.Account {
	store := NewCookieManager(r)
	role, ok := store.Session.Values["role"].(string)
	if !ok {
		return nil
	}
	if role == "hr" {
		hr_id := store.Session.Values["user_id"].(int)
		username := store.Session.Values["username"].(string)
		email := store.Session.Values["email"].(string)

		return &models.Account{
			HR: &models.HR{
				ID:       hr_id,
				Username: username,
				Email:    email,
			},
			User: nil,
		}
	} else if role == "users" {
		user_id := store.Session.Values["user_id"].(int)
		email := store.Session.Values["email"].(string)

		return &models.Account{
			HR: nil,
			User: &models.User{
				ID:    user_id,
				Email: email,
			},
		}
	}
	return nil

}
