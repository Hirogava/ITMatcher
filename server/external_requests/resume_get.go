package externalrequests

import (
	"encoding/json"
	"gaspr/db"
	"log"
	"net/http"

)

type Finder struct {
	HRUsername string `json:"hr_username,string"`
	FirstName string `json:"first_name,string"`
	LastName string `json:"last_name,string"`
	Surname string `json:"surname,string"`
	Email string `json:"email,string"`
	Phone string `json:"phone,string"`
	Portfolio bool `json:"portfolio"`
}

func GetResume(w http.ResponseWriter, r *http.Request, db *db.DBManager){
	var finderData Finder
	if err := json.NewDecoder(r.Body).Decode(&finderData); err != nil {
        log.Printf("Ошибка при получении данных: %v", err)
        http.Error(w, "Ошибка при получении данных", http.StatusBadRequest)
        return
    }

	row := db.DB.QueryRow("SELECT id FROM hr WHERE username = $1", finderData.HRUsername)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		log.Printf("Ошибка при получении данных: %v", err)
		http.Error(w, "Ошибка при получении данных", http.StatusBadRequest)
		return
	}

	query := "INSERT INTO finders (hr_id, first_name, last_name, surname, email, phone, portfolio) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	_, err = db.DB.Exec(query, id, finderData.FirstName, finderData.LastName, finderData.Surname, finderData.Email, finderData.Phone, finderData.Portfolio)
	if err != nil {
		log.Printf("Ошибка при записи в базу данных: %v", err)
		http.Error(w, "Ошибка при записи в базу данных", http.StatusBadRequest)
		return
	}

	// os.MkdirAll("resumes/")
}