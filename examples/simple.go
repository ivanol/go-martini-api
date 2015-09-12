package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ivanol/go-martini-api"
)

type Widget struct {
	ID   uint   `gorm:"primary_key" json:"id"`
	Name string `json:"name"`
}

func main() {
	// Create a DB with the test table and a seed widget
	db, _ := gorm.Open("sqlite3", "./api-example.db")
	db.CreateTable(&Widget{})
	db.Create(&Widget{Name: "Test Widget"})

	// Create an API server with default options
	a := api.New(api.Options{Db: &db})

	// Add the Default REST routes to it (GET, POST, PATCH, DELETE)
	a.AddDefaultRoutes(&Widget{})

	// Run the server.
	a.Martini().RunOnAddr("127.0.0.1:3000")
}
