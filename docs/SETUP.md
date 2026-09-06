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
never printed: the named list is drawn from the public set alone. The language
mix is the exception the table above records — with a PAT it is a distribution
over private code too, so it says something about work the page does not name.
Pass `-include-private=false` if that is more than you want to publish.

To add the PAT: create a **fine-grained** token owned by your own account, set
**repository access to all repositories**, and leave every repository
permission at *no access*. The *Metadata: read* that every fine-grained token
carries and cannot drop is the only one this needs. Nothing else — not
*Contents*, not *Followers* — buys anything: the generator reads repository
metadata and the user-level contribution graph, and never opens a file.

Repository access is the setting that does the work, not the permissions.
Narrow it to selected repositories and the `all:` set collapses onto the public
one, taking the private half of every total with it.

The token lives on the `prod` environment, which is restricted to `main`:

```
gh secret set PROFILE_TOKEN --env prod
```

An environment secret is only visible to a job that names the environment, so
`profile.yml` declares `environment: prod`. Drop that line and the workflow
still passes — `secrets.PROFILE_TOKEN` resolves to empty, the `||` falls
through to `GITHUB_TOKEN`, and the readout quietly loses its private numbers.
The same silence follows an expired token, so if the totals shrink overnight,
suspect the token before you suspect the arithmetic.

If private commits still read as zero with the PAT in place, check **Settings →
Public profile → Include private contributions on my profile**. The API gates
`restrictedContributionsCount` on that switch as well as on the token.

The job only ever reads, and the commit it makes is pushed by the workflow's
own `GITHUB_TOKEN`, so the PAT never needs write. Leaked, a metadata-only token
discloses the names, descriptions, sizes, languages and push times of private
repositories, and who collaborates on them — but no code, no issues, no
Actions, and nothing owned by an organisation. A classic token with `repo`
would leak all of it and grant write besides, which is why this is fine-grained.

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

The banner draws `assets/frames/donut.txt` — a single frame captured from
[wakeart](https://github.com/verdantran/wakeart), committed here rather than
regenerated on every run. To change the picture, drop in another frame and
point the workflow's `-art` at it:

```
wakeart --once --scene ridge --seed 7 > assets/frames/ridge.txt
```

## If the panel goes stale

The workflow runs once a day. GitHub disables a scheduled workflow after 60
days without repository activity, and a push made by `GITHUB_TOKEN` does not
reset that clock. Run the `profile` workflow by hand from the Actions tab and
the schedule resumes — **from `main`**, since the `prod` environment refuses
every other branch.
