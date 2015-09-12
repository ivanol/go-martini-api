// API test framework. This file contains setup code, and utility functions
// for the testing, but no actual tests.
package api

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"

	"testing"
)

var verboseMartini = flag.Bool("verbose", false, "Log all martini requests, and API internals")

// Define some sample structures for the DB
type User struct {
	ID       uint   `gorm:"primary_key" json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`

	PrivateWidgets []PrivateWidget `json:"private_widgets"`
}

type PrivateWidget struct {
	ID     uint   `gorm:"primary_key" json:"id"`
	UserID uint   `json:"user_id"`
	Name   string `json:"name"`
}

type Widget struct {
	ID   uint   `gorm:"primary_key" json:"id"`
	Name string `json:"name"`
}

type WidgetClone struct {
	ID   uint   `gorm:"primary_key" json:"id"`
	Name string `json:"name"`
}

func createAndSeedDb() *gorm.DB {
	db, err := gorm.Open("sqlite3", "./api-test.db")

	if err != nil {
		log.Panicf("Error opening sqlite3 database in test %v\n", err)
	}

	db.DropTable(&User{})
	db.DropTable(&PrivateWidget{})
	db.DropTable(&Widget{})
	db.CreateTable(&User{})
	db.CreateTable(&PrivateWidget{})
	db.CreateTable(&Widget{})

	var private_widgets []PrivateWidget
	db.Model(&User{}).Related(&private_widgets)

	db.Create(&User{Name: "admin", Password: "password"})

	db.Create(&Widget{ID: 1, Name: "Widget 1"})
	db.Create(&Widget{ID: 2, Name: "Widget 2"})
	db.Create(&Widget{ID: 3, Name: "Widget 3"})

	return &db
}

// getSilentMartini returns a ClassicMartini with logging disabled.
func getSilentMartini() *martini.ClassicMartini {
	r := martini.NewRouter()
	m := martini.New()
	if *verboseMartini {
		m.Use(martini.Logger())
	}
	m.Use(martini.Recovery())
	m.Use(martini.Static("public"))
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return &martini.ClassicMartini{m, r}
}

var test_api API

// Returns a singleton instance of API, intialised with an empty DB containing a
// single user and a silent version of ClassicMartini
func getTestApi() API {
	if test_api != nil {
		return test_api
	}

	// Set Log level here. This should only be called once, and near the start
	// of the test run
	if !*verboseMartini {
		log.SetLevel(log.ErrorLevel)
	}

	//Register Privatewidget for automatic inclusion before loading api.
	UseStandardRest(&PrivateWidget{}, RouteOptions{Authenticate: true})
	UseStandardRest(&Widget{})
	UseStandardRest(&Widget{}, RouteOptions{UriModelName: "other_widgets"})

	db := createAndSeedDb()
	a := New(Options{JwtKey: "RandomString", Db: db, Martini: getSilentMartini()})

	a.AddDefaultRoutes("/users", &User{})
	a.SetAuth(&User{}, "/auth")

	test_api = a
	return a
}

// Test a request to the api.
func testReq(t *testing.T, name string, method string, path string, body string, expectedCode int) string {
	api := getTestApi()
	payload := strings.NewReader(body)
	req, err := http.NewRequest(method, path, payload)
	if err != nil {
		t.Errorf("Error creating request for %v: %v\n", path, err)
		return ""
	}
	httpRecorder := httptest.NewRecorder()
	api.Martini().ServeHTTP(httpRecorder, req)
	if httpRecorder.Code == expectedCode {
		t.Logf("SUCCESS - %v returned code %v and body %s\n", name, httpRecorder.Code, httpRecorder.Body.String())
	} else {
		t.Errorf("%v should have code %v. Got %v and body %s\n", name, expectedCode, httpRecorder.Code, httpRecorder.Body.String())
	}
	return httpRecorder.Body.String()
}

// ensurePanic is A deferrable function that fails the test with msg if there
// is no panic.
func ensurePanic(t *testing.T, msg string) {
	p := recover()
	if p == nil {
		t.Errorf(msg)
	} else {
		t.Logf("SUCCESS - We tried '%s' and panicked appropriately", msg)
	}
}