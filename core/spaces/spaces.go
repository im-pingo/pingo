package spaces

import (
	"sync"
)

var serversTable map[string]*Server
var serversLock sync.RWMutex
var domain2Server map[string]*Server

var defaultServer *Server
var autoCreateServer bool

const DefaultSpaceID = "--default--"

func init() {
	serversTable = make(map[string]*Server)
	domain2Server = make(map[string]*Server)
	defaultServer = NewServer(DefaultSpaceID)
}

func SetAutoCreateServer(autoCreate bool) {
	autoCreateServer = autoCreate
}

func GetOrCreateServerByDomain(domain string) *Server {
	serversLock.Lock()
	defer serversLock.Unlock()

	if s, exists := domain2Server[domain]; exists {
		return s
	}

	if autoCreateServer {
		newServer := NewServer(domain)
		newServer.AddDomain(domain)
		serversTable[domain] = newServer
		domain2Server[domain] = newServer
		return newServer
	}

	return defaultServer
}

func AddServerToTable(serverID string, s *Server) {
	serversLock.Lock()
	defer serversLock.Unlock()

	serversTable[serverID] = s
}

func GetServerByID(serverID string) *Server {
	serversLock.RLock()
	defer serversLock.RUnlock()

	if s, exists := serversTable[serverID]; exists {
		return s
	}
	return nil
}

func GetServerByDomain(domain string) *Server {
	serversLock.RLock()
	defer serversLock.RUnlock()

	if s, exists := domain2Server[domain]; exists {
		return s
	}

	if !autoCreateServer {
		return defaultServer
	}
	return nil
}

func DeleteServerFromTable(serverID string) {
	serversLock.Lock()
	defer serversLock.Unlock()

	delete(serversTable, serverID)
}

func IsDomainInMultipleServers(domain string) bool {
	serversLock.RLock()
	defer serversLock.RUnlock()

	count := 0
	for _, s := range serversTable {
		for _, d := range s.domains {
			if d == domain {
				count++
				if count > 1 {
					return true
				}
			}
		}
	}
	return false
}
