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

type OperationsHandler struct {
	DB *gorm.DB
}

func NewOperationsHandler(db *gorm.DB) *OperationsHandler {
	return &OperationsHandler{DB: db}
}

// Index GET /operations
func (h *OperationsHandler) Index(c *gin.Context) {
	var operations []models.Operation
	// Pagination is simplified here, can be extended if needed.
	// Rails code: Operation.all.order("date DESC").paginate(page: params[:page], :per_page => 10)
	h.DB.Order("date DESC").Limit(50).Find(&operations)

	var opItems []map[string]interface{}
	for _, op := range operations {
		var drinkName string
		if op.DrinkID != 0 {
			var drink models.Drink
			h.DB.First(&drink, op.DrinkID)
			drinkName = drink.Name
		}
		var user models.User
		h.DB.First(&user, op.UserID)

		opItems = append(opItems, map[string]interface{}{
			"Operation":     op,
			"DrinkName":     drinkName,
			"UserFirstName": user.FirstName,
			"UserLastName":  user.LastName,
			"UserAlias":     user.Alias,
		})
	}

	c.HTML(http.StatusOK, "operations/index.html", gin.H{
		"opItems": opItems,
	})
}

// Show GET /operations/:id
func (h *OperationsHandler) Show(c *gin.Context) {
	var operation models.Operation
	if err := h.DB.First(&operation, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Operation not found"})
		return
	}
	c.HTML(http.StatusOK, "operations/show.html", gin.H{
		"operation": operation,
	})
}

// New GET /operations/new
func (h *OperationsHandler) New(c *gin.Context) {
	userID := c.Query("userid")

	var user models.User
	h.DB.First(&user, userID)

	var kegs []models.Keg
	h.DB.Where("(\"endDate\" IS NULL OR \"endDate\" < '2000-01-01')").Find(&kegs)
	var kegItems []map[string]interface{}
	for _, keg := range kegs {
		var drink models.Drink
		h.DB.First(&drink, keg.DrinkID)
		kegItems = append(kegItems, map[string]interface{}{
			"DrinkID":   keg.DrinkID,
			"DrinkName": drink.Name,
		})
	}

	c.HTML(http.StatusOK, "operations/new.html", gin.H{
		"operation": models.Operation{},
		"userid":    userID,
		"user":      user,
		"kegs":      kegItems,
	})
}

// Add GET /operations/add
func (h *OperationsHandler) Add(c *gin.Context) {
	userID := c.Query("userid")
	var user models.User
	h.DB.First(&user, userID)

	// Preload drinks for the view to build the select box
	var drinks []models.Drink
	h.DB.Find(&drinks)

	c.HTML(http.StatusOK, "operations/add.html", gin.H{
		"operation": models.Operation{},
		"userid":    userID,
		"user":      user,
		"drinks":    drinks,
	})
}

// Exchange GET /operations/exchange
func (h *OperationsHandler) Exchange(c *gin.Context) {
	userID := c.Query("userid")
	var user models.User
	h.DB.First(&user, userID)

	var users []models.User
	h.DB.Order("\"lastName\" asc").Find(&users)

	c.HTML(http.StatusOK, "operations/exchange.html", gin.H{
		"operation": models.Operation{},
		"userid":    userID,
		"user":      user,
		"users":     users,
	})
}

// Edit GET /operations/:id/edit
func (h *OperationsHandler) Edit(c *gin.Context) {
	var operation models.Operation
	if err := h.DB.First(&operation, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Operation not found"})
		return
	}

	var user models.User
	h.DB.First(&user, operation.UserID)

	c.HTML(http.StatusOK, "operations/edit.html", gin.H{
		"operation": operation,
		"user":      user,
	})
}

func (h *OperationsHandler) recalculateUserAmount(userID uint) {
	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		return
	}

	var operations []models.Operation
	h.DB.Where("user_id = ?", userID).Find(&operations)

	amount := user.InitAmount
	for _, op := range operations {
		amount += op.Sum
	}

	user.Amount = amount
	h.DB.Save(&user)
}

// Create POST /operations
func (h *OperationsHandler) Create(c *gin.Context) {
	postExchangeUser := c.PostForm("post_exchange_user")

	if postExchangeUser != "" {
		// Exchange logic
		var sum float64
		fmt.Sscanf(c.PostForm("operation[sum]"), "%f", &sum)

		var userID1 uint
		fmt.Sscanf(c.PostForm("operation[user_id]"), "%d", &userID1)

		var userID2 uint
		fmt.Sscanf(postExchangeUser, "%d", &userID2)

		op1 := models.Operation{
			Date:    time.Now(), // Assuming date is passed or filled now
			Sum:     -sum,
			UserID:  userID1,
			Comment: c.PostForm("operation[comment]"),
		}

		layout := "2006-01-02T15:04"
		if dateStr := c.PostForm("operation[date]"); dateStr != "" {
			if t, err := time.Parse(layout, dateStr); err == nil {
				op1.Date = t
			}
		}

		op2 := models.Operation{
			Date:    op1.Date,
			Sum:     sum,
			UserID:  userID2,
			Comment: op1.Comment,
		}

		h.DB.Create(&op1)
		h.DB.Create(&op2)

		h.recalculateUserAmount(userID1)
		h.recalculateUserAmount(userID2)

		c.Redirect(http.StatusFound, "/users")
		return

	} else {
		// Regular operation logic
		var op models.Operation
		op.Comment = c.PostForm("operation[comment]")

		layout := "2006-01-02T15:04"
		if dateStr := c.PostForm("operation[date]"); dateStr != "" {
			if t, err := time.Parse(layout, dateStr); err == nil {
				op.Date = t
			}
		} else {
			op.Date = time.Now()
		}

		fmt.Sscanf(c.PostForm("operation[user_id]"), "%d", &op.UserID)

		if postDrink := c.PostForm("post[drink]"); postDrink != "" {
			var drink models.Drink
			h.DB.First(&drink, postDrink)
			fmt.Sscanf(c.PostForm("operation[numberDrink]"), "%f", &op.NumberDrink)
			op.DrinkID = float64(drink.ID)
			op.Sum = -op.NumberDrink * drink.Price
		} else {
			fmt.Sscanf(c.PostForm("operation[sum]"), "%f", &op.Sum)
		}

		if err := h.DB.Create(&op).Error; err != nil {
			c.HTML(http.StatusUnprocessableEntity, "operations/new.html", gin.H{
				"error": err.Error(),
			})
			return
		}

		h.recalculateUserAmount(op.UserID)

		c.Redirect(http.StatusFound, "/users")
	}
}

// Update PATCH/PUT /operations/:id
func (h *OperationsHandler) Update(c *gin.Context) {
	adminPassword := c.PostForm("admin_password")
	if adminPassword != os.Getenv("ADMIN_PASSWORD") {
		c.Redirect(http.StatusFound, c.Request.Referer())
		return
	}

	var op models.Operation
	if err := h.DB.First(&op, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Operation not found"})
		return
	}

	userID := op.UserID

	if c.PostForm("op[delete]") == "yes" {
		h.DB.Delete(&op)
		h.recalculateUserAmount(userID)
		c.Redirect(http.StatusFound, "/operations")
		return
	}

	layout := "2006-01-02T15:04"
	if dateStr := c.PostForm("operation[date]"); dateStr != "" {
		if t, err := time.Parse(layout, dateStr); err == nil {
			op.Date = t
		}
	}
	fmt.Sscanf(c.PostForm("operation[sum]"), "%f", &op.Sum)
	fmt.Sscanf(c.PostForm("operation[user_id]"), "%d", &op.UserID)
	fmt.Sscanf(c.PostForm("operation[numberDrink]"), "%f", &op.NumberDrink)
	op.Comment = c.PostForm("operation[comment]")

	if err := h.DB.Save(&op).Error; err != nil {
		c.HTML(http.StatusUnprocessableEntity, "operations/edit.html", gin.H{
			"error": err.Error(),
		})
		return
	}

	// Recalculate for both old user and new user in case user_id changed
	h.recalculateUserAmount(userID)
	if op.UserID != userID {
		h.recalculateUserAmount(op.UserID)
	}

	c.Redirect(http.StatusFound, "/operations")
}

// Destroy DELETE /operations/:id
func (h *OperationsHandler) Destroy(c *gin.Context) {
	adminPassword := c.PostForm("admin_password")
	if adminPassword != os.Getenv("ADMIN_PASSWORD") {
		c.Redirect(http.StatusFound, c.Request.Referer())
		return
	}

	var op models.Operation
	if err := h.DB.First(&op, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Operation not found"})
		return
	}

	userID := op.UserID
	h.DB.Delete(&op)
	h.recalculateUserAmount(userID)

	c.Redirect(http.StatusFound, "/operations")
}
