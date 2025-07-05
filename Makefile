PREFIX = /usr/local
BINDIR = $(PREFIX)/bin
DATADIR = /var/lib/cerca
CONFDIR = /etc/cerca
CONF_FILE = '${CONFDIR}/cerca.toml'

cerca:
	go build ./cmd/cerca

install: cerca
	@# Install config
	mkdir -p '$(CONFDIR)'
	mkdir -p '$(DATADIR)'
	@# run cerca once to output the default config
	./cerca --config '${CONF_FILE}' || true
	@# use sed to populate field auth_key of default config
	sed -i \
		's/auth_key = ""/auth_key = "$(shell ./cerca genauthkey)"/' \
		'$(CONF_FILE)'
	@# use sed to populate data_dir of default config to what is defined by this makefile
	sed -i \
		's|data_dir = "/var/lib/cerca"|data_dir = "$(DATADIR)"|' \
		'$(CONF_FILE)'
	@# Install data
	@# run cerca for 1 sec to populate the default data files
	timeout 1 ./cerca --port 9999 --config '${CONF_FILE}' || true
	find '$(DATADIR)' '$(CONFDIR)' -type f -exec chmod 644 {} +
	find '$(DATADIR)' '$(CONFDIR)' -type d -exec chmod 755 {} +
	id cerca && chown -R cerca:cerca '$(DATADIR)' '$(CONFDIR)'
	@# Install binary
	install -m755 'cerca' '$(BINDIR)/cerca'
