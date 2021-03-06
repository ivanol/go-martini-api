package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"testing"
)

// Setup a route handler which returns a record of the handlers used!
type testRec struct {
	handlers string
}

func (t *testRec) record(handler string) {
	if len(t.handlers) > 0 {
		t.handlers += ":"
	}
	t.handlers += handler
}

func TestCallbacks(t *testing.T) {
	api := getTestApi()
	api.AddDefaultRoutes(&PrivateWidget{},
		RouteOptions{
			UriModelName: "recordRoutes",
			Authenticate: func(req *Request, c martini.Context) {
				tr := testRec{req.Method}
				tr.record("Authenticate")
				c.Map(&tr)
			},
			Authorize:   func(tr *testRec) { tr.record("Authorize") },
			Query:       func(tr *testRec) { tr.record("Query") },
			CheckUpload: func(tr *testRec) { tr.record("CheckUpload") },
			EditResult: func(tr *testRec, req *Request) {
				tr.record("EditResult")
				req.Result = tr.handlers
			}})
	// Note expected result is a marshalled json string - hence the `""` not ""
	testMethodHandlers(t, "TestCallbacks(GET)", "GET", `"GET:Authenticate:Authorize:Query:EditResult"`)
	testMethodHandlers(t, "TestCallbacks(POST)", "POST", `"POST:Authenticate:Authorize:CheckUpload:EditResult"`)
	testMethodHandlers(t, "TestCallbacks(PATCH)", "PATCH", `"PATCH:Authenticate:Authorize:Query:CheckUpload:EditResult"`)
	testMethodHandlers(t, "TestCallbacks(DELETE)", "DELETE", `"DELETE:Authenticate:Authorize:Query:EditResult"`)
}

// Test our we callback NeedsValidation interfaces appropriately.
func TestNeedsValidation(t *testing.T) {
	body := testReq(t, "PostItem", "POST", "/api/verified_widgets", `{"must_be_hello_world":"NewWidget"}`, 422)
	if body != `{"errors":{"must_be_hello_horld":"Is not equal to \"Hello World!!\""}}` {
		t.Errorf("Didn't receive correct error message for unverified widget: %s\n", body)
	}
	testReq(t, "PostItem", "POST", "/api/verified_widgets", `{"must_be_hello_world":"Hello World!!"}`, 200)
}

// helper function for TestCallbacks. Call the request, and check the expected
// series of callbacks is returned.
func testMethodHandlers(t *testing.T, name string, method string, expected string) {
	body := ""
	if method == "POST" || method == "PUT" || method == "PATCH" {
		body = `{"name":"testname"}`
	}
	uri := "/api/recordRoutes"
	if method == "DELETE" || method == "PATCH" {
		newWidget := PrivateWidget{Name: "ToDelete"}
		getTestApi().DB().Create(&newWidget)
		uri = fmt.Sprintf("%s/%d", uri, newWidget.ID)
	}
	handlers := testReq(t, name, method, uri, body, 200)
	if handlers != expected {
		t.Errorf("For %s expected handler list %s, got %s", method, expected, handlers)
	} else {
		t.Logf("SUCCESS - Got handler list %s for %s", handlers, method)
	}
}

// itemHandlers returns a handler for returning a single item. Test with some requests
func TestItemHandlers(t *testing.T) {
	testReq(t, "GetItem", "GET", "/api/widgets/42", "", 404)
	body := testReq(t, "GetItem", "GET", "/api/widgets/2", "", 200)
	result := Widget{}
	json.Unmarshal([]byte(body), &result)
	if result.Name != "Widget 2" {
		t.Errorf("Failed to retrieve correct single item from the db: %v", result)
	}
}

// indexHandlers returns a handler for returning a list of items. Test with some requests
func TestIndexHandlers(t *testing.T) {
	body := testReq(t, "GetItem", "GET", "/api/widgets", "", 200)
	result := make([]Widget, 0)
	json.Unmarshal([]byte(body), &result)
	if len(result) != 3 {
		t.Errorf("Failed to retrieve correct numbers of widgets: %v", result)
	}
	for _, w := range result {
		if w.Name != fmt.Sprintf("Widget %d", w.ID) {
			t.Errorf("Didn't retrieve correct widget: %v", w)
		}
	}
}

// test post handlers
func TestPostHandlers(t *testing.T) {
	testReq(t, "PostItem(Malformed JSON)", "POST", "/api/widgets", `{"name""NewWidget"}`, 422)
	testReq(t, "PostItem(Existing Item ID)", "POST", "/api/widgets", `{"name":"NewWidget", "id":1}`, 422)
	body := testReq(t, "PostItem", "POST", "/api/widgets", `{"name":"NewWidget"}`, 200)
	newWidget := Widget{}
	checkWidget := Widget{}
	json.Unmarshal([]byte(body), &newWidget)
	getTestApi().DB().Where("id = ?", newWidget.ID).Find(&checkWidget)
	if newWidget.Name != "NewWidget" {
		t.Errorf("Didn't retrieve new object in POST request: %v", newWidget)
	}
	if checkWidget.Name != "NewWidget" {
		t.Errorf("Didn't save new object to DB in apparently successful POST request: %v", newWidget)
	}
	// Clear up
	getTestApi().DB().Delete(&checkWidget)
}

// patchHandlers returns a handler for patching an item. Test with some requests
func TestPatchHandlers(t *testing.T) {
	newWidget := Widget{Name: "ToEdit"}
	getTestApi().DB().Create(&newWidget)

	testReq(t, "EditItem(Doesn'tExist)", "PATCH", "/api/widgets/42", "", 404)
	testReq(t, "EditItem(MalformedJson)", "PATCH", fmt.Sprintf("/api/widgets/%v", newWidget.ID), `{"name:EditedName"}`, 422)
	testReq(t, "EditItem(EditID)", "PATCH", fmt.Sprintf("/api/widgets/%v", newWidget.ID), `{"id":0,"name":"EditedName"}`, 422)
	body := testReq(t, "EditItem", "PATCH", fmt.Sprintf("/api/widgets/%v", newWidget.ID), `{"name":"EditedName"}`, 200)
	checkWidget := Widget{}
	json.Unmarshal([]byte(body), &checkWidget)
	if checkWidget.Name != "EditedName" || checkWidget.ID != newWidget.ID {
		t.Errorf("Failed to return edited item on edit: %v != %v", checkWidget, newWidget)
	}
	if getTestApi().DB().Where("id = ?", newWidget.ID).Find(&checkWidget).RecordNotFound() {
		t.Errorf("The record we edited disappeared")
	} else {
		if checkWidget.Name != "EditedName" {
			t.Errorf("Failed to edit record in DB despite apparent success: %v != %v", newWidget, checkWidget)
		} else {
			t.Logf("PATCH widget succeeded")
		}
	}

}

// deleteHandlers returns a handler for deleting a single item. Test with some requests
func TestDeleteHandlers(t *testing.T) {
	newWidget := Widget{Name: "ToDelete"}
	getTestApi().DB().Create(&newWidget)

	testReq(t, "DeleteItem", "DELETE", "/api/widgets/42", "", 404)
	body := testReq(t, "DeleteItem", "DELETE", fmt.Sprintf("/api/widgets/%v", newWidget.ID), "", 200)
	checkWidget := Widget{}
	json.Unmarshal([]byte(body), &checkWidget)
	if checkWidget.Name != "ToDelete" {
		t.Errorf("Failed to return deleted item on delete: %v", checkWidget)
	}
	if getTestApi().DB().Where("id = ?", newWidget.ID).Find(&checkWidget).RecordNotFound() {
		t.Logf("SUCCESS - Record successfully deleted")
	} else {
		t.Errorf("Record not deleted when it should have been")
	}

}
