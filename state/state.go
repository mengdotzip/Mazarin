package state

import (
	"net"
	"sync"
)

//State works for now but I want to move away from this in the long run

// Shared data across modules
var (
	Mutex          = sync.RWMutex{}
	WhitelistedIPs = make(map[string]bool)
	ActiveConns    = make(map[string][]net.Conn)
)
