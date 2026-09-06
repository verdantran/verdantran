package render

import (
	"fmt"
	"strings"

	"github.com/verdantran/verdantran/internal/content"
)

// A Theme is the whole palette of the readout: one colour per content class,
// and the screen those colours sit on. Both renderers take one, so the banner
// and the text block on the same page cannot drift apart.
//
// The ANSI side stays inside the basic sixteen. Those are the codes a viewer
// themes for itself, which is what keeps the block legible when the page is
// being read in light mode.
type Theme struct {
	Name string

	Fill map[string]string // content class -> SVG colour
	SGR  map[string]string // content class -> ANSI select-graphic-rendition

	Screen   string    // the dark glass everything sits on
	Vignette string    // the pool of light at its centre
	Edge     string    // border and rules
	Bar      string    // title bar
	Chrome   string    // title text
	Glow     string    // scanline and sweep tint
	Art      string    // the watermark behind the readout
	Lights   [3]string // the three dots on the title bar
	Frame    string    // ANSI code for the block's box drawing
}

// themes are offered in this order; the first is the default.
var themes = []Theme{{
	Name: "neon",
	Fill: map[string]string{
		content.Prompt: "#ff5fd2", content.Cmd: "#7ef7ff",
		content.Label: "#a7bccc", content.Dot: "#2b4353",
		content.Value: "#7ef7ff", content.BarOn: "#3dff92",
		content.BarOff: "#223946", content.Accent: "#ffc94a",
		content.Dim: "#5e7a8a", content.OK: "#3dff92",
	},
	SGR: map[string]string{
		content.Prompt: "95", content.Cmd: "96",
		content.Label: "37", content.Dot: "90",
		content.Value: "96", content.BarOn: "32",
		content.BarOff: "90", content.Accent: "33",
		content.Dim: "90", content.OK: "92",
	},
	Screen: "#04070c", Vignette: "#0b2233", Edge: "#164457",
	Bar: "#08111c", Chrome: "#4d7f95", Glow: "#7ef7ff", Art: "#1f7f9c",
	Lights: [3]string{"#ff5f8f", "#ffc94a", "#3dff92"}, Frame: "36",
}, {
	Name: "amber",
	Fill: map[string]string{
		content.Prompt: "#ff9a3c", content.Cmd: "#ffc76b",
		content.Label: "#c9a06a", content.Dot: "#4a3418",
		content.Value: "#ffc76b", content.BarOn: "#ffb454",
		content.BarOff: "#3a2a12", content.Accent: "#ffe08a",
		content.Dim: "#8a6a3a", content.OK: "#ffd166",
	},
	SGR: map[string]string{
		content.Prompt: "93", content.Cmd: "33",
		content.Label: "37", content.Dot: "90",
		content.Value: "33", content.BarOn: "33",
		content.BarOff: "90", content.Accent: "93",
		content.Dim: "90", content.OK: "93",
	},
	Screen: "#0a0703", Vignette: "#33200b", Edge: "#5a3a12",
	Bar: "#140d05", Chrome: "#9a6b2a", Glow: "#ffb454", Art: "#8a5a12",
	Lights: [3]string{"#ff9a3c", "#ffc94a", "#ffe08a"}, Frame: "33",
}, {
	Name: "phosphor",
	Fill: map[string]string{
		content.Prompt: "#7dffb0", content.Cmd: "#3dff92",
		content.Label: "#9ccdae", content.Dot: "#1c3a26",
		content.Value: "#3dff92", content.BarOn: "#6bff9e",
		content.BarOff: "#17301f", content.Accent: "#d2ff6b",
		content.Dim: "#5a8a68", content.OK: "#3dff92",
	},
	SGR: map[string]string{
		content.Prompt: "92", content.Cmd: "32",
		content.Label: "37", content.Dot: "90",
		content.Value: "32", content.BarOn: "92",
		content.BarOff: "90", content.Accent: "93",
		content.Dim: "90", content.OK: "92",
	},
	Screen: "#030805", Vignette: "#0b2a14", Edge: "#14522a",
	Bar: "#061109", Chrome: "#3f8a58", Glow: "#6bff9e", Art: "#1a7a45",
	Lights: [3]string{"#7dffb0", "#a8ff5f", "#3dff92"}, Frame: "32",
}, {
	Name: "violet",
	Fill: map[string]string{
		content.Prompt: "#ff7ad9", content.Cmd: "#c6a3ff",
		content.Label: "#b0a8cc", content.Dot: "#2e2450",
		content.Value: "#c6a3ff", content.BarOn: "#c77dff",
		content.BarOff: "#241c40", content.Accent: "#ffd36b",
		content.Dim: "#7a6f9c", content.OK: "#7dffc4",
	},
	SGR: map[string]string{
		content.Prompt: "95", content.Cmd: "95",
		content.Label: "37", content.Dot: "90",
		content.Value: "95", content.BarOn: "95",
		content.BarOff: "90", content.Accent: "33",
		content.Dim: "90", content.OK: "92",
	},
	Screen: "#07050e", Vignette: "#241a44", Edge: "#3b2a6b",
	Bar: "#0d0918", Chrome: "#7a68b0", Glow: "#b98cff", Art: "#5a3fa0",
	Lights: [3]string{"#ff7ad9", "#ffd36b", "#7dffc4"}, Frame: "35",
}, {
	Name: "mono",
	Fill: map[string]string{
		content.Prompt: "#ffffff", content.Cmd: "#dfe6ee",
		content.Label: "#9aa4b0", content.Dot: "#2a3038",
		content.Value: "#dfe6ee", content.BarOn: "#eef3f8",
		content.BarOff: "#232a32", content.Accent: "#c3ccd6",
		content.Dim: "#6b7480", content.OK: "#eef3f8",
	},
	SGR: map[string]string{
		content.Prompt: "97", content.Cmd: "37",
		content.Label: "37", content.Dot: "90",
		content.Value: "37", content.BarOn: "97",
		content.BarOff: "90", content.Accent: "97",
		content.Dim: "90", content.OK: "97",
	},
	Screen: "#06070a", Vignette: "#1b2028", Edge: "#333a44",
	Bar: "#0b0d12", Chrome: "#6b7480", Glow: "#d6dde6", Art: "#4a525c",
	Lights: [3]string{"#dfe6ee", "#9aa4b0", "#6b7480"}, Frame: "37",
}}

// LookupTheme finds a theme by name. The error names the alternatives, because
// a mistyped theme is otherwise a silent fall back to a palette nobody asked
// for.
func LookupTheme(name string) (Theme, error) {
	for _, t := range themes {
		if strings.EqualFold(name, t.Name) {
			return t, nil
		}
	}
	return Theme{}, fmt.Errorf("no theme named %q (have: %s)", name, ThemeNames())
}

// ThemeNames lists the themes on offer, default first.
func ThemeNames() string {
	names := make([]string, 0, len(themes))
	for _, t := range themes {
		names = append(names, t.Name)
	}
	return strings.Join(names, ", ")
}

// DefaultTheme is the palette used when none is asked for.
func DefaultTheme() Theme { return themes[0] }
