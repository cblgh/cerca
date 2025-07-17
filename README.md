# Cerca
_lean forum software_

Meaning:
* to search, quest, run _(it)_
* near, close, around, nearby, nigh _(es)_
* approximately, roughly _(en; from **circa**)_

This piece of software was created after a long time of pining for a new wave of forums hangs.
The reason it exists are many. To harbor longer form discussions, and for crawling through
threads and topics. For habitually visiting the site to see if anything new happened, as
opposed to being obtrusively notified when in the middle of something else. For that sweet
tinge of nostalgia that comes with the terrain, from having grown up in pace with the sprawling
phpBB forum communities of the mid noughties.

It was written for the purpose of powering the nascent [Merveilles community forums](https://forum.merveilles.town).

## Features

* **Customizable**: Many of Cerca's facets are customizable and the structure is intentionally simple to enable DIY modification
* **Private**: Threads are public viewable by default but new threads may be set as private, restricting views to logged-in users only
* **Easy admin**: A simple admin panel lets you add users, reset passwords, and remove old accounts. Impactful actions require two admins to perform, or a week of time to pass without a veto from any admin
* **Invites**: Fully-featured system for creating both one-time and multi-use invites. Admins can monitor invite redemption by batch as well as issue and delete batches of invites. Accessible using the same simple type of web interface that services the rest of the forum's administration tasks.
* **Transparency**: Actions taken by admins are viewable by any logged-in user in the form of a moderation log
* **Low maintenance**: Cerca is architected to minimize maintenance and hosting costs by carefully choosing which features it supports, how they work, and which features are intentionally omitted
* **RSS**: Receive updates when threads are created or new posts are made by subscribing to the forum RSS feed

## Installation

Assuming you're using Linux with systemd, which is the most common scenario, follow these steps.

1. Create a user: `useradd -r cerca`
1. Make sure user can't log in: `usermod -s /bin/false cerca`
1. Build `cerca`: `make`
1. Install `cerca`: `sudo make install`
1. Copy nginx config from `contrib/nginx.conf` to the appropriate place in `/etc/nginx/`
   and adjust according to your needs
1. Reload nginx config: `sudo systemctl reload nginx`
1. Copy service file: `cp contrib/cerca.service /etc/systemd/system/cerca.service`
1. Reload service file: `systemctl daemon-reload`
1. Enable and immediately start cerca: `systemctl enable --now cerca`
1. Add a user for yourself: `cerca adduser -database /var/lib/cerca/forum.db -username <username>`
1. Make yourself an admin: `cerca makeadmin -database /var/lib/cerca/forum.db -username <username>`

Feel free to inspect logs with `journalctl -feu cerca`.

## Other features

Here is the complete help for the `cerca` command:

```
USAGE:
  run the forum

  cerca write-defaults -config <path-to-cerca.toml> --data-dir <dir-to-store-files-and-database>
  cerca -config <path-to-cerca.toml>
  cerca -config <path-to-cerca.toml> -dev

COMMANDS:
  adduser        create a new user
  makeadmin      make an existing user an admin
  migrate        manage database migrations
  resetpw        reset a user's password
  genauthkey     generate and output an authkey for use with `cerca run`
  version        output version information
  write-defaults output and save a default cerca config file and associated content files

OPTIONS:
  -config string
        config and settings file containing cerca's customizations (default "cerca.toml")
  -dev
        trigger development mode
  -port int
        port to run the forum on (default 8272)
```

For example, you can reset a user's password with
`cerca resetpw -database /var/lib/cerca/forum.db -username <username>`.

## Config

Cerca supports community customization.

* Write a custom [about text](/defaults/sample-about.md) describing the community inhabiting the forum
* Define your own [registration rules](/defaults/sample-rules.md),
  [instructions on getting an invite code to register](/defaults/sample-registration.md),
  and link to an existing code of conduct
* Set your own [custom logo](/defaults/sample-logo.html) (whether svg, png or emoji)
* Create your own theme by writing plain, frameworkless [css](/html/assets/theme.css)

The installation process will create a config file in `/etc/cerca/config.toml`, which you are
free to customise. You can also specify your own config file location with the `-config`
option.

The installation process also copies sample content files to `/var/lib/cerca/docs`, which you can then edit.

In general, after running the `cerca` process once, you will find that all the customizable
files are located relative to the `data_dir` specified in the config file:

```
├── assets
│   ├── logo.html
│   └── theme.css
├── docs
│   ├── about.md
│   ├── registration.md
│   └── rules.md
└── forum.db
```

Change any of these files and then restart the `cerca` process to serve the changes (force-refresh your
browser to see `theme.css` changes).

## Contributing

If you want to join the fun, first have a gander at the [CONTRIBUTING.md](/CONTRIBUTING.md)
document. It lays out the overall idea of the project, and outlines what kind of contributions
will help improve the project.

### Translations

Cerca supports use with different natural languages. To translate Cerca into your language, please
have a look at the existing [translations (i18n.go)](/i18n/i18n.go) and submit yours as a
[pull request](https://github.com/cblgh/cerca/compare).

## Local development

Install [golang](https://go.dev/).

To launch a local instance of the forum on Linux:

First output and configure the default config and the required content files by running the
`cerca write-defaults` command:

```
go run ./cmd/cerca write-defaults --config ./cerca.toml --data-dir ./cerca-data
```

Then run the forum:

```
go run ./cmd/cerca -dev -config ./cerca.toml
```

It should respond `Serving forum on :8277`. You can now go to [http://localhost:8277](http://localhost:8277).

### Building a binary

```
go build ./cmd/cerca
```

### Building with reduced size

This is optional, but if you want to minimize the size of the binary follow the instructions
below. Less useful for active development, more useful for sending binaries to other computers.

Pass `-ldflags="-s -w"` when building your binary:

```
go build -ldflags="-s -w" ./cmd/cerca
```

Additionally, run [upx](https://upx.github.io) on any generated binary:

```
upx --lzma cerca
```
