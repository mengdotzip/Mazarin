# Mazarin

**Go-based proxy server with web authentication, firewall capabilities, TLS, routing and more to come :)**

https://github.com/user-attachments/assets/209af405-0cf4-453d-a745-81dba677f82b

## Backstory

Originally, Mazarin was created out of necessity: a friend wanted to connect to my game server, but was behind a VPN, making IP-based whitelisting impossible. Without the budget for a proper VPS or commercial solutions, I built Mazarin as a self-hosted proxy with authentication and firewall features to whitelist users dynamically.

Since then, Mazarin has evolved into a more robust proxy server with domain routing, SSL support and plans for token-based authentication, advanced firewall features like auto-blacklisting and rate limiting.


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
git clone https://github.com/mengdotzip/Mazarin.git
cd mazarin
```

2. **Configure**

Create a `config.json` file. Please check out the [DOCS](docs/README.md) for the different configuration options.

3. **Set up authentication**

Create a `keys.json` file in your keys directory. Please check out the [DOCS](docs/Authentication.md) for the json format.

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
