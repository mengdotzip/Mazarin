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
	"mazarin/webserver"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	fmt.Println("v0.0.7")

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

	var wg sync.WaitGroup

	if cfg.Webserver.EnableWebServer {
		webRoute := config.ProxyConfig{
			ListenUrl: cfg.Webserver.ListenURL,
			Port:      cfg.Webserver.ListenPort,
			Type:      "func",
			Protocol:  "web",
		}
		cfg.Proxy = append(cfg.Proxy, webRoute)
		webserver.Init(webserver.LoadKeys(cfg.Webserver.KeysDir))
	}
	//-----

	//Start listen servers
	listenerMap, toBeRouted, err := config.ParseProxies(cfg.Proxy, &cfg.TLS) //I really  like how I propegate the error here, I will do this more often probably
	if err != nil {
		log.Println(err)
		return
	}
	for _, srv := range listenerMap {
		switch srv.Protocol {
		case "web":
			if !srv.TLS {
				wg.Add(1)
				go listeners.ListenWeb(ctx, &cfg.TLS, &cfg.Firewall, srv.LinkedProxies[0], &cfg.Webserver, &wg)
				continue
			}
			go listeners.ListenWebTLS(ctx, &cfg.TLS, &cfg.Firewall, &srv, &cfg.Webserver, &wg)

		case "tcp/udp":
			wg.Add(1)
			go func() {
				if err := listeners.ListenProxy(ctx, &cfg.Firewall, srv.LinkedProxies[0], &wg); err != nil {
					log.Println("Proxy server failed starting up, starting a shutdown")
					stop() //Signal with the main ctx to start a clean shutdown
					return
				}
			}()
		}
	}

	if len(toBeRouted) > 0 {
		router.InitRouter(toBeRouted)
	}
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
