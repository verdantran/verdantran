package render

import (
	"fmt"
	"strings"

	"github.com/verdantran/verdantran/internal/content"
)

// ANSI draws the readout as a framed terminal panel, in the theme's own codes.
// With colour off it emits no escape sequences, for the case where a renderer
// shows them raw.
func ANSI(lines []content.Line, title string, colour bool, th Theme) string {
	inner := 60
	for _, l := range lines {
		inner = max(inner, l.Width()+2)
	}

	paint := func(text, class string) string {
		code, ok := th.SGR[class]
		if !colour || !ok || text == "" {
			return text
		}
		return "\x1b[" + code + "m" + text + "\x1b[0m"
	}
	frame := func(s string) string {
		if !colour {
			return s
		}
		return "\x1b[" + th.Frame + "m" + s + "\x1b[0m"
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
