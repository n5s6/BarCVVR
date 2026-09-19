package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/nmontes/barcvvr-go/handlers"
	"github.com/nmontes/barcvvr-go/models"
)

func TestKegsIndex(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	drink := models.Drink{Name: "Soda"}
	db.Create(&drink)
	db.Create(&models.Keg{DrinkID: float64(drink.ID), Capacity: 30})

	kegsHandler := handlers.NewKegsHandler(db)
	r.GET("/kegs", kegsHandler.Index)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/kegs", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", w.Code)
	}
	if !strings.Contains(w.Body.String(), "1") {
		t.Errorf("Expected response body to contain '1'")
	}
}

func TestKegsCreate(t *testing.T) {
	t.Setenv("ADMIN_PASSWORD", "secret")
	db := SetupTestDB()
	r := SetupTestRouter()

	drink := models.Drink{Name: "Soda"}
	db.Create(&drink)

	kegsHandler := handlers.NewKegsHandler(db)
	r.POST("/kegs", kegsHandler.Create)

	form := url.Values{}
	form.Add("admin_password", "secret")
	form.Add("post[drink]", "1")
	form.Add("keg[capacity]", "20.5")
	form.Add("keg[price]", "50.0")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/kegs", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %v", w.Code)
	}

	var keg models.Keg
	db.First(&keg, "capacity = ?", 20.5)
	if keg.Capacity != 20.5 {
		t.Errorf("Expected to find keg with capacity 20.5 in DB")
	}
}
