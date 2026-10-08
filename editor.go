package main

// The internal editor follows the modal command conventions of BusyBox vi.
// It is a fresh Go implementation and does not include BusyBox source.
import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
)

type editorMode uint8

const (
	editorNormal editorMode = iota
	editorInsert
	editorReplace
	editorCommand
	editorSearch
)

type editorSnapshot struct {
	lines [][]rune
	row   int
	col   int
}

type internalEditor struct {
	path          string
	lines         [][]rune
	row           int
	col           int
	top           int
	hscroll       int
	mode          editorMode
	input         []rune
	pending       rune
	search        string
	searchForward bool
	yank          [][]rune
	undo          []editorSnapshot
	modified      bool
	finalNewline  bool
	eol           string
	status        string
}

func cloneEditorLines(lines [][]rune) [][]rune {
	out := make([][]rune, len(lines))
	for i := range lines {
		out[i] = append([]rune(nil), lines[i]...)
	}
	return out
}

func openInternalEditor(path string) (*internalEditor, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("file is not valid UTF-8")
	}

	eol := "\n"
	if bytes.Contains(data, []byte("\r\n")) {
		eol = "\r\n"
	}
	finalNewline := len(data) > 0 && data[len(data)-1] == '\n'
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))

	parts := strings.Split(string(data), "\n")
	if finalNewline && len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) == 0 {
		parts = []string{""}
	}

	lines := make([][]rune, len(parts))
	for i, line := range parts {
		lines[i] = []rune(line)
	}

	return &internalEditor{
		path:         path,
		lines:        lines,
		eol:          eol,
		finalNewline: finalNewline,
		status:       "NORMAL",
	}, nil
}

func (e *internalEditor) save() error {
	info, err := os.Stat(e.path)
	if err != nil {
		return err
	}

	var b strings.Builder
	for i, line := range e.lines {
		b.WriteString(string(line))
		if i+1 < len(e.lines) || e.finalNewline {
			b.WriteByte('\n')
		}
	}

	data := strings.ReplaceAll(b.String(), "\n", e.eol)
	dir := filepath.Dir(e.path)
	base := filepath.Base(e.path)

	tmp, err := os.CreateTemp(dir, "."+base+".lf-edit-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if err := tmp.Chmod(info.Mode().Perm()); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.WriteString(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, e.path); err != nil {
		return err
	}

	e.modified = false
	e.status = "written"
	return nil
}

func (e *internalEditor) snapshot() {
	e.undo = append(e.undo, editorSnapshot{
		lines: cloneEditorLines(e.lines),
		row:   e.row,
		col:   e.col,
	})
	if len(e.undo) > 64 {
		e.undo = e.undo[len(e.undo)-64:]
	}
}

func (e *internalEditor) undoLast() {
	if len(e.undo) == 0 {
		e.status = "already at oldest change"
		return
	}
	last := e.undo[len(e.undo)-1]
	e.undo = e.undo[:len(e.undo)-1]
	e.lines = cloneEditorLines(last.lines)
	e.row = min(last.row, len(e.lines)-1)
	e.col = min(last.col, len(e.lines[e.row]))
	e.modified = true
	e.pending = 0
	e.status = "undo"
}

func (e *internalEditor) current() []rune {
	if len(e.lines) == 0 {
		e.lines = [][]rune{{}}
	}
	e.row = min(max(e.row, 0), len(e.lines)-1)
	return e.lines[e.row]
}

func (e *internalEditor) clampCol() {
	line := e.current()
	e.col = min(max(e.col, 0), len(line))
	if e.mode == editorNormal && len(line) > 0 {
		e.col = min(e.col, len(line)-1)
	}
}

func (e *internalEditor) ensureVisible(height, width int) {
	height = max(height, 1)
	width = max(width, 1)

	if e.row < e.top {
		e.top = e.row
	}
	if e.row >= e.top+height {
		e.top = e.row - height + 1
	}
	e.top = min(max(e.top, 0), max(len(e.lines)-1, 0))

	line := e.current()
	e.hscroll = min(max(e.hscroll, 0), len(line))
	for e.hscroll < e.col && displaywidth.String(string(line[e.hscroll:e.col])) >= width {
		e.hscroll++
	}
	for e.hscroll > 0 && displaywidth.String(string(line[e.hscroll-1:e.col])) < width {
		e.hscroll--
	}
}

func (e *internalEditor) moveVertical(delta int) {
	e.row = min(max(e.row+delta, 0), len(e.lines)-1)
	e.clampCol()
}

func (e *internalEditor) moveLeft() {
	if e.col > 0 {
		e.col--
	}
}

func (e *internalEditor) moveRight() {
	line := e.current()
	if e.mode == editorInsert || e.mode == editorReplace {
		e.col = min(e.col+1, len(line))
		return
	}
	if len(line) > 0 {
		e.col = min(e.col+1, len(line)-1)
	}
}

func (e *internalEditor) lineStart() {
	e.col = 0
}

func (e *internalEditor) lineEnd() {
	line := e.current()
	if e.mode == editorInsert || e.mode == editorReplace {
		e.col = len(line)
	} else {
		e.col = max(len(line)-1, 0)
	}
}

func (e *internalEditor) openLine(after bool) {
	e.snapshot()
	at := e.row
	if after {
		at++
	}
	e.lines = append(e.lines, nil)
	copy(e.lines[at+1:], e.lines[at:len(e.lines)-1])
	e.lines[at] = []rune{}
	e.row = at
	e.col = 0
	e.mode = editorInsert
	e.modified = true
	e.status = "-- INSERT --"
}

func (e *internalEditor) deleteChar(before bool) {
	line := e.current()
	if before {
		if e.col == 0 {
			return
		}
		e.snapshot()
		line = append(line[:e.col-1], line[e.col:]...)
		e.col--
	} else {
		if e.col >= len(line) {
			return
		}
		e.snapshot()
		line = append(line[:e.col], line[e.col+1:]...)
	}
	e.lines[e.row] = line
	e.modified = true
	e.clampCol()
}

func (e *internalEditor) deleteLine() {
	if len(e.lines) == 1 && len(e.lines[0]) == 0 {
		return
	}
	e.snapshot()
	e.yank = [][]rune{append([]rune(nil), e.current()...)}
	e.lines = append(e.lines[:e.row], e.lines[e.row+1:]...)
	if len(e.lines) == 0 {
		e.lines = [][]rune{{}}
	}
	e.row = min(e.row, len(e.lines)-1)
	e.clampCol()
	e.modified = true
	e.status = "1 line deleted"
}

func (e *internalEditor) yankLine() {
	e.yank = [][]rune{append([]rune(nil), e.current()...)}
	e.status = "1 line yanked"
}

func (e *internalEditor) put(after bool) {
	if len(e.yank) == 0 {
		return
	}
	e.snapshot()

	at := e.row
	if after {
		at++
	}
	copied := cloneEditorLines(e.yank)
	oldLen := len(e.lines)
	e.lines = append(e.lines, nil)
	copy(e.lines[at+len(copied):], e.lines[at:oldLen])
	copy(e.lines[at:], copied)
	e.row = at
	e.col = 0
	e.modified = true
	e.status = fmt.Sprintf("%d line(s) pasted", len(copied))
}

func (e *internalEditor) joinLine() {
	if e.row+1 >= len(e.lines) {
		e.status = "already at last line"
		return
	}
	e.snapshot()
	line := e.current()
	if len(line) > 0 {
		line = append(line, ' ')
	}
	line = append(line, e.lines[e.row+1]...)
	e.lines[e.row] = line
	e.lines = append(e.lines[:e.row+1], e.lines[e.row+2:]...)
	e.modified = true
	e.clampCol()
	e.status = "joined"
}

func (e *internalEditor) insertRune(r rune) {
	line := e.current()
	if e.mode == editorReplace {
		if e.col < len(line) {
			line[e.col] = r
		} else {
			line = append(line, r)
		}
		e.col++
	} else {
		line = append(line, 0)
		copy(line[e.col+1:], line[e.col:])
		line[e.col] = r
		e.col++
	}
	e.lines[e.row] = line
	e.modified = true
}

func (e *internalEditor) insertNewline() {
	e.snapshot()
	line := e.current()
	left := append([]rune(nil), line[:e.col]...)
	right := append([]rune(nil), line[e.col:]...)
	e.lines[e.row] = left
	e.lines = append(e.lines, nil)
	copy(e.lines[e.row+2:], e.lines[e.row+1:])
	e.row++
	e.lines[e.row] = right
	e.col = 0
	e.modified = true
}

func (e *internalEditor) backspace() {
	if e.col > 0 {
		e.snapshot()
		line := e.current()
		line = append(line[:e.col-1], line[e.col:]...)
		e.lines[e.row] = line
		e.col--
		e.modified = true
		return
	}
	if e.row == 0 {
		return
	}
	e.snapshot()
	prevLen := len(e.lines[e.row-1])
	e.lines[e.row-1] = append(e.lines[e.row-1], e.lines[e.row]...)
	e.lines = append(e.lines[:e.row], e.lines[e.row+1:]...)
	e.row--
	e.col = prevLen
	e.modified = true
}

func (e *internalEditor) replaceRune(r rune) {
	line := e.current()
	if len(line) == 0 {
		return
	}
	e.snapshot()
	line[e.col] = r
	e.lines[e.row] = line
	e.modified = true
	e.col = min(e.col+1, len(line)-1)
}

func (e *internalEditor) enterInsert(replace bool) {
	e.snapshot()
	if replace {
		e.mode = editorReplace
		e.status = "-- REPLACE --"
	} else {
		e.mode = editorInsert
		e.status = "-- INSERT --"
	}
}

func (e *internalEditor) leaveInsert() {
	if e.mode == editorInsert || e.mode == editorReplace {
		if e.mode == editorInsert && e.col > 0 {
			e.col--
		}
		e.mode = editorNormal
		e.clampCol()
		e.pending = 0
		e.status = "NORMAL"
	}
}

func isEditorWord(r rune) bool {
	return r == '_' || r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z'
}

func (e *internalEditor) wordForward() {
	line := e.current()
	for e.col < len(line) && !isEditorWord(line[e.col]) {
		e.col++
	}
	for e.col < len(line) && isEditorWord(line[e.col]) {
		e.col++
	}
	if e.col >= len(line) && e.row+1 < len(e.lines) {
		e.row++
		e.col = 0
	}
	e.clampCol()
}

func (e *internalEditor) wordBackward() {
	line := e.current()
	for e.col > 0 && !isEditorWord(line[e.col-1]) {
		e.col--
	}
	for e.col > 0 && isEditorWord(line[e.col-1]) {
		e.col--
	}
}

func (e *internalEditor) wordEnd() {
	line := e.current()
	if e.col >= len(line) && e.row+1 < len(e.lines) {
		e.row++
		e.col = 0
		line = e.current()
	}
	for e.col < len(line) && !isEditorWord(line[e.col]) {
		e.col++
	}
	if e.col == len(line) && e.col > 0 {
		e.col--
	}
}

func (e *internalEditor) searchFor(query string, forward bool) bool {
	if query == "" {
		return false
	}
	start := e.row
	for n := 1; n <= len(e.lines); n++ {
		offset := n
		if !forward {
			offset = -n
		}
		row := (start + offset + len(e.lines)) % len(e.lines)
		line := string(e.lines[row])
		var col int
		if forward {
			col = strings.Index(line, query)
		} else {
			col = strings.LastIndex(line, query)
		}
		if col >= 0 {
			e.row = row
			e.col = col
			e.search = query
			e.searchForward = forward
			e.clampCol()
			e.status = "/" + query
			return true
		}
	}
	e.status = "pattern not found: " + query
	return false
}

func (e *internalEditor) executeCommand(cmd string) bool {
	switch strings.TrimSpace(cmd) {
	case "w":
		if err := e.save(); err != nil {
			e.status = "write: " + err.Error()
		}
	case "q":
		if e.modified {
			e.status = "no write since last change (use :q! or ZZ)"
		} else {
			return true
		}
	case "q!":
		e.modified = false
		return true
	case "wq", "x":
		if err := e.save(); err != nil {
			e.status = "write: " + err.Error()
		} else {
			return true
		}
	default:
		e.status = "unknown command: :" + strings.TrimSpace(cmd)
	}
	return false
}

func (e *internalEditor) handleKey(ev *tcell.EventKey, width, height int) bool {
	if ev.Key() == tcell.KeyF10 {
		e.modified = false
		return true
	}

	switch e.mode {
	case editorInsert, editorReplace:
		switch ev.Key() {
		case tcell.KeyEscape:
			e.leaveInsert()
		case tcell.KeyEnter:
			e.insertNewline()
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			e.backspace()
		case tcell.KeyDelete:
			e.deleteChar(false)
		case tcell.KeyLeft:
			e.moveLeft()
		case tcell.KeyRight:
			e.moveRight()
		case tcell.KeyUp:
			e.moveVertical(-1)
		case tcell.KeyDown:
			e.moveVertical(1)
		case tcell.KeyHome:
			e.lineStart()
		case tcell.KeyEnd:
			e.lineEnd()
		case tcell.KeyTab:
			e.insertRune('	')
		case tcell.KeyRune:
			if ev.Modifiers() == tcell.ModNone || ev.Modifiers() == tcell.ModShift {
				for _, r := range ev.Str() {
					if r >= ' ' || r == '	' {
						e.insertRune(r)
					}
				}
			}
		}
	case editorCommand, editorSearch:
		switch ev.Key() {
		case tcell.KeyEscape:
			e.mode = editorNormal
			e.input = nil
			e.status = "NORMAL"
		case tcell.KeyEnter:
			input := string(e.input)
			mode := e.mode
			e.input = nil
			e.mode = editorNormal
			if mode == editorCommand {
				if e.executeCommand(input) {
					return true
				}
			} else {
				e.searchFor(input, true)
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if len(e.input) > 0 {
				e.input = e.input[:len(e.input)-1]
			}
		case tcell.KeyRune:
			if ev.Modifiers() == tcell.ModNone || ev.Modifiers() == tcell.ModShift {
				e.input = append(e.input, []rune(ev.Str())...)
			}
		}
		if e.mode == editorCommand {
			e.status = ":" + string(e.input)
		} else if e.mode == editorSearch {
			e.status = "/" + string(e.input)
		}
	default:
		for _, r := range []rune(ev.Str()) {
			if ev.Key() != tcell.KeyRune {
				break
			}
			if ev.Modifiers() != tcell.ModNone && ev.Modifiers() != tcell.ModShift {
				return false
			}

			if e.pending != 0 {
				pending := e.pending
				e.pending = 0
				switch pending {
				case 'd':
					if r == 'd' {
						e.deleteLine()
					} else {
						e.status = "operator pending: dd"
					}
				case 'y':
					if r == 'y' {
						e.yankLine()
					} else {
						e.status = "operator pending: yy"
					}
				case 'g':
					if r == 'g' {
						e.row = 0
						e.clampCol()
						e.status = "top"
					}
				case 'r':
					if len(e.current()) > 0 {
						e.replaceRune(r)
					}
				case 'Z':
					switch r {
					case 'Z':
						if err := e.save(); err != nil {
							e.status = "write: " + err.Error()
							return false
						}
						return true
					case 'Q':
						e.modified = false
						return true
					default:
						e.status = "Z? only ZZ or ZQ"
					}
				}
				continue
			}

			switch r {
			case 'h':
				e.moveLeft()
			case 'j':
				e.moveVertical(1)
			case 'k':
				e.moveVertical(-1)
			case 'l':
				e.moveRight()
			case '0', '^':
				e.lineStart()
			case '$':
				e.lineEnd()
			case 'G':
				e.row = len(e.lines) - 1
				e.clampCol()
			case 'g':
				e.pending = 'g'
			case 'd':
				e.pending = 'd'
			case 'y':
				e.pending = 'y'
			case 'p':
				e.put(true)
			case 'P':
				e.put(false)
			case 'x':
				e.deleteChar(false)
			case 'X':
				e.deleteChar(true)
			case 'D':
				e.snapshot()
				line := e.current()
				e.lines[e.row] = append([]rune(nil), line[:e.col]...)
				e.modified = true
				e.clampCol()
			case 'J':
				e.joinLine()
			case 'u':
				e.undoLast()
			case 'i':
				e.enterInsert(false)
			case 'a':
				if e.col < len(e.current()) {
					e.col++
				}
				e.enterInsert(false)
			case 'A':
				e.lineEnd()
				e.enterInsert(false)
			case 'I':
				e.lineStart()
				e.enterInsert(false)
			case 'o':
				e.openLine(true)
			case 'O':
				e.openLine(false)
			case 'R':
				e.enterInsert(true)
			case 'r':
				e.pending = 'r'
			case ':':
				e.mode = editorCommand
				e.input = nil
				e.status = ":"
			case '/':
				e.mode = editorSearch
				e.input = nil
				e.status = "/"
			case 'n':
				e.searchFor(e.search, e.searchForward)
			case 'N':
				e.searchFor(e.search, !e.searchForward)
			case 'w':
				e.wordForward()
			case 'b':
				e.wordBackward()
			case 'e':
				e.wordEnd()
			case 'f':
				e.moveVertical(height)
			case 'F':
				e.moveVertical(-height)
			case ' ':
				e.moveRight()
			case 'Z':
				e.pending = 'Z'
			}
		}
	}

	e.ensureVisible(height, width)
	return false
}

func (ui *ui) openEditor(path string) error {
	e, err := openInternalEditor(path)
	if err != nil {
		return err
	}
	ui.editor = e
	ui.pager = nil
	ui.fileInfoPath = ""
	ui.fileInfoReadyState = false
	ui.sxScreen.forceClear = true
	return nil
}

func (ui *ui) closeEditor() {
	if ui.editor == nil {
		return
	}
	ui.editor = nil
	ui.fileInfoPath = ""
	ui.fileInfoReadyState = false
	ui.sxScreen.forceClear = true
}

func (ui *ui) handleEditorEvent(ev tcell.Event) bool {
	if ui.editor == nil {
		return false
	}
	key, ok := ev.(*tcell.EventKey)
	if !ok {
		return true
	}
	if ui.editor.handleKey(key, ui.screen.Size()) {
		ui.closeEditor()
	}
	return true
}

func (ui *ui) drawEditor() {
	if ui.editor != nil {
		ui.editor.draw(ui)
	}
}

func (e *internalEditor) draw(ui *ui) {
	w, h := ui.screen.Size()
	viewport := max(h-1, 1)
	e.ensureVisible(viewport, w)

	ui.screen.Clear()
	for screenRow := 0; screenRow < viewport; screenRow++ {
		row := e.top + screenRow
		if row >= len(e.lines) {
			ui.screen.PutStrStyled(0, screenRow, "~", tcell.StyleDefault)
			continue
		}
		line := e.lines[row]
		if e.hscroll >= len(line) {
			continue
		}
		text := truncateRight(string(line[e.hscroll:]), w)
		ui.screen.PutStrStyled(0, screenRow, text, tcell.StyleDefault)
	}

	mode := "NORMAL"
	switch e.mode {
	case editorInsert:
		mode = "INSERT"
	case editorReplace:
		mode = "REPLACE"
	case editorCommand:
		mode = "COMMAND"
	case editorSearch:
		mode = "SEARCH"
	}

	status := fmt.Sprintf(" %s %s", mode, e.path)
	if e.modified {
		status += " [+]"
	}
	if e.mode == editorCommand || e.mode == editorSearch {
		status = " " + e.status
	} else if e.status != "" && e.status != "NORMAL" {
		status += "  " + e.status
	}
	status += fmt.Sprintf("  %d,%d", e.row+1, e.col+1)
	ui.screen.PutStrStyled(0, h-1, truncateRight(status, w), tcell.StyleDefault.Reverse(true))

	line := e.current()
	cursorCol := 0
	col := min(e.col, len(line))
	if e.hscroll <= col {
		cursorCol = displaywidth.String(string(line[e.hscroll:col]))
	}
	if cursorCol >= w {
		cursorCol = max(w-1, 0)
	}
	cursorRow := min(max(e.row-e.top, 0), max(viewport-1, 0))
	ui.screen.ShowCursor(cursorCol, cursorRow)
	ui.screen.Show()
}
