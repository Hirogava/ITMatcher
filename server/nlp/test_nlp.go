package nlp

import (
	"encoding/json"
	"gaspr/db"
	"io"
	"log"
	"net/http"
	"os"
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

	jsonFile, _, err := r.FormFile("resume_data")
	if err != nil {
		log.Printf("Ошибка при получении JSON-файла: %v", err)
		http.Error(w, "Ошибка при открытии json файла", http.StatusBadRequest)
		return
	}
	defer jsonFile.Close()

	var resume ResumeData
	err = json.NewDecoder(jsonFile).Decode(&resume)
	if err != nil {
		log.Printf("Ошибка при декодировании JSON: %v", err)
		http.Error(w, "Ошибка при получении файла", http.StatusBadRequest)
		return
	}

	dir := "nlpData/" + strconv.Itoa(resume.ResumeId)
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

			if _, err := io.Copy(dst, file); err != nil {
				log.Printf("Ошибка при копировании файла: %v", err)
				http.Error(w, "Ошибка при копировании файла", http.StatusBadRequest)
				return
			}
			log.Println("Файл успешно загружен")

			file.Close()
			dst.Close()
		}


	}
}