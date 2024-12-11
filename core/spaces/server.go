package spaces

import (
	"sync"

	"github.com/pingostack/pingos/core/router"
)

type Server struct {
	routers map[string]*router.Router
	domains []string
	lock    sync.RWMutex
	id      string
}

func NewServer(id string, domains []string) *Server {
	return &Server{
		id:      id,
		routers: make(map[string]*router.Router),
		domains: domains,
	}
}

func (s *Server) AddRouter(name string, r *router.Router) {
	s.routers[name] = r
}

func (s *Server) GetRouter(name string) *router.Router {
	if r, ok := s.routers[name]; ok {
		return r
	}
	return nil
}

func (s *Server) DeleteRouter(name string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.routers, name)
}

func (s *Server) AddDomain(domain string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, d := range s.domains {
		if d == domain {
			return
		}
	}

	s.domains = append(s.domains, domain)
}

func (s *Server) DeleteDomain(domain string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for i, d := range s.domains {
		if d == domain {
			s.domains = append(s.domains[:i], s.domains[i+1:]...)
			return
		}
	}
}

func (s *Server) MatchDomain(domain string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, d := range s.domains {
		if d == domain {
			return true
		}
	}
	return false
}

func (s *Server) Close() {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, r := range s.routers {
		r.Close()
	}
}
