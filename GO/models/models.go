package models

import "time"

// Drink maps to the drinks table
type Drink struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"type:varchar"`
	DrinkType string    `gorm:"type:varchar"`
	Price     float64   `gorm:"type:decimal"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (Drink) TableName() string {
	return "drinks"
}

// User maps to the users table
type User struct {
	ID             uint      `gorm:"primaryKey"`
	FirstName      string    `gorm:"column:firstName;type:varchar"`
	LastName       string    `gorm:"column:lastName;type:varchar"`
	Alias          string    `gorm:"column:alias;type:varchar"`
	PasswordDigest string    `gorm:"column:password_digest;type:varchar"`
	InitAmount     float64   `gorm:"column:initAmount;type:decimal"`
	Amount         float64   `gorm:"column:amount;type:decimal"`
	Email          string    `gorm:"column:email;type:varchar"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}

// Keg maps to the kegs table
type Keg struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"column:name;type:varchar"`
	DrinkID   float64   `gorm:"column:drink_id;type:decimal"` // Rails schema says decimal for drink_id here
	StartDate time.Time `gorm:"column:startDate;type:timestamp"`
	EndDate   time.Time `gorm:"column:endDate;type:timestamp"`
	Capacity  float64   `gorm:"column:capacity;type:decimal"`
	Price     float64   `gorm:"column:price;type:decimal"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (Keg) TableName() string {
	return "kegs"
}

// Operation maps to the operations table
type Operation struct {
	ID          uint      `gorm:"primaryKey"`
	Date        time.Time `gorm:"column:date;type:timestamp"`
	Sum         float64   `gorm:"column:sum;type:decimal"`
	UserID      uint      `gorm:"column:user_id;type:integer"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	DrinkID     float64   `gorm:"column:drink_id;type:decimal"` // Decimal in schema
	NumberDrink float64   `gorm:"column:numberDrink;type:decimal"`
	Comment     string    `gorm:"column:comment;type:varchar"`
}

func (Operation) TableName() string {
	return "operations"
}

// Beerflow maps to the beerflows table
type Beerflow struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Quantity  float64   `gorm:"type:decimal" json:"quantity"`
	DrinkID   uint      `gorm:"type:integer" json:"drink_id"` // Integer in schema
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Beerflow) TableName() string {
	return "beerflows"
}
