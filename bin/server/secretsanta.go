package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/lcserny/go-secretsanta"
)

func main() {
    logFile := initLogging()
    defer logFile.Close()

	initHttpServer()
}

func initLogging() *os.File {
    logFile, err := os.OpenFile("secretsanta.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("Failed to open log file: %v", err)
    }
    log.SetOutput(logFile)
    return logFile
}

func initHttpServer() {
	router := mux.NewRouter()

	// add more controllers as needed
	gosecretsanta.InitMatchesController(router)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
