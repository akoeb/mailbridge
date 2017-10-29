package main

import (
	"net"
	"fmt"
	"net/http"
	"errors"
	"time"
	"log"
	"sync"
)
type TarpitValue struct {
	counter int
	expires time.Time
}
type Tarpit struct {
	tick int
	IPAddresses  map[string]*TarpitValue
	sync.RWMutex
}

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
	if !found {
		value = &TarpitValue{}
		tp.IPAddresses[ip] = value
	}
	sleep := value.counter
	value.counter ++
	value.expires = time.Now().Add(time.Duration(tp.tick * value.counter) * time.Second)
	tp.Unlock()
	log.Printf("Incremented counter for ip %s to %d, expires %v", ip, sleep + 1, value.expires)
	// now do the actual sleep
	if sleep > 0 {
		log.Printf("Sleeping for %d seconds", sleep * tp.tick)
		time.Sleep(time.Duration(sleep * tp.tick) * time.Second)
	}
	// done
	return nil
}
func (tp *Tarpit) Decrement() int {
	i := 0
	// this sort of global lock might lead to scaling issues at some
	// point. It would be wise to monitor execution time of this method
	tp.Lock()
	for ip, value := range tp.IPAddresses {
		if time.Now().After(value.expires) {
			i++

			// decrement counter
			value.counter -= 1

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

// schedule that the counter vor an IP address decrements over time
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

func (tp *Tarpit) getIP(request *http.Request) (string, error) {
	// This will only be defined when site is accessed via non-anonymous proxy
	// and takes precedence over RemoteAddr
	// Header.Get is case-insensitive
	var ip string
	ip = request.Header.Get("X-Forwarded-For")
	if ip == "" {
		remoteIp, _, err := net.SplitHostPort(request.RemoteAddr)
		if err != nil {
			return "", errors.New(fmt.Sprintf("userip: %q is not IP:port", request.RemoteAddr))
		}
		ip = remoteIp
	}

	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return "", errors.New("Could not parse client IP: " + ip)
	}
	return clientIP.String(), nil
}

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

