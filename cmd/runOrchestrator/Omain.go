package main

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"

	mux "github.com/gorilla/mux"
	orch "github.com/nk-BH-D/BH_Lu_3/bak/Orchestrator"
)

func main() {
	md := orch.NewMemoryData(5 * time.Minute)

	r := mux.NewRouter()

	r.HandleFunc("/calc", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("паника при обработке /calc: %v\n%s", err, debug.Stack())
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}
		}()
		log.Println("Запрос /calc")
		orch.CalculateHandler(w, r, md)
		log.Println("Запрос /calc обработан")
	})

	r.HandleFunc("/inLCF", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("паника при обработке /calc: %v\n%s", err, debug.Stack())
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}
		}()
		log.Println("Запрос /inLCf")
		orch.LCF_Otvet(w, r, md)
		log.Println("Запрос /inLCf обработан")
	})
	log.Printf("Оркестратор запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
