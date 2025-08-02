PREFIX = /usr/local
BINDIR = $(PREFIX)/bin
DATADIR = /var/lib/cerca
CONFDIR = /etc/cerca
CONF_FILE = '${CONFDIR}/cerca.toml'

cerca:
	go build ./cmd/cerca

install: cerca
	@# Run cerca's command to output a default config (with data dir and authkey set) and create the associated default content files 
	./cerca write-defaults --config '${CONF_FILE}' --data-dir '${DATADIR}'
	find '$(DATADIR)' '$(CONFDIR)' -type f -exec chmod 644 {} +
	find '$(DATADIR)' '$(CONFDIR)' -type d -exec chmod 755 {} +
	id cerca && chown -R cerca:cerca '$(DATADIR)' '$(CONFDIR)'
	@# Install binary
	install -m755 'cerca' '$(BINDIR)/cerca'
