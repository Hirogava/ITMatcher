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

	/*
		Общие
	*/
	r.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		handlers.Register(manager, w, r)
	}).Methods(http.MethodPost)

	r.HandleFunc("/api/auth", func(w http.ResponseWriter, r *http.Request) {
		handlers.Login(manager, w, r)
	}).Methods(http.MethodPost)

	r.Handle("/api/logout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.Logout(w, r)
	})).Methods(http.MethodGet)

	/*
		AI
	*/
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

	/*
		Внешние
	*/
	r.HandleFunc("/api/resume", func(w http.ResponseWriter, r *http.Request) {
		handlers.SendResume(w, r, manager)
	}).Methods(http.MethodPost)

	/*
		HR
	*/
	r.Handle("/api/hr/resume/{id}", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.ResumeById(w, r, manager)
		}))).Methods(http.MethodGet)

	r.Handle("/api/hr/resumes", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.ResumesList(w, r, manager)
		}))).Methods(http.MethodGet)

	r.Handle("/api/hr/vacancy", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.AddVacancy(w, r, manager)
		}))).Methods(http.MethodPost)

	r.Handle("/api/hr/finder", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.AddFinder(w, r, manager)
		}))).Methods(http.MethodPost)

	r.Handle("/api/hr/finder/{finder_id}/vacancy/{vacancy_id}", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.GetAnalizedResume(w, r, manager)
		}))).Methods(http.MethodGet)

	r.Handle("/api/hr/vacancy/{vacancy_id}", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.GetVacancy(w, r, manager)
		}))).Methods(http.MethodGet)

	/*
		Finder
	*/
	r.Handle("/api/finder/resumes", middleware.AuthRequired("users",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			list, _ := manager.GetUserResumes(*cookies.GetId(r)) // можно и в темплейт сунуть наверное не знаю
			json.NewEncoder(w).Encode(list)
		}))).Methods(http.MethodGet)

	r.Handle("/api/finder/resume/{resume_id}", middleware.AuthRequired("users",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.GetFinderResume(w, r, manager)
		}))).Methods(http.MethodGet)

	r.Handle("/api/finder/resume", middleware.AuthRequired("users",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.AddFinderResume(w, r, manager)
		}))).Methods(http.MethodPost)

}
