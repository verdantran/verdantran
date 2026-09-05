// Package ghstats pulls the numbers the profile advertises from GitHub's
// GraphQL API in a single request.
package ghstats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const endpoint = "https://api.github.com/graphql"

// Two repository sets come back: the public one, which is the only source of
// names the page is allowed to print, and everything the token can see, which
// feeds the aggregates. With a plain GITHUB_TOKEN the two are identical.
const queryHead = `query($login: String!) {
  user(login: $login) {
    login
    name
    createdAt
    followers { totalCount }
    public: repositories(first: 100, ownerAffiliations: OWNER, privacy: PUBLIC, isFork: false, orderBy: {field: PUSHED_AT, direction: DESC}) {
      totalCount
      nodes { ...bits }
    }
    all: repositories(first: 100, ownerAffiliations: OWNER, isFork: false, orderBy: {field: PUSHED_AT, direction: DESC}) {
      totalCount
      nodes { ...bits }
    }
    contributionsCollection {
      totalCommitContributions
      restrictedContributionsCount
      totalPullRequestContributions
    }
%s
  }
}

fragment bits on Repository {
  name
  description
  stargazerCount
  pushedAt
  languages(first: 8, orderBy: {field: SIZE, direction: DESC}) {
    edges { size node { name } }
  }
}`

// firstYear is GitHub's own founding year — a safe floor for the all-time
// walk, since a year before the account existed simply returns zero.
const firstYear = 2008

// buildQuery aliases one contributionsCollection per calendar year. The API
// caps a single collection at a twelve-month span, so an all-time commit count
// has to be summed year by year.
func buildQuery() string {
	var years strings.Builder
	for y := firstYear; y <= time.Now().UTC().Year(); y++ {
		fmt.Fprintf(&years, "    y%d: contributionsCollection(from: %q, to: %q) { totalCommitContributions restrictedContributionsCount }\n",
			y, fmt.Sprintf("%d-01-01T00:00:00Z", y), fmt.Sprintf("%d-12-31T23:59:59Z", y))
	}
	return fmt.Sprintf(queryHead, years.String())
}

type Repo struct {
	Name        string
	Description string
	Stars       int
	PushedAt    time.Time
}

type Lang struct {
	Name    string
	Percent float64
}

type Stats struct {
	Login       string
	Name        string
	CreatedAt   time.Time
	Followers   int
	PublicRepos int
	OtherRepos  int // private and internal, when the token can see them
	Stars       int
	Commits     int // last 365 days
	AllCommits  int
	PRs         int
	TopRepos    []Repo
	Langs       []Lang
	FetchedAt   time.Time
}

type repoNode struct {
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	StargazerCount int       `json:"stargazerCount"`
	PushedAt       time.Time `json:"pushedAt"`
	Languages      struct {
		Edges []struct {
			Size int `json:"size"`
			Node struct {
				Name string `json:"name"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"languages"`
}

type repoSet struct {
	TotalCount int        `json:"totalCount"`
	Nodes      []repoNode `json:"nodes"`
}

type contributions struct {
	TotalCommitContributions      int `json:"totalCommitContributions"`
	RestrictedContributionsCount  int `json:"restrictedContributionsCount"`
	TotalPullRequestContributions int `json:"totalPullRequestContributions"`
}

func (c contributions) commits() int {
	return c.TotalCommitContributions + c.RestrictedContributionsCount
}

type response struct {
	Data struct {
		User struct {
			Login     string    `json:"login"`
			Name      string    `json:"name"`
			CreatedAt time.Time `json:"createdAt"`
			Followers struct {
				TotalCount int `json:"totalCount"`
			} `json:"followers"`
			Public                  repoSet       `json:"public"`
			All                     repoSet       `json:"all"`
			ContributionsCollection contributions `json:"contributionsCollection"`
		} `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// Opts shapes the readout. SkipLangs drops languages whose byte counts lie —
// Jupyter notebooks carry their own rendered output, so they swamp everything
// else in a mixed account.
type Opts struct {
	TopRepos       int
	TopLangs       int
	SkipLangs      []string
	IncludePrivate bool // count private work in the aggregates; never name it
}

func Fetch(login, token string, o Opts) (*Stats, error) {
	body, err := json.Marshal(map[string]any{
		"query":     buildQuery(),
		"variables": map[string]string{"login": login},
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "verdantran-profile")

	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api: %s: %s", resp.Status, clip(raw, 200))
	}
	var r response
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, err
	}
	if len(r.Errors) > 0 {
		return nil, fmt.Errorf("github api: %s", r.Errors[0].Message)
	}
	u := r.Data.User
	if u.Login == "" {
		return nil, fmt.Errorf("github api: no user %q in response", login)
	}

	// The year aliases are decoded separately: their names are built at
	// request time, so they cannot be struct fields.
	var envelope struct {
		Data struct {
			User map[string]json.RawMessage `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, err
	}
	allCommits := 0
	for key, val := range envelope.Data.User {
		if !isYearAlias(key) {
			continue
		}
		var c contributions
		if err := json.Unmarshal(val, &c); err != nil {
			return nil, fmt.Errorf("github api: %s: %w", key, err)
		}
		allCommits += c.commits()
	}

	s := &Stats{
		Login:       u.Login,
		Name:        u.Name,
		CreatedAt:   u.CreatedAt,
		Followers:   u.Followers.TotalCount,
		PublicRepos: u.Public.TotalCount,
		OtherRepos:  max(u.All.TotalCount-u.Public.TotalCount, 0),
		Commits:     u.ContributionsCollection.commits(),
		AllCommits:  allCommits,
		PRs:         u.ContributionsCollection.TotalPullRequestContributions,
		FetchedAt:   time.Now().UTC(),
	}

	// Stars only ever come from public repos; a private star count means nothing.
	for _, n := range u.Public.Nodes {
		s.Stars += n.StargazerCount
		if len(s.TopRepos) < o.TopRepos {
			s.TopRepos = append(s.TopRepos, Repo{n.Name, n.Description, n.StargazerCount, n.PushedAt})
		}
	}

	mix := u.Public.Nodes
	if o.IncludePrivate {
		mix = u.All.Nodes
	}
	s.Langs = languages(mix, o.SkipLangs, o.TopLangs)
	return s, nil
}

func languages(nodes []repoNode, skipList []string, top int) []Lang {
	skip := map[string]bool{}
	for _, l := range skipList {
		if l = strings.ToLower(strings.TrimSpace(l)); l != "" {
			skip[l] = true
		}
	}

	sizes := map[string]int{}
	total := 0
	for _, n := range nodes {
		for _, e := range n.Languages.Edges {
			if skip[strings.ToLower(e.Node.Name)] {
				continue
			}
			sizes[e.Node.Name] += e.Size
			total += e.Size
		}
	}

	out := make([]Lang, 0, len(sizes))
	for name, size := range sizes {
		out = append(out, Lang{name, 100 * float64(size) / float64(max(total, 1))})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Percent != out[j].Percent {
			return out[i].Percent > out[j].Percent
		}
		return out[i].Name < out[j].Name
	})
	if len(out) > top {
		out = out[:top]
	}
	return out
}

// Uptime is the account age, worn as a terminal would wear it.
func (s *Stats) Uptime() string {
	d := time.Since(s.CreatedAt)
	years := int(d.Hours() / 24 / 365.25)
	days := int(d.Hours()/24) - int(float64(years)*365.25)
	return fmt.Sprintf("%dy %dd", years, days)
}

// Ago renders a timestamp the way a status line would.
func Ago(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", max(int(d.Minutes()), 1))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return fmt.Sprintf("%dy ago", int(d.Hours()/24/365))
	}
}

// isYearAlias matches the yNNNN keys buildQuery emits.
func isYearAlias(k string) bool {
	if len(k) != 5 || k[0] != 'y' {
		return false
	}
	for _, c := range k[1:] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func clip(b []byte, n int) string {
	if len(b) > n {
		return string(b[:n]) + "..."
	}
	return string(b)
}
