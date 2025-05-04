// main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"mazarin/config"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	//"mazarin/firewall"
	"mazarin/proxy"
	"mazarin/webserver"
)

func main() {
	fmt.Println("v0.0.5")

	//cmd flag, Generate hashed key and exit.
	shouldExit := parseArgs()
	if shouldExit {
		return
	}
	//

	//OS exit signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
		wg.Add(1)
		go webserver.Start(ctx, &cfg.Webserver, webserver.LoadKeys(cfg.Webserver.KeysDir), &wg)
	}

	for _, srvs := range cfg.Proxy {
		wg.Add(1)
		go func() {
			if err := proxy.ProxyListener(ctx, &cfg.Firewall, &srvs, &wg); err != nil {
				log.Println("Proxy server failed starting up, starting a shutdown")
				stop() //Signal with the main ctx to start a clean shutdown
			}
		}()
	}

	//Clean shutdown portion
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		log.Println("Main thread shutdown signal received, starting shutdown timer")
		fmt.Println("Main thread shutdown signal received, starting shutdown timer")
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
