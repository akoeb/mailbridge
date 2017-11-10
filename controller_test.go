package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

// create mock object for mailServer
type MockMailServer struct {
}

func (ms *MockMailServer) Send(m *EmailMessage) error {
	return nil
}

// create mock object for activeTokens
type MockActiveTokens struct {
	calledNew int
	lastToken *Token
	mockNew   func() (*Token, error)
}

// New delegates to mockable
func (at *MockActiveTokens) New() (*Token, error) {
	if at.mockNew != nil {
		return at.mockNew()
	}
	return nil, fmt.Errorf("No mocked function found")
}
func (at *MockActiveTokens) Validate(key string) error {
	return nil
}
func (at *MockActiveTokens) Clean() int {
	return 0
}
func (at *MockActiveTokens) SetupTicker() {

}

// create mock object for Tarpit
type MockTarpit struct {
	calledWait int
}

func (tp *MockTarpit) Wait(request *http.Request) error {
	tp.calledWait++
	return nil
}
func (tp *MockTarpit) Decrement() int {
	return 0
}
func (tp *MockTarpit) SetupTicker() {}
func (tp *MockTarpit) getIP(request *http.Request) (string, error) {
	return "", nil
}

// ****************************************************************** //
//                Now we start the actual tests
// ****************************************************************** //

// test the positive case
func TestController_GetTokenOK(t *testing.T) {

	mockMailServer := &MockMailServer{}
	mockActiveTokens := &MockActiveTokens{}
	// define the mock method for New
	mockActiveTokens.mockNew = func() (*Token, error) {
		// record the call
		mockActiveTokens.calledNew++

		// create new token
		token := &Token{}

		// new token with lifetime of 3 secs
		err := token.Init(3)
		if err != nil {
			return nil, err
		}
		mockActiveTokens.lastToken = token
		return token, nil
	}

	mockTarpit := &MockTarpit{}

	c := InitController(mockMailServer, mockActiveTokens, mockTarpit)
	router := httprouter.New()
	router.GET("/token", c.GetToken)

	req, _ := http.NewRequest("GET", "/token", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// check status
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Wrong status: %d, should be %d", status, http.StatusCreated)
	}
	// content type
	if cType := rr.Header().Get("Content-Type"); cType != "application/json" {
		t.Errorf("content type header does not match: got %v want %v", cType, "application/json")
	}
	// make sure that the New() Method has been called on the activeTokens mock
	if mockActiveTokens.calledNew != 1 {
		t.Errorf("method New() on ActiveTokens has been called wrongly. expected %d, got %d", 1, mockActiveTokens.calledNew)
	}
	// make sure that the Wait() Method has been called on the tarpit mock
	if mockTarpit.calledWait != 1 {
		t.Errorf("method Wait() on Tarpit has been called wrongly. expected %d, got %d", 1, mockTarpit.calledWait)
	}
	// TODO compare body with token returned by activeToken Mock
	body, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Errorf("Error in reading token response: %v", err)
	}
	var response TokenResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Errorf("Error in un-marshalling token response: %v", err)
	}
	if response.Expires != mockActiveTokens.lastToken.Expires.Unix() {
		t.Errorf("Invalid expiration in token response, expect: %v, got: %v",
			mockActiveTokens.lastToken.Expires.Unix(), response.Expires)
	}
	if response.Token != mockActiveTokens.lastToken.String() {
		t.Errorf("Invalid token string in token response, expect: %v, got: %v",
			mockActiveTokens.lastToken.Expires.Unix(), response.Expires)
	}
}

// edge case Wait() returns error
func TestController_GetTokenWaitError(t *testing.T) {

	mockMailServer := &MockMailServer{}
	mockActiveTokens := &MockActiveTokens{}
	// define the mock method for New
	mockActiveTokens.mockNew = func() (*Token, error) {
		return nil, fmt.Errorf("returning error from New()")
	}

	mockTarpit := &MockTarpit{}

	c := InitController(mockMailServer, mockActiveTokens, mockTarpit)
	router := httprouter.New()
	router.GET("/token", c.GetToken)

	req, _ := http.NewRequest("GET", "/token", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// check status
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Wrong status: %d, should be %d", status, http.StatusBadRequest)
	}
	if strings.TrimSpace(rr.Body.String()) != "ERROR" {
		t.Errorf("GetTokenWaitError: Expected response body: %s, got %s", "ERROR", rr.Body.String())
	}

}

// TODO * New() returns error
// TODO * marshalling or response object creation fails

func TestController_SendMail(t *testing.T) {
	// TODO needs mocks: active tokens to validate token and mailserver to send mail
}

// TODO other tests:
// send mail with body size too big
// send mail with wrong json
// send mail with json fields missing
// send mail with invalid token
// send mail with error in mail sending
