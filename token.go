package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"time"
)

/**
Token represents a unique token that the client must request first and provide later
after expiration, the token can not be used any more
*/
type Token struct {
	a       [4]byte
	b       [2]byte
	c       [2]byte
	d       [2]byte
	e       [6]byte
	Expires time.Time
}

/**
String representation of the token
*/
func (token *Token) String() string {
	return fmt.Sprintf("%X-%X-%X-%X-%X", token.a, token.b, token.c, token.d, token.e)
}

/**
Initialize a new random token and set expiration
*/
func (token *Token) Init(lifetime int) error {

	// identifier
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return err
	}

	copy(token.a[:], buf[0:4])
	copy(token.b[:], buf[4:6])
	copy(token.c[:], buf[6:8])
	copy(token.d[:], buf[8:10])
	copy(token.e[:], buf[10:])

	// expiration
	token.Expires = time.Now().Add(time.Duration(lifetime) * time.Second)
	return nil
}

/**
This is an in memory structure that will hold all active tokens

Beware: If this application dies, all active tokens die with it...
*/
type ActiveTokens struct {
	Tokens          map[string]*Token
	lifetime        int
	cleanupInterval int
}

/**
add and return a new random token
*/
func (at *ActiveTokens) New() (*Token, error) {

	// create new token
	token := &Token{}
	err := token.Init(at.lifetime)
	if err != nil {
		return nil, err
	}
	// get string representation
	key := token.String()

	// check whether the Token exists already:
	_, exists := at.Tokens[key]
	if exists {
		return nil, errors.New("freshly initialized key exists already in the table")
	}

	// add to map
	at.Tokens[key] = token
	return token, nil
}

/**
Get and delete a token from the map of active tokens
will return error if the token did not exist or is expired
will return nil if all is fine
*/
func (at *ActiveTokens) Validate(key string) error {
	token, ok := at.Tokens[key]
	if !ok {
		return errors.New("token did not exist")
	}
	if time.Now().After(token.Expires) {
		return errors.New("token already expired")
	}
	delete(at.Tokens, key)
	return nil
}

/**
Iterate all tokens in the map and delete the expired ones
*/
func (at *ActiveTokens) Clean() int {
	i := 0
	for key, token := range at.Tokens {
		if time.Now().After(token.Expires) {
			delete(at.Tokens, key)
			i++
		}
	}
	return i
}

func (at *ActiveTokens) SetupTicker() {
	// create a ticker to clean up expired tokens in regular intervals
	ticker := time.NewTicker(time.Second * time.Duration(at.cleanupInterval))
	go func() {
		for t := range ticker.C {
			deleted := at.Clean()
			if deleted > 0 {
				log.Printf("[%s] Cleaning up %d active tokens", t, deleted)
			}
		}
	}()
}

/**
Factory function to initialize the ActiveTokens map
*/
func InitActiveTokens(config *ApplicationConfig) *ActiveTokens {
	at := &ActiveTokens{
		lifetime:        config.Lifetime,
		cleanupInterval: config.CleanupInterval,
	}
	if at.Tokens == nil {
		at.Tokens = make(map[string]*Token)
	}
	at.SetupTicker()
	return at
}
