# go-martini-api
This package uses [Martini](https://github.com/go-martini/martini),
[gorm] (https://github.com/jinzhu/gorm) and Golang to allow building RESTful
APIs with the minimum of time and boilerplate code.

## Simple Example
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
	ID     uint   `gorm:"primary_key" json:"id"`
	Name   string `json:"name"`
    Public bool   `json:"public"`
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

## Customisation

Routes are added with `a.AddDefaultRoutes(modelPointer, api.RouteOptions{}...`. 
A RouteOptions structure can contain a series of Martini Handler definitions which
will be called in turn, and can be used to authenticate, authorize, and otherwise
limit a route.

|            |GET|POST|PATCH|DELETE|
|------------|---|----|-----|------|
|Authenticate|Yes|Yes |Yes  |Yes   |
|Authorize   |Yes|Yes |Yes  |Yes   |
|Query       |Yes|    |Yes  |Yes   |
|CheckUpload |   |Yes |Yes  |      |
|EditResult  |Yes|Yes |Yes  |Yes   |

All of these Handlers have access to a `*api.Request` object which they can
modify. This contains `DB *gorm.DB` which is the database object for this request.
This can be used to scope a query, which will limit GET, PATCH, and DELETE
requests. The actual query is made immediately after calling the Query handler, and
this should be used for any such scoping. eg:

```go
a.AddDefaultRoutes(&Widget{}, api.RouteOptions{
  Query: func(req *api.Request) {
            req.DB = req.DB.Where("public = ?", true)
            } })
```

CheckUpload is called for POST and PATCH calls. `req.Uploaded` will contain a pointer
to the uploaded object, and this can be inspected, used to deny the request, or edited
before it is saved to the database.

EditResult is the last customisable handler and is called for all
calls. It has access to req.Result, which will be a pointer to the
retrieved, added, edited, or deleted model depending on the call. For the
index it will be a slice of the model. This will by default be marshalled
to JSON and returned to the user. In EditResult it can be edited first,
or an entirely different result can be returned if wished.

## Authentication

The Authenticate handler method of RouteOptions can be used to carry
out a custom authentication strategy. If instead you wish to use the
builtin authentication you must you define an implemention of the
LoginModel interface. This can then be used with
`a.SetAuth('/login/, &MyLoginModel{})` to use jwt based authentication by simply setting
`RouteOptions{Authenticate: true}`.  The successfully logged in user
will be bound to all subsequent handlers as LoginModel.

## Detailed Example

A [detailed example](https://github.com/ivanol/go-martini-api/blob/master/examples/detailed.go)
is in the examples folder. Run it with `go run detailed.go`, and then visit
http://localhost:3000/ to access it.
