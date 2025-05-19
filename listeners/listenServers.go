package listeners

import (
	"context"
	"crypto/tls"
	"log"
	"mazarin/config"
	"mazarin/firewall"
	"mazarin/proxy"
	"mazarin/router"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

func ListenProxy(ctx context.Context, fw *config.FirewallConfig, proxyConf *config.ProxyConfig, wg *sync.WaitGroup) error {
	defer wg.Done()

	listener, err := net.Listen(proxyConf.Protocol, proxyConf.ListenAddr)
	if err != nil {
		log.Printf("PROXY: %v %v failed to start: %v", proxyConf.Protocol, proxyConf.ListenAddr, err)
		return err
	}
	defer listener.Close()
	log.Printf("PROXY: %v %v server started", proxyConf.Protocol, proxyConf.ListenAddr)

	var listenWG sync.WaitGroup

	listenWG.Add(1)
	go func() {
		defer listenWG.Done()
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("PROXY: %v %v failed to accept connection: %v", proxyConf.Protocol, proxyConf.ListenAddr, err)
				if ne, ok := err.(*net.OpError); ok {
					if ne.Op == "accept" && strings.Contains(ne.Error(), "use of closed network connection") {
						log.Printf("PROXY: %v %v accept loop exiting, listener has been closed", proxyConf.Protocol, proxyConf.ListenAddr)
						return
					}
				}
				log.Printf("PROXY: %v %v failed to accept connection: %v", proxyConf.Protocol, proxyConf.ListenAddr, err)
				continue
			}

			clientIP, _, err := net.SplitHostPort(conn.RemoteAddr().String())
			if err != nil {
				log.Printf("PROXY: Failed to parse client IP: %v", err)
				continue
			}

			//INSERT ROUTER CODE
			//If the router is on, hand off the work and continue the for loop
			/*if routerConf.EnableRouter {
				router.Route(conn)
				continue
			}*/

			allowed := true
			if fw.EnableFirewall {
				allowed = fw.DefaultAllow
				if !allowed {
					allowed = firewall.CheckWhitelistAddConn(clientIP, conn)
				}
			}

			if allowed {
				log.Printf("PROXY: %v %v Starting proxy for %v to dest %v", proxyConf.Protocol, proxyConf.ListenAddr, clientIP, proxyConf.TargetAddr)
				go proxy.HandleProxyConnection(ctx, conn, proxyConf.TargetAddr, clientIP, proxyConf.Protocol)
			} else {
				log.Printf("PROXY: %v %v Blocked connection from: %v", proxyConf.Protocol, proxyConf.ListenAddr, clientIP)
				conn.Close()
			}
		}
	}()

	<-ctx.Done()
	stopServer(listener)
	listenWG.Wait()
	return nil
}

//WEB LISTEN----------

func ListenWebTLS(parentCtx context.Context, tlsConf *config.TLSConfig, fw *config.FirewallConfig, srv *config.RoutesConfig, webConf *config.WebserverConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	//Some handy tips: https://blog.cloudflare.com/exposing-go-on-the-internet/

	//Load tls cert and key is currently not working, I think the standard certs domain registrars give you have to be edited for this
	/*cert, err := tls.X509KeyPair([]byte(tlsConf.Cert), []byte(tlsConf.Key))
	if err != nil {
		log.Printf("HTTPS Listener: Loading X509KeyPair error: %v", err)
		cancel()
	}*/
	cfg := &tls.Config{
		//Certificates:             []tls.Certificate{cert},
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,      //fallback
			tls.X25519MLKEM768, //quantum-safe hybrid, if you want you can put this first for security, but be ware of the performance trade off!
		},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}

	// Check certificate expiration
	/*parsedCert, err := x509.ParseCertificate(cert.Certificate[0])
	if err == nil && time.Until(parsedCert.NotAfter) < time.Hour*24*7 {
		log.Printf("HTTPS Listener: Certificate expires soon: %v", cert.Leaf.NotAfter)
	}*/

	mux := http.NewServeMux()
	server := &http.Server{
		//ReadTimeout:  5 * time.Second,
		//WriteTimeout: 5 * time.Second,
		IdleTimeout: 40 * time.Second,
		TLSConfig:   cfg,
		Addr:        srv.Port,
		Handler:     mux,
	}

	//Let the router handle everything
	mux.HandleFunc("/", router.RouteWithCfg(ctx, webConf, fw))

	var webWG sync.WaitGroup
	webWG.Add(1)
	go func() {
		defer webWG.Done()

		log.Printf("HTTPS Listener: %v %v server started", srv.ListenUrl, srv.Port)
		err := server.ListenAndServeTLS(tlsConf.Cert, tlsConf.Key)
		if err != nil && err != http.ErrServerClosed {
			log.Printf("HTTPS Listener: ListenAndServeTLS error: %v", err)
			cancel()
		}
	}()

	listenForExit(ctx, server, &webWG)
}

func ListenWeb(parentCtx context.Context, tlsConf *config.TLSConfig, fw *config.FirewallConfig, srv *config.RoutesConfig, webConf *config.WebserverConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	mux := http.NewServeMux()
	server := &http.Server{
		//ReadTimeout:  5 * time.Second,
		//WriteTimeout: 5 * time.Second,
		IdleTimeout: 40 * time.Second,
		Addr:        srv.Port,
		Handler:     mux,
	}

	//Let the router handle everything
	mux.HandleFunc("/", router.RouteWithCfg(ctx, webConf, fw))

	var webWG sync.WaitGroup
	webWG.Add(1)
	go func() {
		defer webWG.Done()

		log.Printf("HTTP Listener: %v %v server started", srv.ListenUrl, srv.Port)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP Listener: ListenAndServeTLS error: %v", err)
			cancel()
		}
	}()

	listenForExit(ctx, server, &webWG)
}

func listenForExit(ctx context.Context, server *http.Server, webWG *sync.WaitGroup) {

	<-ctx.Done()
	log.Printf("Listener: Shutdown signal received")
	stopWebListener(server)

	//added this to make sure a deadlock on shutdown would be contained to this func
	webWGDone := make(chan struct{})
	go func() {
		webWG.Wait()
		close(webWGDone)
	}()

	select {
	case <-webWGDone:
		return
	case <-time.After(4 * time.Second):
		log.Printf("Listener: Timed out waiting for web server goroutines to finish")
		return
	}
}

func stopWebListener(server *http.Server) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Listener: forced shutdown of Listener on port:%v err: %v", server.Addr, err)
		server.Close()
		return
	}
	log.Printf("Listener: Server shut down successfully on port: %v", server.Addr)
}

func stopServer(listener net.Listener) {
	if err := listener.Close(); err != nil {
		log.Printf("PROXY: Listen server shutdown error: %v", err)
		return
	}
	log.Printf("PROXY: Listen server shut down successfully")
}
