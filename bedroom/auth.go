package bedroom

import (
	"net/http"

	"github.com/trianglehasfoursides/bedroompop/flags"
)

func MiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "can't authenticate", http.StatusInternalServerError)
			return
		}

		isValid := (username == flags.Username) && (password == flags.Password)
		if !isValid {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
