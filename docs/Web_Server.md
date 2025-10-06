# Web Server

### Basic HTTP Proxy
---

A simple HTTP proxy that forwards all traffic directly to your web server:

```json
{
  "proxies": [
    {
      "port": ":80",
      "target_addr": "192.168.129.88:80",
      "protocol": "tcp"
    }
  ]
}
```
This configuration forwards all HTTP traffic from port 80 to your internal web server. No domain filtering or TLS

### Serve a folder
---

With this config you can serve a folder over http or https:

```json
{
  "proxies": [
    {
      "listen_url": "static.domain.com",
      "port": ":443",
      "target_addr": "./static",
      "type": "static",
      "protocol": "web"
    }
  ]
}
```
Watch out, everything in this folder (in this case ./static) will be public to the user.

### HTTPS with TLS
---

Mazarin can provide TLS encryption for your HTTP services, upgrading plain HTTP backends to HTTPS:
```json
{
  "proxies": [
    {
      "listen_url": "vault.domain.com",
      "port": ":443",
      "target_addr": "192.168.129.88:80",
      "type": "proxy",
      "protocol": "web"
    }
  ],
  "tls": {
    "enable_tls": true,
    "cert_file": "./tls/domain.pem",
    "key_file": "./tls/priv.pem",
    "domains": [
      "vault.domain.com"
    ]
  }
}
  ```

- **TLS:** Encrypts all traffic between clients and your server using HTTPS, protecting sensitive data and ensuring secure connections. Mazarin handles certificate management and SSL termination automatically.

### Multiple Domains with Single Certificate
---

Host multiple web services under different subdomains:

```json
{
  "proxies": [
    {
      "listen_url": "vault.domain.com",
      "port": ":443",
      "target_addr": "192.168.129.88:80",
      "type": "proxy",
      "protocol": "web"
    },
    {
      "listen_url": "api.domain.com",
      "port": ":443",
      "target_addr": "192.168.129.88:8080",
      "type": "proxy",
      "protocol": "web"
    }
  ],
  "tls": {
    "enable_tls": true,
    "cert_file": "./tls/domain.pem",
    "key_file": "./tls/priv.pem",
    "domains": [
      "vault.domain.com",
      "api.domain.com"
    ]
  }
}
```

### Multiple Proxies on the same url
---

You can define multiple proxies with the same url, as long as their port is different.

```json
{
  "proxies": [
    {
      "listen_url": "api.domain.com",
      "port": ":443",
      "target_addr": "192.168.129.88:443",
      "type": "proxy",
      "protocol": "web"
    },
    {
      "listen_url": "api.domain.com",
      "port": ":8080",
      "target_addr": "192.168.129.88:8080",
      "type": "proxy",
      "protocol": "web"
    }
  ],
  "tls": {
    "enable_tls": true,
    "cert_file": "./tls/domain.pem",
    "key_file": "./tls/priv.pem",
    "domains": [
      "api.domain.com"
    ]
  }
}
  ```

### With Authentication & Firewall
---

Protect your web services with authentication and IP filtering:

```json
{
  "proxies": [
    {
      "listen_url": "vault.domain.com",
      "port": ":443",
      "target_addr": "192.168.129.88:80",
      "type": "proxy",
      "protocol": "web"
    },
    {
      "listen_url": "api.domain.com",
      "port": ":443",
      "target_addr": "192.168.129.88:8080",
      "type": "proxy",
      "protocol": "web"
    }
  ],
  "tls": {
    "enable_tls": true,
    "cert_file": "./tls/domain.pem",
    "key_file": "./tls/priv.pem",
    "domains": [
      "vault.domain.com",
      "api.domain.com",
      "auth.domain.com"
    ]
  },
  "firewall": {
    "enable_firewall": true,
    "default_allow": false
  },
  "webserver": {
    "enable_webserver": true,
    "listen_port": ":443",
    "listen_url": "auth.domain.com",
    "static_dir": "./static",
    "keys_dir": "./keys"
  }
}
  ```

- **Firewall:** Only users who authenticate or are explicitly allowed will be able to connect. All others are blocked by default.
- **Webserver:** Hosts the authentication portal so users can log in and be granted access.
    - **Static Content:** If you need the contents of the static folder for the web interface, you can find them at [`/webserver/static`](../webserver/static).
### Insecure & Self Signed Certificates 
---

```json
{
  "proxies": [
    {
      "listen_url": "proxmox.domain.com",
      "port": ":443",
      "target_addr": "192.168.129.88:8006",
      "type": "proxy",
      "protocol": "web",
      "allow_insecure": true,
      "no_headers": true
    },
  ],
  "tls": {
    "enable_tls": true,
    "cert_file": "./tls/domain.pem",
    "key_file": "./tls/priv.pem",
    "domains": [
      "proxmox.domain.com",
      "auth.domain.com"
    ]
  },
  "firewall": {
    "enable_firewall": true,
    "default_allow": false
  },
  "webserver": {
    "enable_webserver": true,
    "listen_port": ":443",
    "listen_url": "auth.domain.com",
    "static_dir": "./static",
    "keys_dir": "./keys"
  }
}
  ```
- **Firewall:** 
  - **allow_insecure:** Allow a proxy connection to insecure and self signed certificates. (Note: This makes you vulnerable to man in the middle attacks)
  - **no_headers:** Dont allow mazarin to set security headers. Most of the time you should not touch this, but in the case of eg; proxmox, you will have to set no_headers to true.