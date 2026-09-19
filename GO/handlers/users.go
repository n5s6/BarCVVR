package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nmontes/barcvvr-go/models"
	"gorm.io/gorm"
)

type UsersHandler struct {
	DB *gorm.DB
}

func NewUsersHandler(db *gorm.DB) *UsersHandler {
	return &UsersHandler{DB: db}
}

// Index GET /users
func (h *UsersHandler) Index(c *gin.Context) {
	var kegs []models.Keg
	h.DB.Where("(\"endDate\" IS NULL OR \"endDate\" < '2000-01-01')").Find(&kegs) // Fixed missing quoted column

	var users []models.User
	h.DB.Order("\"lastName\" asc").Find(&users)

	totalAmount := 0.0
	operationLastMonth := make(map[uint]float64)

	lastMonth := time.Now().AddDate(0, -1, 0)

	for _, user := range users {
		totalAmount += user.Amount

		var operations []models.Operation
		h.DB.Where("user_id = ? AND \"numberDrink\" IS NOT NULL AND created_at > ?", user.ID, lastMonth).Find(&operations)

		totalOpLastMonth := 0.0
		for _, op := range operations {
			if op.Sum < 0 {
				totalOpLastMonth += op.Sum
			}
		}
		operationLastMonth[user.ID] = totalOpLastMonth
	}

	// View Models
	type UserItem struct {
		User        models.User
		Class       string
		WarningHTML template.HTML // Using template.HTML prevents Gin from escaping the icon
	}
	var userItems []UserItem

	for _, user := range users {
		item := UserItem{User: user}
		if user.Amount <= -50 {
			item.Class = "red-text"
			item.WarningHTML = template.HTML(`<a class="btn btn-floating pulse red"><i class="material-icons">warning</i></a>`)
		} else if user.Amount <= -40 && user.Amount > -50 {
			item.Class = "orange-text text-darken-4"
		} else if user.Amount <= -30 && user.Amount > -40 {
			item.Class = "orange-text text-darken-3"
		} else if user.Amount <= -20 && user.Amount > -30 {
			item.Class = "orange-text text-darken-2"
		} else if user.Amount <= -10 && user.Amount > -20 {
			item.Class = "orange-text text-darken-1"
		} else if user.Amount <= 0 && user.Amount > -10 {
			item.Class = "orange-text"
		} else if user.Amount >= 50 {
			item.Class = "green-text"
		}
		userItems = append(userItems, item)
	}

	// In Go, mapping sorted by values requires creating a struct slice to hold the kv pairs
	type KV struct {
		Key   uint
		Value float64
	}
	var sortedOps []KV
	for k, v := range operationLastMonth {
		sortedOps = append(sortedOps, KV{k, v})
	}
	sort.Slice(sortedOps, func(i, j int) bool {
		return sortedOps[i].Value < sortedOps[j].Value
	})

	// Create another map or array to preserve order in templates if necessary,
	// though Go templates iterate maps randomly. We might want to pass the sorted slice instead.

	// Sort userItems based on operationLastMonth ascending sum (which is what Rails did implicitly with @userSorted)
	// The original ruby code: @operationLastMouth = @operationLastMouth.sort_by {|_key, value| value}.to_h
	// But it actually iterated `@users` to build the table or `@operationLastMouth`?
	// The view did: `<% @operationLastMouth.each do |key, value| %><%= running_user = @users.find(key) %>...`
	// So the table is sorted by the operationLastMonth value ascending!

	var sortedUserItems []UserItem
	for _, kv := range sortedOps {
		for _, item := range userItems {
			if item.User.ID == kv.Key {
				sortedUserItems = append(sortedUserItems, item)
				break
			}
		}
	}

	// Keg view models
	type KegItem struct {
		Keg       models.Keg
		DrinkName string
		OpNum     float64
		Percent   float64
	}
	var kegItems []KegItem
	for _, keg := range kegs {
		var operations []models.Operation
		h.DB.Where("\"drink_id\" = ? AND \"date\" >= ?", int(keg.DrinkID), keg.StartDate).Find(&operations)
		var opNum float64
		for _, op := range operations {
			if op.NumberDrink != 0 {
				opNum += op.NumberDrink
			}
		}

		var drink models.Drink
		h.DB.First(&drink, uint(keg.DrinkID))

		volGlass := 0.0
		if drink.DrinkType == "Bière Cidre" {
			volGlass = 0.3
		} else if drink.DrinkType == "Vin" {
			volGlass = 0.15
		} else if drink.DrinkType == "Whisky Pastis Ricard Rhum" {
			volGlass = 0.04
		}

		percent := 0.0
		if volGlass > 0 && keg.Capacity > 0 {
			percent = 100 - (100 / (keg.Capacity / volGlass) * opNum)
		}

		kegItems = append(kegItems, KegItem{
			Keg:       keg,
			DrinkName: drink.Name,
			OpNum:     float64(int(opNum)),
			Percent:   percent,
		})
	}

	var drinkBlonde models.Drink
	var priceBlonde float64
	var kegBlondeID uint
	if h.DB.First(&drinkBlonde, 1).Error == nil {
		priceBlonde = drinkBlonde.Price
		for _, keg := range kegs {
			if keg.DrinkID == 1 {
				kegBlondeID = keg.ID
			}
		}
	}

	var drinkSpecial models.Drink
	var priceSpecial float64
	var kegSpecialID uint
	if h.DB.First(&drinkSpecial, 9).Error == nil {
		priceSpecial = drinkSpecial.Price
		for _, keg := range kegs {
			if keg.DrinkID == 9 {
				kegSpecialID = keg.ID
			}
		}
	}

	c.HTML(http.StatusOK, "users/index.html", gin.H{
		"totalAmount":   totalAmount,
		"userItems":     sortedUserItems,
		"kegItems":      kegItems,
		"priceBlonde":   priceBlonde,
		"priceSpecial":  priceSpecial,
		"kegBlondeID":   kegBlondeID,
		"kegSpecialID":  kegSpecialID,
	})
}

// Show GET /users/:id
func (h *UsersHandler) Show(c *gin.Context) {
	var user models.User
	if err := h.DB.First(&user, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "User not found"})
		return
	}

	var operations []models.Operation
	h.DB.Where("user_id = ?", user.ID).Order("date desc").Find(&operations)

	c.HTML(http.StatusOK, "users/show.html", gin.H{
		"user":       user,
		"operations": operations,
	})
}

// Lost GET /users/:id/lost
func (h *UsersHandler) Lost(c *gin.Context) {
	// Rails code:
	// user = User.find(params["id"])
	// UserNotifierMailer.send_lost_email(user).deliver
	// redirect_to User, notice

	// In Go, sending email is deferred. I'll just redirect for now.
	c.Redirect(http.StatusFound, "/users")
}

// New GET /users/new
func (h *UsersHandler) New(c *gin.Context) {
	c.HTML(http.StatusOK, "users/new.html", gin.H{
		"user": models.User{},
	})
}

// Edit GET /users/:id/edit
func (h *UsersHandler) Edit(c *gin.Context) {
	var user models.User
	if err := h.DB.First(&user, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "User not found"})
		return
	}
	c.HTML(http.StatusOK, "users/edit.html", gin.H{
		"user": user,
	})
}

// Create POST /users
func (h *UsersHandler) Create(c *gin.Context) {
	user := models.User{
		FirstName:      c.PostForm("user[firstName]"),
		LastName:       c.PostForm("user[lastName]"),
		Alias:          c.PostForm("user[alias]"),
		PasswordDigest: c.PostForm("user[password_digest]"),
		Email:          c.PostForm("user[email]"),
	}

	fmt.Sscanf(c.PostForm("user[initAmount]"), "%f", &user.InitAmount)
	user.Amount = user.InitAmount

	if err := h.DB.Create(&user).Error; err != nil {
		c.HTML(http.StatusUnprocessableEntity, "users/new.html", gin.H{
			"user":  user,
			"error": err.Error(),
		})
		return
	}

	// Emulate sending signup email implicitly
	c.Redirect(http.StatusFound, "/users")
}

// Update PATCH/PUT /users/:id
func (h *UsersHandler) Update(c *gin.Context) {
	adminPassword := c.PostForm("admin_password")
	if adminPassword != os.Getenv("ADMIN_PASSWORD") {
		c.Redirect(http.StatusFound, c.Request.Referer())
		return
	}

	var user models.User
	if err := h.DB.First(&user, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "User not found"})
		return
	}

	// Check if intended to delete directly from update (Rails logic: params[:usr][:delete] == "yes")
	if c.PostForm("usr[delete]") == "yes" {
		h.DB.Where("user_id = ?", user.ID).Delete(&models.Operation{})
		h.DB.Delete(&user)
		c.Redirect(http.StatusFound, "/users")
		return
	}

	user.FirstName = c.PostForm("user[firstName]")
	user.LastName = c.PostForm("user[lastName]")
	user.Alias = c.PostForm("user[alias]")
	user.PasswordDigest = c.PostForm("user[password_digest]")
	user.Email = c.PostForm("user[email]")
	fmt.Sscanf(c.PostForm("user[initAmount]"), "%f", &user.InitAmount)
	fmt.Sscanf(c.PostForm("user[amount]"), "%f", &user.Amount)

	if err := h.DB.Save(&user).Error; err != nil {
		c.HTML(http.StatusUnprocessableEntity, "users/edit.html", gin.H{
			"user":  user,
			"error": err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/users")
}

// Destroy DELETE /users/:id
func (h *UsersHandler) Destroy(c *gin.Context) {
	adminPassword := c.PostForm("admin_password")
	if adminPassword != os.Getenv("ADMIN_PASSWORD") {
		c.Redirect(http.StatusFound, c.Request.Referer())
		return
	}

	var user models.User
	if err := h.DB.First(&user, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "User not found"})
		return
	}

	h.DB.Where("user_id = ?", user.ID).Delete(&models.Operation{})
	h.DB.Delete(&user)

	c.Redirect(http.StatusFound, "/users")
}
