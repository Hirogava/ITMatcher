package routes

import (
	"gaspr/db"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

func Init(r *mux.Router, manager *db.Manager) {
	LandingRoute(r)
	LoginRoute(r)
	RegisterRoute(r)
	FindersRoute(r)
	HrAccRoute(r)
	VacanciesRoute(r)
	ApiAuthRoute(r, manager)
	ApiLogoutRoute(r)
	ApiRegisterRoute(r, manager)
	ApiExternalHrServiceRoute(r, manager)
	ApiFinderResumeByIDRoute(r, manager)
	ApiGetAllResumesRoute(r, manager)
	ApiNLPRoute(r)
	ApiSaveUserSkillsRoute(r, manager)
}

func LandingRoute(r *mux.Router) {
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./static/html/landing.html"))
		tmpl.Execute(w, nil)
	})
}

func LoginRoute(r *mux.Router) {
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./static/html/login.html"))
		tmpl.Execute(w, nil)
	})
}

func RegisterRoute(r *mux.Router) {
	r.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./static/html/registration.html"))
		tmpl.Execute(w, nil)
	})
}

func FindersRoute(r *mux.Router) {
	r.HandleFunc("/finders", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./static/html/finders.html"))
		tmpl.Execute(w, nil)
	})
}

func HrAccRoute(r *mux.Router) {
	r.HandleFunc("/hracc", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./static/html/hr_account.html"))
		tmpl.Execute(w, nil)
	})
}

func VacanciesRoute(r *mux.Router) {
	r.HandleFunc("/vacancies", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./static/html/vacancies.html"))
		tmpl.Execute(w, nil)
	})
}
