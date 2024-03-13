package main

import (
	"bao2803/photo_gallery/middleware"
	"bao2803/photo_gallery/models"
	"fmt"
	"net/http"

	"bao2803/photo_gallery/controllers"

	"github.com/gorilla/mux"
)

// TODO: move to config file
const (
	host     = "localhost"
	port     = 5432
	user     = "bao2803"
	password = "bao28032003"
	dbname   = "photo_gallery_dev"
)

func main() {
	// Create a DB connection string and then use it to create our model services.
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s"+
		" sslmode=disable", host, port, user, password, dbname)
	services, err := models.NewServices(psqlInfo)
	if err != nil {
		panic(err)
	}
	defer services.Close()
	services.AutoMigrate()

	r := mux.NewRouter()

	// Controllers
	usersC := controllers.NewUsers(services.User)
	galleriesC := controllers.NewGalleries(services.Gallery, r)
	staticC := controllers.NewStatic()

	// User routes
	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.HandleFunc("/cookietest", usersC.CookieTest).Methods("GET")

	// Middlewares for gallery
	requireUserMw := middleware.RequireUser{
		UserService: services.User,
	}

	// Gallery routes
	r.Handle("/galleries/new", requireUserMw.Apply(galleriesC.New)).Methods("GET")
	r.HandleFunc("/galleries", requireUserMw.ApplyFn(galleriesC.Create)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesC.Show).Methods("GET").Name(controllers.ShowGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", requireUserMw.ApplyFn(galleriesC.Edit)).Methods("GET")

	// Misc
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/faq", staticC.Faq).Methods("GET")
	r.NotFoundHandler = staticC.NotFound

	http.ListenAndServe(":3000", r)
}
