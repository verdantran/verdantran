FRAME ?= assets/frames/donut.txt

render:
	GITHUB_TOKEN=$$(gh auth token) go run ./cmd/profile -art $(FRAME)

preview: render
	./scripts/preview.sh assets/terminal.svg -8s /tmp/terminal-preview.png
	open /tmp/terminal-preview.png

show:
	GITHUB_TOKEN=$$(gh auth token) go run ./cmd/profile -dry-run

.PHONY: render preview show
