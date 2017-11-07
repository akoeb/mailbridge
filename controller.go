package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// Controller is a object that holds all our handlers and their dependencies
type Controller struct {
	activeTokens ActiveTokensInterface
	mailServer   MailServerInterface
	tarpit       TarpitInterface
	bodyLimit    int64
}

// InitController is the factory method for the controller
func InitController(m MailServerInterface, a ActiveTokensInterface, t TarpitInterface) *Controller {
	c := &Controller{
		activeTokens: a,
		mailServer:   m,
		tarpit:       t,
		bodyLimit:    1048576,
	}
	return c
}

// GetToken is the handler for the /token endpoint
func (c *Controller) GetToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	// tarpit the client if we had an earlier request:
	err := c.tarpit.Wait(r)
	if err != nil {
		log.Printf("ERROR tarpitting user: %v", err)
		http.Error(w, "ERROR", http.StatusBadRequest)
		return
	}

	token, err := c.activeTokens.New()
	if err != nil {
		log.Printf("ERROR Token Creation: %v", err)
		http.Error(w, "ERROR", http.StatusBadRequest)
		return
	}

	// Marshal provided interface into JSON structure
	o, err := ResponseObjectFromToken(token)
	if err != nil {
		log.Printf("ERROR Token Creation: %v", err)
		http.Error(w, "ERROR", http.StatusBadRequest)
		return
	}
	response, _ := json.Marshal(&o)

	// Write content-type, status code, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s\n", response)
}

// SendMail is the Handler for the /send endpoint
func (c *Controller) SendMail(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, c.bodyLimit))
	if err != nil {
		log.Printf("ERROR ReadBodyStream: %v", err)
		http.Error(w, "ERROR", http.StatusBadRequest)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Printf("ERROR CloseBodyStream: %v", err)
		http.Error(w, "ERROR", http.StatusBadRequest)
		return
	}
	var request SendMailRequest
	if err := json.Unmarshal(body, &request); err != nil {
		log.Printf("ERROR InvalidSendMailRequest: %v", err)
		http.Error(w, "ERROR", http.StatusBadRequest)
		return
	}
	// Input Validation
	if err := request.Validate(); err != nil {
		log.Printf("ERROR Failed Validation: %v", err)
		http.Error(w, "ERROR", http.StatusBadRequest)
		return
	}
	// validate token:
	if err := c.activeTokens.Validate(request.Token); err != nil {
		log.Printf("ERROR Invalid Token %v: %v", request.Token, err)
		http.Error(w, "ERROR", http.StatusBadRequest)
		return
	}
	if err := c.mailServer.Send(MessageObjectFromRequest(request)); err != nil {
		log.Printf("ERROR Mail Sending: %v", err)
		http.Error(w, "ERROR", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Request and Response Objects

// TokenResponse represents the response object returned by the token endpoint
type TokenResponse struct {
	Token   string
	Expires int64
}

// ResponseObjectFromToken is a mapper method that returns a TokenResponse for the HTTP endpoint from a given Token
func ResponseObjectFromToken(token *Token) (*TokenResponse, error) {
	if token == nil {
		return nil, fmt.Errorf("No Token created")
	}
	return &TokenResponse{
		Token:   token.String(),
		Expires: token.Expires.Unix(),
	}, nil
}

// SendMailRequest represents the accepted structure that clients send to the send endpoint
type SendMailRequest struct {
	Token   string `json:"Token"`
	From    string `json:"From"`
	To      string `json:"To"`
	Subject string `json:"Subject"`
	Body    string `json:"Body"`
}

// Validate checks whether all required fields are set
func (in *SendMailRequest) Validate() error {
	var msg []string

	if in.Token == "" {
		msg = append(msg, "Token must be set")
	}
	if in.From == "" {
		msg = append(msg, "From must be set")
	}
	if in.To == "" {
		msg = append(msg, "To must be set")
	}
	if in.Subject == "" {
		msg = append(msg, "Subject must be set")
	}
	if in.Body == "" {
		msg = append(msg, "Message Body must be set")
	}
	if len(msg) > 0 {
		return errors.New(strings.Join(msg, "\n"))
	}
	return nil
}

// MessageObjectFromRequest is a mapper method that returns a email message object from a given SendMailRequest
func MessageObjectFromRequest(request SendMailRequest) *EmailMessage {
	return &EmailMessage{
		from:        request.From,
		recipientID: request.To,
		subject:     request.Subject,
		body:        request.Body,
	}
}
