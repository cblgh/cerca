# Cerca
_a forum software_

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

## Config
Cerca supports community customization.

* Write a custom [about text](/defaults/sample-about.md) describing the community inhabiting the forum
* Define your own [registration rules](/defaults/sample-rules.md), [how to verify one's account](/defaults/sample-verification-instructions.md), and link to an existing code of conduct
* Set your own [custom logo](/defaults/sample-logo.html) (whether svg, png or emoji)
* Create your own theme by writing plain, frameworkless [css](/html/assets/theme.css)

To enable these customizations, there's a config file. To choose a config file, run cerca with
the `--config` option; the default config file is set to `./cerca.toml`.

```
cerca --config ./configs/cerca.toml
```

The configuration format is [TOML](https://toml.io/en/) and the config is populated with the following
defaults:

```TOML
[general]	
name = "" # whatever you want to name your forum; primarily used as display in tab titles
conduct_url = "" # optional + recommended: if omitted, the CoC checkboxes in /register will be hidden
language = "English" # Swedish, English. contributions for more translations welcome!

[rss]
feed_name = "" # defaults to [general]'s name if unset
feed_description = ""
forum_url = "" # should be forum index route https://example.com. used to generate post routes for feed, must be set to generate a feed

[documents]
logo =  "content/logo.html" # can contain emoji, <img>, <svg> etc. see defaults/sample-logo.html in repo for instructions
about = "content/about.md"
rules = "content/rules.md"
verification_explanation = "content/verification-instructions.md"
```

Content documents that are not found will be prepopulated using Cerca's [sample content
files](/defaults). The easiest thing to do is to run Cerca once and let it populate content
files using the samples, and then edit the files in `content/*` after the fact, before running
Cerca again to see your changes.

Either write your own configuration following the above format, or run cerca once to populate it and
then edit the created config.

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

To launch a local instance of the forum, run those commands (linux):

- `go run run.go --dev`

It should respond `Serving forum on :8277`. Just go on [http://localhost:8277](http://localhost:8277).

### Building binaries with reduced size 
This is optional, but if you want to minimize the size of any binary (whether it be the `cerca` executable 
or any of the binaries in [`cmd/`](/cmd/) follow the instructions below. Less useful for active development, more 
useful for sending binaries to other computers.

Pass `-ldflags="-s -w"` when building your binary:

```
go build -ldflags="-s -w" .
```

Additionally, run [upx](https://upx.github.io) on any generated binary:

```
upx --lzma cerca
```
