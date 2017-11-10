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

	// initialize mocks
	ms := &MockMailServer{}
	at := &MockActiveTokens{}
	tp := &MockTarpit{}

	// define the mock method for New
	at.mockNew = func() (*Token, error) {
		// record the call
		at.calledNew++

		// create new token
		token := &Token{}

		// new token with lifetime of 3 secs
		err := token.Init(3)
		if err != nil {
			return nil, err
		}
		at.lastToken = token
		return token, nil
	}

	// Do the request
	req, _ := http.NewRequest("GET", "/token", nil)
	rr := doRequest(req, ms, at, tp)

	// check status
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Wrong status: %d, should be %d", status, http.StatusCreated)
	}
	// content type
	if cType := rr.Header().Get("Content-Type"); cType != "application/json" {
		t.Errorf("content type header does not match: got %v want %v", cType, "application/json")
	}
	// make sure that the New() Method has been called on the activeTokens mock
	if at.calledNew != 1 {
		t.Errorf("method New() on ActiveTokens has been called wrongly. expected %d, got %d", 1, at.calledNew)
	}
	// make sure that the Wait() Method has been called on the tarpit mock
	if tp.calledWait != 1 {
		t.Errorf("method Wait() on Tarpit has been called wrongly. expected %d, got %d", 1, tp.calledWait)
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
	if response.Expires != at.lastToken.Expires.Unix() {
		t.Errorf("Invalid expiration in token response, expect: %v, got: %v",
			at.lastToken.Expires.Unix(), response.Expires)
	}
	if response.Token != at.lastToken.String() {
		t.Errorf("Invalid token string in token response, expect: %v, got: %v",
			at.lastToken.Expires.Unix(), response.Expires)
	}
}

// edge case New() returns error
func TestController_GetTokenNewError(t *testing.T) {

	// initialize mocks
	ms := &MockMailServer{}
	at := &MockActiveTokens{}
	tp := &MockTarpit{}

	// define the mock method for New
	at.mockNew = func() (*Token, error) {
		return nil, fmt.Errorf("returning error from New()")
	}

	// Do the request
	req, _ := http.NewRequest("GET", "/token", nil)
	rr := doRequest(req, ms, at, tp)

	// check status
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Wrong status: %d, should be %d", status, http.StatusBadRequest)
	}
	if strings.TrimSpace(rr.Body.String()) != "ERROR" {
		t.Errorf("GetTokenWaitError: Expected response body: %s, got %s", "ERROR", rr.Body.String())
	}

}

// TODO * Wait() returns error
// TODO * marshalling or response object creation fails

func TestController_SendMail_OK(t *testing.T) {
	msg := string(`{"Token": "TOKEN","From": "FROM", "To": "TO", "Subject": "SUBJECT", "Body": "BODY"}`)
	req, _ := http.NewRequest("POST", "/send", strings.NewReader(msg))
	rr := doRequestDefault(req)
	// check status
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Wrong status: %d, should be %d", status, http.StatusCreated)
	}
}
func TestController_SendMail_InvalidJson(t *testing.T) {
	msg := string(`"wrong":"json"`)
	req, _ := http.NewRequest("POST", "/send", strings.NewReader(msg))
	rr := doRequestDefault(req)
	// check status
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Wrong status: %d, should be %d", status, http.StatusBadRequest)
	}
}
func TestController_SendMail_NilBody(t *testing.T) {
	req, _ := http.NewRequest("POST", "/send", nil)
	rr := doRequestDefault(req)
	// check status
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Wrong status: %d, should be %d", status, http.StatusBadRequest)
	}
}
func TestController_SendMail_EmptyBody(t *testing.T) {
	req, _ := http.NewRequest("POST", "/send", strings.NewReader(""))
	rr := doRequestDefault(req)
	// check status
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Wrong status: %d, should be %d", status, http.StatusBadRequest)
	}
}

// TODO other tests:
// send mail with body size too big
// send mail with json fields missing
// send mail with invalid token
// send mail with error in mail sending

// ************************************************************************** //

// HELPER METHODS
func doRequestDefault(req *http.Request) *httptest.ResponseRecorder {
	// empty mocks to initialize controller
	ms := &MockMailServer{}
	at := &MockActiveTokens{}
	tp := &MockTarpit{}

	return doRequest(req, ms, at, tp)
}
func doRequest(req *http.Request, ms MailServerInterface, at ActiveTokensInterface, tp TarpitInterface) *httptest.ResponseRecorder {
	c := InitController(ms, at, tp)
	router := httprouter.New()
	router.GET("/token", c.GetToken)
	router.POST("/send", c.SendMail)

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	return rr
}
