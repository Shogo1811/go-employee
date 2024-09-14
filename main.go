package main

import (
	"encoding/json"
	// "fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
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
}

var (
	employees = make(map[int]Employee)
	nextID    = 1
	mu        sync.Mutex
)

func CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var newEmployee Employee
	err := json.NewDecoder(r.Body).Decode(&newEmployee)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	newEmployee.ID = nextID
	employees[nextID] = newEmployee
	nextID++
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newEmployee)
}

func IndexEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(employees)
}

func EmployeeDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	employee, exists := employees[id]
	mu.Unlock()
	if !exists {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employee)
}

func UpdateEmployee(w http.ResponseWriter, r *http.Request) {
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

	mu.Lock()
	defer mu.Unlock()
	if _, exists := employees[id]; !exists {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	updatedEmployee.ID = id
	employees[id] = updatedEmployee

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedEmployee)
}

func DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	if _, exists := employees[id]; !exists {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	delete(employees, id)
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	http.HandleFunc("/index", IndexEmployee)
	http.HandleFunc("/index/create_employee", CreateEmployee)
	http.HandleFunc("/index/detail", EmployeeDetail)
	http.HandleFunc("/index/update", UpdateEmployee)
	http.HandleFunc("/index/delete", DeleteEmployee)
	log.Println("Starting server on :8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
