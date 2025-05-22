package webserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mazarin/config"
	"mazarin/firewall"
	"mazarin/state"
	"net"
	"net/http"
	"time"
)

var userData *UsersData

type AuthRequest struct {
	Username string `json:"username"`
	Key      string `json:"key"`
}

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("WEBSERVER: Failed to parse client IP: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("WEBSERVER: IP %v contacted /auth", clientIP)

	var authReq AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&authReq); err != nil {
		log.Printf("WEBSERVER: Invalid request body from IP %v: %v", clientIP, err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if !firewall.ValidateInput(authReq.Username, "username") {
		log.Printf("WEBSERVER: Client IP %v invalid username characters", clientIP)
		http.Error(w, "Invalid characters in input", http.StatusBadRequest)
		return
	}
	if !firewall.ValidateInput(authReq.Key, "password") {
		log.Printf("WEBSERVER: Client IP %v invalid password characters", clientIP)
		http.Error(w, "Invalid characters in input", http.StatusBadRequest)
		return
	}

	//TODO change this to a map
	authenticated := false
	for _, users := range userData.Users {
		if authReq.Username == users.Name {
			auth, err := validateUserHash(authReq.Key, users.Hash)
			if err != nil {
				log.Printf("WEBSERVER: Input validation failed from IP %v: %v", clientIP, err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if auth {
				authenticated = true
				break
			} else {
				log.Printf("WEBSERVER: Invalid login from IP %v with username %v", clientIP, users.Name)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
	}

	if !authenticated {
		log.Printf("WEBSERVER: User not found: %v from IP %v", authReq.Username, clientIP)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("WEBSERVER: Successful auth for %v from %v", authReq.Username, clientIP)

	state.Mutex.Lock()
	state.WhitelistedIPs[clientIP] = true
	state.Mutex.Unlock()
	log.Printf("WEBSERVER: IP %v got whitelisted in the firewall", clientIP)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Successfully authenticated. You can now establish an SSE connection.",
	})
}

func SseHandler(ctx context.Context, webConf *config.WebserverConfig, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract client IP
	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("WEBSERVER: Failed to parse client IP: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("WEBSERVER: IP %v contacted /sse", clientIP)

	allowed := firewall.CheckWhitelist(clientIP)

	if !allowed {
		log.Printf("WEBSERVER: Unauthorized SSE connection attempt from IP %v", clientIP)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", webConf.ListenURL)

	fmt.Fprintf(w, ":ok\n\n") // Flush headers
	flusher.Flush()

	state.Mutex.Lock()
	state.WhitelistedIPs[clientIP] = true
	state.Mutex.Unlock()

	defer cleanupConnection(clientIP)
	log.Printf("WEBSERVER: IP %v allowed to connect", clientIP)

	sseCTX := r.Context()
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Send periodic pings
	for {
		select {
		//main loop context
		case <-ctx.Done():
			log.Printf("WEBSERVER: Shutdown detected sse session %v closed", clientIP)
			closeSSE(w, flusher)
			return

		case <-sseCTX.Done():
			log.Printf("WEBSERVER: Session %v closed: %v", clientIP, sseCTX.Err())
			return

		case <-pingTicker.C:
			if err := sendPing(w, flusher); err != nil {
				log.Printf("WEBSERVER: Failed to send ping to %v: %v", clientIP, err)
				return
			}
		}
	}
}

func closeSSE(w http.ResponseWriter, flusher http.Flusher) {
	_, err := fmt.Fprintf(w, "event: close\ndata: {\"reason\":\"server shutdown\"}\n\n")
	if err != nil {
		return
	}
	flusher.Flush()
	//give clients some time to close the con rather then us closing it for them.
	time.Sleep(100 * time.Millisecond)
}

func cleanupConnection(ip string) {
	state.Mutex.Lock()
	defer state.Mutex.Unlock()

	delete(state.WhitelistedIPs, ip)
	for _, conn := range state.ActiveConns[ip] {
		conn.Close()
	}
	delete(state.ActiveConns, ip)
	log.Printf("WEBSERVER: Removed IP %v from whitelist", ip)
}

func sendPing(w http.ResponseWriter, flusher http.Flusher) error {
	_, err := fmt.Fprintf(w, "event: ping\ndata: %d\n\n", time.Now().UnixMilli())
	if err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func Init(uD *UsersData) {
	userData = uD

}
