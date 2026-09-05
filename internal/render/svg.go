package render

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/verdantran/verdantran/internal/content"
)

// The SVG is drawn on a fixed character grid: every span is positioned by
// column and forced to an exact width, so the layout survives whatever
// monospace font the viewer happens to have.
const (
	svgW      = 1000
	titleH    = 34
	padTop    = 26
	padBottom = 26
	padLeft   = 26
	fontSize  = 14
	lineH     = 21
	charW     = 8.4

	artX      = 560
	artW      = 410
	artFont   = 9.0
	artLineH  = 10.0
	artCharW  = 5.4
	artOffset = 0.62 // scale of the fill colour's opacity

	loopSecs = 11.0
	typeSecs = 0.17 // per line
	holdSecs = 4.5
)

var fill = map[string]string{
	content.Prompt: "#ff5fd2",
	content.Cmd:    "#7ef7ff",
	content.Label:  "#a7bccc",
	content.Dot:    "#2b4353",
	content.Value:  "#7ef7ff",
	content.BarOn:  "#3dff92",
	content.BarOff: "#223946",
	content.Accent: "#ffc94a",
	content.Dim:    "#5e7a8a",
	content.OK:     "#3dff92",
}

// SVG draws the readout as an animated CRT terminal: lines type themselves in
// sequence, then the loop holds on the prompt and restarts.
func SVG(lines []content.Line, art []string, title string) string {
	bodyH := len(lines) * lineH
	h := titleH + padTop + bodyH + padBottom

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">`,
		svgW, h, svgW, h, esc(title))

	b.WriteString(defs(h))
	b.WriteString(styles(lines))

	// Screen.
	fmt.Fprintf(&b, `<rect x="1" y="1" width="%d" height="%d" rx="10" fill="#04070c" stroke="#164457"/>`, svgW-2, h-2)
	fmt.Fprintf(&b, `<rect x="1" y="1" width="%d" height="%d" rx="10" fill="url(#vignette)"/>`, svgW-2, h-2)

	// Title bar.
	fmt.Fprintf(&b, `<path d="M1 11a10 10 0 0 1 10-10h%d a10 10 0 0 1 10 10v%d H1Z" fill="#08111c"/>`, svgW-22, titleH-11)
	fmt.Fprintf(&b, `<line x1="1" y1="%d" x2="%d" y2="%d" stroke="#164457"/>`, titleH, svgW-1, titleH)
	for i, c := range []string{"#ff5f8f", "#ffc94a", "#3dff92"} {
		fmt.Fprintf(&b, `<circle cx="%d" cy="17" r="5" fill="%s" opacity=".85"/>`, 24+i*18, c)
	}
	fmt.Fprintf(&b, `<text x="%d" y="22" class="chrome">%s</text>`, svgW/2-90, esc(title))

	b.WriteString(`<g filter="url(#glow)">`)
	b.WriteString(artGroup(art, bodyH))

	// Readout.
	base := float64(titleH + padTop + fontSize)
	for i, l := range lines {
		if len(l) == 0 {
			continue
		}
		fmt.Fprintf(&b, `<text class="l l%d" y="%.1f" xml:space="preserve">`, i, base+float64(i)*lineH)
		col := 0
		for _, s := range l {
			n := utf8.RuneCountInString(s.Text)
			if n == 0 {
				continue
			}
			if strings.TrimSpace(s.Text) != "" {
				fmt.Fprintf(&b, `<tspan x="%.1f" textLength="%.1f" lengthAdjust="spacingAndGlyphs" fill="%s">%s</tspan>`,
					float64(padLeft)+float64(col)*charW, float64(n)*charW, fill[s.Class], esc(s.Text))
			}
			col += n
		}
		b.WriteString(`</text>`)
	}

	// The cursor sits where the final prompt leaves off.
	cx := float64(padLeft) + float64(lines[len(lines)-1].Width())*charW
	cy := base + float64(len(lines)-1)*lineH - float64(fontSize) + 2
	fmt.Fprintf(&b, `<rect class="cursor" x="%.1f" y="%.1f" width="%.1f" height="%d" fill="#7ef7ff"/>`,
		cx, cy, charW, fontSize+3)
	b.WriteString(`</g>`)

	fmt.Fprintf(&b, `<rect x="1" y="1" width="%d" height="%d" rx="10" fill="url(#scanlines)"/>`, svgW-2, h-2)
	fmt.Fprintf(&b, `<rect class="sweep" x="1" width="%d" height="90" fill="url(#sweep)"/>`, svgW-2)
	b.WriteString(`</svg>`)
	return b.String()
}

func defs(h int) string {
	return fmt.Sprintf(`<defs>`+
		`<pattern id="scanlines" width="1" height="3" patternUnits="userSpaceOnUse">`+
		`<rect width="1" height="1" fill="#8ff" opacity=".055"/></pattern>`+
		`<radialGradient id="vignette" cx="50%%" cy="42%%" r="78%%">`+
		`<stop offset="0%%" stop-color="#0b2233" stop-opacity=".55"/>`+
		`<stop offset="100%%" stop-color="#04070c" stop-opacity="0"/></radialGradient>`+
		`<linearGradient id="sweep" x1="0" y1="0" x2="0" y2="1">`+
		`<stop offset="0%%" stop-color="#7ef7ff" stop-opacity="0"/>`+
		`<stop offset="50%%" stop-color="#7ef7ff" stop-opacity=".045"/>`+
		`<stop offset="100%%" stop-color="#7ef7ff" stop-opacity="0"/></linearGradient>`+
		`<filter id="glow" x="-4%%" y="-4%%" width="108%%" height="108%%">`+
		`<feGaussianBlur stdDeviation="1.1" result="b"/>`+
		`<feMerge><feMergeNode in="b"/><feMergeNode in="SourceGraphic"/></feMerge></filter>`+
		`<clipPath id="artclip"><rect x="%d" y="%d" width="%d" height="%d"/></clipPath>`+
		`</defs>`, artX, titleH+padTop-14, artW, h-titleH-padTop-padBottom+14)
}

// styles emits one keyframe set per line: hidden, then a stepped wipe across
// its own character count, then held for the rest of the loop.
func styles(lines []content.Line) string {
	var b strings.Builder
	b.WriteString(`<style>`)
	fmt.Fprintf(&b, `text{font-family:ui-monospace,SFMono-Regular,"SF Mono",Menlo,Consolas,"DejaVu Sans Mono",monospace;font-size:%dpx}`, fontSize)
	b.WriteString(animation(lines))
	b.WriteString(`</style>`)
	return b.String()
}

func animation(lines []content.Line) string {
	var b strings.Builder
	b.WriteString(`.chrome{fill:#4d7f95;font-size:12px;letter-spacing:1.5px}`)
	fmt.Fprintf(&b, `.art{fill:#1f7f9c;font-size:%.1fpx;opacity:%.2f}`, artFont, artOffset*0.55)
	fmt.Fprintf(&b, `.l{animation-duration:%.2fs;animation-iteration-count:infinite;animation-fill-mode:both}`, loopSecs)

	// Lines type in source order; blanks still consume a beat so the rhythm
	// matches a real shell scrolling.
	elapsed := 0.0
	for i, l := range lines {
		start := elapsed
		dur := typeSecs
		if len(l) == 0 {
			elapsed += typeSecs * 0.35
			continue
		}
		elapsed += dur
		n := max(l.Width(), 1)
		a := pct(start, loopSecs)
		c := pct(start+dur, loopSecs)
		fmt.Fprintf(&b, `@keyframes k%d{0%%,%.3f%%{clip-path:inset(-6px 100%% -6px 0)}%.3f%%,100%%{clip-path:inset(-6px -14px -6px 0)}}`, i, a, c)
		fmt.Fprintf(&b, `.l%d{animation-name:k%d;animation-timing-function:steps(%d,end)}`, i, i, n)
	}
	if elapsed+holdSecs > loopSecs {
		// Keep the hold honest if the readout ever outgrows the loop.
		fmt.Fprintf(&b, `/* typing %.1fs of %.1fs loop */`, elapsed, loopSecs)
	}

	done := pct(elapsed, loopSecs)
	fmt.Fprintf(&b, `@keyframes reveal{0%%,%.3f%%{visibility:hidden}%.3f%%,100%%{visibility:visible}}`, done, done)
	b.WriteString(`@keyframes blink{0%,49%{opacity:1}50%,100%{opacity:0}}`)
	fmt.Fprintf(&b, `.cursor{animation:reveal %.2fs step-end infinite,blink 1.06s step-end infinite}`, loopSecs)
	fmt.Fprintf(&b, `@keyframes sweep{0%%{transform:translateY(-120px)}100%%{transform:translateY(%dpx)}}`, 1200)
	b.WriteString(`.sweep{animation:sweep 7.5s linear infinite}`)
	return b.String()
}

// artGroup places a trimmed wakeart frame in the right-hand panel, scaled to
// fit and dimmed to a watermark.
func artGroup(art []string, bodyH int) string {
	if len(art) == 0 {
		return ""
	}
	w := float64(widest(art)) * artCharW
	h := float64(len(art)) * artLineH
	if w <= 0 || h <= 0 {
		return ""
	}
	scale := min(float64(artW)/w, float64(bodyH-10)/h)
	scale = min(scale, 2.4)

	x := float64(artX) + (float64(artW)-w*scale)/2
	y := float64(titleH+padTop) + (float64(bodyH)-h*scale)/2

	var b strings.Builder
	fmt.Fprintf(&b, `<g clip-path="url(#artclip)"><g transform="translate(%.1f,%.1f) scale(%.3f)">`, x, y, scale)
	for i, line := range art {
		if strings.TrimSpace(line) == "" {
			continue
		}
		n := utf8.RuneCountInString(line)
		fmt.Fprintf(&b, `<text class="art" x="0" y="%.1f" textLength="%.1f" lengthAdjust="spacingAndGlyphs" xml:space="preserve">%s</text>`,
			float64(i+1)*artLineH, float64(n)*artCharW, esc(line))
	}
	b.WriteString(`</g></g>`)
	return b.String()
}

func pct(t, total float64) float64 { return min(t/total*100, 100) }

var escaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")

func esc(s string) string { return escaper.Replace(s) }
