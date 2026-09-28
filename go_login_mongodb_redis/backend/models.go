package main

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	Email        string        `bson:"email" json:"email"`
	PasswordHash string        `bson:"passwordHash" json:"-"`
}
type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type Product struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      bson.ObjectID `bson:"userId" json:"-"`
	Name        string        `bson:"name" json:"name"`
	Description string        `bson:"description" json:"description"`
	Price       float64       `bson:"price" json:"price"`
	Quantity    int           `bson:"quantity" json:"quantity"`
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt" json:"updatedAt"`
}
type ProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}
type PurchaseRequest struct {
	ProductID     string `json:"productId"`
	Quantity      int    `json:"quantity"`
	PaymentMethod string `json:"paymentMethod"`
}
type OrderItem struct {
	ProductID bson.ObjectID `bson:"productId" json:"productId"`
	Name      string        `bson:"name" json:"name"`
	Price     float64       `bson:"price" json:"price"`
	Quantity  int           `bson:"quantity" json:"quantity"`
}
type Order struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        bson.ObjectID `bson:"userId" json:"-"`
	Items         []OrderItem   `bson:"items" json:"items"`
	Total         float64       `bson:"total" json:"total"`
	PaymentMethod string        `bson:"paymentMethod" json:"paymentMethod"`
	Status        string        `bson:"status" json:"status"`
	CreatedAt     time.Time     `bson:"createdAt" json:"createdAt"`
}
