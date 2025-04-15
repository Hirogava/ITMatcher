package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gaspr/db"
	"gaspr/models"
	"gaspr/services/ai"
	"gaspr/services/analysis"
	"gaspr/services/cookies"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

func ResumesList(w http.ResponseWriter, r *http.Request, manager *db.Manager) {
	resumes, err := manager.GetAllResumesForHr(*cookies.GetId(r))
	if err != nil {
		log.Printf("Проблема с получением резюме: %v", err)
		http.Error(w, fmt.Sprintf("Проблема с получением резюме: %v", err), http.StatusInternalServerError)
		return
	}

	jsonResumes := make([]map[string]interface{}, 0)
	for _, resume := range resumes {
		resumeMap := map[string]interface{}{
			"id":           resume.Id,
			"finder_id":    resume.FinderId,
			"first_name":   resume.FirstName,
			"last_name":    resume.LastName,
			"surname":      resume.Surname,
			"email":        resume.Email,
			"phone_number": resume.PhoneNumber,
			"vacancy_id":   resume.VacancyId,
			"percent_match": resume.Percent,
		}

		vacancy, err := manager.GetVacancyByIdForHr(resume.VacancyId)
		if err != nil {
			log.Printf("Не удалось получить вакансию для резюме: %v", err)
			http.Error(w, "Не удалось получить вакансию для резюме", http.StatusInternalServerError)
			return
		}
		resumeMap["vacancy_name"] = vacancy.Name
		// излишне
		// resumeFilePath := fmt.Sprintf("finder/%d/resume/resume.txt", resume.FinderId)
		// resumeFileData, err := os.ReadFile(resumeFilePath)
		// if err != nil {
		// 	log.Printf("Ошибка при чтении файла резюме: %v", err)
		// 	resumeMap["resume_content"] = ""
		// } else {
		// 	resumeMap["resume_content"] = string(resumeFileData)
		// }

		jsonResumes = append(jsonResumes, resumeMap)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(jsonResumes); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при кодировании JSON: %v", err), http.StatusInternalServerError)
	}
}

func ResumeById(w http.ResponseWriter, r *http.Request, manager *db.Manager) {
	vars := mux.Vars(r)

	resumeId, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный формат ID", http.StatusBadRequest)
		return
	}

	resume, err := manager.GetResumeByIdForHr(resumeId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении данных из БД: %v", err), http.StatusInternalServerError)
		return
	}

	resumeMap := map[string]interface{}{
		"id":           resumeId,
		"finder_id":    resume.FinderId,
		"first_name":   resume.FirstName,
		"last_name":    resume.LastName,
		"surname":      resume.Surname,
		"email":        resume.Email,
		"phone_number": resume.PhoneNumber,
		"vacancy_id":   resume.VacancyId,
	}

	resumeFilePath := fmt.Sprintf("finder/%d/resume/resume.txt", resume.FinderId)
	resumeFileData, err := os.ReadFile(resumeFilePath)
	if err != nil {
		log.Printf("Ошибка при чтении файла резюме: %v", err)
		resumeMap["resume_content"] = ""
	} else {
		resumeMap["resume_content"] = string(resumeFileData)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resumeMap); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при кодировании JSON: %v", err), http.StatusInternalServerError)
	}
}

func SendResume(w http.ResponseWriter, r *http.Request, manager *db.Manager) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 мегабайт
		http.Error(w, fmt.Sprintf("Ошибка обработки multipart-запроса: %v", err), http.StatusBadRequest)
		return
	}

	jsonFile, _, err := r.FormFile("finder")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении JSON-файла: %v", err), http.StatusBadRequest)
		return
	}
	defer jsonFile.Close()

	// лишнее
	type FinderForm struct {
		HRUsername  string `json:"hr_username"`
		VacancyName string `json:"vacancy_name"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		Surname     string `json:"surname"`
		Email       string `json:"email"`
		Phone       string `json:"phone"`
		Portfolio   bool   `json:"portfolio"`
	}
	var finderData FinderForm
	if err := json.NewDecoder(jsonFile).Decode(&finderData); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка разбора JSON: %v", err), http.StatusBadRequest)
		return
	}

	var hrId, vacId, resumeId int
	hrId, err = manager.GetHRIdByUsername(finderData.HRUsername)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении данных из БД: %v", err), http.StatusInternalServerError)
		return
	}

	finderId, err := manager.CreateFinder(finderData.Portfolio, hrId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при записи в базу данных: %v", err), http.StatusInternalServerError)
		return
	}

	var vacSkills models.VacancySkills
	vacId, err = manager.GetVacancyIdByName(finderData.VacancyName, hrId)
	if err == sql.ErrNoRows {
		vacancyFile, _, err := r.FormFile("vacancy")
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка при получении vacancy: %v", err), http.StatusBadRequest)
			return
		}
		vacancyFileData, err := io.ReadAll(vacancyFile)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка при чтении vacancy: %v", err), http.StatusBadRequest)
			return
		}
		vacancyFile.Close()

		vacSkills, err = CreateVacancy(manager, finderData.VacancyName, vacancyFileData, hrId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка создания вакансии: %v", err), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении данных из БД: %v", err), http.StatusInternalServerError)
		return
	}

	var resSkills models.ResumeSkills
	resumeId, err = manager.CreateResumeForHr(finderId, finderData.FirstName, finderData.LastName, finderData.Surname, finderData.Email, finderData.Phone, vacId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при записи в базу данных: %v", err), http.StatusInternalServerError)
		return
	}

	resumeFile, _, err := r.FormFile("resume")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении файла: %v", err), http.StatusBadRequest)
		return
	}
	defer resumeFile.Close()

	resumeFileData, err := SaveResume(resumeFile, finderId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при сохранении файла: %v", err), http.StatusBadRequest)
		return
	}

	resumeData, err := ai.Request(string(resumeFileData))
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка AI модуля: %v", err), http.StatusInternalServerError)
		return
	}

	resSkills, err = saveResumeSkills(resumeData["hard_skills"], resumeData["soft_skills"], resumeId, manager, "hr")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при записи в базу данных: %v", err), http.StatusInternalServerError)
		return
	}

	analyzedSkills, err := analysis.AnalyseResumeSkills(resSkills, vacSkills)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при анализе резюме: %v", err), http.StatusInternalServerError)
		return
	}

	err = manager.SaveAnalyzedData("hr", resumeId, vacId, analyzedSkills)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при записи в базу данных: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}

// Ужасная функция на 130 строчек, надо с ней что-то сделать
func saveResumeSkills(hardSkills []string, softSkills []string, resumeId int, manager *db.Manager, role string) (models.ResumeSkills, error) {
	var result models.ResumeSkills

	switch role {
	case "hr":
		for _, skill := range hardSkills {
			hardSkillID, err := manager.GetHardSkillByName(skill)
			if err != nil {
				if err == sql.ErrNoRows {
					hardSkillID, err = manager.CreateHardSkill(skill)
					if err != nil {
						log.Printf("Ошибка при добавлении hard_skill: %v", err)
						continue
					}
				} else {
					log.Printf("Ошибка при проверке hard_skill: %v", err)
					continue
				}
			}
			err = manager.CreateResumeHardSkill(resumeId, hardSkillID)
			if err != nil {
				log.Printf("Ошибка при добавлении hard_skill: %v", err)
				continue
			}
			result.HardSkills = append(result.HardSkills, models.ResumeHardSkill{
				Id:        hardSkillID,
				SkillName: skill,
			})
		}

		for _, skill := range softSkills {
			softSkillID, err := manager.GetSoftSkillByName(skill)
			if err != nil {
				if err == sql.ErrNoRows {
					softSkillID, err = manager.CreateSoftSkill(skill)
					if err != nil {
						log.Printf("Ошибка при добавлении soft_skill: %v", err)
						continue
					}
				} else {
					log.Printf("Ошибка при проверке soft_skill: %v", err)
					continue
				}
			}
			err = manager.CreateResumeSoftSkill(resumeId, softSkillID)
			if err != nil {
				log.Printf("Ошибка при добавлении soft_skill: %v", err)
				continue
			}
			result.SoftSkills = append(result.SoftSkills, models.ResumeSoftSkill{
				Id:        softSkillID,
				SkillName: skill,
			})
		}
		return result, nil

	case "users":
		for _, skill := range hardSkills {
			hardSkillID, err := manager.GetHardSkillByName(skill)
			if err != nil {
				if err == sql.ErrNoRows {
					hardSkillID, err = manager.CreateHardSkill(skill)
					if err != nil {
						log.Printf("Ошибка при добавлении hard_skill: %v", err)
						continue
					}
				} else {
					log.Printf("Ошибка при проверке hard_skill: %v", err)
					continue
				}
			}
			err = manager.CreateUserResumeHardSkill(resumeId, hardSkillID)
			if err != nil {
				log.Printf("Ошибка при добавлении hard_skill: %v", err)
				continue
			}
			result.HardSkills = append(result.HardSkills, models.ResumeHardSkill{
				Id:        hardSkillID,
				SkillName: skill,
			})
		}

		for _, skill := range softSkills {
			softSkillID, err := manager.GetSoftSkillByName(skill)
			if err != nil {
				if err == sql.ErrNoRows {
					softSkillID, err = manager.CreateSoftSkill(skill)
					if err != nil {
						log.Printf("Ошибка при добавлении soft_skill: %v", err)
						continue
					}
				} else {
					log.Printf("Ошибка при проверке soft_skill: %v", err)
					continue
				}
			}
			err = manager.CreateUserResumeSoftSkill(resumeId, softSkillID)
			if err != nil {
				log.Printf("Ошибка при добавлении soft_skill: %v", err)
				continue
			}
			result.SoftSkills = append(result.SoftSkills, models.ResumeSoftSkill{
				Id:        softSkillID,
				SkillName: skill,
			})
		}
		return result, nil

	default:
		return result, fmt.Errorf("неверно указана роль: %s", role)
	}
}
