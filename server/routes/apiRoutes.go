package routes

import (
	"encoding/json"
	"fmt"
	"gaspr/db"
	"gaspr/handlers"
	middleware "gaspr/handlers/middlewares"
	"gaspr/services/ai"
	"gaspr/services/cookies"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func ApiRoutes(r *mux.Router, manager *db.Manager) {
	api := r.PathPrefix("/api").Subrouter()
	apiHr := api.PathPrefix("/hr").Subrouter()
	apiHr.Use(middleware.AuthRequired("hr"))
	apiFinder := api.PathPrefix("/finder").Subrouter()
	apiFinder.Use(middleware.AuthRequired("users"))

	/*
		Общие
	*/

	api.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		handlers.Register(manager, w, r)
	}).Methods(http.MethodPost)

	api.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		handlers.Login(manager, w, r)
	}).Methods(http.MethodPost)

	api.Handle("/logout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.Logout(w, r)
	})).Methods(http.MethodGet)

	/*
		AI
	*/

	api.HandleFunc("/nlp", func(w http.ResponseWriter, r *http.Request) {
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

	/*
		Внешние
	*/

	api.HandleFunc("/resume", func(w http.ResponseWriter, r *http.Request) {
		handlers.SendResume(w, r, manager)
	}).Methods(http.MethodPost)

	/*
		HR
	*/

	apiHr.HandleFunc("/resume/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.ResumeById(w, r, manager)
	}).Methods(http.MethodGet)

	apiHr.HandleFunc("/resumes", func(w http.ResponseWriter, r *http.Request) {
		handlers.ResumesList(w, r, manager)
	}).Methods(http.MethodGet)

	apiHr.HandleFunc("/vacancy", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddVacancy(w, r, manager)
	}).Methods(http.MethodPost)

	apiHr.HandleFunc("/finder", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddFinder(w, r, manager)
	}).Methods(http.MethodPost)

	apiHr.HandleFunc("/finder/{finder_id}/vacancy/{vacancy_id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetAnalizedResume(w, r, manager)
	}).Methods(http.MethodGet)

	apiHr.HandleFunc("/vacancy/{vacancy_id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetVacancy(w, r, manager)
	}).Methods(http.MethodGet)

	/*
		Finder
	*/
	apiFinder.HandleFunc("/resumes", func(w http.ResponseWriter, r *http.Request) {
		list, _ := manager.GetUserResumes(*cookies.GetId(r)) // можно и в темплейт сунуть наверное не знаю
		json.NewEncoder(w).Encode(list)
	}).Methods(http.MethodGet)

	apiFinder.HandleFunc("/resume/{resume_id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetFinderResume(w, r, manager)
	}).Methods(http.MethodGet)

	apiFinder.HandleFunc("/resume", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddFinderResume(w, r, manager)
	}).Methods(http.MethodPost)
}
