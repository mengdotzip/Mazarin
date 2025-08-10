package sessions

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

//Not in use yet

var (
	mu       = sync.RWMutex{}
	sessions = make(map[string]*Session)
	cleanup  = make(chan struct{})
)

type Session struct {
	Username  string
	CreatedAt time.Time
	ExpiresAt time.Time
	IPAddress string
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func CreateSession(username, ipAddress string, duration time.Duration) string {
	token := generateToken()
	session := &Session{
		Username:  username,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(duration),
		IPAddress: ipAddress,
	}
	mu.Lock()
	defer mu.Unlock()
	sessions[token] = session
	return token
}

func ValidateSession(token string) (string, string, bool) {
	mu.RLock()
	defer mu.RUnlock()
	session, ok := sessions[token]
	if !ok {
		return "", "", false
	}
	if time.Now().After(session.ExpiresAt) {
		// Cleanup goroutine will handle expired sessions
		return "", "", false
	}
	return session.Username, session.IPAddress, true
}
