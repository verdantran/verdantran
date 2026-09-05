# How the readout works

`cmd/profile` asks GitHub's GraphQL API for the numbers, lays them out once in
`internal/content`, and hands that single layout to two renderers:

- `internal/render.SVG` draws `assets/terminal.svg` — a CRT terminal that types
  itself out, holds on the prompt, and loops.
- `internal/render.ANSI` draws the same lines as an ANSI-coloured block, which
  the README carries inside a `<details>` for anyone who wants selectable text.

The README is rewritten only between the `<!-- stats:start -->` and
`<!-- stats:end -->` markers. Everything else on the page is yours to edit.

## Running it locally

```
make show      # print the block to the terminal
make render    # rewrite README.md and assets/terminal.svg
make preview   # render, then screenshot the SVG mid-loop
```

`make` borrows a token from `gh auth token`. The generator needs `GITHUB_TOKEN`
set however you run it.

## The token

The workflow uses `secrets.PROFILE_TOKEN` if it exists and the built-in
`GITHUB_TOKEN` otherwise. The difference matters:

| | `GITHUB_TOKEN` | a PAT in `PROFILE_TOKEN` |
|---|---|---|
| public repos, stars, followers | yes | yes |
| commits in private repos | **no — reports 0** | yes |
| all-time commit count | public only | everything |
| language mix across private work | no | yes |

Private repositories are only ever counted. Their names and descriptions are
never printed: the named list is drawn from the public set alone.

To add the PAT: create a fine-grained or classic token with `repo` and
`read:user`, then `gh secret set PROFILE_TOKEN`.

Pass `-include-private=false` if you would rather the totals only ever describe
public work.

## Flags

```
-login            GitHub user to read (default verdantran)
-art              file holding a wakeart frame for the banner
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

## The art

The workflow installs `wakeart` and takes a live frame. While that repository
is private the install fails, and it falls back to one of the frames committed
under `assets/frames/`. Regenerate those any time with:

```
wakeart --once --scene globe --seed 7 > assets/frames/globe.txt
```

## If the panel goes stale

GitHub disables a scheduled workflow after 60 days without repository activity,
and a push made by `GITHUB_TOKEN` does not reset that clock. Run the `profile`
workflow by hand from the Actions tab and the schedule resumes.
