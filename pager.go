package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gdamore/tcell/v3"
)

const pagerMaxLineBytes = 1 << 16
const pagerReadAhead = 128

type pager struct {
	path   string
	file   *os.File
	reader *bufio.Reader
	lines  []string
	top    int
	eof    bool
	binary bool
}

func openPager(path string) (*pager, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if !info.Mode().IsRegular() {
		f.Close()
		return nil, fmt.Errorf("not a regular file")
	}

	p := &pager{
		path:   path,
		file:   f,
		reader: bufio.NewReaderSize(f, 64*1024),
	}
	p.fill(pagerReadAhead)
	return p, nil
}

func (p *pager) close() {
	if p.file != nil {
		_ = p.file.Close()
		p.file = nil
	}
}

func pagerBinaryByte(b byte) bool {
	return b < 0x07 || (b > 0x0D && b < 0x1B) || (b > 0x1B && b < 0x20) || b == 0x7F
}

func (p *pager) fill(target int) {
	if p.eof || p.binary {
		return
	}

	for len(p.lines) < target {
		line, err := p.reader.ReadString('\\n')
		if len(line) > 0 {
			if len(line) > pagerMaxLineBytes {
				line = line[:pagerMaxLineBytes]
			}

			for i := 0; i < len(line); i++ {
				if pagerBinaryByte(line[i]) {
					p.binary = true
					p.lines = nil
					return
				}
			}

			line = strings.TrimSuffix(line, "\\n")
			line = strings.TrimSuffix(line, "\\r")
			p.lines = append(p.lines, sanitizePreview(line))
		}

		if err != nil {
			if err == io.EOF {
				p.eof = true
			}
			return
		}
	}
}

func (p *pager) maxTop(height int) int {
	if height <= 0 {
		return 0
	}

	p.fill(p.top + height + 1)
	limit := len(p.lines) - height
	return max(limit, 0)
}

func (p *pager) move(delta, height int) {
	target := max(p.top+delta, 0)
	p.fill(target + height + 1)
	if p.eof {
		target = min(target, p.maxTop(height))
	}
	p.top = target
}

func (p *pager) draw(ui *ui) {
	screen := ui.screen
	w, h := screen.Size()
	screen.Clear()

	viewport := max(h-1, 1)
	if p.binary {
		screen.PutStrStyled(0, 0, "\033[7mbinary file\033[0m", tcell.StyleDefault)
	} else {
		p.fill(p.top + viewport + 1)
		for i := 0; i < viewport; i++ {
			index := p.top + i
			if index >= len(p.lines) {
				break
			}
			line := truncateRight(p.lines[index], w)
			screen.PutStrStyled(0, i, line, tcell.StyleDefault)
		}
	}

	end := min(p.top+viewport, len(p.lines))
	status := fmt.Sprintf(" %s  %d-%d", sanitizeName(p.path), min(p.top+1, max(len(p.lines), 1)), end)
	if !p.eof && !p.binary {
		status += " ..."
	}
	status += "   j/k ↑/↓  PgUp/PgDn  Space/B  g/G  q"
	screen.PutStrStyled(0, h-1, truncateRight(status, w), tcell.StyleDefault.Reverse(true))
	screen.Show()
}

func (ui *ui) openPager(path string) error {
	p, err := openPager(path)
	if err != nil {
		return err
	}

	ui.pager = p
	ui.sxScreen.forceClear = true
	return nil
}

func (ui *ui) closePager() {
	if ui.pager == nil {
		return
	}
	ui.pager.close()
	ui.pager = nil
	ui.fileInfoPath = ""
	ui.fileInfoReadyState = false
}

func (ui *ui) handlePagerEvent(ev tcell.Event) bool {
	if ui.pager == nil {
		return false
	}

	key, ok := ev.(*tcell.EventKey)
	if !ok {
		return true
	}

	_, h := ui.screen.Size()
	viewport := max(h-1, 1)

	switch key.Key() {
	case tcell.KeyEscape:
		ui.closePager()
	case tcell.KeyUp:
		ui.pager.move(-1, viewport)
	case tcell.KeyDown:
		ui.pager.move(1, viewport)
	case tcell.KeyPgUp:
		ui.pager.move(-viewport, viewport)
	case tcell.KeyPgDn:
		ui.pager.move(viewport, viewport)
	case tcell.KeyHome:
		ui.pager.top = 0
	case tcell.KeyEnd:
		ui.pager.fill(len(ui.pager.lines) + pagerReadAhead)
		ui.pager.top = ui.pager.maxTop(viewport)
	case tcell.KeyRune:
		keyName := key.Str()
		if key.Modifiers()&tcell.ModAlt != 0 {
			switch strings.ToLower(keyName) {
			case "j":
				ui.pager.move(1, viewport)
			case "k":
				ui.pager.move(-1, viewport)
			case "f":
				ui.pager.move(viewport, viewport)
			case "b":
				ui.pager.move(-viewport, viewport)
			default:
				return true
			}
			break
		}
		switch keyName {
		case 'q':
			ui.closePager()
		case 'k':
			ui.pager.move(-1, viewport)
		case 'j':
			ui.pager.move(1, viewport)
		case 'b':
			ui.pager.move(-viewport, viewport)
		case ' ':
			ui.pager.move(viewport, viewport)
		case 'g':
			ui.pager.top = 0
		case 'G':
			ui.pager.fill(len(ui.pager.lines) + pagerReadAhead)
			ui.pager.top = ui.pager.maxTop(viewport)
		default:
			return true
		}
	default:
		return true
	}

	return true
}
