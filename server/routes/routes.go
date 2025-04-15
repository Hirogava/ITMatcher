package routes

import (
	"fmt"
	"gaspr/db"
	middleware "gaspr/handlers/middlewares"
	"gaspr/models"
	"gaspr/services/cookies"
	"html/template"
	"log"
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
			"pageTitle": "Главная",
			"account":   cookies.GetAccount(r),
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

	r.Handle("/hr/finders", middleware.AuthRequired("hr",
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

				findersData[i] = map[string]interface{}{
					"resume":  resume,
					"vacancy": vacancy,
				}
			}
			allVacancies, err := manager.GetAllVacancies()
			data := map[string]interface{}{
				"pageTitle":    "Вакансии",
				"finders":      findersData,
				"vacancies":    allVacancies,
				"account":      cookies.GetAccount(r),
				"current_page": "finders",
			}

			tmpl.ExecuteTemplate(w, "base", data)
		})))

	// хряк WW
	r.Handle("/hr/acc", middleware.AuthRequired("hr",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tmpl := GetTemplate("hr_account")

			data := map[string]interface{}{
				"pageTitle":    "Аккаунт",
				"account":      cookies.GetAccount(r),
				"current_page": "hr_account",
			}

			tmpl.ExecuteTemplate(w, "base", data)
		})))

	r.Handle("/hr/vacancies", middleware.AuthRequired("hr",
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
				"account":      cookies.GetAccount(r),
			}

			tmpl.ExecuteTemplate(w, "base", data)
		})))

	/*
		User сторона
	*/

	r.Handle("/user/acc", middleware.AuthRequired("users",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tmpl := GetTemplate("finder_account")
			data := map[string]interface{}{
				"pageTitle":    "Профиль",
				"current_page": "user_account",
				"account":      cookies.GetAccount(r),
			}

			log.Println(data)

			tmpl.ExecuteTemplate(w, "base", data)

		})))

	r.Handle("/user/resumes", middleware.AuthRequired("users",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tmpl := GetTemplate("finder_resumes")
			data := map[string]interface{}{
				"pageTitle":    "Все резюме",
				"current_page": "user_resumes",
				"account":      cookies.GetAccount(r),
			}

			log.Println(data)

			tmpl.ExecuteTemplate(w, "base", data)
		})))
}
