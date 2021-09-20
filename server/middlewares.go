package server

import (
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"sync"
	"time"
)

type Visitor struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

type VLimiter struct {
	Visitors map[string]*Visitor
	mu       sync.Mutex
}

type VisitorsPurger struct {
	limiter     VLimiter
	off         chan struct{}
	done        chan struct{}
	timeout     time.Duration
	maxErrCount int
}

func SimpleAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		user, pass, _ := request.BasicAuth()

		if user != "pupa" || pass != "pupa" {
			writer.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		}

		next.ServeHTTP(writer, request)
	})
}

func (vl *VLimiter) GetVisitor(ip string) *rate.Limiter {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	visitor, ex := vl.Visitors[ip]
	if !ex {
		limit := rate.NewLimiter(5, 10)
		vl.Visitors[ip] = &Visitor{
			Limiter:  limit,
			LastSeen: time.Now(),
		}
		return limit

	}
	visitor.LastSeen = time.Now()
	return visitor.Limiter
}

func Limiter(next http.Handler, vl *VLimiter) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ip, _, err := net.SplitHostPort(request.RemoteAddr)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		lim := vl.GetVisitor(ip)
		if lim.Allow() == false {
			http.Error(writer, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func (vl *VLimiter) VisitorsCleaner() {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	for ip, v := range vl.Visitors {
		if time.Since(v.LastSeen) > 3*time.Minute {
			delete(vl.Visitors, ip)
		}
	}
}

func NewPurger(vl VLimiter, timeout time.Duration, maxErrCount int) *VisitorsPurger {
	return &VisitorsPurger{
		limiter:     vl,
		off:         make(chan struct{}, 1),
		done:        make(chan struct{}, 1),
		timeout:     timeout,
		maxErrCount: maxErrCount,
	}
}

func (p *VisitorsPurger) Purge() {
	t := time.NewTicker(p.timeout)
	go func() {
		defer func() {
			t.Stop()
			close(p.done)
		}()

		for {
			select {
			case <-t.C:
				p.limiter.VisitorsCleaner()
			case <-p.off:
				return
			}
		}
	}()
}

func (p *VisitorsPurger) Stop() chan<- struct{} {
	return p.off
}

func (p *VisitorsPurger) Done() <-chan struct{} {
	return p.done
}
