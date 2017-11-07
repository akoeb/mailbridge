package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"time"
)

// EmailMessage represents the mail to be sent
type EmailMessage struct {
	from        string
	subject     string
	body        string
	recipientID string
}

// MailServer is the object for all sending things
type MailServer struct {
	host         string
	port         string
	authUser     string
	authPassword string
	recipientMap map[string]string
}

// InitMailServer is the factory method to initialize a MailServer
func InitMailServer(config *ApplicationConfig) *MailServer {

	return &MailServer{
		host:         config.SMTPHost,
		port:         config.SMTPPort,
		authUser:     config.SMTPAuthUser,
		authPassword: config.SMTPAuthPassword,
		recipientMap: config.RecipientMap,
	}

}

// Send does the actual sending of the mail
func (server *MailServer) Send(mail *EmailMessage) error {

	// check that we are allowed to send email to this recipient
	// and we know who that is
	to, ok := server.recipientMap[mail.recipientID]
	if !ok {
		return fmt.Errorf("No email for id %v", mail.recipientID)
	}

	// construct the data block
	message := fmt.Sprintf("From: %s\r\n", mail.from)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	message += fmt.Sprintf("Subject: %s\r\n", mail.subject)
	message += "\r\n" + mail.body

	// setup Authentication and TLS Configuration
	auth := smtp.PlainAuth("", server.authUser, server.authPassword, server.host)

	// Gmail will reject connection if it's not secure
	// TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         server.host,
	}

	// Connect to the remote SMTP server.
	connStr := fmt.Sprintf("%s:%s", server.host, server.port)
	client, err := smtp.Dial(connStr)
	if err != nil {
		log.Printf("Dial\n")
		return err
	}

	client.StartTLS(tlsConfig)

	// step 1: Use Auth
	if err = client.Auth(auth); err != nil {
		log.Printf("Auth")
		return err
	}
	// Set the sender and recipient first
	if err := client.Mail(mail.from); err != nil {
		log.Printf("Mail")
		return err
	}
	if err := client.Rcpt(to); err != nil {
		log.Printf("Rcpt")
		return err
	}

	// Send the email body.
	wc, err := client.Data()
	if err != nil {
		log.Printf("Data")
		return err
	}
	_, err = fmt.Fprintf(wc, message)
	if err != nil {
		log.Printf("print body")
		return err
	}
	err = wc.Close()
	if err != nil {
		log.Printf("close")
		return err
	}

	// close the connection.
	err = client.Quit()
	if err != nil {
		log.Printf("quit")
		return err
	}
	log.Printf("Mail Sent: %v\n", to)
	return nil
}
