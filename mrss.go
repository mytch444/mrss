package main

import (
	"fmt"
	"os"
	"unicode/utf8"
	"sync"

	"github.com/SlyMarbo/rss"
	"github.com/nsf/termbox-go"
)

var ConfigSuffixUrl = "/urls"

var lineMax = 80

var ctrlKeys = map[termbox.Key](func()) {
	termbox.KeyEnter:     open,
	termbox.KeyArrowDown: moveDown,
	termbox.KeyArrowUp:   moveUp,
	termbox.KeyPgdn:      pageDown,
	termbox.KeyCtrlF:     pageDown,
	termbox.KeySpace:     pageDown,
	termbox.KeyPgup:      pageUp,
	termbox.KeyCtrlB:     pageUp,
	termbox.KeyHome:      gotoTop,
	termbox.KeyEnd:       gotoBottom,
	termbox.KeyCtrlC:     exit,
}

var normKeys = map[rune](func()) {
	'j': moveDown,
	'k': moveUp,
	'o': open,
	'q': close,
	'r': updateFeeds,
	'g': gotoTop,
	'G': gotoBottom,
}

type Feed struct {
	Feed *rss.Feed
	Url string
}

var configPath string

var width, height int

var ModeFeeds = 1
var ModeItems = 2
var ModeItem = 3

var mode int
var feeds []*Feed
var cfeed int
var citem int
var offset int

var itemTitle string
var itemBody []string

var drawLock *sync.Mutex

func exit() {
	termbox.Close()
	os.Exit(0)
}

func open() {
	switch (mode) {
	case ModeFeeds:
		if feeds[cfeed].Feed != nil {
			offset = 0
			citem = 0
			mode = ModeItems
		}
	case ModeItems:
		if len(feeds[cfeed].Feed.Items) > 0 {
			offset = 0
			itemTitle, itemBody = fmtItem(feeds[cfeed].Feed.Items[citem])
			mode = ModeItem
		}
	case ModeItem:
		return
	}
}

func close() {
	switch (mode) {
	case ModeFeeds: exit()
	case ModeItems:
		mode = ModeFeeds
		offset = 0
	case ModeItem:
		offset = 0
		mode = ModeItems
	}
}

func moveDown() {
	switch (mode) {
	case ModeFeeds:
		if cfeed < len(feeds) - 1 {
			cfeed++
			if offset < height - 1 {
				offset++
			}
		}
	case ModeItems:
		if citem < len(feeds[cfeed].Feed.Items) - 1 {
			citem++
			if offset < height - 1 {
				offset++
			}
		}
	case ModeItem:
		if offset < len(itemBody) - 1 {
			offset++
		}
	}
}

func moveUp() {
	if offset > 0 {
		offset--
	}
	switch (mode) {
	case ModeFeeds:
		if cfeed > 0 {
			cfeed--
		}
	case ModeItems:
		if citem > 0 {
			citem--
		}
	}
}

func pageDown() {
	switch (mode) {
	case ModeFeeds:
		cfeed += height - 2
		if cfeed > len(feeds) - 1 {
			cfeed = len(feeds) - 1
		}
	case ModeItems:
		citem += height - 2
		if citem > len(feeds[cfeed].Feed.Items) - 1 {
			citem = len(feeds[cfeed].Feed.Items)
		}
	case ModeItem:
		offset += height - 3
		if offset > len(itemBody) - 1 {
			offset = len(itemBody) - 1
		}
	}
}

func pageUp() {
	switch (mode) {
	case ModeFeeds:
		cfeed -= height - 2
		if cfeed < 0 {
			cfeed = 0
		}
		if cfeed < offset {
			offset = cfeed
		}
	case ModeItems:
		citem -= height - 2
		if citem < 0 {
			citem = 0
		}
		if citem < offset {
			offset = citem
		}

	case ModeItem:
		offset -= height - 3
		if offset < 0 {
			offset = 0
		}
	}
}

func gotoTop() {
	offset = 0
	switch (mode) {
	case ModeFeeds:
		cfeed = 0
	case ModeItems:
		citem = 0
	}
}

func gotoBottom() {
	offset = height - 1
	switch (mode) {
	case ModeFeeds:
		cfeed = len(feeds) - 1
		if cfeed < 0 {
			cfeed = 0
		}
		if offset > cfeed {
			offset = cfeed
		}
	case ModeItems:
		citem = len(feeds[cfeed].Feed.Items) - 1
		if citem < 0 {
			citem = 0
		}
		if offset > citem {
			offset = citem
		}
	case ModeItem:
		offset = len(itemBody) - 1
		if offset < 0 {
			offset = 0
		}
	}

}

func updateFeeds() {
	for i := 0; i < len(feeds); i++ {
		go feeds[i].Feed.Update()
	}
}

func fmtItem(item *rss.Item) (string, []string) {
	body := make([]string, 0)
	body = append(body, "Date: " + item.Date.Format("2006 Jan 02 15:04:05 -0700 MST"))
	body = append(body, "Link: " + item.Link)
	body = append(body, "")
	
	str := item.Content
	line := make([]rune, lineMax)

	i := 0
	l := 0
	for r, s := utf8.DecodeRuneInString(str)
	    l < len(str);    
	    r, s = utf8.DecodeRuneInString(str[l:]) {
	
		if r == '\n' {
			body = append(body, string(line[:i]))
			i = 0
		} else if i >= lineMax {
			body = append(body, string(line[:i]))
			line[0] = r
			i = 1
		} else {
			line[i] = r
			i++
		}
		
		l += s;
	}

	return item.Title, body
}

func putString(str string, x, y int, fg, bg termbox.Attribute) {
	len := 0
	i := 0
	for r, s := utf8.DecodeRuneInString(str);
	    s > 0;
	    r, s = utf8.DecodeRuneInString(str[len:]) {
		termbox.SetCell(x+i, y, r, fg, bg)
		len += s
		i++
	}
}

func setLine(y int, bg termbox.Attribute) {
	for i := 0; i < width; i++ {
		termbox.SetCell(i, y, ' ', bg, bg)
	}
}

func displayMessage(mesg string) {
	fg := termbox.ColorWhite | termbox.AttrBold
	bg := termbox.ColorBlack
	
	setLine(height - 1, bg)
	putString("-> " + mesg, 0, height - 1, fg, bg)
}

func getFeed(f *Feed) {
	var err error
	f.Feed, err = rss.Fetch(f.Url)
	if err == nil {
		go redraw()
	}
}

func readFeeds() {
	var i, r int
	var err error
	
	file, err := os.Open(configPath + ConfigSuffixUrl)
	if err != nil {
		termbox.Close()
		fmt.Println("Error reading url file")
		os.Exit(1)
	}
	
	data := make([]byte, 2048)
	cfeed = 0
	for {
		r, err = file.Read(data)
		if err != nil || r == 0 {
			break
		}
		
		/* Find end of line */
		for i = 0; i < r; i++ {
			if data[i] == '\n' {
				break
			}
		}
		
		if i > r {
			fmt.Println("line too long!")
			exit()
		} else if i == 0 {
			file.Seek(int64(1 - r), 1)
			continue
		}
		
		n := new(Feed)
		n.Url = string(data[:i])
		feeds = append(feeds, n)
		
		go getFeed(n)
		
		file.Seek(int64(1 + i - r), 1)
	}
	
	cfeed = 0
}

func redrawFeeds() {
	fg := termbox.ColorBlack
	bg := termbox.ColorWhite
	var str string
	start:= cfeed - offset
	for i := start; i < len(feeds) && i - start < height; i++ {
		f := feeds[i]
		if f.Feed != nil {
			str = f.Feed.Title
		} else {
			str = f.Url
			termbox.SetCell(0, i - start, ' ', fg, fg)
		}
		
		if i == cfeed {
			setLine(i - start, fg)
			putString(str, 2, i - start, bg, fg) 
		} else {
			putString(str, 2, i - start, fg, bg) 
		}
	}
}

func redrawItems() {
	fg := termbox.ColorBlack
	bg := termbox.ColorWhite
	start := citem - offset
	for i := start; i < len(feeds[cfeed].Feed.Items) &&
	    i - start < height;
	    i++ {
		
		item := feeds[cfeed].Feed.Items[i]
		if i == citem {
			setLine(i - start, fg)
			putString(item.Title, 2, i - start, bg, fg) 
		} else {
			putString(item.Title, 2, i - start, fg, bg) 
		}
	}
}

func redrawItem() {
	setLine(0, termbox.ColorBlack)
	putString(itemTitle, 1, 0, termbox.ColorWhite|termbox.AttrBold,
	          termbox.ColorBlack)
	
	for i := offset;
	    i < len(itemBody) && i + 1 - offset < height;
	    i++ {
		putString(itemBody[i], 0, 1 + i - offset,
		          termbox.ColorBlack, termbox.ColorWhite)
	}
}

func redraw() {
	drawLock.Lock()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	switch (mode) {
	case ModeFeeds:
		redrawFeeds()
	case ModeItems:
		redrawItems()
	case ModeItem:
		redrawItem()
	}
	termbox.Flush()
	drawLock.Unlock()
}

func main() {
	var err error
	
	configPath = os.Getenv("HOME") + "/.config/mrss"
	
	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	
	termbox.SetInputMode(termbox.InputEsc)
	termbox.HideCursor()
	
	width, height = termbox.Size()
	offset = 0
	
	readFeeds()
	
	mode = ModeFeeds
	drawLock = new(sync.Mutex)
	
	for {
		redraw()
		
		ev := termbox.PollEvent()
		drawLock.Lock()
		switch ev.Type {
		case termbox.EventKey:
			if ev.Ch > 0 {
				f := normKeys[ev.Ch]
				if f != nil {
					f()
				}
			} else {
				f := ctrlKeys[ev.Key]
				if f != nil {
					f()
				}
			}
		case termbox.EventResize:
			width = ev.Width
			height = ev.Height
		
		case termbox.EventError:
			panic(ev.Err)
		}
		drawLock.Unlock()
	}
	
	termbox.Close()
}