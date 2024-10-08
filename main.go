package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	_ "github.com/lib/pq"
)

type Employee struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`       // 氏名
	Gender     string `json:"gender"`     // 性別
	HireYear   int    `json:"hire_year"`  // 入社年度
	Address    string `json:"address"`    // 住所
	Department string `json:"department"` // 部署
	Others     string `json:"others"`     // その他
	Image      []byte `json:"image"`      // 画像
	Email      string `json:"email"`      //メールアドレス
	Password   string `json:"password"`   //パスワード
}

var (
	db        *sql.DB
	employees = make(map[int]Employee)
	nextID    = 1
	mu        sync.Mutex
	err       error
)

var jwtKey = []byte("my_secret_key")

// JWTを生成するためのクレーム
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func init() {
	// Docker上のPostgreSQLに接続
	connStr := "host=localhost port=5432 user=suser password=spass dbname=company sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// データベース接続確認
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("Successfully connected to PostgreSQL database")
}

func Login(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var loginReq LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// データベースからユーザー情報を取得
	var dbPassword string
	query := `SELECT password FROM employee WHERE email = $1`
	err = db.QueryRow(query, loginReq.Email).Scan(&dbPassword)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}

	// パスワードの照合
	if loginReq.Password != dbPassword {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// JWTトークンの生成
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: loginReq.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// トークンを返す
	loginRes := LoginResponse{
		Token:   tokenString,
		Message: "Login successful",
		Success: true,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginRes)
}

func CreateEmployee(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	var newEmployee Employee
	err := json.NewDecoder(r.Body).Decode(&newEmployee)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	query := `INSERT INTO employee (name, gender, hire_year, address, department, others, image, email, password)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`
	err = db.QueryRow(query, newEmployee.Name, newEmployee.Gender, newEmployee.HireYear, newEmployee.Address, newEmployee.Department, newEmployee.Others, newEmployee.Image, newEmployee.Email, newEmployee.Password).Scan(&newEmployee.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newEmployee)
}

func IndexEmployee(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	mu.Lock()
	defer mu.Unlock()

	rows, err := db.Query("SELECT id, name, gender, hire_year, address, department, others, image FROM employee")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var employees []Employee
	for rows.Next() {
		var employee Employee
		err := rows.Scan(&employee.ID, &employee.Name, &employee.Gender, &employee.HireYear, &employee.Address, &employee.Department, &employee.Others, &employee.Image)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		employees = append(employees, employee)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employees)
}

func DetailEmployee(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var employee Employee
	query := "SELECT id, name, gender, hire_year, address, department, others, image FROM employee WHERE id = $1"
	err = db.QueryRow(query, id).Scan(&employee.ID, &employee.Name, &employee.Gender, &employee.HireYear, &employee.Address, &employee.Department, &employee.Others, &employee.Image)
	if err == sql.ErrNoRows {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employee)
}

func UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedEmployee Employee
	err = json.NewDecoder(r.Body).Decode(&updatedEmployee)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// SQLクエリでデータを更新
	query := `
		UPDATE employee
		SET name = $1, gender = $2, hire_year = $3, address = $4, department = $5, others = $6, image = $7
		WHERE id = $8`
	_, err = db.Exec(query, updatedEmployee.Name, updatedEmployee.Gender, updatedEmployee.HireYear, updatedEmployee.Address, updatedEmployee.Department, updatedEmployee.Others, updatedEmployee.Image, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedEmployee)
}

func DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// SQLクエリでデータを削除
	query := `DELETE FROM employee WHERE id = $1`
	result, err := db.Exec(query, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 削除が実行された行数をチェック
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	http.HandleFunc("/login", Login)
	http.HandleFunc("/index", IndexEmployee)
	http.HandleFunc("/index/create", CreateEmployee)
	http.HandleFunc("/index/detail", DetailEmployee)
	http.HandleFunc("/index/update", UpdateEmployee)
	http.HandleFunc("/index/delete", DeleteEmployee)
	// サーバー起動コマンド
	log.Fatal(http.ListenAndServe(":8081", nil))
}
