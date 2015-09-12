package api

import (
	"encoding/json"
	"errors"
	"log"

	//"github.com/go-martini/martini"

	"testing"
)

// Add CheckLoginDetails so User implements LoginModel
func (_ *User) CheckLoginDetails(j *JsonBody, a API) (uint, error) {
	user := User{}
	if a.DB().Where("name = ? AND password = ?", (*j)["name"], (*j)["password"]).Find(&user).RecordNotFound() {
		return 0, errors.New("Not authenticated")
	} else {
		return user.ID, nil
	}
}

// This can be set to false to make GetById fail to test the framework.
var disableGetUserById bool

// Add GetById so User implements LoginModel
func (_ *User) GetById(id uint, a API) (LoginModel, error) {
	user := User{}
	if disableGetUserById || a.DB().Where("id = ?", id).Find(&user).RecordNotFound() {
		return &user, errors.New("User not found")
	}
	return &user, nil
}

func TestAuthenticatication(t *testing.T) {
	// A jwt token signed with the 'none' algorithm
	jwt_none := "eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoxfQ."
	// A jwt token created with the wrong hmac key
	wrong_hs256 := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoxfQ.0C1bI67LwtiXcxTbF8Rx3u4StxIF2cUKMdUa3oQqNiE"
	// An rsa token
	jwt_rsa := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJ1c2VyX2lkIjoxfQ.bpAF_DQ2lHtZgq7YNHTWs1tNUzPxSOQ0XYtyIbJG00Uf402l96vcnX0SgkP093ZG6N7ESv4wVwIhqRbHm8K_4r1oMCOQo2zbNkrhsDpvMJ8iKKAga9ZfYBiVPgl1Mu6FYy7uRY-11P4QH6BEgQoLOoLiQaZy9FEoOc7sdmQhV4re7p6TOWEtSqDQF4aYM0ALp4GRTB91UwJTiz9Gy3uAYQOHPqFPupmCKTMkW5H077YAyNHGu5lu4zIzWQfPbcv2zPiLhyNhSPMQxYppCIu-Qpp0X3PnuxAASBCHSrd-s2fWly547dPu2rhBrz1Z8CWSY6A0RC8EC2BQ0G8bUxaqng"

	testReq(t, "Login(Malformed JSON)", "POST", "/auth", `{"name: "admin", "password": "password"}`, 422)
	testReq(t, "Login(Wrong password)", "POST", "/auth", `{"name": "admin", "password": "wrongpassword"}`, 403)
	token := getToken(testReq(t, "Login", "POST", "/auth", `{"name": "admin", "password": "password"}`, 200))
	tokenq := "?access_token=" + token
	testReq(t, "Auth(No token)", "GET", "/api/private_widgets", "", 401)
	testReq(t, "Auth(Invalid token)", "GET", "/api/private_widgets?access_token=PleaseLetMeIn", "", 401)
	testReq(t, "Auth(jwt alg==none token)", "GET", "/api/private_widgets?access_token="+jwt_none, "", 401)
	testReq(t, "Auth(Token wrong signature)", "GET", "/api/private_widgets?access_token="+wrong_hs256, "", 401)
	testReq(t, "Auth(Token wrong algorithm)", "GET", "/api/private_widgets?access_token="+jwt_rsa, "", 401)
	disableGetUserById = true
	testReq(t, "Auth(User doesn't exist)", "GET", "/api/private_widgets"+tokenq, "", 401)
	disableGetUserById = false
	testReq(t, "Auth(Valid token)", "GET", "/api/private_widgets"+tokenq, "", 200)
}

// Extract jwt token from response
func getToken(body string) string {
	var f interface{}
	if err := json.Unmarshal([]byte(body), &f); err != nil {
		log.Panicf("Receieved malformed JSON body, %v\n", body)
	}
	m := f.(map[string]interface{})
	return m["token"].(string)
}
