package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// CONFIG
// ----
type ProxyConfig struct {
	ListenUrl     string            `json:"listen_url"`
	ListenUrls    []string          `json:"listen_urls"`
	Port          string            `json:"port"`
	Ports         []string          `json:"ports"`
	TargetAddr    string            `json:"target_addr"`
	Type          string            `json:"type"`
	Protocol      string            `json:"protocol"`
	AllowInsecure bool              `json:"allow_insecure"`
	NoHeaders     bool              `json:"no_headers"`
	Headers       map[string]string `json:"headers"`
}

// ----
type TLSConfig struct {
	EnableTLS bool     `json:"enable_tls"`
	Cert      string   `json:"cert_file"`
	Key       string   `json:"key_file"`
	Domains   []string `json:"domains"`
}

// ----
type FirewallConfig struct {
	EnableFirewall bool `json:"enable_firewall"`
	DefaultAllow   bool `json:"default_allow"`
}

// ----
type WebserverConfig struct {
	EnableWebServer bool   `json:"enable_webserver"`
	ListenPort      string `json:"listen_port"`
	ListenURL       string `json:"listen_url"`
	StaticDir       string `json:"static_dir"`
	KeysDir         string `json:"keys_dir"`
}

// ----
type LoggingConfig struct {
	EnableLogging bool   `json:"enable_logging"`
	LogDir        string `json:"log_dir"`
	logFile       *os.File
}

// ----
type Config struct {
	Proxy     []ProxyConfig   `json:"proxies"`
	TLS       TLSConfig       `json:"tls"`
	Firewall  FirewallConfig  `json:"firewall"`
	Logging   LoggingConfig   `json:"logging"`
	Webserver WebserverConfig `json:"webserver"`
}

func LoadConfig() (Config, error) {
	var cfg Config

	file, err := os.Open("config.json")
	if err != nil {
		return cfg, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return cfg, err
	}

	cfg.Proxy, err = ParseMulti(cfg.Proxy)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

// parse multi will just take ports and listen_urls and then parse them to multiple configs (per port/url)
func ParseMulti(cfg []ProxyConfig) ([]ProxyConfig, error) {

	// URLS Parsing
	var resultAfterUrls []ProxyConfig
	for _, conf := range cfg {
		if len(conf.ListenUrls) > 0 {
			for _, url := range conf.ListenUrls {
				expanded := conf
				expanded.ListenUrl = url
				resultAfterUrls = append(resultAfterUrls, expanded)
			}
		} else {
			resultAfterUrls = append(resultAfterUrls, conf)
		}
	}

	// PORTS Parsing
	var resultAfterPorts []ProxyConfig
	for _, conf := range resultAfterUrls {
		if len(conf.Ports) > 0 {
			for _, port := range conf.Ports {
				expanded := conf
				expanded.Port = port
				resultAfterPorts = append(resultAfterPorts, expanded)
			}
		} else {
			resultAfterPorts = append(resultAfterPorts, conf)
		}
	}

	// Ports Ranges Parsing
	var resultAfterRanges []ProxyConfig
	for _, conf := range resultAfterPorts {
		splitResult := strings.Split(conf.Port, "-")
		if len(splitResult) > 1 {

			fromInt, err := strconv.Atoi(splitResult[0])
			if err != nil {
				return nil, err
			}
			untilInt, err := strconv.Atoi(splitResult[1])
			if err != nil {
				return nil, err
			}

			// check that the range is legal
			if fromInt >= untilInt {
				return nil, errors.New("cannot have a range starting at a higher or the same number")
			}

			for i := fromInt; i <= untilInt; i++ {
				expanded := conf
				expanded.Port = ":" + strconv.Itoa(i)
				resultAfterRanges = append(resultAfterRanges, expanded)
			}

		} else {
			resultAfterRanges = append(resultAfterRanges, conf)
		}
	}

	return resultAfterRanges, nil
}

//-----------

// /Logging
func (conf *LoggingConfig) InitLog() {
	logDir := conf.LogDir
	fmt.Println("Logging starting in the dir: ", logDir)
	err := os.MkdirAll(logDir, os.ModePerm) // Create logs dir if it doesn't exist
	if err != nil {
		fmt.Println("Failed to create log directory:", err)
		return
	}

	logFilePath := filepath.Join(logDir, "mazarin.log")

	file, err := openLogFile(logFilePath)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
		return
	}

	conf.logFile = file

	// Set log output and flags
	log.SetOutput(file)
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	log.Println("Logging started at ", time.Now().UnixMilli())
}

func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func (conf *LoggingConfig) Close() error {
	if conf.logFile != nil {
		log.Println("Closing log file")
		return conf.logFile.Close()
	}
	return nil
}

// parsing
type ParsedProxy struct {
	Port          string
	Protocol      string
	TLS           bool
	LinkedProxies []*ProxyConfig
}

func ParseProxies(toParse []ProxyConfig, tlsConf *TLSConfig) (map[string]ParsedProxy, []ProxyConfig, error) {
	parsedProxyMap := make(map[string]ParsedProxy)
	var toBeRouted []ProxyConfig
	for _, proxies := range toParse {
		switch proxies.Protocol {
		case "web":
			toBeRouted = append(toBeRouted, proxies)
			allowed, ok := parsedProxyMap[proxies.Port]
			if ok {
				if allowed.Protocol != "web" {
					return parsedProxyMap, toBeRouted, errors.New("PARSER ERROR: Cant have a tcp/udp proxy and a web proxy on the same port, both need to be web proxies")
				}
				if tlsConf.EnableTLS && slices.Contains(tlsConf.Domains, proxies.ListenUrl) && !allowed.TLS {
					return parsedProxyMap, toBeRouted, errors.New("PARSER ERROR: Cant have a http and https proxy on the same port")
				}
				if allowed.TLS && !slices.Contains(tlsConf.Domains, proxies.ListenUrl) {
					return parsedProxyMap, toBeRouted, errors.New("PARSER ERROR: Cant have a https and http proxy on the same port")
				}

				allowed.LinkedProxies = append(allowed.LinkedProxies, &proxies)
				continue
			}

			newProxy := ParsedProxy{
				Port:     proxies.Port,
				Protocol: proxies.Protocol,
				TLS:      slices.Contains(tlsConf.Domains, proxies.ListenUrl),
			}
			newProxy.LinkedProxies = append(newProxy.LinkedProxies, &proxies) //could find how to define the array in the struct creation
			parsedProxyMap[newProxy.Port] = newProxy

		case "tcp", "udp":
			allowed, ok := parsedProxyMap[proxies.Port]
			if ok {
				if allowed.Protocol != "tcp/udp" {
					return parsedProxyMap, toBeRouted, errors.New("PARSER ERROR: Cant have a tcp/udp proxy and a web proxy on the same port, both need to be web proxies")
				}
				return parsedProxyMap, toBeRouted, errors.New("PARSER ERROR: Cant have multiple tcp/udp proxies on the same port, use type: web for this")
			}

			newProxy := ParsedProxy{
				Port:     proxies.Port,
				Protocol: "tcp/udp",
			}
			newProxy.LinkedProxies = append(newProxy.LinkedProxies, &proxies)
			parsedProxyMap[newProxy.Port] = newProxy

		}
	}
	return parsedProxyMap, toBeRouted, nil
}
