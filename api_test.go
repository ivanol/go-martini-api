// Test api.go. Most of api.go is tested by getTestApi() (in helpers_test.go)
// which is called by all the other tests. Here we test failure cases to
// complete our code coverage.
package api

import (
	"testing"
)

// Check a martini is created if not passed.
func TestNilMartini(t *testing.T) {
	api := New(Options{Db: getTestDb()})
	if api.Martini() == nil {
		t.Errorf("API doesn't create it's own Martini when not passed one")
	}
}

// Check we panic when started with a nil DB
func TestNilDb(t *testing.T) {
	defer ensurePanic(t, "API started without a DB")
	_ = New(Options{})
}

// Check we don't allow authorization to be set up without initialising a
// secret key.
func TestEmptyJWTKey(t *testing.T) {
	api := New(Options{Db: getTestDb()})
	defer ensurePanic(t, "API allowed setting up authorization without a secret key")
	api.SetAuth(&User{}, "/login")
}

// Check adding routes with different options for Read/Write
func TestReadWriteOptions(t *testing.T) {
	api := getTestApi()
	api.AddDefaultRoutes(&PrivateWidget{},
		RouteOptions{UriModelName: "trwo_ro"},
		RouteOptions{Authenticate: true, UriModelName: "trwo_ro"})
	api.AddDefaultRoutes(&PrivateWidget{},
		RouteOptions{UriModelName: "trwo_rw"},
		RouteOptions{UriModelName: "trwo_rw"},
		RouteOptions{Authenticate: true, UriModelName: "trwo_rw"})

	testReq(t, "ReadOnly(Read)", "GET", "/api/trwo_ro", "", 200)
	testReq(t, "ReadOnly(WRITE)", "POST", "/api/trwo_ro", `{"name":"sqlinjector"}`, 401)
	testReq(t, "ReadOnly(DELETE)", "DELETE", "/api/trwo_ro/1", "", 401)
	testReq(t, "ReadWrite(Read)", "GET", "/api/trwo_rw", "", 200)
	testReq(t, "ReadWrite(WRITE)", "POST", "/api/trwo_rw", `{"name":"important widget"}`, 200)
	testReq(t, "ReadWrite(DELETE)", "DELETE", "/api/trwo_rw/1", "", 401)

	defer ensurePanic(t, "AddDefaultRoute accepted 4 route options")
	api.AddDefaultRoutes(&PrivateWidget{}, RouteOptions{}, RouteOptions{}, RouteOptions{}, RouteOptions{})
}

// Check getOptions fails for unknown route type
func TestGetOptions(t *testing.T) {
	defer ensurePanic(t, "getOptions should panic for unknown route type")
	options := make([]RouteOptions, 2)
	getOptions(options, -1)
}
