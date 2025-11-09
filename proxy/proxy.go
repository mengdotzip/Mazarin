package proxy

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"mazarin/config"
	"mazarin/state"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func HandleProxyConnection(ctx context.Context, clientConn net.Conn, targetAddr, clientIP string, protocol string) {
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
		defer state.Mutex.Unlock()
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

func HandleHTTPProxy(w http.ResponseWriter, r *http.Request, template *config.ProxyConfig) {
	if !strings.HasPrefix(template.TargetAddr, "http://") && !strings.HasPrefix(template.TargetAddr, "https://") {
		if template.AllowInsecure {
			template.TargetAddr = "https://" + template.TargetAddr
		} else {
			template.TargetAddr = "http://" + template.TargetAddr
		}
	}

	target, err := url.Parse(template.TargetAddr)
	if err != nil {
		log.Printf("HTTP PROXY: Invalid target URL: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	// Create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	if template.AllowInsecure && strings.HasPrefix(template.TargetAddr, "https://") { //allow insecure https connections ONLY USE THIS IN DEV PLEASE
		proxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Customize the Director function to modify the request
	//The Director will now call the old func and add our headers
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Origin-Host", target.Host)
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("HTTP PROXY: Error proxying request: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("HTTP PROXY: Failed to parse client IP: %v", err)
		clientIP = "ERROR"
	}
	log.Printf("HTTP PROXY: Forwarding request from %v to %v%v", clientIP, target.Host, r.URL.Path)

	// Serve the request
	proxy.ServeHTTP(w, r)
}

func HandleStaticServe(w http.ResponseWriter, r *http.Request, routeInfo *config.ProxyConfig) {

	fi, err := os.Stat(routeInfo.TargetAddr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		serveFolder(w, r, routeInfo.TargetAddr, routeInfo.Path)
	case mode.IsRegular():
		serveFile(w, r, routeInfo.TargetAddr)
	}
}

// go 1.22 introduced a safer way to handle file/folder serving by using openRoot: https://go.dev/doc/go1.22
func serveFile(w http.ResponseWriter, r *http.Request, target string) {

	// open root wants a dir so well have to split it
	dir := filepath.Dir(target)
	filename := filepath.Base(target)

	root, err := os.OpenRoot(dir)
	if err != nil {
		http.Error(w, "Route does not exist", http.StatusBadRequest)
		return
	}
	defer root.Close()

	fsys := root.FS()
	http.ServeFileFS(w, r, fsys, filename)
}

func serveFolder(w http.ResponseWriter, r *http.Request, target string, stripPath string) {
	root, err := os.OpenRoot(target)
	if err != nil {
		http.Error(w, "Route does not exist", http.StatusBadRequest)
		return
	}
	defer root.Close()

	fsys := root.FS()
	fileServer := http.FileServerFS(fsys)
	log.Println(stripPath)

	// Strip the path so we can serve a target that has path:/foo defined
	if stripPath != "" {
		fileServer = http.StripPrefix(stripPath, fileServer)
		fileServer.ServeHTTP(w, r)
	} else {
		fileServer.ServeHTTP(w, r)
	}
}
