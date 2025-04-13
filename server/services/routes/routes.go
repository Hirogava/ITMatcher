package routes

import (
	"fmt"
	"gaspr/db"
	middleware "gaspr/handlers/middlewares"
	"gaspr/services/cookies"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func Init(r *mux.Router, manager *db.Manager) {
	LandingRoute(r)
	LoginRoute(r)
	RegisterRoute(r)
	FindersRoute(r, manager)
	HrAccRoute(r, manager)
	VacanciesRoute(r, manager)
	ApiAuthRoute(r, manager)
	ApiLogoutRoute(r)
	ApiRegisterRoute(r, manager)
	ApiExternalHrServiceRoute(r, manager)
	ApiFinderResumeByIDRoute(r, manager)
	ApiGetAllResumesRoute(r, manager)
	ApiNLPRoute(r)
	ApiSaveUserSkillsRoute(r, manager)
}

func join(elements []db.VacancyHardSkill) string {
	var skillNames []string
	for _, element := range elements {
		skillNames = append(skillNames, element.SkillName)
	}
	return strings.Join(skillNames, ", ")
}

func GetTemplate(name string) *template.Template {
	funcMap := template.FuncMap{
		"join": join,
	}
	tmpl := template.Must(template.New("example").Funcs(funcMap).ParseFiles(
		"./static/html/templates/header.html",
		"./static/html/templates/footer.html",
		"./static/html/templates/platform_header.html",
		fmt.Sprintf("./static/html/%s.html", name),
	))
	return tmpl
}

func LandingRoute(r *mux.Router) {
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := GetTemplate("landing")
		data := map[string]interface{}{
			"pageTitle":  "Главная",
			"hr_account": cookies.GetUsernameCookie(r),
		}
		tmpl.ExecuteTemplate(w, "base", data)
	})
}

func LoginRoute(r *mux.Router) {
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		tmpl := GetTemplate("login")
		data := map[string]interface{}{
			"pageTitle": "Вход",
		}
		tmpl.ExecuteTemplate(w, "base", data)
	})
}

func RegisterRoute(r *mux.Router) {
	r.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		tmpl := GetTemplate("registration")
		data := map[string]interface{}{
			"pageTitle": "Регистрация",
		}
		tmpl.ExecuteTemplate(w, "base", data)
	})
}

func FindersRoute(r *mux.Router, manager *db.Manager) {
	r.Handle("/finders", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tmpl := GetTemplate("finders")

			hrAccount := cookies.GetHrAccountCookie(r, manager)
			if hrAccount == nil {
				http.Error(w, "Не удалось получить данные аккаунта", http.StatusUnauthorized)
				return
			}

			findResumes, err := manager.GetAllResumesForHr(hrAccount.ID)
			if err != nil {
				fmt.Println("Ошибка при получении резюме:", err)
				http.Error(w, "Не удалось получить резюме", http.StatusUnauthorized)
				return
			}

			findersData := make([]map[string]interface{}, len(findResumes))
			for i, resume := range findResumes {
				vacancy, err := manager.GetVacancyByIdForHr(resume.VacancyId)
				if err != nil {
					http.Error(w, "Не удалось получить вакансию для резюме", http.StatusInternalServerError)
					return
				}

				// Добавляем словарь в массив
				findersData[i] = map[string]interface{}{
					"resume":  resume,
					"vacancy": vacancy,
				}
			}

			data := map[string]interface{}{
				"pageTitle":    "Вакансии",
				"finders":      findersData,
				"current_page": "finders",
			}

			tmpl.ExecuteTemplate(w, "base", data)
		})))
}

func HrAccRoute(r *mux.Router, manager *db.Manager) {
	r.Handle("/hracc", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tmpl := GetTemplate("hr_account")
			hrAccount := cookies.GetHrAccountCookie(r, manager)
			if hrAccount == nil {
				http.Error(w, "Не удалось получить данные аккаунта", http.StatusUnauthorized)
				return
			}

			data := map[string]interface{}{
				"pageTitle": "Аккаунт",
				"hr_account": map[string]string{
					"username": hrAccount.Username,
					"email":    hrAccount.Email,
				},
				"current_page": "hr_account",
			}

			tmpl.ExecuteTemplate(w, "base", data)
		})))
}

func VacanciesRoute(r *mux.Router, manager *db.Manager) {
	r.Handle("/vacancies", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tmpl := GetTemplate("vacancies")

			hrAccount := cookies.GetHrAccountCookie(r, manager)
			if hrAccount == nil {
				http.Error(w, "Не удалось получить данные аккаунта", http.StatusUnauthorized)
				return
			}

			vacancies, err := manager.GetAllHrVacancies(hrAccount.ID)
			if err != nil {
				fmt.Println("Ошибка при получении вакансий:", err)
				http.Error(w, "Не удалось получить вакансии", http.StatusInternalServerError)
				return
			}
			data := map[string]interface{}{
				"pageTitle":    "Вакансии",
				"vacancies":    vacancies,
				"current_page": "vacancies",
			}

			tmpl.ExecuteTemplate(w, "base", data)
		})))
}
