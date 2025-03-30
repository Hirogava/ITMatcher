package externalrequests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gaspr/db"
	"gaspr/ai"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
)

type Finder struct {
	HRUsername string `json:"hr_username"`
	Vacancy string `json:"vacancy"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Surname string `json:"surname"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Portfolio bool `json:"portfolio"`
}

type ResumeData struct{
	ResumeId int `json:"resume_id"`
	VacancyId int `json:"vacancy_id"`
}

func GetResume(w http.ResponseWriter, r *http.Request, db *db.DBManager){
	idCh := make(chan int, 1)
	errCh := make(chan error, 1)
	db.WG.Add(1)
	var resume ResumeData
	go func (chan<- int) {
		defer close(idCh)
		defer close(errCh)
		defer db.WG.Done()

		// для тестового запроса json и резюме, запрос - curl -X POST http://localhost:8080/get_resume -F "data=@test.json;type=application/json" -F "file=@resume.txt"
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			log.Printf("Ошибка обработки multipart-запроса: %v", err)
			errCh <- err
			return
		}
	
		jsonFile, _, err := r.FormFile("data")
		if err != nil {
			log.Printf("Ошибка при получении JSON-файла: %v", err)
			errCh <- err
			return
		}
		defer jsonFile.Close()
	
		var finderData Finder
		if err := json.NewDecoder(jsonFile).Decode(&finderData); err != nil {
			log.Printf("Ошибка разбора JSON: %v", err)
			errCh <- err
			return
		}
	
		log.Printf("Полученные данные: %+v", finderData)


		row := db.DB.QueryRow("SELECT id FROM hr WHERE username = $1", finderData.HRUsername)
		var id int
		err = row.Scan(&id)
		if err != nil {
			log.Printf("Ошибка при получении данных из БД: %v", err)
			errCh <- err
			return
		}

		query := "INSERT INTO finders (portfolio, hr_id) VALUES ($1, $2) RETURNING id"
		var idFinder int
		err = db.DB.QueryRow(query, finderData.Portfolio, id).Scan(&idFinder)
		if err!= nil {
			log.Printf("Ошибка при записи в базу данных: %v", err)
			errCh <- err
			return
		}
		idCh <- idFinder

		var vacId int
		err = db.DB.QueryRow("SELECT id FROM vacancies WHERE name = $1", finderData.Vacancy).Scan(&vacId)
		if err != nil {
			log.Printf("Ошибка при получении данных из БД: %v", err)
			errCh <- err
			return
		}
		resume.VacancyId = vacId

		var resumeId int
		query = "INSERT INTO resumes (finder_id, first_name, last_name, surname, email, phone_number, vacancy_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id"
		err = db.DB.QueryRow(query, idFinder, finderData.FirstName, finderData.LastName, finderData.Surname, finderData.Email, finderData.Phone, vacId).Scan(&resumeId)
		if err != nil {
			log.Printf("Ошибка при записи в базу данных: %v", err)
			errCh <- err
			return
		}
		resume.ResumeId = resumeId
	}(idCh)

	files := []string{}

	select {
		case err := <-errCh:
			log.Print(err)
			http.Error(w, "Ошибка обработки запроса", http.StatusBadRequest)
			return
		case id := <-idCh:
			dir := "finders/" + strconv.Itoa(id) + "/resume"
			os.MkdirAll(dir, os.ModePerm)
			for _, fileHeaders := range r.MultipartForm.File {
				for _, fileHeader := range fileHeaders {
					file, err := fileHeader.Open()
					if err != nil {
						log.Printf("Ошибка при получении файла: %v", err)
						http.Error(w, "Ошибка при получении файла", http.StatusBadRequest)
						return
					}

					if fileHeader.Header.Get("Content-Type") == "application/json" {
						log.Println("JSON файл, пропускаем сохранение")
						continue
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

					files = append(files, dstPath)

					file.Close()
					dst.Close()
				}
			}
	}
	db.WG.Wait()

	if err := sendFilesAndMetadata(files, resume, "/nlp", db); err != nil {
		log.Printf("Ошибка отправки файлов в NLP API: %v", err)
		http.Error(w, "Ошибка отправки данных", http.StatusInternalServerError)
		return
	}
}

func sendFilesAndMetadata(files []string, metadata ResumeData, apiURL string, db *db.DBManager) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("ошибка сериализации JSON: %v", err)
	}
	part, err := writer.CreateFormFile("resume_data", "resume_data.json")
	if err != nil {
		return fmt.Errorf("ошибка при создании JSON-part: %v", err)
	}
	if _, err := part.Write(jsonData); err != nil {
		return fmt.Errorf("ошибка записи JSON в multipart: %v", err)
	}

	for _, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("ошибка открытия файла %s: %v", filePath, err)
		}
		defer file.Close()

		filePart, err := writer.CreateFormFile("files", filepath.Base(filePath))
		if err != nil {
			return fmt.Errorf("ошибка создания part для файла %s: %v", filePath, err)
		}

		if _, err := io.Copy(filePart, file); err != nil {
			return fmt.Errorf("ошибка копирования файла %s: %v", filePath, err)
		}
	}

	writer.Close()

	req, err := http.NewRequest("POST", "/nlp", body)
    if err != nil {
        log.Printf("Ошибка создания запроса к /nlp: %v", err)
		return err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

	recorder := httptest.NewRecorder()
    nlp.SaveFiles(recorder, req, db)

    if recorder.Code != http.StatusOK {
        log.Printf("Ошибка при вызове /nlp: %s", recorder.Body.String())
        return err
    }
	return nil
}
