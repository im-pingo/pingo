package main

import (
	"github.com/pingostack/pingos/core/spaces"
	"github.com/pingostack/pingos/pkg/logger"
)

func main() {
	err := spaces.AddServer("test", spaces.NewServer("test", []string{"test"}))
	if err != nil {
		logger.Fatal("add server to table failed", err)
	}

	err = spaces.AddServer("test", spaces.NewServer("test", []string{"test"}))
	if err == nil {
		logger.Fatal("add server to table should failed")
	}

	server := spaces.GetServerByDomain("test")
	if server == nil {
		logger.Fatal("get server from table failed")
	}

	server = spaces.GetServerByDomain("test2")
	if server != nil {
		logger.Fatal("get server from table should failed")
	}

	spaces.DeleteServerByDomain("test")
	server = spaces.GetServerByDomain("test")
	if server != nil {
		logger.Fatal("get server from table should failed")
	}

	spaces.SetUseDefaultServer(false)
	spaces.SetAutoCreateServer(true)
	server, exists := spaces.GetOrCreateServerByDomain("test3")
	if server == nil {
		logger.Fatal("get or create server[test3] should not be nil")
	}
	if exists {
		logger.Fatal("get or create server[test3] should not exists")
	}

	spaces.DeleteServerByDomain("test3")
	server = spaces.GetServerByDomain("test3")
	if server != nil {
		logger.Fatal("get server[test3] should be nil")
	}

	spaces.SetAutoCreateServer(false)
	server, exists = spaces.GetOrCreateServerByDomain("test4")
	if server != nil {
		logger.Fatal("get or create server[test4] should be nil")
	}
	if exists {
		logger.Fatal("get or create server[test4] should not exists")
	}

	spaces.SetUseDefaultServer(true)
	server, exists = spaces.GetOrCreateServerByDomain("test5")
	if server == nil {
		logger.Fatal("get or create server[test5] should not be nil")
	}
	if !exists {
		logger.Fatal("get or create server[test5] should exists")
	}

	spaces.DeleteServerByDomain("test5")
	server = spaces.GetServerByDomain("test5")
	if server == nil {
		logger.Fatal("get server[test5] should not be nil")
	}

	logger.Info("test success")
}
