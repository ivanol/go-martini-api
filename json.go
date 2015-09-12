package api

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/go-martini/martini"
)

type JsonBody map[string]interface{}

//ParseJsonBody is middleware to read the body of the http request, and bind it
//to the request object.
func ParseJsonBody(w http.ResponseWriter, r *http.Request, c martini.Context) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	var f interface{}
	if err := json.Unmarshal(body, &f); err != nil {
		log.Println("Receieved malformed JSON body")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		w.Write([]byte("{'error':'Malformed JSON'}"))
	} else {
		m := f.(map[string]interface{})
		j := JsonBody(m)
		c.Map(&j)
	}
}
