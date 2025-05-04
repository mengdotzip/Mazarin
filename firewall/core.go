package firewall

import (
	"mazarin/state"
	"net"
)

func CheckWhitelist(ip string, conn net.Conn) bool {
	state.Mutex.Lock()
	allowed := state.WhitelistedIPs[ip]
	if allowed {
		state.ActiveConns[ip] = append(state.ActiveConns[ip], conn)
		state.Mutex.Unlock()
		return true
	}
	state.Mutex.Unlock()
	return false
}
