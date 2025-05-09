package handlers

import (
	"encoding/json"
	"fmt"
	"gaspr/db"
	"gaspr/models"
	"gaspr/services/ai"
	"gaspr/services/analysis"
	"gaspr/services/cookies"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
)

func AddFinder(w http.ResponseWriter, r *http.Request, manager *db.Manager) {
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	surName := r.FormValue("surname")
	phone := r.FormValue("phone_number")
	email := r.FormValue("email")
	vacId, err := strconv.Atoi(r.FormValue("vacancy"))
	if err != nil {
		log.Printf("Ошибка получения ID вакансии: %v", err)
		http.Error(w, "Ошибка получения ID вакансии", http.StatusInternalServerError)
		return
	}
	hrId := cookies.GetId(r)

	var vacSkills models.VacancySkills
	var resSkills models.ResumeSkills

	r.ParseMultipartForm(10 << 20)
	resumeFile, _, err := r.FormFile("resume_file")
	if err != nil {
		log.Printf("Ошибка загрузки файла: %v", err)
		http.Error(w, "Ошибка загрузки файла", http.StatusInternalServerError)
		return
	}
	defer resumeFile.Close()

	finderId, err := manager.CreateFinder(false, *hrId)
	if err != nil {
		log.Printf("Ошибка создание пользователя: %v", err)
		http.Error(w, "Ошибка сохранения пользователя", http.StatusInternalServerError)
		return
	}

	resumeId, err := manager.CreateResumeForHr(finderId, firstName, lastName, surName, email, phone, vacId)
	if err != nil {
		log.Printf("Ошибка создание резюме: %v", err)
		http.Error(w, "Ошибка создания резюме", http.StatusInternalServerError)
		return
	}

	resumeFileData, err := SaveResume(resumeFile, finderId)
	if err != nil {
		log.Printf("Ошибка сохранения резюме: %v", err)
		http.Error(w, "Ошибка сохранения резюме", http.StatusInternalServerError)
		return
	}

	resumeData, err := ai.Request(string(resumeFileData))
	if err != nil {
		log.Printf("Ошибка запроса к AI: %v", err)
		http.Error(w, "Ошибка запроса к AI", http.StatusInternalServerError)
		return
	}

	resSkills, err = saveResumeSkills(resumeData["hard_skills"], resumeData["soft_skills"], resumeId, manager, "hr")
	if err != nil {
		log.Printf("Ошибка сохранения навыков резюме: %v", err)
		http.Error(w, "Ошибка сохранения навыков резюме", http.StatusInternalServerError)
		return
	}

	vacSkills.HardSkills, vacSkills.SoftSkills, err = manager.GetVacancySkills(vacId, *cookies.GetRole(r))
	if err != nil {
		log.Printf("Ошибка получение навыков вакансии: %v", err)
		http.Error(w, "Ошибка получения навыков вакансии", http.StatusInternalServerError)
		return
	}

	analyzedSkills, err := analysis.AnalyseResumeSkills(resSkills, vacSkills)
	if err != nil {
		log.Printf("Ошибка анализа навыков: %v", err)
		http.Error(w, "Ошибка анализа навыков", http.StatusInternalServerError)
		return
	}

	err = manager.SaveAnalyzedData("hr", resumeId, vacId, analyzedSkills)
	if err != nil {
		log.Printf("Ошибка сохранения анализа навыков: %v", err)
		http.Error(w, "Ошибка сохранения анализа навыков", http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("Ошибка ID: %v", err), http.StatusBadRequest)
		return
	}
	intVacId, err := strconv.Atoi(vacId)
	if err != nil {
		log.Printf("Неправильный ID: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка ID: %v", err), http.StatusBadRequest)
		return
	}

	res, err := manager.GetAnalizedData(intFinId, intVacId)
	if err != nil {
		log.Printf("Ошибка получения анализа резюме: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка при получении анализа резюме: %v", err), http.StatusBadRequest)
		return
	}

	resume, err := os.Open(filepath.Join("finder", strconv.Itoa(intFinId), "resume", "resume.txt"))
	if err != nil {
		log.Printf("Ошибка открытия файла: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка при открытии файла: %v", err), http.StatusBadRequest)
		return
	}
	defer resume.Close()

	resumeFileData, err := io.ReadAll(resume)
	if err != nil {
		log.Printf("Ошибка чтения файла: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка при чтении файла: %v", err), http.StatusBadRequest)
		return
	}

	jsonResume := map[string]interface{}{
		"resume_text": string(resumeFileData),
		"mismatch":    res.Mismatch,
		"coincidence": res.Coincidence,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(jsonResume); err != nil {
		log.Printf("Ошибка вывода резюме: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка при выводе резюме: %v", err), http.StatusBadRequest)
		return
	}
}

func AddFinderResume(w http.ResponseWriter, r *http.Request, manager *db.Manager) {
	resumeFile, _, err := r.FormFile("resume")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении файла: %v", err), http.StatusBadRequest)
		return
	}
	defer resumeFile.Close()

	resumeFileData, err := io.ReadAll(resumeFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при чтении файла: %v", err), http.StatusBadRequest)
		return
	}

	resumeDir := fmt.Sprintf("user/%d/resume", *cookies.GetId(r))
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

	resumeId, err := manager.CreateResumeForUser(*cookies.GetId(r))
	if err != nil {
		log.Println(err)
		return
	}

	resumeSkills, err := saveResumeSkills(resumeData["hard_skills"], resumeData["soft_skills"], resumeId, manager, "users")
	if err != nil {
		log.Println(err)
		return
	}

	topVacs, err := GetTopMatchingVacancies(resumeSkills, manager, *cookies.GetRole(r))
	if len(topVacs) >= 3 {
		err = manager.UpdateUserResumesWithTopVacancies(resumeId, topVacs)
		if err != nil {
			log.Printf("Ошибка сохранения топ-3 вакансий: %v", err)
			http.Error(w, fmt.Sprintf("Ошибка сохранения топ-3 вакансий: %v", err), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		log.Printf("Ошибка получения топ-3 вакансий: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка получения топ-3 вакансий: %v", err), http.StatusInternalServerError)
		return
	}
	for _, vac := range topVacs {
		err = manager.SaveAnalyzedData("users", resumeId, vac.VacancyId, vac.FinalSkills)
		if err != nil {
			log.Printf("Ошибка сохранения данных аналитики: %v", err)
			http.Error(w, fmt.Sprintf("Ошибка сохранения данных аналитики: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topVacs)
}

func GetTopMatchingVacancies(resumeSkills models.ResumeSkills, manager *db.Manager, role string) ([]models.VacancyMatchResult, error) {
	vacancies, err := manager.GetAllVacancies(role)
	if err != nil {
		log.Printf("Ошибка получения списка вакансий: %v", err)
		return nil, err
	}

	var results []models.VacancyMatchResult

	for _, vacancy := range vacancies {
		var vacSkills models.VacancySkills
		vacSkills.HardSkills, vacSkills.SoftSkills, err = manager.GetVacancySkills(vacancy.Id, role)
		if err != nil {
			log.Printf("Ошибка получения навыков вакансии %d: %v", vacancy.Id, err)
			continue
		}

		analyzedSkills, err := analysis.AnalyseResumeSkills(resumeSkills, vacSkills)
		if err != nil {
			log.Printf("Ошибка анализа навыков для вакансии %d: %v", vacancy.Id, err)
			continue
		}

		results = append(results, models.VacancyMatchResult{
			VacancyId:   vacancy.Id,
			FinalSkills: analyzedSkills,
			MatchRate:   analyzedSkills.Percent,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].MatchRate > results[j].MatchRate
	})

	return results, err
}

type UserResumeInfo struct {
	HardSkills []string
	SoftSkills []string
	Vacancies  []struct {
		Name    string
		Percent int
		Skills  models.AnalyzedSkills
	}
}

func GetFinderResume(w http.ResponseWriter, r *http.Request, manager *db.Manager) {
	vars := mux.Vars(r)
	resume_id := vars["resume_id"]
	resumeId, _ := strconv.Atoi(resume_id)

	var result UserResumeInfo
	result.HardSkills, result.SoftSkills, _ = manager.GetSkillsFromUserResume(resumeId)
	
	vacancies_id, _ := manager.GetResumeUserVacancies(resumeId)
	for _, vacancy_id := range vacancies_id {
		percent, name, _ := manager.GetSuperDuperSecretAnonymousBitcoinWalletUnderUSAProtectionSkillAssPercentMatch(resumeId, vacancy_id)
		skills, _ := manager.GetAnalizedUserData(resumeId, vacancy_id)
		result.Vacancies = append(result.Vacancies, struct {
			Name    string
			Percent int
			Skills  models.AnalyzedSkills
		}{
			Name:    name,
			Percent: percent,
			Skills:  skills,
		})
	}
	json.NewEncoder(w).Encode(result)
}
