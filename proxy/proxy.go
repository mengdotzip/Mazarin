package proxy

import (
	"context"
	"io"
	"log"
	"mazarin/config"
	"mazarin/firewall"
	"mazarin/state"
	"net"
	"strings"
	"sync"
)

func handleProxyConnection(ctx context.Context, clientConn net.Conn, targetAddr, clientIP string, protocol string) {
	targetConn, err := net.Dial(protocol, targetAddr)
	if err != nil {
		log.Println("PROXY: Failed to connect to target:", err)
		clientConn.Close()
		return
	}

	defer func() {
		// .Close() redundancy should be fine bcs its a no-op
		clientConn.Close()
		targetConn.Close()

		state.Mutex.Lock()
		conns := state.ActiveConns[clientIP]
		for i, c := range conns {
			if c == clientConn {
				state.ActiveConns[clientIP] = append(conns[:i], conns[i+1:]...)
				break
			}
		}
		if len(state.ActiveConns[clientIP]) == 0 {
			delete(state.ActiveConns, clientIP)
		}
		state.Mutex.Unlock()
		log.Printf("PROXY: connection closed for %s", clientIP)
	}()

	// Create a context that will be canceled when either the parent context is canceled or when one of the copy operations completes
	copyCtx, cancelCopy := context.WithCancel(ctx)
	defer cancelCopy()

	//this goroutine waits for ctx shutdown from the main loop or this one, for that reason its not in the waitgroup
	go func() {
		<-copyCtx.Done()
		clientConn.Close()
		targetConn.Close()
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancelCopy()
		io.Copy(targetConn, clientConn)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancelCopy()
		io.Copy(clientConn, targetConn)
	}()

	wg.Wait()
}

func ProxyListener(ctx context.Context, fw *config.FirewallConfig, proxy *config.ProxyConfig, wg *sync.WaitGroup) error {
	defer wg.Done()

	listener, err := net.Listen(proxy.Protocol, proxy.ListenAddr)
	if err != nil {
		log.Printf("PROXY: %v %v failed to start: %v", proxy.Protocol, proxy.ListenAddr, err)
		return err
	}
	defer listener.Close()
	log.Printf("PROXY: %v %v server started", proxy.Protocol, proxy.ListenAddr)

	var listenWG sync.WaitGroup

	listenWG.Add(1)
	go func() {
		defer listenWG.Done()
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("PROXY: %v %v failed to accept connection: %v", proxy.Protocol, proxy.ListenAddr, err)
				if ne, ok := err.(*net.OpError); ok {
					if ne.Op == "accept" && strings.Contains(ne.Error(), "use of closed network connection") {
						log.Printf("PROXY: %v %v accept loop exiting, listener has been closed", proxy.Protocol, proxy.ListenAddr)
						return
					}
				}
				log.Printf("PROXY: %v %v failed to accept connection: %v", proxy.Protocol, proxy.ListenAddr, err)
				continue
			}

			clientIP, _, err := net.SplitHostPort(conn.RemoteAddr().String())
			if err != nil {
				log.Printf("PROXY: Failed to parse client IP: %v", err)
				continue
			}

			allowed := true
			if fw.EnableFirewall {
				allowed = fw.DefaultAllow
				if !allowed {
					allowed = firewall.CheckWhitelist(clientIP, conn)
				}
			}

			if allowed {
				log.Printf("PROXY: %v %v Starting proxy for %v to dest %v", proxy.Protocol, proxy.ListenAddr, clientIP, proxy.TargetAddr)
				go handleProxyConnection(ctx, conn, proxy.TargetAddr, clientIP, proxy.Protocol)
			} else {
				log.Printf("PROXY: %v %v Blocked connection from: %v", proxy.Protocol, proxy.ListenAddr, clientIP)
				conn.Close()
			}
		}
	}()

	<-ctx.Done()
	stopServer(listener)
	listenWG.Wait()
	return nil
}

func stopServer(listener net.Listener) {
	if err := listener.Close(); err != nil {
		log.Printf("PROXY: Listen server shutdown error: %v", err)
		return
	}
	log.Printf("PROXY: Listen server shut down successfully")
}
