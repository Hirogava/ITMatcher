package main

import (
	"fmt"
	"gaspr/db"
	"gaspr/routes"
	"gaspr/services"
	"gaspr/services/cookies"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	/*
		Initialization
	*/
	services.LoadEnvFile(".env")
	manager := db.NewDBManager("postgres", os.Getenv("DB_CONNECTION_STRING"))
	db.Migrate(manager)

	log.Println("База данных успешно инициализирована и мигрирована.")
	defer manager.Close()

	cookies.Init(os.Getenv("SESSION_KEY"))
	r := mux.NewRouter()

	routes.Init(r, manager)

	/*
		Server initialization
	*/
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	serverPort := os.Getenv("SERVER_PORT")
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", serverPort),
		Handler: r,
	}

	log.Println("Сервер запущен на порту " + serverPort)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
