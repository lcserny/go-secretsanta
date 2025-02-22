package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lcserny/go-secretsanta"
)

func main() {
	initHttpServer()
}

func initHttpServer() {
	router := mux.NewRouter()

	// add more controllers as needed
	gosecretsanta.InitMatchesController(router)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
