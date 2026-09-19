package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/nmontes/barcvvr-go/handlers"
	"github.com/nmontes/barcvvr-go/models"
)

func TestOperationsIndex(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	user := models.User{FirstName: "John"}
	db.Create(&user)
	db.Create(&models.Operation{UserID: user.ID, Sum: 10.5, Comment: "Test Op"})

	opsHandler := handlers.NewOperationsHandler(db)
	r.GET("/operations", opsHandler.Index)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/operations", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Test Op") {
		t.Errorf("Expected response body to contain 'Test Op'")
	}
}

func TestOperationsCreate(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	user := models.User{FirstName: "John", InitAmount: 50.0, Amount: 50.0}
	db.Create(&user)

	opsHandler := handlers.NewOperationsHandler(db)
	r.POST("/operations", opsHandler.Create)

	form := url.Values{}
	form.Add("operation[user_id]", fmt.Sprintf("%d", user.ID))
	form.Add("operation[sum]", "-10.0")
	form.Add("operation[date]", time.Now().Format("2006-01-02T15:04"))
	form.Add("operation[comment]", "Bought something")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/operations", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %v", w.Code)
	}

	var op models.Operation
	db.First(&op, "comment = ?", "Bought something")
	if op.Comment != "Bought something" {
		t.Errorf("Expected to find operation with comment 'Bought something' in DB")
	}

	// Verify user balance recalculated
	var updatedUser models.User
	db.First(&updatedUser, user.ID)
	if updatedUser.Amount != 40.0 { // 50 (init) + (-10) = 40
		t.Errorf("Expected user amount to be 40.0, got %v", updatedUser.Amount)
	}
}
