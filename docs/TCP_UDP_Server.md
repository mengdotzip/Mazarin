# Game Server Example

Setting up a proxy for any TCP/UDP program.



### Simple Config
---

This is the most basic configuration. All data will be proxied directly to the `target_addr` ,no authentication or filtering.

```json
{
  "proxies": [
    {
      "port": ":25565",
      "target_addr": "192.168.129.88:25565",
      "protocol": "tcp"
    }
  ]
}
```




### With Webserver & Firewall
---

A more secure way of setting up a game server proxy is by enabling the webserver for authentication and the firewall to block or whitelist users.

```json
{
  "proxies": [
    {
      "port": ":25565",
      "target_addr": "192.168.129.88:25565",
      "protocol": "tcp"
    }
  ],
  "firewall": {
    "enable_firewall": true,
    "default_allow": false
  },
  "webserver": {
    "enable_webserver": true,
    "listen_port": ":47319",
    "listen_url": "192.168.0.100",
    "static_dir": "./static",
    "keys_dir": "./keys"
  }
}
```

- **Firewall:** Only users who authenticate or are explicitly allowed will be able to connect. All others are blocked by default.
- **Webserver:** Hosts the authentication portal so users can log in and be granted access.
  - **Static Content:** If you need the contents of the static folder for the web interface, you can find them at [`/webserver/static`](../webserver/static).

In a prod environment you will want to make sure the webserver is also running TLS, you can configure this by writing:

```json
{
  "proxies": [
    {
      "port": ":25565",
      "target_addr": "192.168.129.88:25565",
      "protocol": "tcp"
    }
  ],
  "firewall": {
    "enable_firewall": true,
    "default_allow": false
  },
  "webserver": {
    "enable_webserver": true,
    "listen_port": ":47319",
    "listen_url": "proxy.yourdomain.com",
    "static_dir": "./static",
    "keys_dir": "./keys"
  },
  "tls": {
    "enable_tls": true,
    "cert_file": "./tls/domain.pem",
    "key_file": "./tls/priv.pem",
    "domains": [
      "proxy.yourdomain.com"
    ]
  }
}
```