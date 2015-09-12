# go-martini-api
This package uses Martini and go to allow building RESTful APIs with the
minimum of time and boilerplate code.

This is not yet ready for use in production.

## Example
This example creates a webserver which serves (unauthenticated) REST
endpoints for Widget. These are:
* GET /api/widgets Get full widget list
* GET /api/widgets/1 Get widget with id==1
* POST /api/widgets Create a new widget
* PATCH /api/widgets/1 Update widget with id==1
* DELETE /api/widgets/1 Delete widget wit id==1
'''
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
    // Create a DB with the test tables and a seed user
    db, _ := gorm.Open("sqlite3", "./api-test.db")
    db.CreateTable(&Widget{})

    api.UseStandardRest(&Widget{})

    a := api.New(api.Options{})

    a.Martini().RunOnAddr("127.0.0.1:3000")
}
'''
User authentication and authorization is also possible, but without
documentation except in the source code so far. The API is still likely
to change, so don't rely on it. In particular the UseStandardRest()
interface for registering models before the api.New call will probably
disappear.

