package nlp

import (
	"bytes"
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

type ResumeData struct{
	ResumeId int `json:"resume_id"`
	VacancyId int `json:"vacancy_id"`
}

func SaveFiles(w http.ResponseWriter, r *http.Request, db *db.DBManager){
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
			
			dstPath := filepath.Join(dir, fileHeader.Filename)
			dst, err := os.Create(dstPath)
			if err != nil {
				log.Printf("Ошибка при создании файла: %v", err)
				http.Error(w, "Ошибка при создании файла", http.StatusBadRequest)
				return
			}

			resume, err := io.ReadAll(file)
			if err != nil {
				log.Println(err)
				return
			}

			resumeData, err := aiRequest(string(resume))
			if err != nil {
				log.Println(err)
				return
			}
			dst.WriteString(resumeData)

			log.Println("Файл успешно загружен")
			file.Close()
			dst.Close()
		}
	}
}

func aiRequest (resume string) (string, error) {
	cmd := exec.Command("python", "ai\\main.py")
    cmd.Stdin = bytes.NewBufferString(resume)

    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("AI request failed: %v, output: %s", err, string(output))
    }

	return string(output), nil
}