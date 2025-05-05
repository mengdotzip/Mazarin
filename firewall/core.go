package firewall

import (
	"mazarin/state"
	"net"
)

func CheckWhitelist(ip string, conn net.Conn) bool {
	state.Mutex.Lock()
	defer state.Mutex.Unlock()
	allowed := state.WhitelistedIPs[ip]
	if allowed {
		state.ActiveConns[ip] = append(state.ActiveConns[ip], conn)
		return true
	}
	return false
}
