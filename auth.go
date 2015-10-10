package api

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-martini/martini"
)

// Objects that implement LoginModel can be used to login. We don't use any locally defined types so
// that any package implementing this is not directly dependent on ourselves.
type LoginModel interface {
	CheckLoginDetails(json *map[string]interface{}) (uint, error)
	//GetById returns an interface (which should be of type LoginModel) by it's id
	GetById(id uint) (interface{}, error)
}

// getLoginHandler() returns the handler function to respond to the login request.
// The handler defers checking the logindetails to loginModel's CheckLoginDetails.
// On success we create a JWT web token using user_id
func (api *apiServer) getLoginHandler() func(*JsonBody, http.ResponseWriter, martini.Context) []byte {
	return func(j *JsonBody, w http.ResponseWriter, c martini.Context) []byte {
		msi := map[string]interface{}(*j)
		user_id, err := api.loginModel.CheckLoginDetails(&msi)
		if err != nil {
			log.Println("Login failed", err)
			w.WriteHeader(403)
			return []byte("Login failed")
		} else {
			log.Println("Logged in user", user_id)
			token := getJWTToken(user_id, api.options.JwtKey)
			return []byte("{\"token\":\"" + token + "\"}")
		}
	}
}

// IsAuthenticated middleware function checks for a jwt token in the request object, and
// either returns a 401 Unauthorized, or continues after mapping  the LoginModel object into
// the request context
func (api *apiServer) IsAuthenticated() interface{} {
	return func(w http.ResponseWriter, r *http.Request, c martini.Context) {
		token, tokerr := jwt.ParseFromRequest(r, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.WithFields(log.Fields{"method": token.Header["alg"]}).Warn("JWT Auth: Unexpected signing method.")
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(api.options.JwtKey), nil
		})
		if token != nil && token.Valid {
			guser, err := api.loginModel.GetById(uint(token.Claims["id"].(float64)))
			if err != nil {
				w.WriteHeader(401)
				fmt.Fprintf(w, "Unauthorized")
				log.WithFields(log.Fields{"id": token.Claims["id"]}).Warn("Cannot find logged in user")
				return
			}
			user := guser.(LoginModel)
			c.Map(user)
		} else {
			log.WithFields(log.Fields{"error": tokerr}).Warn("Auth: JWT token did not validate")
			w.WriteHeader(401)
			fmt.Fprintf(w, "Unauthorized")
		}
	}
}

//Create a JWT token with id=id and expiring in 1 hour
func getJWTToken(id uint, key string) string {
	token := jwt.New(jwt.SigningMethodHS256)
	// Set some claims
	token.Claims["id"] = id
	token.Claims["exp"] = time.Now().Add(time.Hour * 1).Unix()
	log.WithFields(log.Fields{"expiry": token.Claims["exp"], "id": id}).Info("Signing token.")
	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(key))
	log.Printf("Token: %s, error %v", tokenString, err)
	return tokenString
}
