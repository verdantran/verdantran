package render

import (
	"strings"
	"unicode/utf8"
)

// frameSep separates one frame from the next inside an art file: a form feed,
// the character a printer took as "start a new page".
const frameSep = "\f"

// TrimFrames splits an art file into its frames and crops them all to one
// bounding box. Frames arrive padded out to the full 80x24 viewport, which is
// mostly empty space. The box has to be shared: cropping each frame to its own
// contents would re-centre a moving shape on every frame, so a spin would come
// out as a shudder.
func TrimFrames(s string) [][]string {
	frames := splitFrames(s)

	top, bottom := -1, -1
	left, right := 1<<30, -1
	for _, f := range frames {
		for i, line := range f {
			r := []rune(strings.TrimRight(line, " \t"))
			first := 0
			for first < len(r) && r[first] == ' ' {
				first++
			}
			if first == len(r) {
				continue
			}
			if top < 0 || i < top {
				top = i
			}
			bottom = max(bottom, i)
			left = min(left, first)
			right = max(right, len(r))
		}
	}
	if top < 0 {
		return nil
	}

	out := make([][]string, 0, len(frames))
	for _, f := range frames {
		out = append(out, crop(f, top, bottom, left, right))
	}
	return out
}

func splitFrames(s string) [][]string {
	var out [][]string
	for _, part := range strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), frameSep) {
		part = strings.Trim(part, "\n")
		if strings.TrimSpace(part) == "" {
			continue
		}
		out = append(out, strings.Split(part, "\n"))
	}
	return out
}

// crop cuts one frame to the shared box, padding short lines out to nothing
// rather than dropping them, so every frame keeps the same line count.
func crop(lines []string, top, bottom, left, right int) []string {
	out := make([]string, 0, bottom-top+1)
	for i := top; i <= bottom; i++ {
		if i >= len(lines) {
			out = append(out, "")
			continue
		}
		r := []rune(strings.TrimRight(lines[i], " \t"))
		if len(r) < left {
			out = append(out, "")
			continue
		}
		out = append(out, strings.TrimRight(string(r[left:min(right, len(r))]), " "))
	}
	return out
}

// widest measures the whole set, not one frame: the art is placed once and the
// frames swap underneath it, so they all have to be laid out to the same width.
func widest(frames [][]string) int {
	w := 0
	for _, f := range frames {
		for _, l := range f {
			w = max(w, utf8.RuneCountInString(l))
		}
	}
	return w
}
