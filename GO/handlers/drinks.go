package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/nmontes/barcvvr-go/models"
	"gorm.io/gorm"
)

type DrinksHandler struct {
	DB *gorm.DB
}

func NewDrinksHandler(db *gorm.DB) *DrinksHandler {
	return &DrinksHandler{DB: db}
}

// Index GET /drinks
func (h *DrinksHandler) Index(c *gin.Context) {
	var drinks []models.Drink
	if err := h.DB.Find(&drinks).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "drinks/index.html", gin.H{
		"drinks": drinks,
	})
}

// Show GET /drinks/:id
func (h *DrinksHandler) Show(c *gin.Context) {
	var drink models.Drink
	if err := h.DB.First(&drink, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Drink not found"})
		return
	}
	c.HTML(http.StatusOK, "drinks/show.html", gin.H{
		"drink": drink,
	})
}

// New GET /drinks/new
func (h *DrinksHandler) New(c *gin.Context) {
	c.HTML(http.StatusOK, "drinks/new.html", gin.H{
		"drink": models.Drink{},
	})
}

// Edit GET /drinks/:id/edit
func (h *DrinksHandler) Edit(c *gin.Context) {
	var drink models.Drink
	if err := h.DB.First(&drink, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Drink not found"})
		return
	}
	c.HTML(http.StatusOK, "drinks/edit.html", gin.H{
		"drink": drink,
	})
}

// Create POST /drinks
func (h *DrinksHandler) Create(c *gin.Context) {
	adminPassword := c.PostForm("admin_password")
	if adminPassword != os.Getenv("ADMIN_PASSWORD") {
		c.Redirect(http.StatusFound, c.Request.Referer())
		return
	}

	drink := models.Drink{
		Name:      c.PostForm("drink[name]"),
		DrinkType: c.PostForm("post[drink_type]"), // in Rails, drink_type came from params[:post][:drink_type]
	}

	// parse price
	fmt.Sscanf(c.PostForm("drink[price]"), "%f", &drink.Price)

	if err := h.DB.Create(&drink).Error; err != nil {
		c.HTML(http.StatusUnprocessableEntity, "drinks/new.html", gin.H{
			"drink": drink,
			"error": err.Error(),
		})
		return
	}

	// redirect to drinks index on success
	c.Redirect(http.StatusFound, "/drinks")
}

// Update PATCH/PUT /drinks/:id
func (h *DrinksHandler) Update(c *gin.Context) {
	adminPassword := c.PostForm("admin_password")
	if adminPassword != os.Getenv("ADMIN_PASSWORD") {
		c.Redirect(http.StatusFound, c.Request.Referer())
		return
	}

	var drink models.Drink
	if err := h.DB.First(&drink, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Drink not found"})
		return
	}

	// Check if intended to delete directly from update (Rails logic: params[:drk][:delete] == "yes")
	if c.PostForm("drk[delete]") == "yes" {
		h.DB.Delete(&drink)
		c.Redirect(http.StatusFound, "/drinks")
		return
	}

	drink.Name = c.PostForm("drink[name]")
	drink.DrinkType = c.PostForm("post[drink_type]")
	fmt.Sscanf(c.PostForm("drink[price]"), "%f", &drink.Price)

	if err := h.DB.Save(&drink).Error; err != nil {
		c.HTML(http.StatusUnprocessableEntity, "drinks/edit.html", gin.H{
			"drink": drink,
			"error": err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/drinks")
}

// Destroy DELETE /drinks/:id (Usually rails uses a hidden _method=delete, so in Gin it will route as DELETE or we handle from POST)
func (h *DrinksHandler) Destroy(c *gin.Context) {
	adminPassword := c.PostForm("admin_password")
	if adminPassword != os.Getenv("ADMIN_PASSWORD") {
		c.Redirect(http.StatusFound, c.Request.Referer())
		return
	}

	var drink models.Drink
	if err := h.DB.First(&drink, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Drink not found"})
		return
	}

	h.DB.Delete(&drink)
	c.Redirect(http.StatusFound, "/drinks")
}
