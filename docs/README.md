
# Configure
Below you will find a full config.json file of Mazarin, you should be able to find every config option here with explanation. Please check out our other examples in this folder if you need any clarification.<br>
(Note: You only need to config what you need eg; a config.json file with just "proxies" will work.)

Create a `config.json` file (in the same path as main.go or the executable) with the following structure:

```json
{
  "proxies": [
    {
      "port": ":80",
      "target_addr": "192.168.129.88:80",
      "protocol": "tcp"
    },
    {
      "listen_url": "vault.domain.com",
      "port": ":47319",
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

- **proxies**: Array of proxy and routing configurations
    - **TCP/UDP Proxies**:
        - `port`: The local address and port to listen on (e.g., ":80")
        - `target_addr`: The destination address to forward traffic to
        - `protocol`: "tcp" or "udp"
    - **Domain-based Web Routing**:
        - `listen_url`: Domain name to listen for (e.g., "vault.domain.com")
        - `port`: Port to listen on (e.g., ":47319")
        - `target_addr`: Target address for proxy routes
        - `type`: "proxy" (for HTTP reverse proxy) or "func" (for internal functions)
        - `protocol`: "web" (required for domain-based routing)
- **tls**: TLS/SSL configuration
    - `enable_tls`: Whether to enable TLS
    - `cert_file`: Path to certificate file
    - `key_file`: Path to private key file
    - `domains`: Array of domains covered by the certificate (must include all `listen_url` domains)
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

**Note:** The webserver configuration is automatically added to the proxies array with a "web" protocol. If you want to access your web interface, make sure to include its domain name in the TLS domains list.


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
