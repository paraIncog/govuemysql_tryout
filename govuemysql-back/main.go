package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func initDB() {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		getenv("DB_USER", "root"),
		getenv("DB_PASS", "dubimin"),
		getenv("DB_HOST", "127.0.0.1"),
		getenv("DB_PORT", "3306"),
		getenv("DB_NAME", "sample_db"),
	)
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("DB open error: %v", err)
	}
	for i := 0; i < 30; i++ {
		if err = db.Ping(); err == nil {
			log.Println("✅ Connected to DB")
			return
		}
		log.Println("DB not ready, retrying...", err)
		time.Sleep(2 * time.Second)
	}
	log.Fatalf("DB connection failed: %v", err)
}

func ensureSchema() error {
    dbname := getenv("DB_NAME", "sample_db")
    if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + dbname); err != nil {
        return err
    }
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
          id INT AUTO_INCREMENT PRIMARY KEY,
          name VARCHAR(100) NOT NULL,
          email VARCHAR(255) NOT NULL UNIQUE,
          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `)
    return err
}

func main() {
	initDB()
	if err := ensureSchema(); err != nil { log.Fatalf("schema setup failed: %v", err) }
	s := NewServer(db)
	port := getenv("PORT", "8080")
	log.Printf("🚀 Server on :%s", port)
	if err := http.ListenAndServe(":"+port, s.Router()); err != nil {
		log.Fatal(err)
	}
}
