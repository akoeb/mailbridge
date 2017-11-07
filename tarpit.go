package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// TarpitInterface for being able to mock Tarpit
type TarpitInterface interface {
	Wait(request *http.Request) error
	Decrement() int
	SetupTicker()
	getIP(request *http.Request) (string, error)
}

// TarpitValue is the Value of the synced map used in tarpit. The counter represents the
// number of times that the Wait Method has been called within tick period, and expires
// represents the datetime in future when we start decrementing the counter
type TarpitValue struct {
	counter int
	expires time.Time
}

// Tarpit is used to slow down processing for users who call some method too often.
type Tarpit struct {
	tick        int
	IPAddresses map[string]*TarpitValue
	sync.RWMutex
}

// Wait does the actual counter increment and wait.
// The first call within a specified amount of time, this method will return immediately,
// with every subsequent call however, it will wait longer and longer before returning.
func (tp *Tarpit) Wait(request *http.Request) error {
	// get client ip
	ip, err := tp.getIP(request)
	if err != nil {
		return err
	}
	// is the ip already registered?
	// this must be synced for thread safety
	tp.Lock()
	value, found := tp.IPAddresses[ip]
	// the first call, register the IP
	if !found {
		value = &TarpitValue{}
		tp.IPAddresses[ip] = value
	}
	// how long to sleep depends on the actual counter value
	sleep := value.counter
	// increment counter for the next time
	value.counter++
	// set expiration date, decrement will start only AFTER expiration
	value.expires = time.Now().Add(time.Duration(tp.tick*value.counter) * time.Second)

	// writing to the map is done
	tp.Unlock()
	log.Printf("Incremented counter for ip %s to %d, expires %v", ip, sleep+1, value.expires)

	// now do the actual sleep
	if sleep > 0 {
		log.Printf("Sleeping for %d seconds", sleep*tp.tick)
		time.Sleep(time.Duration(sleep*tp.tick) * time.Second)
	}

	// done
	return nil
}

// Decrement is called by a ticker and iterates over all entries of the map. It fill find the ones
// with expiration in the past and starts decrementing them, one decrement per tick.
func (tp *Tarpit) Decrement() int {
	i := 0
	// this sort of global lock might lead to scaling issues at some
	// point. It would be wise to monitor execution time of this method
	tp.Lock()
	for ip, value := range tp.IPAddresses {
		if time.Now().After(value.expires) {
			i++

			// decrement counter
			value.counter--

			// delete if counter is now 0:
			if value.counter <= 0 {
				log.Printf("Delete entry for IP %s", ip)
				delete(tp.IPAddresses, ip)
			} else {
				log.Printf("Decrement %s to %d", ip, value.counter)
			}
		}
	}
	tp.Unlock()
	return i
}

// SetupTicker schedules that the counter for an IP address decrements over time
func (tp *Tarpit) SetupTicker() {
	ticker := time.NewTicker(time.Second * time.Duration(tp.tick))
	go func() {
		for t := range ticker.C {
			startTime := time.Now()
			decremented := tp.Decrement()
			runtime := time.Since(startTime)
			if decremented > 0 {
				log.Printf("[%s] Processed %d entries in %v seconds", t, decremented, runtime.Seconds())
			}

		}
	}()
}

// Helper method to get the IP address of a client from either the X_FORWARDED_FOR header or the actual
// clients IP address.
func (tp *Tarpit) getIP(request *http.Request) (string, error) {
	// This will only be defined when site is accessed via non-anonymous proxy
	// and takes precedence over RemoteAddr
	// Header.Get is case-insensitive
	var ip string
	ip = request.Header.Get("X-Forwarded-For")
	if ip == "" {
		remoteIP, _, err := net.SplitHostPort(request.RemoteAddr)
		if err != nil {
			return "", fmt.Errorf("userip: %q is not IP:port", request.RemoteAddr)
		}
		ip = remoteIP
	}

	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return "", errors.New("Could not parse client IP: " + ip)
	}
	return clientIP.String(), nil
}

// InitTarpit is the factory function to return a Tarpit
func InitTarpit(config *ApplicationConfig) *Tarpit {
	tp := &Tarpit{
		tick: config.TarpitInterval,
	}
	if tp.IPAddresses == nil {
		tp.IPAddresses = make(map[string]*TarpitValue)
	}
	tp.SetupTicker()
	return tp
}
