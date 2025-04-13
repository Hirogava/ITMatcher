package routes

import (
	"encoding/json"
	"fmt"
	"gaspr/db"
	"gaspr/handlers"
	"gaspr/handlers/middlewares"
	"gaspr/services/ai"
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
	r.HandleFunc("/api/send_resume", func(w http.ResponseWriter, r *http.Request) {
		handlers.SendResume(w, r, manager)
	}).Methods(http.MethodPost)

	/*
		HR
	*/
	r.Handle("/api/hr/resume/{id}", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.ResumeById(w, r, manager)
		}))).Methods(http.MethodGet)

	// Прошлый вариант позволял получать одному hr'у получать доступ к чужим резюме
	r.Handle("/api/hr/resumes", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.ResumesList(w, r, manager)
		}))).Methods(http.MethodGet)

	r.Handle("/api/hr/add_vacancy", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.AddVacancy(w, r, manager)
		}))).Methods(http.MethodPost)

	r.Handle("/api/hr/add_finder", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.AddFinder(w, r, manager)
		}))).Methods(http.MethodPost)

	r.Handle("/api/hr/finder/{finder_id}/{vacancy_id}", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.GetAnalizedResume(w, r, manager)
		}))).Methods(http.MethodGet)

	/*
		Finder
	*/
	r.Handle("/api/finder/add_resume", middleware.AuthRequired("users",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.AddFinderResume(w, r, manager)
		}))).Methods(http.MethodPost)

}
