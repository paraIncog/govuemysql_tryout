package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
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

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		getenv("DB_USER", "root"),
		getenv("DB_PASS", "dubimin"),
		getenv("DB_HOST", "127.0.0.1"),
		getenv("DB_PORT", "3306"),
		getenv("DB_NAME", "sample_db"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil { log.Fatal(err) }
	defer db.Close()
	if err := db.Ping(); err != nil { log.Fatal("DB connection failed:", err) }

	r := gin.Default()
	r.Use(cors(getenv("CORS_ORIGIN", "*")))

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

		api.GET("/users", func(c *gin.Context) {
			rows, err := db.Query("SELECT id, name, email, created_at FROM users ORDER BY id DESC")
			if err != nil { c.JSON(http.StatusInternalServerError, errJSON(err)); return }
			defer rows.Close()

			var users []User
			for rows.Next() {
				var u User
				if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
					c.JSON(http.StatusInternalServerError, errJSON(err)); return
				}
				users = append(users, u)
			}
			c.JSON(http.StatusOK, users)
		})

		api.GET("/users/:id", func(c *gin.Context) {
			id := c.Param("id")
			var u User
			err := db.QueryRow("SELECT id, name, email, created_at FROM users WHERE id = ?", id).
				Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
			if err == sql.ErrNoRows { c.JSON(http.StatusNotFound, gin.H{"error": "not found"}); return }
			if err != nil { c.JSON(http.StatusInternalServerError, errJSON(err)); return }
			c.JSON(http.StatusOK, u)
		})

		api.POST("/users", func(c *gin.Context) {
			var in User
			if err := c.ShouldBindJSON(&in); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
			}
			res, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", in.Name, in.Email)
			if err != nil { c.JSON(http.StatusBadRequest, errJSON(err)); return }
			id, _ := res.LastInsertId()
			var out User
			err = db.QueryRow("SELECT id, name, email, created_at FROM users WHERE id = ?", id).
				Scan(&out.ID, &out.Name, &out.Email, &out.CreatedAt)
			if err != nil { c.JSON(http.StatusInternalServerError, errJSON(err)); return }
			c.JSON(http.StatusCreated, out)
		})

		api.PUT("/users/:id", func(c *gin.Context) {
			id := c.Param("id")
			var in User
			if err := c.ShouldBindJSON(&in); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
			}
			_, err := db.Exec("UPDATE users SET name = ?, email = ? WHERE id = ?", in.Name, in.Email, id)
			if err != nil { c.JSON(http.StatusBadRequest, errJSON(err)); return }
			var out User
			err = db.QueryRow("SELECT id, name, email, created_at FROM users WHERE id = ?", id).
				Scan(&out.ID, &out.Name, &out.Email, &out.CreatedAt)
			if err != nil { c.JSON(http.StatusInternalServerError, errJSON(err)); return }
			c.JSON(http.StatusOK, out)
		})

		api.DELETE("/users/:id", func(c *gin.Context) {
			id := c.Param("id")
			res, err := db.Exec("DELETE FROM users WHERE id = ?", id)
			if err != nil { c.JSON(http.StatusBadRequest, errJSON(err)); return }
			affected, _ := res.RowsAffected()
			if affected == 0 { c.JSON(http.StatusNotFound, gin.H{"error": "not found"}); return }
			c.Status(http.StatusNoContent)
		})
	}

	addr := ":" + getenv("PORT", "8080")
	log.Println("API listening on", addr)
	if err := r.Run(addr); err != nil { log.Fatal(err) }
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}

func errJSON(err error) gin.H { return gin.H{"error": err.Error()} }

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