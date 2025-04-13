package handlers

import (
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
	"path/filepath"
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
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	vacId, err := manager.GetVacancyIdByName(vacancy, hr.ID)
	if err != nil {
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	resumeId, err := manager.CreateResumeForHr(finderId, firstName, lastName, surName, phone, email, vacId)
	if err != nil {
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	resumeFileData, err := SaveResume(resumeFile, finderId)
	if err != nil {
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	resumeData, err := ai.Request(string(resumeFileData))
	if err != nil {
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	resSkills, err = saveResumeSkills(resumeData["hard_skills"], resumeData["soft_skills"], resumeId, manager, "hr")
	if err != nil {
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	vacSkills.HardSkills, vacSkills.SoftSkills, err = manager.GetVacancySkills(vacId)
	if err != nil {
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	analyzedSkills, err := resumeanalysis.AnalizResumeSkills(resSkills, vacSkills)
	if err != nil {
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	err = manager.SaveAnalyzedDataForHr(resumeId, vacId, analyzedSkills)
	if err != nil {
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("Resume uploaded successfully"))
	return
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