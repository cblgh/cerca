# Hosting

## System user

You can use a system user with no login:

```
useradd -r cerca
usermod -s /bin/false cerca
```

## Nginx configuration

```
server {
  listen 80;
  listen 443 ssl;

  server_name <your-domain>;

  location / {
    proxy_set_header X-Real-IP $remote_addr;
    proxy_pass http://127.0.0.1:8272;
  }

  # NOTE: only required if running cerca via a standalone binary
  #       vs. a git clone where it will have access to the assets dir
  location /assets/ {
    root <path-to-your-cerca-assets-dir>;
  }
}
```

## Systemd unit file

This can be placed at `/etc/systemd/system/cerca.service`:

```
[Unit]
Description=cerca
After=syslog.target network.target

[Service]
User=cerca
ExecStart=<path-to-cerca-binary> -config <path-to-cerca.toml> -authkey "<...>" -allowlist <path-to-allowlist.txt> -data <path-to-data-dir>
RemainAfterExit=no
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Then you need to:

```
systemctl daemon-reload
systemctl start cerca
systemctl enable cerca
```

To tail logs:

```
journalctl -fu cerca
```
