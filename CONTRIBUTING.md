# Contributing

Hello! Let's jam! Before that, though, let's slow down and talk about how to make contributions that will be
welcomed with open arms.

The goal of this forum software is to enable small non-commercial communities to have their own
place for longer-flowing conversations, and to maintain a living history. As part of that, it
should be easy to make this software your own. It should also be easy to host on any kind of
computer, given prior experience with hosting a simple html website. Contradicting this is the
messy reality that the current software is written for the explicit use of the [Merveilles]
community!

## Code contributions
In general, it's preferred to keep the code flexible than to impose (unnecessary) hierarchy at
this early stage, and to reach out before you attempt to add anything large-ish. Common sense,
basically.

In a bit more detail and in bullet-point format—the guiding principles:

* Communicate _before_ starting large rewrites or big features
* Keep the existing style and organization
* Flexibility before hierarchy 
* Have fun! There are other places to execute at a megacorp level
* Additions should benefit long-term use of the forum and longer form conversations
* New features should have a reasonable impact on the codebase
    * Said another way: new additions should not have an outsized impact on the _overall_ codebase
* The software should always be easy to host on a variety of devices, from powerful servers to smol memory-drained and storage-depleted computers
* As far as we are able to: avoid client-side javascript
    * Said another way: there should always be a way to do something without a functioning javascript engine
* Don't `go fmt` the entire codebase in the same PR as you're adding a feature; do that separately if it's needed
* The maintainer reserves the right to make final decisions, for example regarding whether
  something:
    * makes the codebase less fun to work with, or understandable
    * goes against the project's idea of benefitting conversations
    * does not compose well with the existing forum experience

At the end of the day, a maintainer must live with decisions made for the project—both good and
bad! That weight of responsibility is taken into account when looking at new contributions.

[Merveilles]: https://now.lectronice.com/notes/merveilles/
