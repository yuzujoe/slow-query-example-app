package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env/v11"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
)

type dbConfig struct {
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT" envDefault:"3306"`
	Name     string `env:"DB_NAME"`
}

type newrelicConfig struct {
	AppName    string `env:"NEW_RELIC_APP_NAME"`
	LicenseKey string `env:"NEW_RELIC_LICENSE_KEY"`
}

type config struct {
	DB       dbConfig
	NewRelic newrelicConfig
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var db *sql.DB

func main() {
	var err error
	cfg := config{}
	if err = env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	db, err = connectDB(cfg.DB)
	if err != nil {
		log.Fatalf("Error on creating database connection: %s", err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to database")

	http.HandleFunc("/users", handleUsers)
	http.HandleFunc("/users/", handleUser)

	fmt.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getUsers(w, r)
	case "POST":
		createUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleUser(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		getUser(w, r)
	case "PUT":
		updateUser(w, r)
	case "DELETE":
		deleteUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getUsers(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query("SELECT id, name, email FROM users")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	json.NewEncoder(w).Encode(users)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := db.QueryRowContext(r.Context(), "INSERT INTO users (name, email) VALUES (?, ?) RETURNING id", u.Name, u.Email).Scan(&u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/users/"):]

	query := "SELECT id, name, email FROM users WHERE id = ?"

	var u User
	err := db.QueryRow(query, id).Scan(&u.ID, &u.Name, &u.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(u)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/users/"):]
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.ExecContext(r.Context(), "UPDATE users SET name = ?, email = ? WHERE id = ?", u.Name, u.Email, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/users/"):]
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func connectDB(cfg dbConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
