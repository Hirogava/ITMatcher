package main

import (
	"fmt"
	// "gaspr/cookies"
	"gaspr/db"
	"gaspr/external_requests"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	db, err := db.NewDBManager("postgres", "user=postgres password=197320 dbname=projectDB sslmode=disable")
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()
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

	r.HandleFunc("/nlp", func(w http.ResponseWriter, r *http.Request){
		log.Println("OK")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("вСЕ ДОШЛО"))
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}