package cookies

import (
	"crypto/rand"
	"encoding/base64"
	"gaspr/db"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

type Manager struct {
	Session *sessions.Session
}

var store *sessions.CookieStore

func Init() {
	key := generateSecretKey()
	store = sessions.NewCookieStore([]byte(key))
	store.Options.HttpOnly = true
	store.Options.Secure = false
	store.Options.SameSite = http.SameSiteStrictMode
}

func NewCookieManager(r *http.Request) *Manager {
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("Ошибка при получении сессии: %v", err)
		panic(err)
	}
	log.Printf("Полученная сессия: %v", session)
	return &Manager{
		Session: session,
	}
}

func generateSecretKey() string {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(key)
}

func GetUsernameCookie(r *http.Request) *string {
	store := NewCookieManager(r)

	username, ok := store.Session.Values["username"].(string)
	if !ok {
		return nil
	}
	return &username
}

func GetHrAccountCookie(r *http.Request, manager *db.Manager) *db.HR {
	store := NewCookieManager(r)
	role, ok := store.Session.Values["role"].(string)
	if !ok {
		return nil
	}
	if role == "hr" {
		username := store.Session.Values["username"].(string)
		email := store.Session.Values["email"].(string)
		hrId, err := manager.GetHRIdByUsername(username)
		if err != nil {
			log.Printf("Ошибка при получении HR аккаунта: %v", err)
			return nil
		}
		return &db.HR{
			ID:       hrId,
			Username: username,
			Email:    email,
		}
	}
	return nil
}
