// Test api.go. Most of api.go is tested by getTestApi() (in helpers_test.go)
// which is called by all the other tests. Here we test failure cases to
// complete our code coverage.
package api

import (
	"testing"
)

// Check a martini is created if not passed.
func TestNilMartini(t *testing.T) {
	api := New(Options{})
	if api.Martini() == nil {
		t.Errorf("API doesn't create it's own Martini when not passed one")
	}
}

// Check we don't allow authorization to be set up without initialising a
// secret key.
func TestEmptyJWTKey(t *testing.T) {
	api := New(Options{})
	defer ensurePanic(t, "API allowed setting up authorization without a secret key")
	api.SetAuth(&User{}, "/login")
}

// Check adding routes with different options for Read/Write
func TestReadWriteOptions(t *testing.T) {
	api := getTestApi()
	api.AddDefaultRoutes("/trwo/ro", &PrivateWidget{}, RouteOptions{}, RouteOptions{Authenticate: true})
	api.AddDefaultRoutes("/trwo/rw", &PrivateWidget{}, RouteOptions{}, RouteOptions{}, RouteOptions{Authenticate: true})

	testReq(t, "ReadOnly(Read)", "GET", "/api/trwo/ro", "", 200)
	testReq(t, "ReadOnly(WRITE)", "POST", "/api/trwo/ro", `{"name":"sqlinjector"}`, 401)
	testReq(t, "ReadOnly(DELETE)", "DELETE", "/api/trwo/ro/1", "", 401)
	testReq(t, "ReadWrite(Read)", "GET", "/api/trwo/rw", "", 200)
	testReq(t, "ReadWrite(WRITE)", "POST", "/api/trwo/rw", `{"name":"important widget"}`, 200)
	testReq(t, "ReadWrite(DELETE)", "DELETE", "/api/trwo/rw/1", "", 401)

	defer ensurePanic(t, "AddDefaultRoute accepted 4 route options")
	api.AddDefaultRoutes("/trwo/fail", &PrivateWidget{}, RouteOptions{}, RouteOptions{}, RouteOptions{}, RouteOptions{})
}

// Check getOptions fails for unknown route type
func TestGetOptions(t *testing.T) {
	defer ensurePanic(t, "getOptions should panic for unknown route type")
	options := make([]RouteOptions, 2)
	getOptions(options, -1)
}
