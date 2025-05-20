// main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"mazarin/config"
	"mazarin/listeners"
	"mazarin/router"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"

	//"mazarin/firewall"

	"mazarin/webserver"
)

func main() {
	fmt.Println("v0.0.6")

	//cmd flag, Generate hashed key and exit.
	shouldExit := parseArgs()
	if shouldExit {
		return
	}
	//----

	//OS exit signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	//----

	//INITS
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Failed to open/decode config.json:", err)
		return
	}

	if cfg.Logging.EnableLogging {
		fmt.Println("Logging is enabled")
		cfg.Logging.InitLog()
		defer cfg.Logging.Close()
	}

	/*if len(cfg.Router) > 0 {
		router.InitRouter(cfg.Router)
	}*/

	var wg sync.WaitGroup

	if cfg.Webserver.EnableWebServer {
		webRoute := config.ProxyConfig{
			ListenUrl: cfg.Webserver.ListenURL,
			Port:      cfg.Webserver.ListenPort,
			Type:      "func",
			Protocol:  "web",
		}
		//router.AddRoute(webRoute)
		cfg.Proxy = append(cfg.Proxy, webRoute)
		webserver.Init(webserver.LoadKeys(cfg.Webserver.KeysDir))
	}
	//-----

	//Start listen servers
	var usedPortsProxy []string
	var usedPortsRouter []string
	var toBeRouted []config.ProxyConfig
	for _, srvs := range cfg.Proxy {
		//web
		if srvs.Protocol == "web" {
			if slices.Contains(usedPortsProxy, srvs.Port) {
				log.Println("ROUTER ERROR: Cant have a tcp/udp proxy and a web proxy on the same port, both need to be web proxies.")
				stop()
				break
			}
			toBeRouted = append(toBeRouted, srvs)
			if !slices.Contains(usedPortsRouter, srvs.Port) {
				usedPortsRouter = append(usedPortsRouter, srvs.Port)
				if cfg.TLS.EnableTLS && slices.Contains(cfg.TLS.Domains, srvs.ListenUrl) {
					wg.Add(1)
					go listeners.ListenWebTLS(ctx, &cfg.TLS, &cfg.Firewall, &srvs, &cfg.Webserver, &wg)
					continue
				}
				wg.Add(1)
				go listeners.ListenWeb(ctx, &cfg.TLS, &cfg.Firewall, &srvs, &cfg.Webserver, &wg) //The web listen func handles its own ctx stop
			}
			continue
		}

		//non web
		if slices.Contains(usedPortsProxy, srvs.Port) {
			log.Println("PROXY ERROR: Cant have multiple tcp/udp proxies on the same port, use web for this.")
			stop()
			break
		}
		usedPortsProxy = append(usedPortsProxy, srvs.Port)
		wg.Add(1)
		go func() {
			if err := listeners.ListenProxy(ctx, &cfg.Firewall, &srvs, &wg); err != nil {
				log.Println("Proxy server failed starting up, starting a shutdown")
				stop() //Signal with the main ctx to start a clean shutdown
			}
		}()
	}

	if len(toBeRouted) > 0 {
		router.InitRouter(toBeRouted)
	}
	//var usedPortsProxy []string
	//var usedPortsRouter []string
	/*for _, srvs := range cfg.Proxy {
		if slices.Contains(usedPortsProxy, srvs.ListenAddr) {
			log.Println("PROXY ERROR: Cant have multiple proxies on the same port, use the router for this.")
			stop()
			break
		}
		usedPortsProxy = append(usedPortsProxy, srvs.ListenAddr)
		wg.Add(1)
		go func() {
			if err := listeners.ListenProxy(ctx, &cfg.Firewall, &srvs, &wg); err != nil {
				log.Println("Proxy server failed starting up, starting a shutdown")
				stop() //Signal with the main ctx to start a clean shutdown
			}
		}()
	}

	//start router servers, the router logic will probbably replace the proxy logic in the future (since the router also has proxy)
	for _, srvs := range cfg.Router {
		if slices.Contains(usedPortsProxy, srvs.Port) {
			log.Println("ROUTER ERROR: Cant have a proxy and a route on the same port, both need to be routes.")
			stop()
			break
		}
		if !slices.Contains(usedPortsRouter, srvs.Port) {
			usedPortsRouter = append(usedPortsRouter, srvs.Port)
			if cfg.TLS.EnableTLS && slices.Contains(cfg.TLS.Domains, srvs.ListenUrl) {
				wg.Add(1)
				go listeners.ListenWebTLS(ctx, &cfg.TLS, &cfg.Firewall, &srvs, &cfg.Webserver, &wg)
				continue
			}
			wg.Add(1)
			go listeners.ListenWeb(ctx, &cfg.TLS, &cfg.Firewall, &srvs, &cfg.Webserver, &wg) //The web listen func handles its own ctx stop
		}
	} */
	//-----

	//Clean shutdown portion
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		log.Println("Main thread shutdown signal received, starting shutdown timer")
		if cfg.Logging.EnableLogging {
			fmt.Println("Main thread shutdown signal received, starting shutdown timer")
		}
		shutdownTimeout := 5 * time.Second

		select {
		case <-done:
			log.Println("All goroutines finished, exiting cleanly")
		case <-time.After(shutdownTimeout):
			log.Println("Shutdown timeout reached, forcing exit")

		}
	case <-done:
		log.Println("All goroutines finished, exiting cleanly")
	}

}

func parseArgs() bool {
	keyPtr := flag.String("key", "", "Generate an hash for a given key and exit")
	flag.Parse()

	if *keyPtr != "" {
		hashKey, err := webserver.HashKey(*keyPtr)
		if err != nil {
			fmt.Println("Error generating hash:", err)
			return true
		}
		fmt.Println(hashKey)
		return true
	}
	return false
}
