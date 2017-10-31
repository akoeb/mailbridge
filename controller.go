package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type controller struct {
	activeTokens *ActiveTokens
	mailServer   *MailServer
	tarpit       *Tarpit
}


func initController(m *MailServer, a *ActiveTokens, t *Tarpit) *controller {
	c := &controller{
		activeTokens: a,
		mailServer:   m,
		tarpit:       t,
	}
	return c
}

func (c *controller) GetToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

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
	o := ResponseObjectFromToken(token)
	response, _ := json.Marshal(&o)

	// Write content-type, status code, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s\n", response)
}
func (c *controller) SendMail(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
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
