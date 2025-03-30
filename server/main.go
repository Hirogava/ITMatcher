package main

import (
	"gaspr/cookies"
	"gaspr/db"
	"gaspr/nlp"
	"gaspr/external_requests"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	dB, err := db.NewDBManager("postgres", "user=postgres password=197320 dbname=projectDB sslmode=disable")
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer dB.Close()

	if err = db.Migrate(dB); err != nil {
		log.Fatalf("Ошибка миграции базы данных: %v", err)
	}
	log.Println("Миграция базы данных завершена успешно")

	r := mux.NewRouter()

	_ = cookies.NewCookieManager()

	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./test.html"))
		tmpl.Execute(w, nil)
	})

	r.HandleFunc("/get_resume", func(w http.ResponseWriter, r *http.Request) {
		externalrequests.GetResume(w, r, dB)
	}).Methods(http.MethodPost)

	r.HandleFunc("/nlp", func(w http.ResponseWriter, r *http.Request){
		nlp.SaveFiles(w, r, dB)
	}).Methods(http.MethodPost)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	log.Println("Сервер запущен на порту 8080")
    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}