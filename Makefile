PREFIX = /usr/local
BINDIR = $(PREFIX)/bin
DATADIR = /var/lib/cerca
CONFDIR = /etc/cerca

cerca:
	go build ./cmd/cerca

install: cerca
	@# Install config
	mkdir -p '$(CONFDIR)'
	sed \
		's/auth_key = ""/auth_key = "$(shell ./cerca genauthkey)"/' \
		'defaults/sample-config.toml' \
		> '/tmp/config.toml'
	cp -i --no-preserve=mode,ownership '/tmp/config.toml' '$(CONFDIR)/config.toml' || true
	rm '/tmp/config.toml'
	@# Install data
	mkdir -p '$(DATADIR)/content'
	cp -ri --no-preserve=mode,ownership 'html/assets' '$(DATADIR)/' || true
	cp -i --no-preserve=mode,ownership 'defaults/sample-logo.html' '$(DATADIR)/content/logo.html' || true
	cp -i --no-preserve=mode,ownership 'defaults/sample-about.md' '$(DATADIR)/content/about.md' || true
	cp -i --no-preserve=mode,ownership 'defaults/sample-rules.md' '$(DATADIR)/content/rules.md' || true
	cp -i --no-preserve=mode,ownership 'defaults/sample-registration-instructions.md' '$(DATADIR)/content/registration-instructions.md' || true
	find '$(DATADIR)' '$(CONFDIR)' -type f -exec chmod 644 {} +
	find '$(DATADIR)' '$(CONFDIR)' -type d -exec chmod 755 {} +
	id cerca && chown -R cerca:cerca '$(DATADIR)' '$(CONFDIR)'
	@# Install binary
	install -m755 'cerca' '$(BINDIR)/cerca'
