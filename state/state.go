package state

import (
	"net"
	"sync"
)

// Shared data across modules
var (
	Mutex          = sync.RWMutex{}
	WhitelistedIPs = make(map[string]bool)
	ActiveConns    = make(map[string][]net.Conn)
)
