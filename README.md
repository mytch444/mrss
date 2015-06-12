mrss
====

Minimal rss reader for the console written in go.

Requires
	[termbox](https://github.com/nsf/termbox-go),
	[rss](https://github.com/SlyMarbo/rss),
	and by extention
	[mahonia](https://github.com/axgle/mahonia)

It is very minimal.

Reads a file with a list of urls that you should create and populate
as `$HOME/.config/mrss/urls`.

Currently doesn't work very well with feeds that contain html. In the
near future I will add features that will strip it out and pressent it
somewhat better.

