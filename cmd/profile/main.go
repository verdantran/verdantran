// Command profile regenerates the readout on the GitHub profile page: an
// animated SVG terminal, and the same numbers as a selectable ANSI block.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/verdantran/verdantran/internal/content"
	"github.com/verdantran/verdantran/internal/ghstats"
	"github.com/verdantran/verdantran/internal/render"
)

const (
	startMark = "<!-- stats:start -->"
	endMark   = "<!-- stats:end -->"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "profile:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		login  = flag.String("login", "verdantran", "GitHub user to read")
		art    = flag.String("art", "", "file holding a wakeart frame for the banner")
		readme = flag.String("readme", "README.md", "README to rewrite between the stats markers")
		svg    = flag.String("svg", "assets/terminal.svg", "where to write the animated terminal")
		repos  = flag.Int("repos", 0, "how many public repositories to list; 0 drops the section")
		langs  = flag.Int("langs", 4, "how many languages to list")
		title  = flag.String("title", "", "terminal title (default <login>@github)")
		skip   = flag.String("exclude-langs", "Jupyter Notebook", "comma-separated languages to leave out of the mix")
		plain  = flag.Bool("plain", false, "emit the block without ANSI colour")
		priv   = flag.Bool("include-private", true, "count private work in the totals (needs a token that can see it); names are never printed")
		dryRun = flag.Bool("dry-run", false, "print the block to stdout, write nothing")
	)
	flag.Parse()

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return errors.New("GITHUB_TOKEN is not set")
	}
	if *title == "" {
		*title = *login + "@github"
	}

	stats, err := ghstats.Fetch(*login, token, ghstats.Opts{
		TopRepos:       *repos,
		TopLangs:       *langs,
		SkipLangs:      strings.Split(*skip, ","),
		IncludePrivate: *priv,
	})
	if err != nil {
		return err
	}
	lines := content.Build(stats)
	block := render.ANSI(lines, *title, !*plain)

	if *dryRun {
		fmt.Println(block)
		return nil
	}

	var frame []string
	if *art != "" {
		raw, err := os.ReadFile(*art)
		if err != nil {
			return err
		}
		frame = render.TrimArt(string(raw))
	}

	if err := write(*svg, render.SVG(lines, frame, *title)); err != nil {
		return err
	}
	patched, err := patchReadme(*readme, block, *plain)
	if err != nil {
		return err
	}
	written := *svg
	if patched {
		written += " and " + *readme
	}
	fmt.Printf("wrote %s (%d public repos, %d stars, %d commits all-time)\n",
		written, stats.PublicRepos, stats.Stars, stats.AllCommits)
	return nil
}

// patchReadme swaps whatever sits between the markers for a fresh fence,
// leaving the hand-written rest of the page alone. A README carrying no
// markers wants only the banner, and is left untouched.
func patchReadme(path, block string, plain bool) (bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	doc := string(raw)
	i := strings.Index(doc, startMark)
	j := strings.Index(doc, endMark)
	if i < 0 || j < 0 {
		return false, nil
	}
	if j < i {
		return false, fmt.Errorf("%s: %s appears before %s", path, endMark, startMark)
	}

	lang := "ansi"
	if plain {
		lang = ""
	}
	fenced := fmt.Sprintf("%s\n```%s\n%s\n```\n%s", startMark, lang, block, endMark)
	return true, write(path, doc[:i]+fenced+doc[j+len(endMark):])
}

func write(path, body string) error {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, []byte(body), 0o644)
}
