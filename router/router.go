package router

import (
	"context"
	"log"
	"mazarin/config"
	"mazarin/firewall"
	"mazarin/proxy"
	"mazarin/webserver"
	"net"
	"net/http"
	"strings"
)

var routes = make(map[string]config.ProxyConfig)

func InitRouter(routConf []config.ProxyConfig) {
	for _, route := range routConf {
		routes[route.ListenUrl] = route
	}
}

func RouteWithCfg(ctx context.Context, webConf *config.WebserverConfig, firewallConf *config.FirewallConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route(ctx, webConf, firewallConf, w, r)
	}
}

func route(ctx context.Context, webConf *config.WebserverConfig, firewallConf *config.FirewallConfig, w http.ResponseWriter, r *http.Request) {
	reqHost := strings.Split(strings.ToLower(r.Host), ":") //TODO check what happens on ipv6

	//FIREWALL
	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("ROUTER: Failed to parse client IP: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if firewallConf.EnableFirewall {
		//Add blacklist/whitelist here in the future
		if !firewallConf.DefaultAllow {
			if !firewall.CheckWhitelist(clientIP) && reqHost[0] != webConf.ListenURL { //Make sure the router still allows the proxy auth page to load :p
				log.Printf("ROUTER: IP: %v access denied for: %v", clientIP, reqHost[0])
				http.Error(w, "Proxy Authentication Required", http.StatusProxyAuthRequired)
				return
			}
		}
	}

	if !firewall.ValidateInput(r.URL.Path, "path") {
		log.Printf("ROUTER: IP: %v invalid path", clientIP)
		http.Error(w, "Domain does not exist", http.StatusBadRequest)
		return
	}
	if !firewall.ValidateInput(reqHost[0], "url") {
		log.Printf("ROUTER: IP: %v invalid url", clientIP)
		http.Error(w, "Domain does not exist", http.StatusBadRequest)
		return
	}
	//------

	//ROUTING
	routeInfo, ok := routes[reqHost[0]]
	if !ok {
		log.Printf("ROUTER: Requested url is not a configured route: %v", reqHost[0])
		http.Error(w, "Domain does not exist", http.StatusBadRequest)
		return
	}

	if !routeInfo.NoHeaders {
		//--Set secure headers---
		//ONLY SET HEADERS FOR WEB, might have to change this to a sperate func in the future
		//Only set HSTS if using HTTPS
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()") //IMPORTANT, if a proxied site needs these then you might want to disable this
		//-------
	}

	log.Printf("ROUTER: IP %v getting routed to %v", clientIP, routeInfo.TargetAddr)
	switch routeInfo.Type {
	case "proxy":
		proxy.HandleHTTPProxy(w, r, &routeInfo)
	case "static":
	case "redirect":
	case "func":
		if webConf.EnableWebServer {
			//Currently only our webserver uses func, the func type is meant for routes that call code in the program
			//TODO make this more configurable
			var stylesCSS = webConf.StaticDir + "/styles.css"
			var scriptJS = webConf.StaticDir + "/script_v2.js"
			switch r.URL.Path {
			case "/":
				http.ServeFile(w, r, webConf.StaticDir)
			case "/styles.css", "/styles.css/":
				http.ServeFile(w, r, stylesCSS)
			case "/script_v2.js":
				http.ServeFile(w, r, scriptJS)
			case "/auth":
				webserver.AuthHandler(w, r)
			case "/sse":
				webserver.SseHandler(ctx, webConf, w, r)
			}
		}
	}
}
