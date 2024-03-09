package middleware

import (
	"bao2803/photo_gallery/context"
	"bao2803/photo_gallery/models"
	"fmt"
	"net/http"
)

type RequireUser struct {
	models.UserService
}

// ApplyFn will return a http.HandlerFunc that will check to see if a user is logged in
// It then either call next(w, r) if they are, or redirect them to the login page if they are not.
func (mw *RequireUser) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("remember_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		user, err := mw.UserService.ByRemember(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		fmt.Println("User found: ", user)
		ctx := r.Context()
		ctx = context.WithValue(ctx, user)
		r = r.WithContext(ctx)
		next(w, r)
	}
}

func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}
