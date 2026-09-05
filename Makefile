FRAME ?= assets/frames/donut.txt

render:
	GITHUB_TOKEN=$$(gh auth token) go run ./cmd/profile -art $(FRAME)

preview: render
	open "$$(./scripts/preview.sh assets/terminal.svg -8s)"

show:
	GITHUB_TOKEN=$$(gh auth token) go run ./cmd/profile -dry-run

.PHONY: render preview show
