# RSS

## Where is my feed?

In the menu, on the top right, you'll see a `rss` menu item.

It's also available as `/rss.xml` on the end of your Cerca URL.

## Are private threads included in the feed?

Not yet. See [`#70`](https://github.com/cblgh/cerca/issues/70) for more.

## How is the feed generated?

The feed is intentionally low volume.

A feed item is generated per-thread. Only the latest poster is included in each feed item. When a new post is made in a thread, the feed item is updated "in place", in other words, the existing feed item is replaced without adding a new feed item.

No post content is included in the item. Instead, the feed is intended as low-tech notification mechanism, a reminder to revisit the forum and to join in on discussions that catch your eye.
