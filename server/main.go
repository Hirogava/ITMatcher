package main

import (
	"encoding/json"
	"fmt"
	"gaspr/db"
	"gaspr/handlers"
	middleware "gaspr/handlers/middlewares"
	"gaspr/services"
	"gaspr/services/ai"
	"gaspr/services/cookies"
	"html/template"
	"io"
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

	cookies.Init()
	r := mux.NewRouter()

	/*
		Static files
	*/
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./static/html/landing.html"))
		tmpl.Execute(w, nil)
	})

	/*
		API
	*/
	r.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		handlers.Register(manager, w, r)
	}).Methods(http.MethodPost)

	r.HandleFunc("/api/auth", func(w http.ResponseWriter, r *http.Request) {
		handlers.Login(manager, w, r)
	}).Methods(http.MethodPost)

	r.Handle("/api/logout",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.Logout(w, r)
		})).Methods(http.MethodPost)

	// пока в main'е
	r.HandleFunc("/api/nlp", func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка при получении файла: %v", err), http.StatusBadRequest)
			return
		}
		defer file.Close()

		fileData, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка при чтении файла: %v", err), http.StatusBadRequest)
			return
		}

		result, err := ai.Request(string(fileData))
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка AI модуля: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}).Methods(http.MethodPost)

	r.HandleFunc("/api/send_resume", func(w http.ResponseWriter, r *http.Request) {
		handlers.SendResume(w, r, manager)
	}).Methods(http.MethodPost)

	/*
		API - HR
	*/
	r.Handle("/api/hr/resume/{id}", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.ResumeById(w, r, manager)
		}))).Methods(http.MethodGet)

	r.Handle("/api/hr/resumes/{hr_id}", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.ResumesList(w, r, manager)
		}))).Methods(http.MethodGet)

	/*
		API - Finder
	*/

	r.Handle("/api/finder/resume/{user_id}", middleware.AuthRequired("finder",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.SaveUserResume(w, r, manager)
		}))).Methods(http.MethodPost)

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
