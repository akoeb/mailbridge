package main

import (
	"strings"
	"errors"
)

type TokenResponse struct {
	Token   string
	Expires int64
}

// mapper
func ResponseObjectFromToken(token *Token) *TokenResponse {
	return &TokenResponse{
		Token:   token.String(),
		Expires: token.Expires.Unix(),
	}
}

type SendMailRequest struct {
	Token   string `json:"Token"`
	From    string `json:"From"`
	To      string `json:"To"`
	Subject string `json:"Subject"`
	Body    string `json:"Body"`
}
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

func MessageObjectFromRequest(request SendMailRequest) *Message {
	return &Message {
		from: request.From,
		recipientId: request.To,
		subject: request.Subject,
		body: request.Body,
	}
}