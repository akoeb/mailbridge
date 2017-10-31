package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"time"
)

// Token represents a unique one-time token with expiration that the client must request first and provide later.
type Token struct {
	a       [4]byte
	b       [2]byte
	c       [2]byte
	d       [2]byte
	e       [6]byte
	Expires time.Time
}

// String returns a string representation of the token without expiration datetime
func (token *Token) String() string {
	return fmt.Sprintf("%X-%X-%X-%X-%X", token.a, token.b, token.c, token.d, token.e)
}


// Init Initializes a new random token and sets the expiration to the provided lifetime parameter [seconds] in future
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

// ActiveTokens is an in memory structure that will hold all active tokens during application lifetime
// Beware: If this application dies, all active tokens die with it...
type ActiveTokens struct {
	Tokens          map[string]*Token
	lifetime        int
	cleanupInterval int
}


// New adds a new random token to the ActiveTokens struct and returns this token
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

// Validate will check that a provided token indeed is in the ActiveTokens map and is not expired.
// It will delete the token so it can not be used a second time.
// An error is returned if  something went wrong or the token did not exist or was expired.
// nil is returned if the token was valid.
func (at *ActiveTokens) Validate(key string) error {

	// check existence
	token, ok := at.Tokens[key]
	if !ok {
		return errors.New("token did not exist")
	}

	// it was found, whether or not it is expired, we will delete it anyway
	delete(at.Tokens, key)

	// check expiration
	if time.Now().After(token.Expires) {
		return errors.New("token already expired")
	}
	return nil
}


// Clean iterates all tokens in the map and deletes the expired ones
// this is called regularly by the ticker
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

// SetupTicker creates a ticker that calls Clean() in regular intervals (config.CleanupInterval)
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

// InitActiveTokens is the factory function to initialize the ActiveTokens map
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
