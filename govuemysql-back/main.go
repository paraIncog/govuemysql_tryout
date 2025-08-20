package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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


type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	initDB()
	if err := ensureSchema(); err != nil {
        log.Fatalf("schema setup failed: %v", err)
    }
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{getenv("CORS_ORIGIN", "*")},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			if err := db.Ping(); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "down", "error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		api.GET("/users", func(c *gin.Context) {
			rows, err := db.Query("SELECT id, name, email FROM users ORDER BY id")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer rows.Close()
			var users []User
			for rows.Next() {
				var u User
				if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				users = append(users, u)
			}
			c.JSON(http.StatusOK, users)
		})

		api.GET("/users/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
				return
			}
			var u User
			err = db.QueryRow("SELECT id, name, email FROM users WHERE id = ?", id).Scan(&u.ID, &u.Name, &u.Email)
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, u)
		})

		api.POST("/users", func(c *gin.Context) {
			var in struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}
			if err := c.ShouldBindJSON(&in); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", in.Name, in.Email)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			id, _ := res.LastInsertId()
			c.JSON(http.StatusCreated, gin.H{"id": id})
		})

		api.PUT("/users/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
				return
			}
			var in struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}
			if err := c.ShouldBindJSON(&in); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := db.Exec("UPDATE users SET name = ?, email = ? WHERE id = ?", in.Name, in.Email, id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			aff, _ := res.RowsAffected()
			if aff == 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.Status(http.StatusNoContent)
		})

		api.DELETE("/users/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
				return
			}
			res, err := db.Exec("DELETE FROM users WHERE id = ?", id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			aff, _ := res.RowsAffected()
			if aff == 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.Status(http.StatusNoContent)
		})
	}

	// Frontend
	r.Static("/assets", "./frontend/assets")
	r.GET("/favicon.ico", func(c *gin.Context) { c.File("./frontend/favicon.ico") })

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.File(filepath.Clean("./frontend/index.html"))
	})

	port := getenv("PORT", "8080")
	log.Printf("🚀 Server running on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

