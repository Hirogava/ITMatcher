package handlers

import (
	"encoding/json"
	"fmt"
	"gaspr/db"
	"gaspr/models"
	"gaspr/services/ai"
	"gaspr/services/cookies"
	"gaspr/services/resumeanalysis"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"log"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
)

func AddFinder(w http.ResponseWriter, r *http.Request, manager *db.Manager) {
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	surName := r.FormValue("sur_name")
	phone := r.FormValue("phone_number")
	email := r.FormValue("email")
	vacancy := r.FormValue("vacancy")
	hr := cookies.GetHrAccountCookie(r, manager)

	var vacSkills models.VacancySkills
	var resSkills models.ResumeSkills

	r.ParseMultipartForm(10 << 20)
	resumeFile, _, err := r.FormFile("resume_file")
	if err != nil {
		w.Write([]byte("Error: " + err.Error()))
		return
	}
	defer resumeFile.Close()

	finderId, err := manager.CreateFinder(false, hr.ID)
	if err != nil {
		log.Printf("Ошибка создание пользователя: %v", err)
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	vacId, err := manager.GetVacancyIdByName(vacancy, hr.ID)
	if err != nil {
		log.Printf("Ошибка получение id вакансии: %v", err)
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	resumeId, err := manager.CreateResumeForHr(finderId, firstName, lastName, surName, phone, email, vacId)
	if err != nil {
		log.Printf("Ошибка создание резюме: %v", err)
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	resumeFileData, err := SaveResume(resumeFile, finderId)
	if err != nil {
		log.Printf("Ошибка сохранения резюме: %v", err)
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	resumeData, err := ai.Request(string(resumeFileData))
	if err != nil {
		log.Printf("Ошибка запроса к AI: %v", err)
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	resSkills, err = saveResumeSkills(resumeData["hard_skills"], resumeData["soft_skills"], resumeId, manager, "hr")
	if err != nil {
		log.Printf("Ошибка сохранения навыков резюме: %v", err)
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	vacSkills.HardSkills, vacSkills.SoftSkills, err = manager.GetVacancySkills(vacId)
	if err != nil {
		log.Printf("Ошибка получение навыков вакансии: %v", err)
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	analyzedSkills, err := resumeanalysis.AnalizResumeSkills(resSkills, vacSkills)
	if err != nil {
		log.Printf("Ошибка анализа навыков: %v", err)
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	err = manager.SaveAnalyzedDataForHr(resumeId, vacId, analyzedSkills)
	if err != nil {
		log.Printf("Ошибка сохранения анализа навыков: %v", err)
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("Resume uploaded successfully"))
}

func SaveResume(resumeFile multipart.File, finderId int) ([]byte, error) {
	resumeFileData, err := io.ReadAll(resumeFile)
	if err != nil {
		return nil, err
	}
	resumeFile.Close()

	resumeDir := fmt.Sprintf("finder/%d/resume", finderId)
	if err := os.MkdirAll(resumeDir, os.ModePerm); err != nil {
		return nil, err
	}

	resumeFilePath := filepath.Join(resumeDir, "resume.txt")
	if err := os.WriteFile(resumeFilePath, resumeFileData, os.ModePerm); err != nil {
		return nil, err
	}

	return resumeFileData, nil
}

func GetAnalizedResume(w http.ResponseWriter, r *http.Request, manager *db.Manager) {
	vars := mux.Vars(r)
	finId := vars["finder_id"]
	vacId := vars["vacancy_id"]
	intFinId, err := strconv.Atoi(finId)
	if err != nil {
		log.Printf("Неправильный ID: %v", err)
		w.Write([]byte("Error: " + err.Error()))
		return
	}
	intVacId, err := strconv.Atoi(vacId)
	if err != nil {
		log.Printf("Неправильный ID: %v", err)
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	res, err := manager.GetAnalizedData(intFinId, intVacId)
	if err != nil {
		log.Printf("Ошибка получения анализа резюме: %v", err)
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Ошибка вывода резюме: %v", err)
		w.Write([]byte("Error: " + err.Error()))
		return
	}
}