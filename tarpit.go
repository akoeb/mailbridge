package main

import (
	"net"
	"fmt"
	"net/http"
	"errors"
	"time"
	"log"
)

type Tarpit struct {
	tick int
	IPAddresses  map[string]int
}

func (tp *Tarpit) Wait(request *http.Request) error {
	ip, err := tp.getIP(request)
	if err != nil {
		return err
	}

	sleep, ok := tp.IPAddresses[ip]
	if ok {
		log.Printf("Sleeping for %d seconds", tp.IPAddresses[ip] * tp.tick)
		time.Sleep(time.Duration(sleep * tp.tick) * time.Second)
		tp.IPAddresses[ip] = sleep + 1

	} else {
		tp.IPAddresses[ip] = 1
	}
	tp.ScheduleDecrement(ip)
	return nil
}
func (tp *Tarpit) ScheduleDecrement(ip string) {
	ticker := time.NewTicker(time.Second * time.Duration(tp.tick))
	go func() {
		for range ticker.C {
			counter := tp.IPAddresses[ip]
			if counter > 0 {
				counter = counter - 1
				tp.IPAddresses[ip] = counter
				log.Printf("Decrement %s to %d", ip, counter)
			} else {
				log.Printf("Stopping ticker for %s", ip)
				delete(tp.IPAddresses, ip)
				ticker.Stop()
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
		tp.IPAddresses = make(map[string]int)
	}
	return tp
}

