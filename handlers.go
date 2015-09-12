package api

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"

	log "github.com/Sirupsen/logrus"
)

// Request is a per request structure, bound to martini, containing a db
// handle and a link to the API. The db handle may be altered by other
// handlers - eg. to add a Where or Limit clause.
// Once there is a result it will be stored in Result. This may be
// checked and/or edited by callbacks, and will then be returned to the
// caller by json.marshal(Result)
//
// For requests where a structure is being uploaded (POST/PUT/PATCH) this
// will be parsed and attached to 'Uploaded' after authentication.
type Request struct {
	DB       *gorm.DB
	API      API
	Result   interface{}
	Uploaded interface{}
}

// options.Authenticate may either be a bool (and if true we return our default auth handler),
// or a handler, in which case we return this.
func (api *apiServer) getAuthenticateHandler(auth interface{}) martini.Handler {
	if auth == nil {
		return nil
	}
	if bauth, ok := auth.(bool); ok {
		if bauth {
			return api.IsAuthenticated()
		}
	} else {
		return auth
	}
	return nil
}

// options.Authenticate may either be a bool (and if true we add our default auth handler),
// or a handler, in which case we add this.
func (api *apiServer) appendAuthenticateHandler(handlers []martini.Handler, auth interface{}) []martini.Handler {
	if auth == nil {
		return handlers
	}
	if bauth, ok := auth.(bool); ok {
		if bauth {
			return append(handlers, api.IsAuthenticated())
		}
	} else {
		return append(handlers, auth)
	}
	return handlers
}

// buildHandlerList returns a list of handlers for a request.
// TODO? have a replaceResult handler (or maybe a options.DontSend) that prevents us
// sending the results and lets us be used as pure middleware
func (api *apiServer) buildHandlerList(options RouteOptions, dbHandler martini.Handler) []martini.Handler {
	return handlerList(
		bindRequestHandler,
		api.getAuthenticateHandler(options.Authenticate),
		options.Authorize,
		options.Query,
		dbHandler,
		options.EditResult,
		sendResult)
}

// Concatenate all non nil arguments into a handler list.
func handlerList(handlers ...martini.Handler) []martini.Handler {
	result := make([]martini.Handler, 0, 5)
	for _, handler := range handlers {
		if handler != nil {
			result = append(result, handler)
		}
	}
	return result
}

// bindRequestHandler creates an empty api request object and binds it to the
// martini
func bindRequestHandler(c martini.Context, a API) {
	req := Request{DB: a.DB(), API: a}
	c.Map(&req)
}

// sendResult takes the item found at req.Result, marshals it to JSON, and returns it
func sendResult(req *Request) []byte {
	j, _ := json.Marshal(req.Result)
	return j
}

// Handler to retrieve a single item by id.
func getItemHandler(itemType reflect.Type) martini.Handler {
	return func(params martini.Params, req *Request, w http.ResponseWriter) {
		id := params["id"]
		item := reflect.New(itemType).Interface()
		if req.DB.Where("id = ?", id).Find(item).RecordNotFound() {
			w.WriteHeader(404)
		} else {
			req.Result = item
		}
	}
}

// itemHandlers returns a handler function list for retrieving a single item from the gorm DB by
// item type.
func (api *apiServer) itemHandlers(itemType reflect.Type, options RouteOptions) []martini.Handler {
	return api.buildHandlerList(options, getItemHandler(itemType))
}

// indexHandlers returns a handler function list for retrieving an index of functions from the gorm DB by
// item type.
func (api *apiServer) indexHandlers(sliceType reflect.Type, options RouteOptions) []martini.Handler {
	indexHandler := func(req *Request) {
		items := getReflectedSlicePtr(sliceType)
		req.DB.Find(items)
		req.Result = items
	}
	return api.buildHandlerList(options, indexHandler)
}

// postHandlers returns a handler function list for posting a single item to the DB
func (api *apiServer) postHandlers(itemType reflect.Type, options RouteOptions) []martini.Handler {
	return handlerList(
		bindRequestHandler,
		api.getAuthenticateHandler(options.Authenticate),
		options.Authorize,
		jsonParseBody(itemType),
		options.CheckUpload,
		doCreate(itemType),
		options.EditResult,
		sendResult)
}

// patchHandlers returns a handler function list for deleting a single item from the DB
func (api *apiServer) patchHandlers(itemType reflect.Type, options RouteOptions) []martini.Handler {
	//unmarshal the uploaded body into req.Result. req.Result should already contain the retrieved item,
	//we will now contain the updated version.
	copyItem := func(req *Request, w http.ResponseWriter, r *http.Request, c martini.Context) {
		body := httpBody(r)
		beforeID, _ := getID(req.Result)
		if err := json.Unmarshal(body, req.Result); err != nil {
			log.WithFields(log.Fields{"error": err}).Warn("Can't parse incoming json")
			w.WriteHeader(422) // unprocessable entity
			return
		}
		afterID, _ := getID(req.Result)
		if beforeID != afterID {
			log.WithFields(log.Fields{"afterID": afterID, "beforeID": beforeID}).Warn("Patch trying to change ID")
			w.WriteHeader(422) // unprocessable entity
			return
		}
		req.Uploaded = req.Result
	}
	patchHandler := func(params martini.Params, req *Request, w http.ResponseWriter) {
		req.DB.Save(req.Uploaded)
		req.Result = req.Uploaded
	}
	return handlerList(
		bindRequestHandler,
		api.getAuthenticateHandler(options.Authenticate),
		options.Authorize,
		options.Query,
		getItemHandler(itemType),
		copyItem,
		options.CheckUpload,
		patchHandler,
		options.EditResult,
		sendResult)
}

// deleteHandlers returns a handler function list for deleting a single item from the DB
func (api *apiServer) deleteHandlers(itemType reflect.Type, options RouteOptions) []martini.Handler {
	//TODO use getItemHandler() as part of deleteHandler
	deleteHandler := func(params martini.Params, req *Request, w http.ResponseWriter) {
		id := params["id"]
		item := reflect.New(itemType).Interface()
		if req.DB.Where("id = ?", id).Find(item).RecordNotFound() {
			w.WriteHeader(404)
		} else {
			req.Result = item
			req.DB.Delete(item)
		}
	}
	return api.buildHandlerList(options, deleteHandler)
}

// jsonParseBody returns a martini handler that deserialises the json body of a request into
// a struct
func jsonParseBody(itemType reflect.Type) martini.Handler {
	return func(req *Request, w http.ResponseWriter, r *http.Request, c martini.Context) {
		body := httpBody(r)
		item := reflect.New(itemType).Interface()
		if err := json.Unmarshal(body, item); err != nil {
			log.WithFields(log.Fields{"error": err}).Warn("Can't parse incoming json")
			w.WriteHeader(422) // unprocessable entity
			return
		}
		req.Uploaded = item
		//log.Errorf("Uploaded after %v", req.Uploaded)
	}
}

// add req.Uploaded to the gorm DB. No checks take place in this function. You should have
// Authorized and Authenticated your user using callbacks, and validated the req.Uploaded
// structure in the CheckUpload callback.
func doCreate(itemType reflect.Type) martini.Handler {
	return func(req *Request, w http.ResponseWriter) {
		uploaded := req.Uploaded
		log.Printf("upload is a %T\n", uploaded)

		//item := reflect.New(itemType).Elem().Interface()
		post := req.DB.Create(req.Uploaded)
		err := post.Error
		if err != nil {
			log.Warn("Error creating in doCreate: ", err)
			w.WriteHeader(422)
		}
		req.Result = req.Uploaded
	}
}
