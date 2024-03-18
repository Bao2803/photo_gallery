package middleware

import (
	"bao2803/photo_gallery/context"
	"bao2803/photo_gallery/models"
	"net/http"
	"strings"
)

// User middleware will look up the current user via their remember_token cookie using the UserService.
// If the user is found, they will be set on the request context.
// Regardless, the next handler is always called.
type User struct {
	models.UserService
}

// ApplyFn will return a http.HandlerFunc that will check to see if a user is logged in.
func (mw *User) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Prevent db lookup for serving static files
		path := r.URL.Path
		if strings.HasPrefix(path, "/assets/") ||
			strings.HasPrefix(path, "/images/") {
			next(w, r)
			return
		}
		// Check if user is logged in
		cookie, err := r.Cookie("remember_token")
		if err != nil { // Continue WITHOUT setting the context if there is no cookie
			next(w, r)
			return
		}
		// Validate the cookie
		user, err := mw.UserService.ByRemember(cookie.Value)
		if err != nil { // Continue WITHOUT setting the context if the cookie is not valid
			next(w, r)
			return
		}
		// Valid cookie; setting context for followed handlers
		ctx := r.Context()
		ctx = context.WithValue(ctx, user)
		r = r.WithContext(ctx)
		next(w, r)
	}
}

// Apply will call ApplyFn on the input's ServeHTTP method to check the user's log in status.
func (mw *User) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

// RequireUser will redirect a user to the /login page if they are not logged in.
// This middleware assumes that User middleware has already been run, otherwise it will always redirect users.
type RequireUser struct{}

// ApplyFn will return a http.HandlerFunc that will redirect the user to the login page if they are not logged in.
func (mw *RequireUser) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context()) // retrieving user from context
		if user == nil {                  // user is not logged in -> redirect
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r) // user is logged in -> continue
	}
}

// Apply will call ApplyFn on the input's ServeHTTP method
// to redirect the user to the login page if they are not logged in.
func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}
