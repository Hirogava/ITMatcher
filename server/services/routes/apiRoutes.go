package routes

import (
	"encoding/json"
	"fmt"
	"gaspr/db"
	"gaspr/handlers"
	"gaspr/services/ai"
	"gaspr/handlers/middlewares"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

/*
Общие роуты
*/

func ApiRegisterRoute(r *mux.Router, manager *db.Manager) {
	r.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		handlers.Register(manager, w, r)
	}).Methods(http.MethodPost)
}

func ApiAuthRoute(r *mux.Router, manager *db.Manager) {
	r.HandleFunc("/api/auth", func(w http.ResponseWriter, r *http.Request) {
		handlers.Login(manager, w, r)
	}).Methods(http.MethodPost)
}

func ApiLogoutRoute(r *mux.Router) {
	r.Handle("/api/logout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.Logout(w, r)
	})).Methods(http.MethodGet)
}

/*
NLP роуты
*/

func ApiNLPRoute(r *mux.Router) {
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
}

/*
Внешние роуты
*/

func ApiExternalHrServiceRoute(r *mux.Router, manager *db.Manager) {
	r.HandleFunc("/api/send_resume", func(w http.ResponseWriter, r *http.Request) {
		handlers.SendResume(w, r, manager)
	}).Methods(http.MethodPost)
}

/*
HR роуты
*/

func ApiFinderResumeByIDRoute(r *mux.Router, manager *db.Manager) {
	r.Handle("/api/hr/resume/{id}", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.ResumeById(w, r, manager)
		}))).Methods(http.MethodGet)
}

func ApiGetAllResumesRoute(r *mux.Router, manager *db.Manager) {
	r.Handle("/api/hr/resumes/{hr_id}", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.ResumesList(w, r, manager)
		}))).Methods(http.MethodGet)
}

func ApiAddFinder(r *mux.Router, manager *db.Manager) {
	r.Handle("/api/hr/add_finder", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.AddFinder(w, r, manager)
		}))).Methods(http.MethodGet)
}

/*
User роуты
*/

func ApiSaveUserSkillsRoute(r *mux.Router, manager *db.Manager) {
	r.Handle("/api/finder/resume/{user_id}", middleware.AuthRequired("finder",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.SaveUserResume(w, r, manager)
		}))).Methods(http.MethodPost)
}