package handlers_test

import (
	"testing"

	"github.com/nmontes/barcvvr-go/handlers"
	"gorm.io/gorm"
)

func TestNewHandlers(t *testing.T) {
	db := &gorm.DB{}

	handlersList := []interface{}{
		handlers.NewUsersHandler(db),
		handlers.NewDrinksHandler(db),
		handlers.NewKegsHandler(db),
		handlers.NewOperationsHandler(db),
		handlers.NewBeerflowsHandler(db),
	}

	for i, h := range handlersList {
		if h == nil {
			t.Errorf("Expected handler at index %d to be initialized, got nil", i)
		}
	}
}
