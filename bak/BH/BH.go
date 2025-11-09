package BH

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	_ "github.com/mattn/go-sqlite3"
)

// присоединяемся к базе данных
func connectDB() (*sql.DB, error) {
	return sql.Open("sqlite3", "user_data.db")
}

type UserData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

var jwtKey = []byte("BH")

func generateJWT(login string) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   login,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("RH: начало работы")
	// проверяем метод
	if r.Method != http.MethodPost {
		log.Printf("метод: %s не поддерживается для этого запроса", r.Method)
		http.Error(w, "метод не разрешон", http.StatusMethodNotAllowed)
		return
	}
	// декодируем JSON
	var req_USD UserData
	log.Println("RH: декодирует JSON")
	if err := json.NewDecoder(r.Body).Decode(&req_USD); err != nil {
		log.Printf("ошибка JSON: %v", err)
		http.Error(w, "ошибки при декодировании JSON", http.StatusBadRequest)
		return
	}
	log.Printf("RH: декодировал JSON успешно: %+v", req_USD)
	// хешируем логин
	hashLogin := sha256.Sum256([]byte(req_USD.Login))
	hashStrLogin := hex.EncodeToString(hashLogin[:])
	// хешируем пароль
	hashPassword := sha256.Sum256([]byte(req_USD.Password))
	hashStrPassword := hex.EncodeToString(hashPassword[:])
	// подключаемся к db
	db, err := connectDB()
	if err != nil {
		log.Printf("ошибка подключения к базе данных: %v", err)
		http.Error(w, "ошибка подключения к базе данных", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	// записываем данные в db
	query := `INSERT INTO user_data (login, password, hashed_login, hashed_password, expressions, expression_results) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = db.Exec(query, req_USD.Login, req_USD.Password, hashStrLogin, hashStrPassword, "", "")
	if err != nil {
		log.Printf("ошибка записи в базу данных: %v", err)
		http.Error(w, "ошибка записи в базу данных", http.StatusInternalServerError)
		return
	}
	//отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message_0":     "Данные успешно сохранены",
		"hash_login":    hashStrLogin,
		"hash_password": hashStrPassword,
	}
	json.NewEncoder(w).Encode(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("LH: начало работы")
	// проверяем метод
	if r.Method != http.MethodPost {
		log.Printf("метод: %s не поддерживается для этого запроса", r.Method)
		http.Error(w, "метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	// декодируем json
	var req_USD UserData
	log.Println("LH: декодирует json")
	if err := json.NewDecoder(r.Body).Decode(&req_USD); err != nil {
		log.Printf("ошибка json: %v", err)
		http.Error(w, "ошибки при декодировании json", http.StatusBadRequest)
		return
	}
	log.Printf("LH: декодировал json успешно: %+v", req_USD)

	// подключаемся к базе данных
	db, err := connectDB()
	if err != nil {
		log.Printf("ошибка подключения к базе данных: %v", err)
		http.Error(w, "ошибка подключения к базе данных", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// извлекаем хеши из базы данных
	var dbHashedPassword, dbHashedLogin string
	query := `SELECT hashed_login, hashed_password FROM user_data WHERE hashed_login = $1`
	err = db.QueryRow(query, req_USD.Login).Scan(&dbHashedLogin, &dbHashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("пользователь с логином %s не найден", req_USD.Login)
			http.Error(w, "неверный логин или пароль", http.StatusUnauthorized)
			return
		}
		log.Printf("ошибка чтения из базы данных: %v", err)
		http.Error(w, "ошибка чтения из базы данных", http.StatusInternalServerError)
		return
	}

	// сравниваем хеши
	if dbHashedPassword != req_USD.Password {
		log.Printf("неверный пароль для логина %s", req_USD.Login)
		http.Error(w, "неверный логин или пароль", http.StatusUnauthorized)
		return
	}

	// генерируем jwt
	token, err := generateJWT(req_USD.Login)
	if err != nil {
		log.Printf("ошибка генерации jwt: %v", err)
		http.Error(w, "ошибка генерации токена", http.StatusInternalServerError)
		return
	}

	// обновляем токен в базе данных
	updateQuery := `UPDATE user_data SET id = $1 WHERE login = $2`
	_, err = db.Exec(updateQuery, token, req_USD.Login)
	if err != nil {
		log.Printf("ошибка обновления токена в базе данных: %v", err)
		http.Error(w, "ошибка обновления токена", http.StatusInternalServerError)
		return
	}

	// отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message": "успешный вход",
		"token":   token,
	}
	json.NewEncoder(w).Encode(response)
}

type Expression_BH_In_Server struct {
	Expression string `json:"expr"`
}

func CalculateHandlerWithAuth(w http.ResponseWriter, r *http.Request) {
	log.Println("CHWA: начал обработку выражения с авторизацией")

	// проверяем заголовок Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "требуется авторизация", http.StatusUnauthorized)
		return
	}

	// извлекаем токен
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "неверный формат токена", http.StatusUnauthorized)
		return
	}
	tokenString := parts[1]

	// проверяем токен
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "недействительный токен", http.StatusUnauthorized)
		return
	}

	// декодируем тело запроса
	var req_EBHIS Expression_BH_In_Server
	log.Println("CHWA: декодирует JSON")
	if err := json.NewDecoder(r.Body).Decode(&req_EBHIS); err != nil {
		log.Printf("ошибка JSON: %v", err)
		http.Error(w, "ошибки при декодировании JSON", http.StatusBadRequest)
		return
	}
	log.Printf("CHWA: декодировал JSON успешно: %+v", req_EBHIS)

	// отправляем выражение в оркестратор
	expressionPayload := map[string]string{
		"expr": req_EBHIS.Expression,
	}

	payloadBytes, err := json.Marshal(expressionPayload)
	if err != nil {
		log.Printf("ошибка при маршалинге JSON: %v", err)
		http.Error(w, "ошибка при подготовке данных для оркестратора", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post("http://orchestrator:8080/calc", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("ошибка отправки в оркестратор: %v", err)
		http.Error(w, "ошибка отправки в оркестратор", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// читаем результат от оркестратора
	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("ошибка декодирования ответа от оркестратора: %v", err)
		http.Error(w, "ошибка обработки ответа от оркестратора", http.StatusInternalServerError)
		return
	}
	log.Printf("CHWA: получил результат от оркестратора: %+v", result)

	// подключаемся к базе данных
	db, err := connectDB()
	if err != nil {
		log.Printf("ошибка подключения к базе данных: %v", err)
		http.Error(w, "ошибка подключения к базе данных", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// обновляем данные в базе данных
	updateQuery := `UPDATE user_data SET expressions = COALESCE(expressions, '') || $1 || ';', expression_results = COALESCE(expression_results, '') || $2 || ';' WHERE hashed_login = $3`
	_, err = db.Exec(updateQuery, req_EBHIS.Expression, result["result"], claims.Subject)
	if err != nil {
		log.Printf("ошибка обновления данных в базе данных: %v", err)
		http.Error(w, "ошибка обновления данных в базе данных", http.StatusInternalServerError)
		return
	}

	// отправляем результат пользователю
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message":  "выражение обработано",
		"result":   result["result"],
		"token_id": claims.Subject,
	}
	json.NewEncoder(w).Encode(response)
}

func GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("GEH: начал обработку запроса на получение выражений")

	// проверяем заголовок Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "требуется авторизация", http.StatusUnauthorized)
		return
	}

	// извлекаем токен
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "неверный формат токена", http.StatusUnauthorized)
		return
	}
	tokenString := parts[1]

	// проверяем токен
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "недействительный токен", http.StatusUnauthorized)
		return
	}

	// подключаемся к базе данных
	db, err := connectDB()
	if err != nil {
		log.Printf("ошибка подключения к базе данных: %v", err)
		http.Error(w, "ошибка подключения к базе данных", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// извлекаем выражения и результаты
	var expressions, results string
	query := `SELECT expressions, expression_results FROM user_data WHERE hashed_login = $1`
	err = db.QueryRow(query, claims.Subject).Scan(&expressions, &results)
	if err != nil {
		log.Printf("ошибка чтения из базы данных: %v", err)
		http.Error(w, "ошибка чтения из базы данных", http.StatusInternalServerError)
		return
	}

	// отправляем данные пользователю
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"expressions": expressions,
		"results":     results,
	}
	json.NewEncoder(w).Encode(response)
}
