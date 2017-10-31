package main

import (
	"regexp"
	"testing"
	"time"
)

func TestActiveTokens_New(t *testing.T) {
	t.Parallel()

	tokenRegexp := `^[A-Z0-9]{8}\-[A-Z0-9]{4}\-[A-Z0-9]{4}\-[A-Z0-9]{4}\-[A-Z0-9]{12}$`

	// init config and active tokens
	lifetime := 2
	config := getConfig(lifetime, 5)
	activeTokens := InitActiveTokens(&config)

	// record time while running MUT
	startTime := time.Now()
	token, err := activeTokens.New()
	endTime := time.Now()

	// asserts
	if err != nil {
		t.Errorf("Error in getting new token: %v", err)
	}
	if token.Expires.Before(startTime) {
		t.Errorf("Error in token creation: Expiration date is in the past")

	}
	if token.Expires.Before(endTime) {
		t.Errorf("Error in token creation: Expiration date is NOW")
	}

	expectedExpirationStart := startTime.Add(time.Duration(lifetime) * time.Second)
	expectedExpirationEnd := endTime.Add(time.Duration(lifetime) * time.Second)
	if token.Expires.Before(expectedExpirationStart) {
		t.Errorf("Error in token creation: Expiration date is too early. Is: %v, expectedStart: %v", token.Expires, expectedExpirationStart)
	}
	if token.Expires.After(expectedExpirationEnd) {
		t.Errorf("Error in token creation: Expiration date is too late. Is: %v, expectedEnd: %v", token.Expires, expectedExpirationEnd)
	}

	// regex validate token
	Re := regexp.MustCompile(tokenRegexp)
	if !Re.MatchString(token.String()) {
		t.Errorf("Error in getting token: Regexp does not match: %v, token %v", tokenRegexp, token.String())
	}

	// validate this token is in the token map:
	exists, found := activeTokens.Tokens[token.String()]
	if !found {
		t.Errorf("Error in getting token: Token not in active tokens map")
	}
	if !compareToken(*token, *exists) {
		t.Errorf("Error in getting token: Token differs from active tokens map")
	}

	// get another token and make sure it is a different one:
	token2, err := activeTokens.New()
	if err != nil {
		t.Errorf("Error in getting second token: %v", err)
	}
	if compareToken(*token, *token2) {
		t.Errorf("Error in getting token: getting identical token a second time")
	}
	if token2.String() == token.String() {
		t.Errorf("Error: Getting two tokens with the same value: %v", token.String())

	}
}
func TestActiveTokens_Clean_TickerExpired(t *testing.T) {
	t.Parallel()

	// init config and active tokens, token expires before cleanup
	lifetime := 1
	config := getConfig(lifetime, 2)
	activeTokens := InitActiveTokens(&config)

	// get two tokens
	token1, err := activeTokens.New()
	if err != nil {
		t.Errorf("Error in getting token: %v", err)
	}
	token2, err := activeTokens.New()
	if err != nil {
		t.Errorf("Error in getting token: %v", err)
	}
	// validate they are recorded in the map
	if len(activeTokens.Tokens) != 2 {
		t.Errorf("Error: active tokens map should be 2 but is %v", len(activeTokens.Tokens))
	}
	if _, found := activeTokens.Tokens[token1.String()]; !found {
		t.Errorf("Error: did not find active token 1: %v", token1.String())
	}
	if _, found := activeTokens.Tokens[token2.String()]; !found {
		t.Errorf("Error: did not find active token 2: %v", token2.String())
	}

	// wait until after cleanup
	time.Sleep(2100 * time.Millisecond)

	// validate they are deleted
	if len(activeTokens.Tokens) != 0 {
		t.Errorf("Error: active tokens map should be 0 but is %v", len(activeTokens.Tokens))
	}
	if _, found := activeTokens.Tokens[token1.String()]; found {
		t.Errorf("Error: active token 1 still exists: %v", token1.String())
	}
	if _, found := activeTokens.Tokens[token2.String()]; found {
		t.Errorf("Error: active token 2 still exists %v", token2.String())
	}
}
func TestActiveTokens_Clean_Direct(t *testing.T) {
	t.Parallel()

	// init config and active tokens, token expires before cleanup
	lifetime := 1
	config := getConfig(lifetime, 5)
	activeTokens := InitActiveTokens(&config)

	// get two tokens
	token1, err := activeTokens.New()
	if err != nil {
		t.Errorf("Error in getting token: %v", err)
	}
	token2, err := activeTokens.New()
	if err != nil {
		t.Errorf("Error in getting token: %v", err)
	}
	// validate they are recorded in the map
	if len(activeTokens.Tokens) != 2 {
		t.Errorf("Error: active tokens map should be 2 but is %v", len(activeTokens.Tokens))
	}
	if _, found := activeTokens.Tokens[token1.String()]; !found {
		t.Errorf("Error: did not find active token 1: %v", token1.String())
	}
	if _, found := activeTokens.Tokens[token2.String()]; !found {
		t.Errorf("Error: did not find active token 2: %v", token2.String())
	}

	// wait until after cleanup
	time.Sleep(1100 * time.Millisecond)

	// then MUT
	deleted := activeTokens.Clean()

	// validate they are deleted
	if len(activeTokens.Tokens) != 0 {
		t.Errorf("Error: active tokens map should be 0 but is %v", len(activeTokens.Tokens))
	}
	if _, found := activeTokens.Tokens[token1.String()]; found {
		t.Errorf("Error: active token 1 still exists: %v", token1.String())
	}
	if _, found := activeTokens.Tokens[token2.String()]; found {
		t.Errorf("Error: active token 2 still exists %v", token2.String())
	}
	if deleted != 2 {
		t.Errorf("clean method returned %v but should have returned 2", deleted)
	}
}

func TestActiveTokens_Clean_TickerNonExpired(t *testing.T) {
	t.Parallel()

	// init config and active tokens, lifetime until after cleanup
	lifetime := 2
	config := getConfig(lifetime, 1)
	activeTokens := InitActiveTokens(&config)

	// get two tokens
	token1, err := activeTokens.New()
	if err != nil {
		t.Errorf("Error in getting token: %v", err)
	}
	token2, err := activeTokens.New()
	if err != nil {
		t.Errorf("Error in getting token: %v", err)
	}
	// validate they are recorded in the map
	if len(activeTokens.Tokens) != 2 {
		t.Errorf("Error: active tokens map should be 2 but is %v", len(activeTokens.Tokens))
	}
	if _, found := activeTokens.Tokens[token1.String()]; !found {
		t.Errorf("Error: did not find active token 1: %v", token1.String())
	}
	if _, found := activeTokens.Tokens[token2.String()]; !found {
		t.Errorf("Error: did not find active token 2: %v", token2.String())
	}

	// wait until after cleanup
	time.Sleep(1200 * time.Millisecond)

	// validate they are NOT deleted
	if len(activeTokens.Tokens) != 2 {
		t.Errorf("Error: active tokens map should be 2 but is %v", len(activeTokens.Tokens))
	}
	if _, found := activeTokens.Tokens[token1.String()]; !found {
		t.Errorf("Error: active token 1 has been cleaned up but should not: %v", token1.String())
	}
	if _, found := activeTokens.Tokens[token2.String()]; !found {
		t.Errorf("Error: active token 2 has been cleaned up but should not: %v", token2.String())
	}
}

func TestActiveTokens_Validate(t *testing.T) {
	t.Parallel()

	// init config and active tokens, lifetime until after cleanup
	lifetime := 2
	config := getConfig(lifetime, 1)
	activeTokens := InitActiveTokens(&config)

	// get a token
	token, err := activeTokens.New()
	if err != nil {
		t.Errorf("Error in getting token: %v", err)
	}
	// token must be available
	if _, found := activeTokens.Tokens[token.String()]; !found {
		t.Errorf("Error: active token not available: %v", token.String())
	}

	// run MUT
	if err := activeTokens.Validate(token.String()); err != nil {
		t.Errorf("Error: Token validation returned error: %v", err)
	}
	// now, token must not be available
	if _, found := activeTokens.Tokens[token.String()]; found {
		t.Errorf("Error: active token is available but should not: %v", token.String())
	}
}

func getConfig(lifetime int, cleanupInterval int) ApplicationConfig {
	config := &ApplicationConfig{}
	config.CleanupInterval = cleanupInterval
	config.Lifetime = lifetime
	return *config
}
func compareToken(a, b Token) bool {
	if &a == &b {
		return true
	}
	if a.Expires != b.Expires {
		return false
	}
	if a.a != b.a || a.b != b.b || a.c != b.c || a.d != b.d || a.e != b.e {
		return false
	}
	return true
}
