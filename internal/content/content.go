// Package content lays out the readout once, as coloured segments, so the
// ANSI block and the SVG terminal can never drift apart.
package content

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/verdantran/verdantran/internal/ghstats"
)

// Class names a colour role. Both renderers map these to their own palette.
const (
	Prompt = "prompt"
	Cmd    = "cmd"
	Label  = "label"
	Dot    = "dot"
	Value  = "value"
	BarOn  = "bar-on"
	BarOff = "bar-off"
	Accent = "accent"
	Dim    = "dim"
	OK     = "ok"
)

type Seg struct {
	Text  string
	Class string
}

type Line []Seg

func (l Line) Width() int {
	n := 0
	for _, s := range l {
		n += utf8.RuneCountInString(s.Text)
	}
	return n
}

func (l Line) Plain() string {
	var b strings.Builder
	for _, s := range l {
		b.WriteString(s.Text)
	}
	return b.String()
}

// labelWidth is the column the values line up on, dot leaders filling the gap.
const labelWidth = 17

// Build turns the stats into the lines both renderers draw.
func Build(s *ghstats.Stats) []Line {
	var out []Line
	add := func(l Line) { out = append(out, l) }
	blank := func() { add(Line{}) }

	cmd := func(c string) {
		add(Line{{"$ ", Prompt}, {c, Cmd}})
	}
	field := func(label, value, class string) {
		add(Line{
			{"  ", Dim},
			{label + " ", Label},
			{strings.Repeat(".", max(labelWidth-utf8.RuneCountInString(label), 1)), Dot},
			{" " + value, class},
		})
	}
	// count prints a metric, or omits the row entirely when there is nothing
	// to report — an empty account reads better than a wall of zeroes.
	count := func(label string, n int) {
		if n == 0 {
			return
		}
		field(label, fmt.Sprintf("%d", n), Value)
	}

	cmd("./identity --scan")
	blank()
	who := s.Login
	if s.Name != "" {
		who = fmt.Sprintf("%s (%s)", safe(s.Name), s.Login)
	}
	field("OPERATOR", who, Value)
	field("UPTIME", s.Uptime(), Value)
	count("STARS", s.Stars)
	count("FOLLOWERS", s.Followers)
	count("COMMITS TOTAL", s.AllCommits)
	count("COMMITS / 365d", s.Commits)
	blank()

	if len(s.Langs) > 0 {
		cmd("./identity --languages")
		blank()
		for _, l := range s.Langs {
			add(Line{
				{"  ", Dim},
				{pad(l.Name, 16), Label},
				{" ", Dim},
			}.append(bar(l.Percent, 18)...).append(
				Seg{fmt.Sprintf(" %4.1f%%", l.Percent), Value},
			))
		}
		blank()
	}

	if len(s.TopRepos) > 0 {
		cmd("./identity --repos --recent " + fmt.Sprint(len(s.TopRepos)))
		blank()
		for _, r := range s.TopRepos {
			add(Line{
				{"  ", Dim},
				{pad(safe(r.Name), 20), Label},
				{pad(badge(r.Stars), 6), Accent},
				{pad(ghstats.Ago(r.PushedAt), 9), Dim},
				{truncate(safe(r.Description), 24), Dim},
			})
		}
		blank()
	}

	field("STATUS", "NOMINAL", OK)
	field("LAST SYNC", s.FetchedAt.Format("2006-01-02 15:04 UTC"), Dim)
	blank()
	add(Line{{"$ ", Prompt}})
	return out
}

func (l Line) append(segs ...Seg) Line { return append(l, segs...) }

// badge is a repo's star count, blank at zero. The column keeps its width
// either way, so the rows after it stay aligned.
func badge(n int) string {
	if n == 0 {
		return ""
	}
	return "★ " + fmt.Sprint(n)
}

// bar draws a proportion in block glyphs, split so each half can be coloured.
func bar(percent float64, width int) []Seg {
	on := int(percent/100*float64(width) + 0.5)
	on = min(max(on, 0), width)
	return []Seg{
		{strings.Repeat("█", on), BarOn},
		{strings.Repeat("░", width-on), BarOff},
	}
}

// safe neutralises free text from the API before it reaches a renderer: a
// backtick would close the README's code fence, and an ESC would be read as
// colour by the ANSI block. Runs before truncate, so widths stay honest.
func safe(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == '`':
			return '\''
		case r == '\n' || r == '\r' || r == '\t':
			return ' '
		case r < 0x20 || r == 0x7f:
			return -1
		}
		return r
	}, s)
}

func pad(s string, w int) string {
	s = truncate(s, w)
	return s + strings.Repeat(" ", max(w-utf8.RuneCountInString(s), 0))
}

func truncate(s string, w int) string {
	if utf8.RuneCountInString(s) <= w {
		return s
	}
	r := []rune(s)
	if w <= 1 {
		return string(r[:max(w, 0)])
	}
	return string(r[:w-1]) + "…"
}
