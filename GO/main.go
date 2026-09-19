package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nmontes/barcvvr-go/handlers"

	"github.com/gin-contrib/multitemplate"
)

func methodOverrideHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			_ = r.ParseForm()
			if method := r.FormValue("_method"); method != "" {
				r.Method = strings.ToUpper(method)
			}
		}
		h.ServeHTTP(w, r)
	})
}

func loadTemplates(templatesDir string) multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	layouts, err := filepath.Glob(templatesDir + "/*.html")
	if err != nil {
		panic(err.Error())
	}

	// Add all nested templates
	includes, err := filepath.Glob(templatesDir + "/*/*.html")
	if err != nil {
		panic(err.Error())
	}

	for _, include := range includes {
		// Layouts must be passed to AddFromFiles first, then the specific template that fills it
		files := make([]string, 0, len(layouts)+1)
		files = append(files, layouts...)
		files = append(files, include)

		// The template name will be like "users/index.html"
		name := strings.TrimPrefix(include, templatesDir+"/")

		// If layout.html has {{define "layout"}}, the root block is named "layout", NOT name.
		// However, multitemplate expects the mapped name to correspond to the file.
		r.AddFromFiles(name, files...)
	}
	return r
}

func main() {
	// Load .env file if it exists
	godotenv.Load()

	// Initialize database
	dsn := "host=localhost user=barCVVR dbname=barCVVR_development port=5432 sslmode=disable TimeZone=Europe/Paris"
	if envDsn := os.Getenv("DATABASE_URL"); envDsn != "" {
		dsn = envDsn
	}

	db, err := InitDB(dsn)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	//db.AutoMigrate(&models.User{}, &models.Drink{}, &models.Keg{}, &models.Operation{}, &models.Beerflow{})

	// Initialize Gin router
	r := gin.Default()

	// Load HTML templates and Serves Static Assets
	r.Static("/assets", "assets")

	// Use multitemplate to render nested directories reliably
	r.HTMLRender = loadTemplates("templates")

	// Handlers
	usersHandler := handlers.NewUsersHandler(db)
	drinksHandler := handlers.NewDrinksHandler(db)
	kegsHandler := handlers.NewKegsHandler(db)
	opsHandler := handlers.NewOperationsHandler(db)
	beerflowsHandler := handlers.NewBeerflowsHandler(db)

	// Routes
	r.GET("/", func(c *gin.Context) {
		c.Redirect(301, "/users")
	})

	// Users
	r.GET("/users", usersHandler.Index)
	r.GET("/users/new", usersHandler.New)
	r.POST("/users", usersHandler.Create)
	r.GET("/users/:id", usersHandler.Show)
	r.GET("/users/:id/edit", usersHandler.Edit)
	r.POST("/users/:id", usersHandler.Update) // Use POST to simulate PATCH/PUT forms
	r.PUT("/users/:id", usersHandler.Update)
	r.PATCH("/users/:id", usersHandler.Update)
	r.GET("/users/:id/lost", usersHandler.Lost)
	r.DELETE("/users/:id", usersHandler.Destroy)

	// Drinks
	r.GET("/drinks", drinksHandler.Index)
	r.GET("/drinks/new", drinksHandler.New)
	r.POST("/drinks", drinksHandler.Create)
	r.GET("/drinks/:id", drinksHandler.Show)
	r.GET("/drinks/:id/edit", drinksHandler.Edit)
	r.POST("/drinks/:id", drinksHandler.Update)
	r.PUT("/drinks/:id", drinksHandler.Update)
	r.PATCH("/drinks/:id", drinksHandler.Update)
	r.DELETE("/drinks/:id", drinksHandler.Destroy)

	// Kegs
	r.GET("/kegs", kegsHandler.Index)
	r.GET("/kegs/new", kegsHandler.New)
	r.POST("/kegs", kegsHandler.Create)
	r.GET("/kegs/:id", kegsHandler.Show)
	r.GET("/kegs/:id/edit", kegsHandler.Edit)
	r.POST("/kegs/:id", kegsHandler.Update)
	r.PUT("/kegs/:id", kegsHandler.Update)
	r.PATCH("/kegs/:id", kegsHandler.Update)
	r.DELETE("/kegs/:id", kegsHandler.Destroy)
	r.GET("/kegs/:id/renew", kegsHandler.Renew)

	// Operations
	r.GET("/operations", opsHandler.Index)
	r.GET("/operations/new", opsHandler.New)
	r.POST("/operations", opsHandler.Create)
	r.GET("/operations/:id", opsHandler.Show)
	r.GET("/operations/:id/edit", opsHandler.Edit)
	r.POST("/operations/:id", opsHandler.Update)
	r.PUT("/operations/:id", opsHandler.Update)
	r.PATCH("/operations/:id", opsHandler.Update)
	r.DELETE("/operations/:id", opsHandler.Destroy)
	r.GET("/operations/add", opsHandler.Add)
	r.GET("/operations/exchange", opsHandler.Exchange)

	// Beerflows
	r.GET("/beerflows", beerflowsHandler.Index)
	r.GET("/beerflows/new", beerflowsHandler.New)
	r.POST("/beerflows", beerflowsHandler.Create)
	r.PATCH("/beerflows", beerflowsHandler.Create)
	r.PUT("/beerflows", beerflowsHandler.Create)
	r.GET("/beerflows/:id", beerflowsHandler.Show)
	r.GET("/beerflows/:id/edit", beerflowsHandler.Edit)
	r.POST("/beerflows/:id", beerflowsHandler.Update)
	r.PUT("/beerflows/:id", beerflowsHandler.Update)
	r.PATCH("/beerflows/:id", beerflowsHandler.Update)
	r.DELETE("/beerflows/:id", beerflowsHandler.Destroy)
	r.POST("/beerflows/:id/delete", beerflowsHandler.Destroy)

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	
	server := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: methodOverrideHandler(r),
	}
	server.ListenAndServe()
}
