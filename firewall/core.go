package firewall

import (
	"log"
	"mazarin/state"
	"net"
)

func CheckWhitelistAddConn(ip string, conn net.Conn) bool {
	state.Mutex.Lock()
	defer state.Mutex.Unlock()
	allowed := state.WhitelistedIPs[ip]
	if allowed {
		state.ActiveConns[ip] = append(state.ActiveConns[ip], conn)
		return true
	}
	return false
}

func CheckWhitelist(ip string) bool {
	state.Mutex.RLock()
	allowed := state.WhitelistedIPs[ip]
	state.Mutex.RUnlock()

	if allowed {
		log.Printf("FIREWALL: Authorized connection from IP %v", ip)
		return true
	}
	return false
}
