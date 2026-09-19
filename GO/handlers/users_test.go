package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/nmontes/barcvvr-go/handlers"
	"github.com/nmontes/barcvvr-go/models"
)

func TestUsersIndex(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	// Insert test data
	db.Create(&models.User{FirstName: "John", LastName: "Doe", Alias: "JD"})

	usersHandler := handlers.NewUsersHandler(db)
	r.GET("/users", usersHandler.Index)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", w.Code)
	}
	if !strings.Contains(w.Body.String(), "John") {
		t.Errorf("Expected response body to contain 'John'")
	}
}

func TestUsersCreate(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	usersHandler := handlers.NewUsersHandler(db)
	r.POST("/users", usersHandler.Create)

	form := url.Values{}
	form.Add("user[firstName]", "Alice")
	form.Add("user[lastName]", "Smith")
	form.Add("user[alias]", "Al")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %v", w.Code)
	}

	var user models.User
	db.First(&user, "first_name = ?", "Alice")
	if user.LastName != "Smith" {
		t.Errorf("Expected to find user 'Alice Smith' in DB")
	}
}

func TestUsersShow(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	user := models.User{FirstName: "Show", LastName: "User", Alias: "SU"}
	db.Create(&user)

	usersHandler := handlers.NewUsersHandler(db)
	r.GET("/users/:id", usersHandler.Show)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/users/%d", user.ID), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Show") {
		t.Errorf("Expected response body to contain 'Show'")
	}
}

func TestUsersLost(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	user := models.User{FirstName: "Lost", LastName: "User", Alias: "LU"}
	db.Create(&user)

	usersHandler := handlers.NewUsersHandler(db)
	r.GET("/users/:id/lost", usersHandler.Lost)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/users/%d/lost", user.ID), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %v", w.Code)
	}
}

func TestUsersNew(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	usersHandler := handlers.NewUsersHandler(db)
	r.GET("/users/new", usersHandler.New)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/new", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", w.Code)
	}
}

func TestUsersEdit(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	user := models.User{FirstName: "Edit", LastName: "User", Alias: "EU"}
	db.Create(&user)

	usersHandler := handlers.NewUsersHandler(db)
	r.GET("/users/:id/edit", usersHandler.Edit)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/users/%d/edit", user.ID), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Edit") {
		t.Errorf("Expected response body to contain 'Edit'")
	}
}

func TestUsersUpdate(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	userToUpdate := models.User{FirstName: "Old", LastName: "User", Alias: "OU"}
	db.Create(&userToUpdate)

	usersHandler := handlers.NewUsersHandler(db)
	r.POST("/users/:id", usersHandler.Update)

	form := url.Values{}
	form.Add("user[firstName]", "NewName")
	form.Add("user[lastName]", "User")
	form.Add("admin_password", "") // blank for tests

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/users/%d", userToUpdate.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %v", w.Code)
	}

	var user models.User
	db.First(&user, userToUpdate.ID)
	if user.FirstName != "NewName" {
		t.Errorf("Expected user to be updated to 'NewName', got %s", user.FirstName)
	}
}

func TestUsersUpdateDelete(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	userToDelete := models.User{FirstName: "ToBeDeleted", LastName: "User", Alias: "TD"}
	db.Create(&userToDelete)

	usersHandler := handlers.NewUsersHandler(db)
	r.POST("/users/:id", usersHandler.Update)

	form := url.Values{}
	form.Add("usr[delete]", "yes")
	form.Add("admin_password", "")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/users/%d", userToDelete.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %v", w.Code)
	}

	var count int64
	db.Model(&models.User{}).Where("id = ?", userToDelete.ID).Count(&count)
	if count != 0 {
		t.Errorf("Expected user to be deleted, count is %d", count)
	}
}

func TestUsersDestroy(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	userToDestroy := models.User{FirstName: "ToBeDestroyed", LastName: "User", Alias: "TD"}
	db.Create(&userToDestroy)

	usersHandler := handlers.NewUsersHandler(db)
	r.DELETE("/users/:id", usersHandler.Destroy)

	form := url.Values{}
	form.Add("admin_password", "")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/users/%d", userToDestroy.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %v", w.Code)
	}

	var count int64
	db.Model(&models.User{}).Where("id = ?", userToDestroy.ID).Count(&count)
	if count != 0 {
		t.Errorf("Expected user to be destroyed, count is %d", count)
	}
}
