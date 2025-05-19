# Mazarin

**Go-based proxy server with web authentication, firewall capabilities, and more to come :)**


## Features

- Forward proxy with TCP/UDP support
- Web-based authentication with hashed keys
- IP whitelisting firewall
- Domain-based routing with TLS support
- HTTP reverse proxy capabilities
- Server-Sent Events (SSE) support for real-time communication
- Graceful shutdown handling
- Configurable via JSON
- Modular Go codebase for easy extension


## Getting Started

1. **Clone the repository**
```bash
git clone https://github.com/Maty-0/Mazarin.git
cd mazarin
```

2. **Configure**

Create a `config.json` file (in the same path as main.go or the executable) with the following structure:

```json
{
  "proxies": [
    {
      "listen_addr": ":80",
      "target_addr": "192.168.129.88:80",
      "protocol": "tcp"
    }
  ],
  "router": {
    "enable_router": true,
    "routes": [
      {
        "listen_url": "vault.domain.com",
        "port": ":47319",
        "target_addr": "192.168.129.88:80",
        "type": "proxy",
        "protocol": "tcp"
      }
    ]
  },
  "tls": {
    "enable_tls": true,
    "cert_file": "./tls/domain.pem",
    "key_file": "./tls/priv.pem",
    "domains": [
      "proxy.domain.com",
      "vault.domain.com"
    ]
  },
  "firewall": {
    "enable_firewall": true,
    "default_allow": false
  },
  "logging": {
    "enable_logging": true,
    "log_dir": "./logs"
  },
  "webserver": {
    "enable_webserver": true,
    "listen_port": ":47319",
    "listen_url": "proxy.domain.com",
    "static_dir": "./static",
    "keys_dir": "./keys"
  }
}
```


### Configuration Options

- **proxies**: Array of proxy configurations
    - `listen_addr`: The local address and port to listen on (e.g., ":80")
    - `target_addr`: The destination address to forward traffic to
    - `protocol`: Either "tcp" or "udp"
- **router**: Domain-based routing configuration
    - `enable_router`: Whether to enable the router
    - `routes`: Array of route configurations
        - `listen_url`: Domain name to listen for (e.g., "vault.domain.com" or "192.168.0.10")
        - `port`: Port to listen on
        - `target_addr`: Target address for proxy routes
        - `type`: Route type ("proxy" or "func")
        - `protocol`: Protocol for proxy routes ("tcp" or "udp")
- **tls**: TLS/SSL configuration
    - `enable_tls`: Whether to enable TLS
    - `cert_file`: Path to certificate file
    - `key_file`: Path to private key file
    - `domains`: Array of domains covered by the certificate (Note, dont forget to add the webserver url to here)
- **firewall**:
    - `enable_firewall`: Whether to enable the firewall
    - `default_allow`: If true, allows all connections by default; if false, only allows whitelisted IPs
- **logging**:
    - `enable_logging`: Whether to enable logging
    - `log_dir`: Directory where logs will be stored
- **webserver**:
    - `enable_webserver`: Whether to enable the web interface
    - `listen_port`: Port for the web interface
    - `listen_url`: Domain name for the web interface
    - `static_dir`: Directory for static web files
    - `keys_dir`: Directory containing authentication keys

3. **Set up authentication**

Create a `keys.json` file in your keys directory:

```json
{
  "users": [
      {
        "name": "test",
        "hash": "$2a$10$f.qQVxQMikTkKZWYekqYfOi17O8f1/83HA5CX8TADYtQGhHmptZha",
        "allowed_sessions": 1
      },
      {
        "name": "user2",
        "hash": "$2a$10$Z1/wTrjFwzWaC60CwQYgVe.M7hcKr0YESo2G6etOSInxkklltcfIO", 
        "allowed_sessions": 1
      }
  ]
}
```
(In this example the password for test is test_password and for user2 is user2_password)

4. **Generate hashed keys**
```bash
go run main.go -key yourpassword
```

Use the output hash in your `keys.json` for authentication.

5. **Run**
```bash
go run main.go
```



## Planned Improvements

- Token-based authentication (JWT) with device and IP binding
- Auto-blacklisting and rate limiting for abusive clients
- Structured logging and metrics integration
- Alert the user on wrong config.json configurations
- PostgreSQL support for user management



## Who is Mazarin

Named after Cardinal Jules Mazarin, the influential 17th-century chief minister and regent of France, who acted as a powerful proxy to King Louis XIV during his minority. Just as Mazarin managed and secured the affairs of the kingdom behind the scenes, this proxy server governs and protects access to your network services - providing controlled, authenticated, and secure connections.



## Backstory

Originally, Mazarin was created out of necessity: a friend wanted to connect to my game server, but was behind a VPN, making IP-based whitelisting impossible. Without the budget for a proper VPS or commercial solutions, I built Mazarin as a self-hosted proxy with authentication and firewall features to whitelist users dynamically.

Since then, Mazarin has evolved into a more robust proxy server with plans for token-based authentication, domain routing, SSL support, and advanced firewall features like auto-blacklisting and rate limiting.
