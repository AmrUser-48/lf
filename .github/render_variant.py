#!/usr/bin/env python3
from pathlib import Path
import sys

variant = sys.argv[1] if len(sys.argv) > 1 else ""
if variant not in {"hard", "light"}:
    raise SystemExit("variant must be hard or light")

path = Path("ui.go")
s = path.read_text()

old_len = """func printLength(s string) int {
\tlength := 0

\tslen := len(s)
\tfor i := 0; i < slen; {
\t\tseq := readTermSequence(s[i:])
\t\tif seq != "" {
\t\t\ti += len(seq)
\t\t\tcontinue
\t\t}

\t\tgc, w := firstGrapheme(s[i:])
\t\ti += len(gc)

\t\tif gc == "\\t" {
\t\t\tlength += gOpts.tabstop - length%gOpts.tabstop
\t\t} else if isPrintable(gc) {
\t\t\tlength += w
\t\t} else {
\t\t\tlength++ // U+FFFD replacement has width 1
\t\t}
\t}

\treturn length
}
"""

new_len_hard = """func printLength(s string) int {
\tlength := 0

\tslen := len(s)
\tfor i := 0; i < slen; {
\t\tseq := readTermSequence(s[i:])
\t\tif seq != "" {
\t\t\ti += len(seq)
\t\t\tcontinue
\t\t}

\t\tif s[i] >= ' ' && s[i] <= '~' {
\t\t\tj := i + 1
\t\t\tfor j < slen && s[j] >= ' ' && s[j] <= '~' {
\t\t\t\tj++
\t\t\t}
\t\t\tlength += j - i
\t\t\ti = j
\t\t\tcontinue
\t\t}

\t\tr, size := utf8.DecodeRuneInString(s[i:])
\t\ti += size

\t\tif r == '\\t' {
\t\t\tlength += gOpts.tabstop - length%gOpts.tabstop
\t\t} else if r == utf8.RuneError && size == 1 || isControlChar(r) {
\t\t\tlength++
\t\t} else {
\t\t\tlength += displaywidth.Rune(r)
\t\t}
\t}

\treturn length
}
"""

new_len_light = """func printLength(s string) int {
\tlength := 0

\tslen := len(s)
\tfor i := 0; i < slen; {
\t\tseq := readTermSequence(s[i:])
\t\tif seq != "" {
\t\t\ti += len(seq)
\t\t\tcontinue
\t\t}

\t\tif s[i] >= ' ' && s[i] <= '~' {
\t\t\tj := i + 1
\t\t\tfor j < slen && s[j] >= ' ' && s[j] <= '~' {
\t\t\t\tj++
\t\t\t}
\t\t\tlength += j - i
\t\t\ti = j
\t\t\tcontinue
\t\t}

\t\tgc, w := firstGrapheme(s[i:])
\t\tif gc == "" {
\t\t\tbreak
\t\t}
\t\ti += len(gc)

\t\tif gc == "\\t" {
\t\t\tlength += gOpts.tabstop - length%gOpts.tabstop
\t\t} else if isPrintable(gc) {
\t\t\tlength += w
\t\t} else {
\t\t\tlength++
\t\t}
\t}

\treturn length
}
"""

old_print = """func (win *win) print(screen tcell.Screen, x, y int, st tcell.Style, s string) tcell.Style {
\tvar b strings.Builder
\toff := 0
\tput := func() {
\t\tif b.Len() > 0 {
\t\t\ts := b.String()
\t\t\tscreen.PutStrStyled(win.x+x+off, win.y+y, s, st)
\t\t\toff += printLength(s)
\t\t\tb.Reset()
\t\t}
\t}

\tslen := len(s)
\tfor i := 0; i < slen; {
\t\tseq := readTermSequence(s[i:])
\t\tif seq != "" {
\t\t\tput()
\t\t\tst = applyTermSequence(seq, st)
\t\t\ti += len(seq)
\t\t\tcontinue
\t\t}

\t\tgc := firstGraphemeCluster(s[i:])
\t\tif gc == "\\t" {
\t\t\tw := gOpts.tabstop - (x+off+printLength(b.String()))%gOpts.tabstop
\t\t\tb.WriteString(strings.Repeat(" ", w))
\t\t} else if isPrintable(gc) {
\t\t\tb.WriteString(gc)
\t\t} else {
\t\t\tb.WriteString("\\uFFFD")
\t\t}

\t\ti += len(gc)
\t}

\tput()
\treturn st
}
"""

new_print_hard = """func (win *win) print(screen tcell.Screen, x, y int, st tcell.Style, s string) tcell.Style {
\toff := 0
\tslen := len(s)
\tfor i := 0; i < slen; {
\t\tseq := readTermSequence(s[i:])
\t\tif seq != "" {
\t\t\tst = applyTermSequence(seq, st)
\t\t\ti += len(seq)
\t\t\tcontinue
\t\t}

\t\tif s[i] >= ' ' && s[i] <= '~' {
\t\t\tj := i + 1
\t\t\tfor j < slen && s[j] >= ' ' && s[j] <= '~' {
\t\t\t\tj++
\t\t\t}
\t\t\tscreen.PutStrStyled(win.x+x+off, win.y+y, s[i:j], st)
\t\t\toff += j - i
\t\t\ti = j
\t\t\tcontinue
\t\t}

\t\tif s[i] == '\\t' {
\t\t\tw := gOpts.tabstop - (x+off)%gOpts.tabstop
\t\t\tscreen.PutStrStyled(win.x+x+off, win.y+y, strings.Repeat(" ", w), st)
\t\t\toff += w
\t\t\ti++
\t\t\tcontinue
\t\t}

\t\tr, size := utf8.DecodeRuneInString(s[i:])
\t\tif r == utf8.RuneError && size == 1 || isControlChar(r) {
\t\t\tscreen.PutStrStyled(win.x+x+off, win.y+y, "\\uFFFD", st)
\t\t\toff++
\t\t} else {
\t\t\tscreen.PutStrStyled(win.x+x+off, win.y+y, string(r), st)
\t\t\toff += displaywidth.Rune(r)
\t\t}
\t\ti += size
\t}
\treturn st
}
"""

new_print_light = """func (win *win) print(screen tcell.Screen, x, y int, st tcell.Style, s string) tcell.Style {
\toff := 0
\tslen := len(s)
\tfor i := 0; i < slen; {
\t\tseq := readTermSequence(s[i:])
\t\tif seq != "" {
\t\t\tst = applyTermSequence(seq, st)
\t\t\ti += len(seq)
\t\t\tcontinue
\t\t}

\t\tif s[i] >= ' ' && s[i] <= '~' {
\t\t\tj := i + 1
\t\t\tfor j < slen && s[j] >= ' ' && s[j] <= '~' {
\t\t\t\tj++
\t\t\t}
\t\t\tif j == slen || s[j] < utf8.RuneSelf {
\t\t\t\tscreen.PutStrStyled(win.x+x+off, win.y+y, s[i:j], st)
\t\t\t\toff += j - i
\t\t\t\ti = j
\t\t\t\tcontinue
\t\t\t}
\t\t\tif j-i > 1 {
\t\t\t\tend := j - 1
\t\t\t\tscreen.PutStrStyled(win.x+x+off, win.y+y, s[i:end], st)
\t\t\t\toff += end - i
\t\t\t\ti = end
\t\t\t}
\t\t}

\t\tgc, w := firstGrapheme(s[i:])
\t\tif gc == "" {
\t\t\tbreak
\t\t}
\t\tif gc == "\\t" {
\t\t\tw = gOpts.tabstop - (x+off)%gOpts.tabstop
\t\t\tscreen.PutStrStyled(win.x+x+off, win.y+y, strings.Repeat(" ", w), st)
\t\t\toff += w
\t\t} else if isPrintable(gc) {
\t\t\tscreen.PutStrStyled(win.x+x+off, win.y+y, gc, st)
\t\t\toff += w
\t\t} else {
\t\t\tscreen.PutStrStyled(win.x+x+off, win.y+y, "\\uFFFD", st)
\t\t\toff++
\t\t}
\t\ti += len(gc)
\t}
\treturn st
}
"""

new_len = new_len_hard if variant == "hard" else new_len_light
new_print = new_print_hard if variant == "hard" else new_print_light

if old_len not in s:
    raise SystemExit("printLength source not found")
s = s.replace(old_len, new_len, 1)
if old_print not in s:
    raise SystemExit("win.print source not found")
s = s.replace(old_print, new_print, 1)
path.write_text(s)
