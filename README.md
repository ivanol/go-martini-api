# go-martini-api
This package uses Martini and go to allow building RESTful APIs with the
minimum of time and boilerplate code.

This is not yet ready for use in production.

## Example
This example creates a webserver which serves (unauthenticated) REST
endpoints for Widget. These are:

| Verb    | URI            | Action                          |
|---------|----------------|-------
| GET     | /api/widgets   | Get full widget list
| GET     | /api/widgets/1 | Get widget with id==1
| POST    | /api/widgets   | Create a new widget
| PATCH   | /api/widgets/1 | Update widget with id==1
| DELETE  | /api/widgets/1 | Delete widget wit id==1

```go
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
```

User authentication and authorization is also possible, but without
documentation except in the source code so far. The API is still likely
to change, so please don't rely on it yet.
