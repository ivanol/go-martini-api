package api

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/gedex/inflector"
	"github.com/serenize/snaker"
)

//Miscellaneous utility functions

// getReflectedSlicePtr takes a Type such as User, and returns an
// interface containing &MakeSlice([]User, 0, 0).
func getReflectedSlicePtr(sliceType reflect.Type) interface{} {
	slice := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(slice.Type())
	slicePtr.Elem().Set(slice)
	return slicePtr.Interface()
}

// pluralCamelName takes an interface, and returns its type converted
// to camel_case and pluralised. eg. pluralCamelName(ImportantPerson{})
// should return "important_people"
func pluralCamelName(i interface{}) string {
	t := fmt.Sprintf("%T", i)
	a := strings.Split(t, ".")
	t1 := a[len(a)-1]
	t2 := snaker.CamelToSnake(t1)
	t3 := inflector.Pluralize(t2)
	return t3
}

// httpBody returns the uploaded http body. On io failure we just return
// an empty array.
func httpBody(r *http.Request) []byte {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil || r.Body.Close() != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Can't read request body")
		return []byte("")
	}
	return body
}
