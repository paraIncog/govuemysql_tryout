// server.go
package main

import (
	"database/sql"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	db     *sql.DB
	router *gin.Engine
}

func NewServer(db *sql.DB) *Server {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// CORS (keeps dev & prod happy)
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))

	s := &Server{db: db, router: r}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	api := s.router.Group("/api")

	api.GET("/users", s.listUsers)
	api.GET("/users/:id", s.getUser)
	api.POST("/users", s.createUser)
	api.PUT("/users/:id", s.updateUser)
	api.DELETE("/users/:id", s.deleteUser)

	// static Vue build
	s.router.Static("/assets", "./frontend/assets")
	s.router.GET("/favicon.ico", func(c *gin.Context) { c.File("./frontend/favicon.ico") })

	// SPA fallback
	s.router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.File(filepath.Clean("./frontend/index.html"))
	})
}

func (s *Server) Router() http.Handler { return s.router }
