package handlers

import (
	"database/sql"
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
