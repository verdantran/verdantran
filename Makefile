FRAME ?= assets/frames/donut.txt
THEME ?= neon

render:
	GITHUB_TOKEN=$$(gh auth token) go run ./cmd/profile -art $(FRAME) -theme $(THEME)

preview: render
	open "$$(./scripts/preview.sh assets/terminal.svg -8s)"

show:
	GITHUB_TOKEN=$$(gh auth token) go run ./cmd/profile -theme $(THEME) -dry-run

.PHONY: render preview show
