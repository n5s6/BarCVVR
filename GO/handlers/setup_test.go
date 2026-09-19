package handlers_test

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"math/big"

	"github.com/gin-gonic/gin"
	"github.com/nmontes/barcvvr-go/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupTestDB() *gorm.DB {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	dbName := fmt.Sprintf("file:memdb%d?mode=memory&cache=shared", n.Int64())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(
		&models.User{},
		&models.Drink{},
		&models.Keg{},
		&models.Operation{},
		&models.Beerflow{},
	)

	return db
}

func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Create dummy templates to satisfy handlers and render basic variables for assertions
	t := template.New("base")
	template.Must(t.New("users/index.html").Parse("{{range .userItems}}{{.User.FirstName}} {{.User.LastName}}{{end}}"))
	template.Must(t.New("users/show.html").Parse("{{.user.FirstName}}"))
	template.Must(t.New("users/new.html").Parse("New User"))
	template.Must(t.New("users/edit.html").Parse("{{.user.FirstName}}"))
	template.Must(t.New("drinks/index.html").Parse("{{range .drinks}}{{.Name}}{{end}}"))
	template.Must(t.New("drinks/new.html").Parse("New Drink"))
	template.Must(t.New("drinks/show.html").Parse("{{.drink.Name}}"))
	template.Must(t.New("drinks/edit.html").Parse("{{.drink.Name}}"))
	template.Must(t.New("kegs/index.html").Parse("{{range .kegItems}}{{.Keg.ID}}{{end}}"))
	template.Must(t.New("kegs/new.html").Parse("New Keg"))
	template.Must(t.New("kegs/show.html").Parse("{{.keg.ID}}"))
	template.Must(t.New("kegs/edit.html").Parse("Edit Keg"))
	template.Must(t.New("operations/index.html").Parse("{{range .opItems}}{{.Operation.Comment}}{{end}}"))
	template.Must(t.New("operations/new.html").Parse("New Operation"))
	template.Must(t.New("operations/show.html").Parse("{{.operation.Comment}}"))
	template.Must(t.New("operations/edit.html").Parse("Edit Operation"))
	template.Must(t.New("operations/add.html").Parse("Add Operation"))
	template.Must(t.New("operations/exchange.html").Parse("Exchange Operation"))
	template.Must(t.New("beerflows/index.html").Parse("{{range .beerflows}}{{.Quantity}}{{end}}"))
	template.Must(t.New("beerflows/new.html").Parse("New Beerflow"))
	template.Must(t.New("beerflows/show.html").Parse("{{.beerflow.ID}}"))
	template.Must(t.New("beerflows/edit.html").Parse("Edit Beerflow"))
	template.Must(t.New("error.html").Parse("Error: {{.error}}"))

	r.SetHTMLTemplate(t)
	return r
}
