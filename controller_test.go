package main

import (
	"net/http"
	"net/http/httptest"
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
}

func (at *MockActiveTokens) New() (*Token, error) {
	// TODO write test that checks what happens if token is nil

	// record the call
	at.calledNew++

	// create new token
	token := &Token{}

	// new token with lifetime of 3 secs
	err := token.Init(3)
	if err != nil {
		return nil, err
	}
	return token, nil
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
}

func (tp *MockTarpit) Wait(request *http.Request) error {
	return nil
}
func (tp *MockTarpit) Decrement() int {
	return 0
}
func (tp *MockTarpit) SetupTicker() {}
func (tp *MockTarpit) getIP(request *http.Request) (string, error) {
	return "", nil
}

func TestController_GetToken(t *testing.T) {

	mockMailServer := &MockMailServer{}
	mockActiveTokens := &MockActiveTokens{}
	mockTarpit := &MockTarpit{}

	c := InitController(mockMailServer, mockActiveTokens, mockTarpit)
	router := httprouter.New()
	router.GET("/token", c.GetToken)

	req, _ := http.NewRequest("GET", "/token", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// TODO make sure following mocked dependencies have been called:
	// err = tarpit.Wait(req)

	// check status
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Wrong status: %d", status)
	}
	// content type
	if cType := rr.Header().Get("Content-Type"); cType != "application/json" {
		t.Errorf("content type header does not match: got %v want %v", cType, "application/json")
	}
	// make sure that the New() Method has been called on the activeTokens mock
	if mockActiveTokens.calledNew != 1 {
		t.Errorf("method New() on ActiveTokens has been called wrongly. expected %d, got %d", 1, mockActiveTokens.calledNew)
	}
	// TODO compare body with token returned by activeToken Mock
}
func TestController_SendMail(t *testing.T) {
	// TODO needs mocks: active tokens to validate token and mailserver to send mail
}

// TODO other tests:
// send mail with body size too big
// send mail with wrong json
// send mail with json fields missing
// send mail with invalid token
// send mail with error in mail sending
