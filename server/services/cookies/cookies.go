package cookies

import (
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
