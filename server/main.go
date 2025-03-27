package main

import (
	"fmt"
	// "gaspr/cookies"
	"gaspr/db"
	"gaspr/external_requests"
	"log"
	"net/http"
	"html/template"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	db, err := db.NewDBManager("postgres", "user=postgres password=197320 dbname=projectDB sslmode=disable")
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	resp, _ := db.DB.Query("SELECT * FROM users")
	fmt.Println(resp)

	r := mux.NewRouter()

	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./test.html"))
		tmpl.Execute(w, nil)
	})

	r.HandleFunc("/get_resume", func(w http.ResponseWriter, r *http.Request) {
		externalrequests.GetResume(w, r, db)
	}).Methods(http.MethodPost)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}