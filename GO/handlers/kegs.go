package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nmontes/barcvvr-go/models"
	"gorm.io/gorm"
)

type KegsHandler struct {
	DB *gorm.DB
}

func NewKegsHandler(db *gorm.DB) *KegsHandler {
	return &KegsHandler{DB: db}
}

// Index GET /kegs
func (h *KegsHandler) Index(c *gin.Context) {
	var kegs []models.Keg
	h.DB.Order("created_at DESC").Limit(50).Find(&kegs) // simplification for pagination

	type KegItem struct {
		Keg       models.Keg
		DrinkName string
		Profit    float64
	}
	var kegItems []KegItem

	for _, keg := range kegs {
		prof := -keg.Price

		var operations []models.Operation
		query := h.DB.Where("\"drink_id\" = ? AND \"date\" >= ?", int(keg.DrinkID), keg.StartDate)
		if keg.EndDate.Year() > 2000 {
			query = query.Where("\"date\" <= ?", keg.EndDate)
		}
		query.Find(&operations)

		for _, op := range operations {
			prof -= op.Sum
		}
		
		var drink models.Drink
		h.DB.First(&drink, int(keg.DrinkID))

		kegItems = append(kegItems, KegItem{
			Keg:       keg,
			DrinkName: drink.Name,
			Profit:    prof,
		})
	}

	c.HTML(http.StatusOK, "kegs/index.html", gin.H{
		"kegItems": kegItems,
	})
}

// Show GET /kegs/:id
func (h *KegsHandler) Show(c *gin.Context) {
	var keg models.Keg
	if err := h.DB.First(&keg, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Keg not found"})
		return
	}

	opNum := 0.0
	profit := -keg.Price

	var operations []models.Operation
	query := h.DB.Where("\"drink_id\" = ? AND \"date\" >= ?", int(keg.DrinkID), keg.StartDate)
	if keg.EndDate.Year() > 2000 {
		query = query.Where("\"date\" <= ?", keg.EndDate)
	}
	query.Find(&operations)

	for _, op := range operations {
		opNum += op.NumberDrink
		profit -= op.Sum
	}

	c.HTML(http.StatusOK, "kegs/show.html", gin.H{
		"keg":    keg,
		"opNum":  opNum,
		"profit": profit,
	})
}

// New GET /kegs/new
func (h *KegsHandler) New(c *gin.Context) {
	var drinks []models.Drink
	h.DB.Find(&drinks)

	c.HTML(http.StatusOK, "kegs/new.html", gin.H{
		"keg":    models.Keg{},
		"drinks": drinks,
	})
}

// Renew GET /kegs/:id/renew
func (h *KegsHandler) Renew(c *gin.Context) {
	var keg models.Keg
	if err := h.DB.First(&keg, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Keg not found"})
		return
	}

	now := time.Now()

	// Update old keg end date
	h.DB.Model(&keg).UpdateColumn("endDate", now)

	// Create new keg with same details, but a new sequential name
	var count int64
	h.DB.Model(&models.Keg{}).Count(&count)
	newName := fmt.Sprintf("Fût %d", count+1)

	newKeg := models.Keg{
		Name:      newName,
		DrinkID:   keg.DrinkID,
		StartDate: now,
		Capacity:  keg.Capacity,
		Price:     keg.Price,
	}
	h.DB.Create(&newKeg)

	c.Redirect(http.StatusFound, "/users")
}

// Edit GET /kegs/:id/edit
func (h *KegsHandler) Edit(c *gin.Context) {
	var keg models.Keg
	if err := h.DB.First(&keg, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Keg not found"})
		return
	}

	var drinks []models.Drink
	h.DB.Find(&drinks)

	c.HTML(http.StatusOK, "kegs/edit.html", gin.H{
		"keg":        keg,
		"kegDrinkID": uint(keg.DrinkID),
		"drinks":     drinks,
	})
}

// Create POST /kegs
func (h *KegsHandler) Create(c *gin.Context) {
	adminPassword := c.PostForm("admin_password")
	if adminPassword != os.Getenv("ADMIN_PASSWORD") {
		c.Redirect(http.StatusFound, c.Request.Referer())
		return
	}

	keg := models.Keg{
		Name: c.PostForm("keg[name]"),
	}

	layout := "2006-01-02T15:04"
	if dateStr := c.PostForm("keg[startDate]"); dateStr != "" {
		if t, err := time.ParseInLocation(layout, dateStr, time.Local); err == nil {
			keg.StartDate = t
		}
	}
	if dateStr := c.PostForm("keg[endDate]"); dateStr != "" {
		if t, err := time.ParseInLocation(layout, dateStr, time.Local); err == nil {
			keg.EndDate = t
		}
	}

	fmt.Sscanf(c.PostForm("post[drink]"), "%f", &keg.DrinkID)
	fmt.Sscanf(c.PostForm("keg[capacity]"), "%f", &keg.Capacity)
	fmt.Sscanf(c.PostForm("keg[price]"), "%f", &keg.Price)

	if err := h.DB.Create(&keg).Error; err != nil {
		c.HTML(http.StatusUnprocessableEntity, "kegs/new.html", gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/kegs")
}

// Update PATCH/PUT /kegs/:id
func (h *KegsHandler) Update(c *gin.Context) {
	adminPassword := c.PostForm("admin_password")
	if adminPassword != os.Getenv("ADMIN_PASSWORD") {
		c.Redirect(http.StatusFound, c.Request.Referer())
		return
	}

	var keg models.Keg
	if err := h.DB.First(&keg, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Keg not found"})
		return
	}

	if c.PostForm("kg[delete]") == "yes" {
		h.DB.Delete(&keg)
		c.Redirect(http.StatusFound, "/kegs")
		return
	}

	keg.Name = c.PostForm("keg[name]")

	layout := "2006-01-02T15:04"
	if dateStr := c.PostForm("keg[startDate]"); dateStr != "" {
		if t, err := time.Parse(layout, dateStr); err == nil {
			keg.StartDate = t
		}
	}
	if dateStr := c.PostForm("keg[endDate]"); dateStr != "" {
		if t, err := time.Parse(layout, dateStr); err == nil {
			keg.EndDate = t
		}
	}

	fmt.Sscanf(c.PostForm("keg[drink_id]"), "%f", &keg.DrinkID)
	fmt.Sscanf(c.PostForm("keg[capacity]"), "%f", &keg.Capacity)
	fmt.Sscanf(c.PostForm("keg[price]"), "%f", &keg.Price)

	if err := h.DB.Save(&keg).Error; err != nil {
		c.HTML(http.StatusUnprocessableEntity, "kegs/edit.html", gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/kegs")
}

// Destroy DELETE /kegs/:id
func (h *KegsHandler) Destroy(c *gin.Context) {
	adminPassword := c.PostForm("admin_password")
	if adminPassword != os.Getenv("ADMIN_PASSWORD") {
		c.Redirect(http.StatusFound, c.Request.Referer())
		return
	}

	var keg models.Keg
	if err := h.DB.First(&keg, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Keg not found"})
		return
	}

	h.DB.Delete(&keg)
	c.Redirect(http.StatusFound, "/kegs")
}
