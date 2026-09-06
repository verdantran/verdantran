#!/usr/bin/env bash
# Screenshot the terminal SVG at a chosen point in its loop. Chrome's virtual
# clock does not drive CSS animations, so the frame is picked with a negative
# animation-delay instead. Prints the path it wrote.
set -euo pipefail
svg="${1:-assets/terminal.svg}"
at="${2:--8s}"

# A private 0700 directory, so neither the scratch page nor the screenshot
# lands on a name anyone else on the machine can predict or pre-create.
work="$(mktemp -d)"
html="$work/preview.html"
out="${3:-$work/terminal-preview.png}"
trap 'rm -f "$html"' EXIT

# The art frames each carry their own delay to take their turn, so the blanket
# override would leave every one of them off screen. Work out which frame the
# requested point in the loop lands on and pin that one open instead.
frames=$(grep -o 'class="art a' "$svg" | wc -l | tr -d ' ')
spin=$(sed -n 's/.*animation:flip \([0-9.]*\)s.*/\1/p' "$svg" | head -1)
pin=""
if [ "$frames" -gt 1 ] && [ -n "$spin" ]; then
  secs=${at#-}
  idx=$(awk -v s="${secs%s}" -v sp="$spin" -v n="$frames" \
    'BEGIN{i=int(s/sp*n)%n; if(i<0)i+=n; print i}')
  pin="svg .art{animation:none!important;visibility:hidden!important}"
  pin="$pin svg .a$idx{visibility:visible!important}"
fi

{
  printf '<!doctype html><meta charset="utf-8"><style>html,body{margin:0;background:#0d1117}'
  printf 'svg *{animation-delay:%s!important}%s</style>' "$at" "$pin"
  cat "$svg"
} > "$html"

w=$(sed -n 's/.*<svg[^>]*width="\([0-9]*\)".*/\1/p' "$svg" | head -1)
h=$(sed -n 's/.*<svg[^>]*height="\([0-9]*\)".*/\1/p' "$svg" | head -1)

"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
  --headless --disable-gpu --hide-scrollbars \
  --window-size="$w,$h" --screenshot="$out" "file://$html" 2>/dev/null
echo "$out"
