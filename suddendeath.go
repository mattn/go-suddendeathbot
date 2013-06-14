package main

import (
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"os"
	"regexp"
	"strings"
)

var reSuddenDeath = regexp.MustCompile(`^(>+)([^<]+)(<+)$`)

func runeWidth(r rune) int {
	if r >= 0x1100 &&
		(r <= 0x115f || r == 0x2329 || r == 0x232a ||
			(r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
			(r >= 0xac00 && r <= 0xd7a3) ||
			(r >= 0xf900 && r <= 0xfaff) ||
			(r >= 0xfe30 && r <= 0xfe6f) ||
			(r >= 0xff00 && r <= 0xff60) ||
			(r >= 0xffe0 && r <= 0xffe6) ||
			(r >= 0x20000 && r <= 0x2fffd) ||
			(r >= 0x30000 && r <= 0x3fffd)) {
		return 2
	}
	return 1
}

func strWidth(str string) int {
	r := 0
	for _, c := range []rune(str) {
		r += runeWidth(c)
	}
	return r
}

func suddenDeath(msg string) string {
	lines := strings.Split(msg, "\n")
	widths := []int{}

	maxWidth := 0
	for _, line := range lines {
		width := strWidth(line)
		widths = append(widths, width)
		if maxWidth < width {
			maxWidth = width
		}
	}

	ret := "＿" + strings.Repeat("人", maxWidth/2+2) + "＿\n"
	for i, line := range lines {
		ret += "＞　" + line + strings.Repeat(" ", maxWidth-widths[i]) + "　＜\n"
	}
	ret += "￣" + strings.Repeat("Ｙ", maxWidth/2+2) + "￣\n"
	return ret
}

func main() {
	c := irc.SimpleClient("suddendeath", "suddendeath")
	c.EnableStateTracking()

	c.AddHandler("connected", func(conn *irc.Conn, line *irc.Line) {
		for _, room := range os.Args[1:] {
			c.Join("#" + room)
		}
	})

	quit := make(chan bool)
	c.AddHandler("disconnected", func(conn *irc.Conn, line *irc.Line) {
		quit <- true
	})

	c.AddHandler("privmsg", func(conn *irc.Conn, line *irc.Line) {
		println(line.Src, line.Args[0], line.Args[1])
		if reSuddenDeath.MatchString(line.Args[1]) {
			m := reSuddenDeath.FindStringSubmatch(line.Args[1])
			if len(m[1]) == len(m[3]) {
				result := m[2]
				nl := len(m[1])
				for n := 0; n < nl; n++ {
					result = suddenDeath(strings.TrimRight(result, "\n"))
				}
				for _, s := range strings.Split(result, "\n") {
					c.Notice(line.Args[0], s)
				}
			}
		}
	})

	for {
		if err := c.Connect("irc.freenode.net:6667"); err != nil {
			fmt.Printf("Connection error: %s\n", err)
			return
		}
		<-quit
	}
}
