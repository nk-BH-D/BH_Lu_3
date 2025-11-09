package BH

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// вспомогательная функция для создания тестовой базы данных
func setupTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:") // создаём базу данных в памяти
	if err != nil {
		return nil, err
	}

	// создаём таблицу user_data
	createTableQuery := `
    CREATE TABLE user_data (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        login TEXT,
        password TEXT,
        hashed_login TEXT,
        token TEXT,
        expressions TEXT,
        expression_results TEXT
    );`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// тест для RegisterHandler
func TestRegisterHandler(t *testing.T) {
	// создаём тестовую базу данных
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("не удалось создать тестовую базу данных: %v", err)
	}
	defer db.Close()

	// создаём тестовый запрос
	userData := UserData{
		Login:    "test_user",
		Password: "test_password",
	}
	body, _ := json.Marshal(userData)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// создаём тестовый HTTP-ответ
	rr := httptest.NewRecorder()

	// вызываем обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RegisterHandler(w, r)
	})
	handler.ServeHTTP(rr, req)

	// проверяем статус ответа
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("RegisterHandler вернул статус %v, ожидался %v", status, http.StatusOK)
	}

	// проверяем тело ответа
	expected := `{"message_0":"Данные успешно сохранены"`
	if !bytes.Contains(rr.Body.Bytes(), []byte(expected)) {
		t.Errorf("RegisterHandler вернул неожиданный ответ: %v", rr.Body.String())
	}
}

// тест для LoginHandler
func TestLoginHandler(t *testing.T) {
	// создаём тестовую базу данных
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("не удалось создать тестовую базу данных: %v", err)
	}
	defer db.Close()

	// добавляем тестового пользователя
	_, err = db.Exec(`INSERT INTO user_data (login, password, hashed_login) VALUES (?, ?, ?)`,
		"test_user", "test_password", "hashed_test_user")
	if err != nil {
		t.Fatalf("не удалось добавить тестового пользователя: %v", err)
	}

	// создаём тестовый запрос
	userData := UserData{
		Login:    "test_user",
		Password: "test_password",
	}
	body, _ := json.Marshal(userData)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// создаём тестовый HTTP-ответ
	rr := httptest.NewRecorder()

	// вызываем обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LoginHandler(w, r)
	})
	handler.ServeHTTP(rr, req)

	// проверяем статус ответа
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("LoginHandler вернул статус %v, ожидался %v", status, http.StatusOK)
	}

	// проверяем тело ответа
	expected := `{"message":"успешный вход"`
	if !bytes.Contains(rr.Body.Bytes(), []byte(expected)) {
		t.Errorf("LoginHandler вернул неожиданный ответ: %v", rr.Body.String())
	}
}

// тест для CalculateHandlerWithAuth
func TestCalculateHandlerWithAuth(t *testing.T) {
	// создаём тестовую базу данных
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("не удалось создать тестовую базу данных: %v", err)
	}
	defer db.Close()

	// добавляем тестового пользователя с токеном
	_, err = db.Exec(`INSERT INTO user_data (login, password, hashed_login, token) VALUES (?, ?, ?, ?)`,
		"test_user", "test_password", "hashed_test_user", "test_token")
	if err != nil {
		t.Fatalf("не удалось добавить тестового пользователя: %v", err)
	}

	// создаём тестовый запрос
	expression := Expression_BH_In_Server{
		Expression: "2+2",
	}
	body, _ := json.Marshal(expression)
	req := httptest.NewRequest(http.MethodPost, "/calculate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test_token")

	// создаём тестовый HTTP-ответ
	rr := httptest.NewRecorder()

	// вызываем обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CalculateHandlerWithAuth(w, r)
	})
	handler.ServeHTTP(rr, req)

	// проверяем статус ответа
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("CalculateHandlerWithAuth вернул статус %v, ожидался %v", status, http.StatusOK)
	}

	// проверяем тело ответа
	expected := `{"message":"выражение обработано"`
	if !bytes.Contains(rr.Body.Bytes(), []byte(expected)) {
		t.Errorf("CalculateHandlerWithAuth вернул неожиданный ответ: %v", rr.Body.String())
	}
}
