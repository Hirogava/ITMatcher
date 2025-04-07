package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gaspr/db"
	"gaspr/services/ai"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

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

func ResumesList(w http.ResponseWriter, r *http.Request, manager *db.Manager) {
	resumes, err := manager.GetAllResumes()
	if err != nil {
		http.Error(w, fmt.Sprintf("Что-то пошло не так: %v", err), http.StatusInternalServerError)
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
		}

		resumeFilePath := fmt.Sprintf("finder/%d/resume/resume.txt", resume.FinderId)
		resumeFileData, err := os.ReadFile(resumeFilePath)
		if err != nil {
			log.Printf("Ошибка при чтении файла резюме: %v", err)
			resumeMap["resume_content"] = ""
		} else {
			resumeMap["resume_content"] = string(resumeFileData)
		}

		jsonResumes = append(jsonResumes, resumeMap)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(jsonResumes); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при кодировании JSON: %v", err), http.StatusInternalServerError)
	}
}

func ResumeById(w http.ResponseWriter, r *http.Request, manager *db.Manager) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID резюме не указан", http.StatusBadRequest)
		return
	}

	resumeId, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Неверный формат ID", http.StatusBadRequest)
		return
	}

	resume, err := manager.GetResumeById(resumeId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении данных из БД: %v", err), http.StatusInternalServerError)
		return
	}

	resumeMap := map[string]interface{}{
		"id":           resume.Id,
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

	vacId, err = manager.GetVacancyIdByName(finderData.VacancyName)
	if err == sql.ErrNoRows {
		vacId, err = manager.CreateVacancy(finderData.VacancyName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка при записи в базу данных: %v", err), http.StatusInternalServerError)
			return
		}
		vacancyFile, _, err := r.FormFile("vacancy")
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка при получении файла: %v", err), http.StatusBadRequest)
			return
		}
		vacancyFileData, err := io.ReadAll(vacancyFile)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка при чтении файла: %v", err), http.StatusBadRequest)
			return
		}
		vacancyFile.Close()

		vacancyDir := fmt.Sprintf("vacancy/%d", vacId)
		if err := os.MkdirAll(vacancyDir, os.ModePerm); err != nil {
			http.Error(w, fmt.Sprintf("Ошибка создания директории: %v", err), http.StatusInternalServerError)
			return
		}

		vacancyFilePath := filepath.Join(vacancyDir, "vacancy.txt")
		if err := os.WriteFile(vacancyFilePath, vacancyFileData, os.ModePerm); err != nil {
			http.Error(w, fmt.Sprintf("Ошибка сохранения файла: %v", err), http.StatusInternalServerError)
			return
		}

		var vacancyData map[string][]string
		vacancyData, err = ai.Request(string(vacancyFileData))
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка AI модуля: %v", err), http.StatusInternalServerError)
			return
		}

		saveVacancySkills(vacId, vacancyData["hard_skills"], vacancyData["soft_skills"], manager)

	} else if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении данных из БД: %v", err), http.StatusInternalServerError)
		return
	}

	resumeId, err = manager.CreateResume(finderId, finderData.FirstName, finderData.LastName, finderData.Surname, finderData.Email, finderData.Phone, vacId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при записи в базу данных: %v", err), http.StatusInternalServerError)
		return
	}

	resumeFile, _, err := r.FormFile("resume")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении файла: %v", err), http.StatusBadRequest)
		return
	}
	resumeFileData, err := io.ReadAll(resumeFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при чтении файла: %v", err), http.StatusBadRequest)
		return
	}
	resumeFile.Close()

	resumeDir := fmt.Sprintf("finder/%d/resume", finderId)
	if err := os.MkdirAll(resumeDir, os.ModePerm); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка создания директории: %v", err), http.StatusInternalServerError)
		return
	}

	resumeFilePath := filepath.Join(resumeDir, "resume.txt")
	if err := os.WriteFile(resumeFilePath, resumeFileData, os.ModePerm); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка сохранения файла: %v", err), http.StatusInternalServerError)
		return
	}

	resumeData, err := ai.Request(string(resumeFileData))
	if err != nil {
		log.Println(err)
		return
	}

	saveResumeSkills(resumeData["hard_skills"], resumeData["soft_skills"], resumeId, manager)

	w.WriteHeader(http.StatusOK)
	return
}

func saveResumeSkills(hardSkills []string, softSkills []string, resumeId int, manager *db.Manager) {

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
			log.Printf("Ошибка при добавлении в resume_hard_skill: %v", err)
		}
	}

	for _, skill := range softSkills {
		var softSkillID int
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
			log.Printf("Ошибка при добавлении в resume_soft_skill: %v", err)
		}
	}
}

func saveVacancySkills(vacancyId int, hardSkills, softSkills []string, manager *db.Manager) {
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

		err = manager.CreateVacancyHardSkill(vacancyId, hardSkillID)
		if err != nil {
			log.Printf("Ошибка при добавлении в vacantion_hard_skill: %v", err)
		}
	}

	for _, skill := range softSkills {
		var softSkillID int
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

		err = manager.CreateVacancySoftSkill(vacancyId, softSkillID)
		if err != nil {
			log.Printf("Ошибка при добавлении в vacantion_soft_skill: %v", err)
		}
	}
}
