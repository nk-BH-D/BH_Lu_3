package main

import (
	"database/sql"
	"log"
	"net/http"
	"runtime/debug"

	mux "github.com/gorilla/mux"
	BH "github.com/nk-BH-D/BH_Lu_3/bak/BH"
)

func CreateTableIfNotExists(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS user_data (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        login TEXT NOT NULL,
        password TEXT NOT NULL,
        hashed_login TEXT NOT NULL,
        hashed_password TEXT NOT NULL,
		expressions TEXT NOT NULL,
		expression_results TEXT NOT NULL
    );`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	log.Println("Таблица user_data успешно создана или уже существует.")
	return nil
}

func main() {
	// Подключаемся к базе данных
	db, err := sql.Open("sqlite3", "user_data.db") // Замените "database_name.db" на имя вашей базы данных
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	// Создаём таблицу, если её ещё нет
	if err := CreateTableIfNotExists(db); err != nil {
		log.Fatalf("Ошибка создания таблицы: %v", err)
	}
	log.Println("База данных созданна")

	r := mux.NewRouter()

	r.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("паника при обработке /register: %v\n%s", err, debug.Stack())
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}
		}()
		log.Println("Запрос /register")
		BH.RegisterHandler(w, r)
		log.Println("Запрос /register обработан")
	})

	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("паника при обработке /register: %v\n%s", err, debug.Stack())
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}
		}()
		log.Println("запрос /login")
		BH.LoginHandler(w, r)
		log.Println("Запрос /login обработан")
	})

	r.HandleFunc("/calculator", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("паника при обработке /register: %v\n%s", err, debug.Stack())
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}
		}()
		log.Println("Запрос /calculator")
		BH.CalculateHandlerWithAuth(w, r)
		log.Println("Запрос /calcualtor обработан")
	})

	r.HandleFunc("/my_data", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("паника при обработке /register: %v\n%s", err, debug.Stack())
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}
		}()
		log.Println("Запрос /my_data")
		BH.GetExpressionsHandler(w, r)
		log.Println("Запрос /my_data обработан")
	})
	log.Printf("Сервер запущен на порту 9090")
	log.Fatal(http.ListenAndServe(":9090", r))
}
