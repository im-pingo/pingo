package spaces

import (
	"errors"
	"sync"
	"sync/atomic"
)

var serversTable map[string]*Server
var serversLock sync.RWMutex

var defaultServer *Server
var autoCreateServer atomic.Bool
var useDefaultServer atomic.Bool

const DefaultSpaceID = "--default--"

func init() {
	serversTable = make(map[string]*Server)
	defaultServer = NewServer(DefaultSpaceID, []string{"*"})
}

func SetAutoCreateServer(autoCreate bool) {
	autoCreateServer.Store(autoCreate)
}

func SetUseDefaultServer(useDefault bool) {
	useDefaultServer.Store(useDefault)
}

func GetOrCreateServerByDomain(domain string) (server *Server, exists bool) {
	serversLock.Lock()
	defer serversLock.Unlock()

	for _, s := range serversTable {
		if s.MatchDomain(domain) {
			return s, true
		}
	}

	if autoCreateServer.Load() {
		newServer := NewServer(domain, []string{domain})
		serversTable[domain] = newServer
		return newServer, false
	}

	if useDefaultServer.Load() {
		return defaultServer, true
	}

	return nil, false
}

func AddServer(serverID string, s *Server) error {
	serversLock.Lock()
	defer serversLock.Unlock()

	// check if domain is already in use
	for _, item := range serversTable {
		for _, domain := range s.domains {
			if item.MatchDomain(domain) {
				return errors.New("domain already in use")
			}
		}
	}

	// check if server already exists
	if _, exists := serversTable[serverID]; exists {
		return errors.New("server already exists")
	}

	serversTable[serverID] = s

	return nil
}

func GetServerByID(serverID string) *Server {
	serversLock.RLock()
	defer serversLock.RUnlock()

	if s, exists := serversTable[serverID]; exists {
		return s
	}

	if useDefaultServer.Load() {
		return defaultServer
	}

	return nil
}

func GetServerByDomain(domain string) *Server {
	serversLock.RLock()
	defer serversLock.RUnlock()

	for _, s := range serversTable {
		if s.MatchDomain(domain) {
			return s
		}
	}

	if useDefaultServer.Load() {
		return defaultServer
	}

	return nil
}

func deleteServer(serverID string) {
	if s, exists := serversTable[serverID]; exists {
		s.Close()
		delete(serversTable, serverID)
	}
}

func DeleteServerByID(serverID string) {
	serversLock.Lock()
	defer serversLock.Unlock()
	deleteServer(serverID)
}

func DeleteServerByDomain(domain string) {
	serversLock.Lock()
	defer serversLock.Unlock()
	for _, s := range serversTable {
		if s.MatchDomain(domain) {
			deleteServer(s.id)
			break
		}
	}
}
