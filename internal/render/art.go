package render

import (
	"strings"
	"unicode/utf8"
)

// TrimArt crops a wakeart frame to its bounding box. The renderer pads every
// frame out to the full 80x24 viewport, which is mostly empty space.
func TrimArt(s string) []string {
	raw := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")

	top, bottom := -1, -1
	left, right := 1<<30, -1
	for i, line := range raw {
		r := []rune(strings.TrimRight(line, " \t"))
		if len(r) == 0 {
			continue
		}
		first := 0
		for first < len(r) && r[first] == ' ' {
			first++
		}
		if first == len(r) {
			continue
		}
		if top < 0 {
			top = i
		}
		bottom = i
		left = min(left, first)
		right = max(right, len(r))
	}
	if top < 0 {
		return nil
	}

	out := make([]string, 0, bottom-top+1)
	for _, line := range raw[top : bottom+1] {
		r := []rune(strings.TrimRight(line, " \t"))
		if len(r) < left {
			out = append(out, "")
			continue
		}
		out = append(out, strings.TrimRight(string(r[left:min(right, len(r))]), " "))
	}
	return out
}

func widest(lines []string) int {
	w := 0
	for _, l := range lines {
		w = max(w, utf8.RuneCountInString(l))
	}
	return w
}
