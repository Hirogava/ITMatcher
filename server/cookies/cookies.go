package cookies

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/gorilla/sessions"
)

type CookieManager struct {
	Session *sessions.Session
}

func NewCookieManager() *CookieManager {
	key := generateSecretKey()
	store := sessions.NewCookieStore([]byte(key))
	session, _ := store.Get(nil, "session-name")
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

