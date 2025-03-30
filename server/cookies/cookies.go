package cookies

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gorilla/sessions"
)

type CookieManager struct {
	Session *sessions.Session
}

var store *sessions.CookieStore 

func init() {
	key := generateSecretKey()
	store = sessions.NewCookieStore([]byte(key))
	store.Options.HttpOnly = true
	store.Options.Secure = true
	store.Options.SameSite = http.SameSiteStrictMode

}

func NewCookieManager(r *http.Request) *CookieManager {
	session, _ := store.Get(r, "session-name")
	return &CookieManager{
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

