// handlers.go
package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (s *Server) listUsers(c *gin.Context) {
	rows, err := s.db.Query(`SELECT id, name, email FROM users ORDER BY id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, u)
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) getUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error":"invalid id"}); return }
	var u User
	err = s.db.QueryRow(`SELECT id,name,email FROM users WHERE id=?`, id).Scan(&u.ID,&u.Name,&u.Email)
	if err == sql.ErrNoRows { c.JSON(http.StatusNotFound, gin.H{"error":"user not found"}); return }
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, u)
}

func (s *Server) createUser(c *gin.Context) {
	var in struct{ Name, Email string }
	if err := c.ShouldBindJSON(&in); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	res, err := s.db.Exec(`INSERT INTO users(name,email) VALUES(?,?)`, in.Name, in.Email)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (s *Server) updateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error":"invalid id"}); return }
	var in struct{ Name, Email string }
	if err := c.ShouldBindJSON(&in); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	res, err := s.db.Exec(`UPDATE users SET name=?, email=? WHERE id=?`, in.Name, in.Email, id)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	aff, _ := res.RowsAffected()
	if aff == 0 { c.JSON(http.StatusNotFound, gin.H{"error":"user not found"}); return }
	c.Status(http.StatusNoContent)
}

func (s *Server) deleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error":"invalid id"}); return }
	res, err := s.db.Exec(`DELETE FROM users WHERE id=?`, id)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	aff, _ := res.RowsAffected()
	if aff == 0 { c.JSON(http.StatusNotFound, gin.H{"error":"user not found"}); return }
	c.Status(http.StatusNoContent)
}
