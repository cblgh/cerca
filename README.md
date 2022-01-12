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
phpBB forum communities of the mid naughties.

It was written for the purpose of powering the nascent [Merveilles community forums](https://forum.merveilles.town).

## Local development

Install [golang](https://go.dev/).

To launch a local instance of the forum, run those commands (linux):

- `touch temp.txt`
- `mkdir data`
- `go run run.go --authkey 0 --allowlist temp.txt --dev`

It should respond `Serving forum on :8277`. Just go on [http://localhost:8272](http://localhost:8272).
