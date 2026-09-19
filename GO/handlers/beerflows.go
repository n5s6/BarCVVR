package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nmontes/barcvvr-go/models"
	"gorm.io/gorm"
)

type BeerflowsHandler struct {
	DB *gorm.DB
}

func NewBeerflowsHandler(db *gorm.DB) *BeerflowsHandler {
	return &BeerflowsHandler{DB: db}
}

// Index GET /beerflows
func (h *BeerflowsHandler) Index(c *gin.Context) {
	var beerflows []models.Beerflow
	h.DB.Find(&beerflows)

	c.HTML(http.StatusOK, "beerflows/index.html", gin.H{
		"beerflows": beerflows,
	})
}

// Show GET /beerflows/:id
func (h *BeerflowsHandler) Show(c *gin.Context) {
	var beerflow models.Beerflow
	if err := h.DB.First(&beerflow, c.Param("id")).Error; err != nil {
        if c.GetHeader("Accept") != "" && strings.Contains(c.GetHeader("Accept"), "application/json") {
            c.JSON(http.StatusNotFound, gin.H{"error": "Beerflow not found"})
        } else {
		    c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Beerflow not found"})
        }
		return
	}

    if c.GetHeader("Accept") != "" && strings.Contains(c.GetHeader("Accept"), "application/json") {
        c.JSON(http.StatusOK, beerflow)
        return
    }

	c.HTML(http.StatusOK, "beerflows/show.html", gin.H{
		"beerflow": beerflow,
	})
}

// New GET /beerflows/new
func (h *BeerflowsHandler) New(c *gin.Context) {
	c.HTML(http.StatusOK, "beerflows/new.html", gin.H{
		"beerflow": models.Beerflow{},
	})
}

// Edit GET /beerflows/:id/edit
func (h *BeerflowsHandler) Edit(c *gin.Context) {
	var beerflow models.Beerflow
	if err := h.DB.First(&beerflow, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Beerflow not found"})
		return
	}
	c.HTML(http.StatusOK, "beerflows/edit.html", gin.H{
		"beerflow": beerflow,
	})
}

// Create POST /beerflows
func (h *BeerflowsHandler) Create(c *gin.Context) {
	var beerflow models.Beerflow

	// Either process form submission or JSON (Rails had respond_to json)
	// For API simplicity if it receives JSON we can bind it, else parse forms.
	if c.ContentType() == "application/json" {
		if err := c.ShouldBindJSON(&beerflow); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		qtyStr := c.PostForm("beerflow[quantity]")
		if qtyStr == "" {
			qtyStr = c.PostForm("quantity")
		}
		fmt.Sscanf(qtyStr, "%f", &beerflow.Quantity)

		drinkIDStr := c.PostForm("beerflow[drink_id]")
		if drinkIDStr == "" {
			drinkIDStr = c.PostForm("drink_id")
		}
		var drinkID uint
		fmt.Sscanf(drinkIDStr, "%d", &drinkID)
		beerflow.DrinkID = drinkID
	}

	if err := h.DB.Create(&beerflow).Error; err != nil {
		c.HTML(http.StatusUnprocessableEntity, "beerflows/new.html", gin.H{
			"error": err.Error(),
		})
		return
	}

	// Rails responds with redirect to @beerflow
	c.Redirect(http.StatusFound, fmt.Sprintf("/beerflows/%d", beerflow.ID))
}

// Update PATCH/PUT /beerflows/:id
func (h *BeerflowsHandler) Update(c *gin.Context) {
	var beerflow models.Beerflow
	if err := h.DB.First(&beerflow, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Beerflow not found"})
		return
	}

	var newQuantity float64
	fmt.Sscanf(c.PostForm("beerflow[quantity]"), "%f", &newQuantity)

	if newQuantity != 0 {
		beerflow.Quantity += newQuantity
	} else {
		beerflow.Quantity = 0
	}

	if err := h.DB.Save(&beerflow).Error; err != nil {
		c.HTML(http.StatusUnprocessableEntity, "beerflows/edit.html", gin.H{
			"error": err.Error(),
		})
		return
	}

	// The rails app had the redirect commented out for update, we'll keep it simple
	c.JSON(http.StatusOK, beerflow)
}

// Destroy DELETE /beerflows/:id
func (h *BeerflowsHandler) Destroy(c *gin.Context) {
	var beerflow models.Beerflow
	if err := h.DB.First(&beerflow, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Beerflow not found"})
		return
	}

	h.DB.Delete(&beerflow)
	c.Redirect(http.StatusFound, "/beerflows")
}
