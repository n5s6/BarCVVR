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

func TestBeerflowsIndex(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	db.Create(&models.Beerflow{Quantity: 50, DrinkID: 1})

	beerflowsHandler := handlers.NewBeerflowsHandler(db)
	r.GET("/beerflows", beerflowsHandler.Index)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/beerflows", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", w.Code)
	}
	if !strings.Contains(w.Body.String(), "50") {
		t.Errorf("Expected response body to contain '50'")
	}
}

func TestBeerflowsCreate(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	beerflowsHandler := handlers.NewBeerflowsHandler(db)
	r.POST("/beerflows", beerflowsHandler.Create)

	form := url.Values{}
	form.Add("beerflow[quantity]", "100.0")
	form.Add("beerflow[drink_id]", "1")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/beerflows", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %v", w.Code)
	}

	var bf models.Beerflow
	db.First(&bf, "quantity = ?", 100.0)
	if bf.Quantity != 100.0 {
		t.Errorf("Expected to find beerflow with quantity 100.0 in DB")
	}
}
