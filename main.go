package main

import (
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
	us, err := models.NewUserService(psqlInfo)
	if err != nil {
		panic(err)
	}
	defer func(us models.UserService) {
		err := us.Close()
		if err != nil {

		}
	}(us)
	err = us.AutoMigrate()
	if err != nil {
		return
	}

	// Controllers
	usersC := controllers.NewUsers(us)
	staticC := controllers.NewStatic()

	// Handlers
	r := mux.NewRouter()
	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/faq", staticC.Faq).Methods("GET")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.HandleFunc("/cookietest", usersC.CookieTest).Methods("GET")

	// 404
	r.NotFoundHandler = staticC.NotFound

	http.ListenAndServe(":3000", r)
}
