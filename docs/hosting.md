# Hosting

## `CERCA_ROOT`

By default certain paths are relative to where `cerca` executes. These are:

* the css theme file that can be customized, `html/assets/theme.css`
* the content files defined by `config.Documents`, which are in `content/` by default
* the config file, called `cerca.toml` by default and can be defined with flag `--config`
* the database, residing in `data/forum.db` by default and an be defined with flag `--data`

In order to make these default assumptions more useful in different deployment environments,
the environment variable `CERCA_ROOT`. When set, it defines a base path from which to resolve
the above relative paths. 

If any path (for example `Config.Document.About` is set to `/var/www/forum-content`) is an
absolute path, or if flags `--config` and `--data` are set to non-default values, then those
paths **will not** be joined with `CERCA_ROOT`.

In normal operation, `CERCA_ROOT` should be the path to the directory storing the following
files for a particular instance of a Cerca forum:

```
├── cerca.toml
├── content
│   ├── about.md
│   ├── logo.html
│   ├── registration-instructions.md
│   └── rules.md
├── data
│   └── forum.db
├── html
│   └── assets
│       └── theme.css
```

If you run multiple forum instances on the same machine, then it is likely you want to also set
different values for the `CERCA_ROOT` to point to the correct forum instance.

You can set the environment variable in many ways, including directly when executing `cerca`:

```
CERCA_ROOT="/var/www/forum-for-friends" cerca run --authkey CHANGEME
```

You can also set the equivalent value in the config:

```
[tooling]
cerca_root = "/your/path"
```

If `CERCA_ROOT` and the config's `cerca_root` are both set, then the config file's `cerca_root`
takes effect over the `CERCA_ROOT` environment variable.

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
