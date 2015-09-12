package api

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gedex/inflector"
	"github.com/serenize/snaker"
)

//Implement WantsInitAlert and register for alerts by calling RequestInitAlert()
//from your init() method. Then you'll get a callback once the api framework is
//running which you can use to register your routes.
type WantsInitAlert interface {
	OnAPIInit(API)
}

type modelRegistration struct {
	model   interface{}
	options []RouteOptions
}

var wantInitAlerts []WantsInitAlert
var useStandardRest []modelRegistration

//RequestInitAlert() adds wantsInitAlert to the queue to be called back once
//the api framework is running.
func RequestInitAlert(wantsInitAlert WantsInitAlert) {
	wantInitAlerts = append(wantInitAlerts, wantsInitAlert)
}

func UseStandardRest(usr interface{}, options ...RouteOptions) {
	useStandardRest = append(useStandardRest, modelRegistration{model: usr, options: options})
}

//doInitAlerts() calls all callbacks registered with RequestInitAlert()
func doInitAlerts(api API) {
	log.Debug("API: Doing init alerts")
	for _, wantsInitAlert := range wantInitAlerts {
		wantsInitAlert.OnAPIInit(api)
	}
	for _, usr := range useStandardRest {
		modelPath := tableNameFromType(usr.model)
		if len(usr.options) > 0 && len(usr.options[0].UriModelName) > 0 {
			modelPath = usr.options[0].UriModelName
		}
		path := fmt.Sprintf("/%v", modelPath)
		api.AddDefaultRoutes(path, usr.model, usr.options...)
	}
}

func tableNameFromType(i interface{}) string {
	t := fmt.Sprintf("%T", i)
	a := strings.Split(t, ".")
	t1 := a[len(a)-1]
	t2 := snaker.CamelToSnake(t1)
	t3 := inflector.Pluralize(t2)
	return t3
}
