package main

import (
	"errors"
	"strings"
)

// TokenResponse represents the response object returned by the token endpoint
type TokenResponse struct {
	Token   string
	Expires int64
}

// ResponseObjectFromToken is a mapper method that returns a TokenResponse for the HTTP endpoint from a given Token
func ResponseObjectFromToken(token *Token) *TokenResponse {
	return &TokenResponse{
		Token:   token.String(),
		Expires: token.Expires.Unix(),
	}
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
