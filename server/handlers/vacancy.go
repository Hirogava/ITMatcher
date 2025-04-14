package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gaspr/db"
	"gaspr/models"
	"gaspr/services/ai"
	"gaspr/services/cookies"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
)

func AddVacancy(w http.ResponseWriter, r *http.Request, manager *db.Manager) {
	vacancyName := r.FormValue("vacancy_name")
	if vacancyName == "" {
		http.Error(w, "Ошибка при получении vacancy_name", http.StatusBadRequest)
		return
	}
	vacancyFile, _, err := r.FormFile("vacancy_file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении vacancy_file: %v", err), http.StatusBadRequest)
		return
	}
	vacancyFileData, err := io.ReadAll(vacancyFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при чтении vacancy: %v", err), http.StatusBadRequest)
		return
	}
	vacancyFile.Close()

	hrId := cookies.GetId(r)
	_, err = CreateVacancy(manager, vacancyName, vacancyFileData, *hrId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка создания вакансии: %v", err), http.StatusInternalServerError)
		return
	}
}

func CreateVacancy(manager *db.Manager, vacancyName string, vacancyFileData []byte, hrId int) (models.VacancySkills, error) {
	vacId, err := manager.CreateVacancy(vacancyName, hrId)
	if err != nil {
		return models.VacancySkills{}, err
	}

	vacancyDir := fmt.Sprintf("vacancy/%d", vacId)
	if err := os.MkdirAll(vacancyDir, os.ModePerm); err != nil {
		return models.VacancySkills{}, err
	}

	vacancyFilePath := filepath.Join(vacancyDir, "vacancy.txt")
	if err := os.WriteFile(vacancyFilePath, vacancyFileData, os.ModePerm); err != nil {
		return models.VacancySkills{}, err
	}

	vacancyData, err := ai.Request(string(vacancyFileData))
	if err != nil {
		return models.VacancySkills{}, err
	}

	vacSkills, err := saveVacancySkills(vacId, vacancyData["hard_skills"], vacancyData["soft_skills"], manager)
	if err != nil {
		return models.VacancySkills{}, err
	}
	return vacSkills, nil
}

func GetVacancy(w http.ResponseWriter, r *http.Request, manager *db.Manager) {
	vars := mux.Vars(r)
	vacancyId, err := strconv.Atoi(vars["vacancy_id"])
	if err != nil {
		log.Printf("Неверно указан Id: %v", err)
		http.Error(w, fmt.Sprintf("Неверно указан Id: %v", err), http.StatusBadRequest)
		return
	}

	vacancy, err := manager.GetVacancyByIdForHr(vacancyId)
	if err != nil {
		log.Printf("Ошибка получения вакансии: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка получения вакансии: %v", err), http.StatusInternalServerError)
		return
	}

	var vacancySoft []string
	var vacancyHard []string
	for _, vacancySkill := range vacancy.HardSkills {
		vacancyHard = append(vacancyHard, vacancySkill.SkillName)
	}
	for _, vacancySkill := range vacancy.SoftSkills {
		vacancySoft = append(vacancySoft, vacancySkill.SkillName)
	}

	vacancyData, err := os.Open(filepath.Join("vacancy", strconv.Itoa(vacancyId), "vacancy.txt"))
	if err != nil {
		log.Printf("Ошибка открытия файла вакансии: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка открытия файла вакансии: %v", err), http.StatusInternalServerError)
		return
	}
	defer vacancyData.Close()

	vacancyDataBytes, err := io.ReadAll(vacancyData)
	if err != nil {
		log.Printf("Ошибка чтения файла вакансии: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка чтения файла вакансии: %v", err), http.StatusInternalServerError)
		return
	}

	vacancyJson := map[string]interface{}{
		"name": vacancy.Name,
		"vacancy_text": string(vacancyDataBytes),
		"hard_skills":  vacancyHard,
		"soft_skills":  vacancySoft,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(vacancyJson); err != nil {
		log.Printf("Ошибка при записи ответа: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка при записи ответа: %v", err), http.StatusInternalServerError)
		return
	}
}

func saveVacancySkills(vacancyId int, hardSkills, softSkills []string, manager *db.Manager) (models.VacancySkills, error) {
	var result models.VacancySkills

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

		_ = manager.CreateVacancyHardSkill(vacancyId, hardSkillID)
		result.HardSkills = append(result.HardSkills, models.VacancyHardSkill{
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

		_ = manager.CreateVacancySoftSkill(vacancyId, softSkillID)
		result.SoftSkills = append(result.SoftSkills, models.VacancySoftSkill{
			Id:        softSkillID,
			SkillName: skill,
		})
	}

	return result, nil
}
