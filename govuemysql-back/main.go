package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	mysqlDrv "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name" binding:"required"`
	Email     string    `json:"email" binding:"required,email"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	_ = godotenv.Load()

	// Build DSN from env
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		getenv("DB_USER", "root"),
		getenv("DB_PASS", "dubimin"),
		getenv("DB_HOST", "127.0.0.1"),
		getenv("DB_PORT", "3306"),
		getenv("DB_NAME", "sample_db"),
	)

	// Connect DB
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("DB connection failed:", err)
	}

	// Ensure schema
	if err := ensureSchema(db); err != nil {
		log.Fatal("ensure schema failed:", err)
	}

	r := gin.Default()
	r.Use(cors(getenv("CORS_ORIGIN", "*")))

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

		api.GET("/users", func(c *gin.Context) {
			rows, err := db.Query(`SELECT id, name, email, created_at FROM users ORDER BY id DESC`)
			if err != nil {
				log.Printf("list users query error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
				return
			}
			defer rows.Close()

			var users []User
			for rows.Next() {
				var u User
				if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
					log.Printf("list users scan error: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
					return
				}
				users = append(users, u)
			}
			if err := rows.Err(); err != nil {
				log.Printf("list users rows error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "rows failed"})
				return
			}
			c.JSON(http.StatusOK, users)
		})

		api.GET("/users/:id", func(c *gin.Context) {
			id := c.Param("id")
			var u User
			err := db.QueryRow(`SELECT id, name, email, created_at FROM users WHERE id=?`, id).
				Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			if err != nil {
				log.Printf("get user error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
				return
			}
			c.JSON(http.StatusOK, u)
		})

		api.POST("/users", func(c *gin.Context) {
			var in User
			if err := c.ShouldBindJSON(&in); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "name and valid email are required"})
				return
			}

			res, err := db.Exec(`INSERT INTO users (name, email) VALUES (?, ?)`, in.Name, in.Email)
			if err != nil {
				if isDuplicate(err) {
					c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
					return
				}
				log.Printf("create user error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
				return
			}

			id, _ := res.LastInsertId()
			var out User
			if err := db.QueryRow(`SELECT id, name, email, created_at FROM users WHERE id=?`, id).
				Scan(&out.ID, &out.Name, &out.Email, &out.CreatedAt); err != nil {
				log.Printf("fetch after create error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch after create failed"})
				return
			}
			c.JSON(http.StatusCreated, out)
		})

		api.PUT("/users/:id", func(c *gin.Context) {
			id := c.Param("id")
			var in User
			if err := c.ShouldBindJSON(&in); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "name and valid email are required"})
				return
			}

			_, err := db.Exec(`UPDATE users SET name=?, email=? WHERE id=?`, in.Name, in.Email, id)
			if err != nil {
				if isDuplicate(err) {
					c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
					return
				}
				log.Printf("update user error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
				return
			}

			var out User
			if err := db.QueryRow(`SELECT id, name, email, created_at FROM users WHERE id=?`, id).
				Scan(&out.ID, &out.Name, &out.Email, &out.CreatedAt); err != nil {
				log.Printf("fetch after update error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch after update failed"})
				return
			}
			c.JSON(http.StatusOK, out)
		})

		api.DELETE("/users/:id", func(c *gin.Context) {
			id := c.Param("id")
			res, err := db.Exec(`DELETE FROM users WHERE id=?`, id)
			if err != nil {
				log.Printf("delete user error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
				return
			}
			affected, _ := res.RowsAffected()
			if affected == 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.Status(http.StatusNoContent)
		})
	}

	addr := ":" + getenv("PORT", "8080")
	log.Println("🚀 API listening on", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func ensureSchema(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS users (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`)
	return err
}

func cors(origin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Duplicate-key detection
func isDuplicate(err error) bool {
	var me *mysqlDrv.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate entry")
}
