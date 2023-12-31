package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {

			if err := recover(); err != nil {

				w.Header().Set("Connection", "close")

				app.serverStatusError(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) reteLimit(next http.Handler) http.Handler {

	type client struct {
		limiter  *rate.Limiter
		lastseen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastseen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if app.Config.limiter.enabled {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
		app.serverStatusError(w, r, err)
		return
				}
		mu.Lock()
		if _, found := clients[ip]; !found {
		clients[ip] = &client{
	
		limiter: rate.NewLimiter(rate.Limit(app.Config.limiter.rps), app.Config.limiter.burst),
					}
				}
		clients[ip].lastseen = time.Now()
		if !clients[ip].limiter.Allow() {
		mu.Unlock()
		app.rateLimitExceededResponse(w, r)
		return
				}
		mu.Unlock()
			}
		next.ServeHTTP(w, r)
	})
}
