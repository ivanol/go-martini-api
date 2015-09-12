package api

import (
	"encoding/json"
	"testing"
)

var userHasBeenInitialised bool

// Make *User a memeber of WantsInitAlert, and record the alert
func (*User) OnAPIInit(a API) {
	userHasBeenInitialised = true
}

func init() {
	RequestInitAlert(&User{})
}

func TestRegisterUriName(t *testing.T) {
	// helpers should have registered /api/other_widgets as a modelname. Check it works.
	body := testReq(t, "GetItem", "GET", "/api/other_widgets/2", "", 200)
	result := Widget{}
	json.Unmarshal([]byte(body), &result)
	if result.Name != "Widget 2" {
		t.Errorf("Failed to retrieve correct item from the db using an alternative model name: %v", result)
	}
}

func TestInitAlerts(t *testing.T) {
	_ = getTestApi()
	if !userHasBeenInitialised {
		t.Errorf("Didn't get an init alert that we requested")
	}
}
