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

func TestDrinksIndex(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	db.Create(&models.Drink{Name: "Cola", DrinkType: "Soda", Price: 2.50})

	drinkHandler := handlers.NewDrinksHandler(db)
	r.GET("/drinks", drinkHandler.Index)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/drinks", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Cola") {
		t.Errorf("Expected response body to contain 'Cola'")
	}
}

func TestDrinksCreate(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	drinkHandler := handlers.NewDrinksHandler(db)
	r.POST("/drinks", drinkHandler.Create)

	form := url.Values{}
	form.Add("drink[name]", "Orange Juice")
	form.Add("drink[price]", "3.0")
	form.Add("drink[type]", "0")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/drinks", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %v", w.Code)
	}

	var drink models.Drink
	db.First(&drink, "name = ?", "Orange Juice")
	if drink.Price != 3.0 {
		t.Errorf("Expected to find drink 'Orange Juice' with price 3.0 in DB")
	}
}

func TestDrinksShow(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	drink := models.Drink{Name: "Lemonade", DrinkType: "Soda", Price: 2.00}
	db.Create(&drink)

	drinkHandler := handlers.NewDrinksHandler(db)
	r.GET("/drinks/:id", drinkHandler.Show)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/drinks/%d", drink.ID), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Lemonade") {
		t.Errorf("Expected response body to contain 'Lemonade'")
	}
}

func TestDrinksNew(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	drinkHandler := handlers.NewDrinksHandler(db)
	r.GET("/drinks/new", drinkHandler.New)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/drinks/new", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", w.Code)
	}
}

func TestDrinksEdit(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	drink := models.Drink{Name: "Water", DrinkType: "Non-Alcoholic", Price: 1.00}
	db.Create(&drink)

	drinkHandler := handlers.NewDrinksHandler(db)
	r.GET("/drinks/:id/edit", drinkHandler.Edit)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/drinks/%d/edit", drink.ID), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Water") {
		t.Errorf("Expected response body to contain 'Water'")
	}
}

func TestDrinksUpdate(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	drinkToUpdate := models.Drink{Name: "Old Water", DrinkType: "Water", Price: 1.00}
	db.Create(&drinkToUpdate)

	drinkHandler := handlers.NewDrinksHandler(db)
	r.POST("/drinks/:id", drinkHandler.Update)

	form := url.Values{}
	form.Add("drink[name]", "New Water")
	form.Add("drink[price]", "1.50")
	form.Add("post[drink_type]", "Water")
	form.Add("admin_password", "") // os.Getenv("ADMIN_PASSWORD") is usually blank in tests

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/drinks/%d", drinkToUpdate.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %v", w.Code)
	}

	var drink models.Drink
	db.First(&drink, drinkToUpdate.ID)
	if drink.Name != "New Water" || drink.Price != 1.50 {
		t.Errorf("Expected drink to be updated to 'New Water' with price 1.50, got name: %s, price: %f", drink.Name, drink.Price)
	}
}

func TestDrinksUpdateDelete(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	drinkToDelete := models.Drink{Name: "To Be Deleted", DrinkType: "Water", Price: 1.00}
	db.Create(&drinkToDelete)

	drinkHandler := handlers.NewDrinksHandler(db)
	r.POST("/drinks/:id", drinkHandler.Update)

	form := url.Values{}
	form.Add("drk[delete]", "yes")
	form.Add("admin_password", "")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/drinks/%d", drinkToDelete.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %v", w.Code)
	}

	var count int64
	db.Model(&models.Drink{}).Where("id = ?", drinkToDelete.ID).Count(&count)
	if count != 0 {
		t.Errorf("Expected drink to be deleted, count is %d", count)
	}
}

func TestDrinksDestroy(t *testing.T) {
	db := SetupTestDB()
	r := SetupTestRouter()

	drinkToDestroy := models.Drink{Name: "To Be Destroyed", DrinkType: "Water", Price: 1.00}
	db.Create(&drinkToDestroy)

	drinkHandler := handlers.NewDrinksHandler(db)
	r.DELETE("/drinks/:id", drinkHandler.Destroy)

	form := url.Values{}
	form.Add("admin_password", "")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/drinks/%d", drinkToDestroy.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %v", w.Code)
	}

	var count int64
	db.Model(&models.Drink{}).Where("id = ?", drinkToDestroy.ID).Count(&count)
	if count != 0 {
		t.Errorf("Expected drink to be destroyed, count is %d", count)
	}
}

