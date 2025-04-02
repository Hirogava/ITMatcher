package nlp

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"gaspr/db"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type ResumeData struct {
	ResumeId  int `json:"resume_id"`
	VacancyId int `json:"vacancy_id"`
}

func SaveFiles(w http.ResponseWriter, r *http.Request, db *db.DBManager) {
	log.Println("Запрос получен")
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("Ошибка обработки multipart-запроса: %v", err)
		http.Error(w, "Ошибка при получении файла", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("Ошибка обработки multipart-запроса: %v", err)
		http.Error(w, "Ошибка при получении файла", http.StatusBadRequest)
		return
	}

	jsonStr := r.FormValue("resume_data")
	if jsonStr == "" {
		log.Println("Ошибка: JSON не передан")
		http.Error(w, "Отсутствует JSON-данные", http.StatusBadRequest)
		return
	}

	var resume ResumeData
	if err := json.Unmarshal([]byte(jsonStr), &resume); err != nil {
		log.Printf("Ошибка при декодировании JSON: %v", err)
		http.Error(w, "Ошибка при обработке JSON", http.StatusBadRequest)
		return
	}

	dir := "aiData/" + strconv.Itoa(resume.ResumeId)
	os.MkdirAll(dir, os.ModePerm)

	for _, fileHeaders := range r.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			file, err := fileHeader.Open()
			if err != nil {
				log.Printf("Ошибка при открытии файла: %v", err)
				http.Error(w, "Ошибка при получении файла", http.StatusBadRequest)
				return
			}
			
			dstPath := filepath.Join(dir, "resume.json")
			dst, err := os.Create(dstPath)
			if err != nil {
				log.Printf("Ошибка при создании файла: %v", err)
				http.Error(w, "Ошибка при создании файла", http.StatusBadRequest)
				return
			}

			resumeFile, err := io.ReadAll(file)
			if err != nil {
				log.Println(err)
				return
			}

			resumeData, err := aiRequest(string(resumeFile))
			if err != nil {
				log.Println(err)
				return
			}
			var result map[string][]string
			if err := json.Unmarshal([]byte(resumeData), &result); err != nil {
				log.Printf("Ошибка при декодировании ответа AI: %v", err)
				http.Error(w, "Ошибка при обработке данных AI", http.StatusInternalServerError)
				return
			}

			saveHardSoftSkills(result["hard_skills"], result["soft_skills"], resume.ResumeId, db)

			dst.WriteString(resumeData)

			log.Println("Файл успешно загружен")
			file.Close()
			dst.Close()
		}
	}
}

func saveHardSoftSkills(hard_skills []string, soft_skills []string, resume_id int, db *db.DBManager) {

	for _, skill := range hard_skills {
		var hardSkillID int
		err := db.DB.QueryRow("SELECT id FROM hard_skills WHERE hard_skill = $1", skill).Scan(&hardSkillID)
		if err != nil {
			if err == sql.ErrNoRows {
				err = db.DB.QueryRow("INSERT INTO hard_skills (hard_skill) VALUES ($1) RETURNING id", skill).Scan(&hardSkillID)
				if err != nil {
					log.Printf("Ошибка при добавлении hard_skill: %v", err)
					continue
				}
			} else {
				log.Printf("Ошибка при проверке hard_skill: %v", err)
				continue
			}
		}

		_, err = db.DB.Exec("INSERT INTO resume_hard_skill (resume_id, hard_skill_id) VALUES ($1, $2)", resume_id, hardSkillID)
		if err != nil {
			log.Printf("Ошибка при добавлении в resume_hard_skill: %v", err)
		}
	}

	for _, skill := range soft_skills {
		var softSkillID int
		err := db.DB.QueryRow("SELECT id FROM soft_skills WHERE soft_skill = $1", skill).Scan(&softSkillID)
		if err != nil {
			if err == sql.ErrNoRows {
				err = db.DB.QueryRow("INSERT INTO soft_skills (soft_skill) VALUES ($1) RETURNING id", skill).Scan(&softSkillID)
				if err != nil {
					log.Printf("Ошибка при добавлении soft_skill: %v", err)
					continue
				}
			} else {
				log.Printf("Ошибка при проверке soft_skill: %v", err)
				continue
			}
		}

		_, err = db.DB.Exec("INSERT INTO resume_soft_skill (resume_id, soft_skill_id) VALUES ($1, $2)", resume_id, softSkillID)
		if err != nil {
			log.Printf("Ошибка при добавлении в resume_soft_skill: %v", err)
		}
	}
}

func aiRequest(resume string) (string, error) {
	cmd := exec.Command("python", "ai/main.py")
	cmd.Stdin = bytes.NewBufferString(resume)
	cmd.Env = append(os.Environ(), "PYTHONIOENCODING=utf-8")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("AI request failed: %v, output: %s", err, string(output))
	}

	return string(output), nil
}
