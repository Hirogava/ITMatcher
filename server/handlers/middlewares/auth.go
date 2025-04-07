package middleware

import (
	"gaspr/services/cookies"
	"net/http"
)

func AuthRequired(userType string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store := cookies.NewCookieManager(r)
		if store.Session.Values["role"] != userType {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
