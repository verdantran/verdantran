# How the readout works

`cmd/profile` asks GitHub's GraphQL API for the numbers, lays them out once in
`internal/content`, and hands that single layout to two renderers:

- `internal/render.SVG` draws `assets/terminal.svg` — a CRT terminal that types
  itself out, holds on the prompt, and loops.
- `internal/render.ANSI` draws the same lines as an ANSI-coloured block. The
  page does not carry it, but `make show` prints it, which is the quickest way
  to see a change without opening a browser.

The page is only the banner. To put the text block on it as well, drop a pair
of `<!-- stats:start -->` / `<!-- stats:end -->` markers into `README.md` and
the generator will keep an ANSI fence between them — GitHub renders those in
colour. Without the markers the README is left alone entirely.

## Running it locally

```
make show      # print the block to the terminal
make render    # rewrite README.md and assets/terminal.svg
make preview   # render, then screenshot the SVG mid-loop
```

`make` borrows a token from `gh auth token`. The generator needs `GITHUB_TOKEN`
set however you run it.

## Flags

```
-login            GitHub user to read (default verdantran)
-art              file holding an ASCII art frame for the banner
-repos            how many public repositories to list; 0 drops the section
-langs            how many languages to list
-exclude-langs    languages to leave out of the mix (default "Jupyter Notebook")
-include-private  count private work in the totals (default true)
-plain            emit the block without ANSI colour
-dry-run          print the block to stdout, write nothing
```

`-exclude-langs` exists because GitHub sizes a language by bytes on disk, and a
Jupyter notebook stores its own rendered output — one notebook can outweigh
every other repository put together.

## The all-time commit count

GitHub caps a `contributionsCollection` at a twelve-month span, so the all-time
number is one aliased collection per calendar year from 2008 to now, summed —
still a single request. A year before the account existed just returns zero.

## If the panel goes stale

The workflow runs once a day. GitHub disables a scheduled workflow after 60
days without repository activity, and a push made by `GITHUB_TOKEN` does not
reset that clock. Run the `profile` workflow by hand from the Actions tab, from
`main`, and the schedule resumes.
