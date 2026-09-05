package render

import (
	"fmt"
	"strings"

	"github.com/verdantran/verdantran/internal/content"
)

// sgr maps a segment class to a select-graphic-rendition code. Only the basic
// sixteen — GitHub themes those, so the block stays legible in light mode.
var sgr = map[string]string{
	content.Prompt: "95",
	content.Cmd:    "96",
	content.Label:  "37",
	content.Dot:    "90",
	content.Value:  "96",
	content.BarOn:  "32",
	content.BarOff: "90",
	content.Accent: "33",
	content.Dim:    "90",
	content.OK:     "92",
}

const frameSGR = "36"

// ANSI draws the readout as a framed terminal panel. With colour off it emits
// no escape sequences, for the case where a renderer shows them raw.
func ANSI(lines []content.Line, title string, colour bool) string {
	inner := 60
	for _, l := range lines {
		inner = max(inner, l.Width()+2)
	}

	paint := func(text, class string) string {
		code, ok := sgr[class]
		if !colour || !ok || text == "" {
			return text
		}
		return "\x1b[" + code + "m" + text + "\x1b[0m"
	}
	frame := func(s string) string {
		if !colour {
			return s
		}
		return "\x1b[" + frameSGR + "m" + s + "\x1b[0m"
	}

	var b strings.Builder
	head := fmt.Sprintf("┌─ %s ", title)
	b.WriteString(frame(head + strings.Repeat("─", max(inner+2-runes(head)-1, 0)) + "┐"))
	b.WriteByte('\n')

	for _, l := range lines {
		var body strings.Builder
		for _, s := range l {
			body.WriteString(paint(s.Text, s.Class))
		}
		body.WriteString(strings.Repeat(" ", max(inner-l.Width(), 0)))
		b.WriteString(frame("│") + " " + body.String() + " " + frame("│"))
		b.WriteByte('\n')
	}

	b.WriteString(frame("└" + strings.Repeat("─", inner+2) + "┘"))
	return b.String()
}

func runes(s string) int { return len([]rune(s)) }
