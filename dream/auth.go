package dream

import (
	"net/http"

	"github.com/trianglehasfoursides/bedroompop/flags"
)

func MiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			write(w, http.StatusInternalServerError, "can't authenticate")
			return
		}

		isValid := (username == flags.Username) && (password == flags.Password)
		if !isValid {
			write(w, http.StatusUnauthorized, "wrong username/password")
			return
		}

		next.ServeHTTP(w, r)
	})
}
