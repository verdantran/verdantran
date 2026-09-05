#!/usr/bin/env bash
# Screenshot the terminal SVG at a chosen point in its loop. Chrome's virtual
# clock does not drive CSS animations, so the frame is picked with a negative
# animation-delay instead.
set -euo pipefail
svg="${1:-assets/terminal.svg}"
at="${2:--8s}"
out="${3:-/tmp/terminal-preview.png}"
html="$(mktemp -t preview).html"

{
  printf '<!doctype html><meta charset="utf-8"><style>html,body{margin:0;background:#0d1117}'
  printf 'svg *{animation-delay:%s!important}</style>' "$at"
  cat "$svg"
} > "$html"

w=$(sed -n 's/.*<svg[^>]*width="\([0-9]*\)".*/\1/p' "$svg" | head -1)
h=$(sed -n 's/.*<svg[^>]*height="\([0-9]*\)".*/\1/p' "$svg" | head -1)

"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
  --headless --disable-gpu --hide-scrollbars \
  --window-size="$w,$h" --screenshot="$out" "file://$html" 2>/dev/null
echo "$out"
