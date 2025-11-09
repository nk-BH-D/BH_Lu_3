package main

import (
	"log"
	"net/http"
	"runtime/debug"

	mux "github.com/gorilla/mux"
	lu "github.com/nk-BH-D/BH_Lu_3/bak/Demon"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/LCF", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("паника при обработке LCf: %v\n%s", err, debug.Stack())
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}
		}()
		log.Println("Запрос /LCF")
		lu.LCF(w, r)
		log.Println("Запрос /LCF обработан")
	})
	log.Println("Демон запущен на порту 8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
