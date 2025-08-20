package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id","name","email"}).
		AddRow(1,"Alice","alice@example.com").
		AddRow(2,"Bob","bob@example.com")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email FROM users ORDER BY id`)).
		WillReturnRows(rows)

	s := NewServer(db)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	s.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	var got []User
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil { t.Fatal(err) }
	if len(got) != 2 || got[0].Name != "Alice" || got[1].Email != "bob@example.com" {
		t.Fatalf("unexpected body: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}
