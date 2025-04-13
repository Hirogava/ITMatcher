package routes

import (
	"fmt"
	"gaspr/db"
	middleware "gaspr/handlers/middlewares"
	"gaspr/models"
	"gaspr/services/cookies"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func Init(r *mux.Router, manager *db.Manager) {
	StaticRoutes(r, manager)
	ApiRoutes(r, manager)
}

func join(elements []models.VacancyHardSkill) string {
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

func StaticRoutes(r *mux.Router, manager *db.Manager) {

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := GetTemplate("landing")
		data := map[string]interface{}{
			"pageTitle":  "Главная",
			"hr_account": cookies.GetUsername(r),
		}
		tmpl.ExecuteTemplate(w, "base", data)
	})

	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		tmpl := GetTemplate("login")
		data := map[string]interface{}{
			"pageTitle": "Вход",
		}
		tmpl.ExecuteTemplate(w, "base", data)
	})

	r.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		tmpl := GetTemplate("registration")
		data := map[string]interface{}{
			"pageTitle": "Регистрация",
		}
		tmpl.ExecuteTemplate(w, "base", data)
	})

	r.Handle("/finders", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tmpl := GetTemplate("finders")

			findResumes, err := manager.GetAllResumesForHr(*cookies.GetId(r))
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

	// хряк WW
	r.Handle("/hracc", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tmpl := GetTemplate("hr_account")
			store := cookies.NewCookieManager(r)

			data := map[string]interface{}{
				"pageTitle": "Аккаунт",
				"hr_account": map[string]string{
					"username": store.Session.Values["username"].(string),
					"email":    store.Session.Values["email"].(string),
				},
				"current_page": "hr_account",
			}

			tmpl.ExecuteTemplate(w, "base", data)
		})))

	r.Handle("/vacancies", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tmpl := GetTemplate("vacancies")

			vacancies, err := manager.GetAllHrVacancies(*cookies.GetId(r))
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
