package middleware

import (
	"bao2803/photo_gallery/context"
	"bao2803/photo_gallery/models"
	"net/http"
)

// User middleware will look up the current user via their remember_token cookie using the UserService.
// If the user is found, they will be set on the request context.
// Regardless, the next handler is always called.
type User struct {
	models.UserService
}

// ApplyFn will return a http.HandlerFunc that will check to see if a user is logged in
// It then either call next(w, r) if they are, or redirect them to the login page if they are not.
func (mw *User) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("remember_token")
		if err != nil {
			//http.Redirect(w, r, "/login", http.StatusFound)
			next(w, r)
			return
		}
		user, err := mw.UserService.ByRemember(cookie.Value)
		if err != nil {
			//http.Redirect(w, r, "/login", http.StatusFound)
			next(w, r)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, user)
		r = r.WithContext(ctx)
		next(w, r)
	}
}

func (mw *User) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

// RequireUser will redirect a user to the /login page if they are not logged in.
// This middleware assumes that User middleware has already been run, otherwise it will always redirect users.
type RequireUser struct{}

func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

func (mw *RequireUser) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}
