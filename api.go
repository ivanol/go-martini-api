package api

import (
	log "github.com/Sirupsen/logrus"
	"reflect"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
)

//func init() { log.SetLevel(log.DebugLevel) }

// Global Options for an API instance.
type Options struct {
	JwtKey  string
	Martini *martini.ClassicMartini
	Db      *gorm.DB
}

// RouteOptions can be applied to a single route or to a model. Pass them as
// one of the options to AddDefaultRoutes.
type RouteOptions struct {
	// A string to prefix all routes with. Eg. if the Model is called 'Employee' and
	// Prefix=="/department/:dept_id" then a list of all employees is found at
	// "/api/department/:dept_id/employees".
	// If you want to limit that list by :dept_id you need to define a callback
	// below
	Prefix string

	// By default the plural camel_case of the model name is used for the uri. eg.
	// if model is UserType the uri is /api/user_types. Override this here.
	UriModelName string

	// Handlers. If present these will be added in the following order. They will all
	// have access to a Request object containing the database handle, and can modify
	// this as required
	Authenticate interface{}     // Use to authenticate. Either a boolean 'true/false', or martini.Handler
	Authorize    martini.Handler // Use to authorize (if this can be done on route alone).
	Query        martini.Handler // Use to edit the db object (eg. add a Where or Preload)
	// Now the DB query will be carried out.
	// Now in a POST / PUT / PATCH request the uploaded object will be bound to req.Uploaded
	CheckUpload martini.Handler //POST/PUT/PATCH only. req.Uploaded will contain the upload object
	// The Request object should now contain a Result. This will be the object retrieved
	// from gorm, or the edited/deleted object. By default it will be marshalled and sent
	// back to the user. You can change that behaviour here.
	EditResult martini.Handler
}

type API interface {
	// Getters
	Martini() *martini.ClassicMartini
	DB() *gorm.DB

	// Add rest routes for model at path. We will add by defult index, GET,
	// POST, PATCH and DELETE routes.  modelPtr should be a pointer to a
	// struct that is a gorm database table. options is optional. It should
	// contain a ModelOptions struct. If two options arguments are given then
	// the first will apply to GET routes, and the second to POST/PATCH/DELETE
	AddDefaultRoutes(modelPtr interface{}, options ...RouteOptions)
	AddIndexRoute(modelP interface{}, options ...RouteOptions)
	AddGetRoute(modelP interface{}, options ...RouteOptions)
	AddPostRoute(modelP interface{}, options ...RouteOptions)
	AddPatchRoute(modelP interface{}, options ...RouteOptions)
	AddDeleteRoute(modelP interface{}, options ...RouteOptions)

	// Set the model used for logging in (eg. User). Path will be added as a
	// POST route to this model, with the LoginModel's AuthenticateJson method
	// called in the handler to determine if authentication passes.
	SetAuth(model LoginModel, path string)

	// Returns a middleware handler for authentication.
	IsAuthenticated() interface{}
}

type apiServer struct {
	db         *gorm.DB
	martini    *martini.ClassicMartini
	loginModel LoginModel
	options    *Options
}

//New returns a new API, initialised with martini and db. It
//also Maps itself into martini, so it will be available by dependency
//injection in callbacks, and calls any callbacks registered with
//RequestInitAlert()
//func New(m *martini.ClassicMartini, db *gorm.DB) API {
func New(options Options) API {
	m := options.Martini
	if m == nil {
		m = martini.Classic()
	}
	if options.Db == nil {
		panic("Can't start API server without a database. Please pass a gorm DB object  (eg. api.New(api.Options{Db: XXX}) )")
	}
	api := apiServer{db: options.Db, martini: m, options: &options}

	api.martini.Use(func(c martini.Context) {
		c.Map(&api)
	})
	return &api
}

func (api *apiServer) DB() *gorm.DB {
	return api.db
}

func (api *apiServer) Martini() *martini.ClassicMartini {
	return api.martini
}

//Implements API interface for AddDefaultRoutes()
func (api *apiServer) AddDefaultRoutes(modelP interface{}, options ...RouteOptions) {
	modelType := reflect.TypeOf(modelP).Elem()
	log.WithFields(log.Fields{"Model": modelType}).Debug("Adding REST routes")
	api.AddIndexRoute(modelP, options...)
	api.AddGetRoute(modelP, options...)
	api.AddPostRoute(modelP, options...)
	api.AddPatchRoute(modelP, options...)
	api.AddDeleteRoute(modelP, options...)
}

const (
	ROUTE_READ   = iota
	ROUTE_WRITE  = iota
	ROUTE_DELETE = iota
)

//Implements API interface for AddIndexRoute()
func (api *apiServer) AddIndexRoute(modelP interface{}, _options ...RouteOptions) {
	options := getOptions(_options, ROUTE_READ)
	finalPath := makePath(modelP, options)
	modelType := reflect.TypeOf(modelP).Elem()
	sliceType := reflect.SliceOf(modelType)
	log.WithFields(log.Fields{"Model": modelType, "path": finalPath}).Info("Adding INDEX route")
	api.martini.Get(finalPath, api.indexHandlers(sliceType, options)...)
}

//Implements API interface for AddGetRoute()
func (api *apiServer) AddGetRoute(modelP interface{}, _options ...RouteOptions) {
	options := getOptions(_options, ROUTE_READ)
	finalPath := makePath(modelP, options) + "/:id"
	modelType := reflect.TypeOf(modelP).Elem()
	log.WithFields(log.Fields{"Model": modelType, "path": finalPath}).Info("Adding GET route")
	api.martini.Get(finalPath, api.itemHandlers(modelType, options)...)
}

//Implements API interface for AddPostRoute()
func (api *apiServer) AddPostRoute(modelP interface{}, _options ...RouteOptions) {
	options := getOptions(_options, ROUTE_WRITE)
	finalPath := makePath(modelP, options)
	modelType := reflect.TypeOf(modelP).Elem()
	log.WithFields(log.Fields{"Model": modelType, "path": finalPath}).Info("Adding POST route")
	api.martini.Post(finalPath, api.postHandlers(modelType, options)...)
}

//Implements API interface for AddPatchRoute()
func (api *apiServer) AddPatchRoute(modelP interface{}, _options ...RouteOptions) {
	options := getOptions(_options, ROUTE_WRITE)
	finalPath := makePath(modelP, options) + "/:id"
	modelType := reflect.TypeOf(modelP).Elem()
	log.WithFields(log.Fields{"Model": modelType, "path": finalPath}).Info("Adding PATCH route")
	api.martini.Patch(finalPath, api.patchHandlers(modelType, options)...)
}

//Implements API interface for AddDeleteRoute()
func (api *apiServer) AddDeleteRoute(modelP interface{}, _options ...RouteOptions) {
	options := getOptions(_options, ROUTE_DELETE)
	finalPath := makePath(modelP, options) + "/:id"
	modelType := reflect.TypeOf(modelP).Elem()
	log.WithFields(log.Fields{"Model": modelType, "path": finalPath}).Info("Adding DELETE route")
	api.martini.Delete(finalPath, api.deleteHandlers(modelType, options)...)
}

//Implements API interface for SetAuth()
func (api *apiServer) SetAuth(model LoginModel, path string) {
	if api.options.JwtKey == "" {
		panic("Can't do authorisation safely unless you provide a random secret string as JwtKey parameter of api.New()")
	}
	api.loginModel = model

	api.martini.Post(path, ParseJsonBody, api.getLoginHandler())
}

// Extract options from slice
func getOptions(options []RouteOptions, routeType int) RouteOptions {
	if len(options) > 3 {
		panic("Too many arguments for route setup function. Should be 1-3 options items")
	}
	if len(options) == 0 {
		return RouteOptions{}
	}
	if len(options) == 1 {
		return options[0]
	}
	switch routeType {
	case ROUTE_READ:
		return options[0]
	case ROUTE_WRITE:
		return options[1]
	case ROUTE_DELETE:
		if len(options) == 3 {
			return options[2]
		} else {
			return options[1]
		}
	}
	panic("unknown route type")
}

// makePath returns the path for the item, modidified as required by any options.
func makePath(modelP interface{}, options RouteOptions) string {
	path := options.UriModelName
	if len(path) == 0 {
		path = pluralCamelName(modelP)
	}
	return "/api" + options.Prefix + "/" + path
}
