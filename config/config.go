package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// ----
type ProxyConfig struct {
	ListenAddr string `json:"listen_addr"`
	TargetAddr string `json:"target_addr"`
	Protocol   string `json:"protocol"`
}

// ----
type RoutesConfig struct {
	ListenUrl  string `json:"listen_url"`
	Port       string `json:"port"`
	TargetAddr string `json:"target_addr"`
	Type       string `json:"type"`
	Protocol   string `json:"protocol"`
}

type RouterConfig struct {
	EnableRouter bool           `json:"enable_router"`
	Routes       []RoutesConfig `json:"routes"`
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
	Router    RouterConfig    `json:"router"`
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

	return cfg, nil
}

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
