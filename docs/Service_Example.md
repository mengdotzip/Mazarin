# Service Example

Here you can find an example of the systemd service config I use
```                
[Unit]
Description=Mazarin Proxy Server
Documentation=https://github.com/mengdotzip/mazarin
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=/home/ublocalproxy
ExecStart=/home/ublocalproxy/mazarin
Restart=on-failure
RestartSec=6s
StandardOutput=journal
StandardError=journal

NoNewPrivileges=true

LimitNOFILE=65535

[Install]
WantedBy=multi-user.target


```

If you also want to run this than you can take a copy of my file and change the following:

```                
WorkingDirectory=/home/ublocalproxy
ExecStart=/home/ublocalproxy/mazarin
```

- **WorkingDirectory** should point to the folder which mazarin is in.
- **ExecStart** should point to the mazarin executable.

You can than move the file you have created into the systemd dir e.g.:
```
sudo mv docs/mazarin.service /etc/systemd/system/
```

**To start** it you do the following:
- Reload systemd
```
sudo systemctl daemon-reload
```
- Enable startup on boot
```
sudo systemctl enable mazarin
```
- Start service
```
sudo systemctl start mazarin
```

You can check the status with
```
sudo systemctl status mazarin
```
You can view logs by using
```
journalctl -u mazarin -f
```
Or reading the defined output file in the Mazarin config.