package dnscache

// Package dnscache caches DNS lookups

import (
	"net"
	"sync"
	"time"
)

type ipaccess struct {
	ips      []net.IP
	rr       uint
	accessed bool
}
type Resolver struct {
	lock  sync.RWMutex
	cache map[string]*ipaccess
}

func New(refreshRate time.Duration) *Resolver {
	resolver := &Resolver{
		cache: make(map[string]*ipaccess, 64),
	}
	if refreshRate > 0 {
		go resolver.autoRefresh(refreshRate)
	}
	return resolver
}

func (r *Resolver) Fetch(address string) (*ipaccess, error) {
	r.lock.RLock()
	e, exists := r.cache[address]
	r.lock.RUnlock()
	if exists {
		e.accessed = true
		return e, nil
	}

	return r.Lookup(address, true)
}

// FetchOne return an IP selected by round-robin
func (r *Resolver) FetchOne(address string) (net.IP, error) {
	e, err := r.Fetch(address)
	if err != nil || len(e.ips) == 0 {
		return nil, err
	}
	e.rr = (e.rr + 1) % uint(len(e.ips))
	return e.ips[e.rr], nil
}

func (r *Resolver) FetchOneString(address string) (string, error) {
	ip, err := r.FetchOne(address)
	if err != nil || ip == nil {
		return "", err
	}
	return ip.String(), nil
}

func (r *Resolver) Refresh() {
	entries := make(map[string]*ipaccess, len(r.cache))
	r.lock.RLock()
	for k, v := range r.cache {
		entries[k] = v
	}
	r.lock.RUnlock()

	for k, e := range entries {
		if e != nil && e.accessed {
			r.Lookup(k, false)
			time.Sleep(time.Second * 2)
		}
	}
}

func (r *Resolver) Remove() {
	r.lock.RLock()
	r.cache = nil
	r.lock.RUnlock()
}

func (r *Resolver) Lookup(address string, accessed bool) (*ipaccess, error) {
	ips, err := net.LookupIP(address)
	if err != nil {
		return nil, err
	}
	e := &ipaccess{
		ips:      ips,
		accessed: accessed,
	}
	r.lock.Lock()
	r.cache[address] = e
	r.lock.Unlock()
	return e, nil
}

func (r *Resolver) autoRefresh(rate time.Duration) {
	for {
		time.Sleep(rate)
		r.Refresh()
	}
}
