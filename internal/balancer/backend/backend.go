package backend

import (
	"errors"
	"log"
	"net"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	URL          *url.URL
	alive        bool
	rwmu         sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func NewBackend(url *url.URL, alive bool, proxy *httputil.ReverseProxy) *Backend {
	return &Backend{
		URL:          url,
		alive:        alive,
		rwmu:         sync.RWMutex{},
		ReverseProxy: proxy,
	}
}

// ISBackendAlive проверяет доступность backend
func (b *Backend) IsBackendAlive() error {
	timeout := 2 * time.Second
	log.Printf("Checking availability of backend %s", b.URL.String())
	conn, err := net.DialTimeout("tcp", b.URL.Host, timeout)
	if err != nil {
		log.Printf("Backend %s is unavailable, error: %v", b.URL.String(), err)
		return errors.New("site is unavailable")
	}
	_ = conn.Close()
	return nil
}

// setAlive устанавливает доступность backend в alive true/false
func (b *Backend) SetAlive(alive bool) {
	b.rwmu.Lock()
	b.alive = alive
	b.rwmu.Unlock()
}

// isAlive возвращает состояние backend
func (b *Backend) IsAlive() bool {
	b.rwmu.RLock()
	defer b.rwmu.RUnlock()
	return b.alive
}
