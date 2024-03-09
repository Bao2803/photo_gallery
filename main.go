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
	dbname   = "lenslocked_dev"
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

	// Controllers
	usersC := controllers.NewUsers(services.User)
	galleriesC := controllers.NewGalleries(services.Gallery)
	staticC := controllers.NewStatic()

	requireUserMw := middleware.RequireUser{
		UserService: services.User,
	}
	newGallery := requireUserMw.Apply(galleriesC.New)
	createGallery := requireUserMw.ApplyFn(galleriesC.Create)

	r := mux.NewRouter()

	// User routes
	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.HandleFunc("/cookietest", usersC.CookieTest).Methods("GET")

	// Gallery routes
	r.Handle("/galleries/new", newGallery).Methods("GET")
	r.HandleFunc("/galleries", createGallery).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesC.Show).Methods("GET")

	// Misc
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/faq", staticC.Faq).Methods("GET")
	r.NotFoundHandler = staticC.NotFound

	http.ListenAndServe(":3000", r)
}
