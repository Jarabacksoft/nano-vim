package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

func keyName(ev *tcell.EventKey) string {
	switch ev.Key() {
	case tcell.KeyCtrlC:
		return "esc"
	case tcell.KeyEscape:
		return "esc"
	case tcell.KeyEnter, tcell.KeyCtrlJ:
		return "enter"
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return "backspace"
	case tcell.KeyDelete:
		return "delete"
	case tcell.KeyTab:
		return "tab"
	case tcell.KeyLeft:
		return "left"
	case tcell.KeyRight:
		return "right"
	case tcell.KeyUp:
		return "up"
	case tcell.KeyDown:
		return "down"
	case tcell.KeyCtrlR:
		return "redo"
	case tcell.KeyCtrlD:
		return "halfdown"
	case tcell.KeyCtrlU:
		return "halfup"
	case tcell.KeyRune:
		return string(ev.Rune())
	}
	return ""
}

type view struct{ top, left int }

func displayRune(r rune, col int) (rune, int) {
	if r == '\t' {
		return ' ', 8 - col%8
	}
	if unicode.IsControl(r) {
		return '�', 1
	}
	w := runewidth.RuneWidth(r)
	return r, w
}
func safeLabel(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return '�'
		}
		return r
	}, s)
}
func put(s tcell.Screen, x, y, limit int, text string, style tcell.Style) {
	for _, r := range text {
		w := runewidth.RuneWidth(r)
		if w < 1 {
			continue
		}
		if x+w > limit {
			return
		}
		s.SetContent(x, y, r, nil, style)
		x += w
	}
}
func (v *view) draw(s tcell.Screen, e *editor) {
	w, h := s.Size()
	s.Clear()
	if w < 1 || h < 1 {
		return
	}
	height := h - 2
	if height < 1 {
		height = 1
	}
	line, col := 0, 0
	for i, r := range e.text {
		if i == e.pos {
			break
		}
		if r == '\n' {
			line++
			col = 0
		} else {
			_, width := displayRune(r, col)
			col += width
		}
	}
	if line < v.top {
		v.top = line
	}
	if line >= v.top+height {
		v.top = line - height + 1
	}
	totalLines := strings.Count(string(e.text), "\n") + 1
	digits := len(strconv.Itoa(totalLines))
	gutter := 0
	if e.number && w >= digits+2 {
		gutter = digits + 1
	}
	textWidth := w - gutter
	if col < v.left {
		v.left = col
	}
	if col >= v.left+textWidth {
		v.left = col - textWidth + 1
	}
	a, b := e.anchor, e.pos
	if a > b {
		a, b = b, a
	}
	if e.lineVisual {
		a = e.start(a)
		b = e.end(b)
	}
	row, x := 0, 0
	base := tcell.StyleDefault
	selected := base.Reverse(true)
	if gutter > 0 {
		for row := v.top; row < totalLines && row < v.top+height; row++ {
			label := row + 1
			style := base.Dim(true)
			if e.relative && row != line {
				label = row - line
				if label < 0 {
					label = -label
				}
			}
			if row == line {
				style = base.Bold(true)
			}
			put(s, 0, row-v.top, gutter, fmt.Sprintf("%*d ", digits, label), style)
		}
	}
	for i, r := range e.text {
		if r == '\n' {
			if row >= v.top && row < v.top+height && e.mode == "VISUAL" && i >= a && i <= b {
				if x-v.left >= 0 && x-v.left < textWidth {
					s.SetContent(gutter+x-v.left, row-v.top, ' ', nil, selected)
				}
			}
			row++
			x = 0
			continue
		}
		ch, width := displayRune(r, x)
		style := base
		if e.mode == "VISUAL" && i >= a && i <= b {
			style = selected
		}
		y := row - v.top
		if y >= 0 && y < height && y < h {
			if width == 0 {
				if x-v.left > 0 {
					main, comb, st, _ := s.GetContent(gutter+x-v.left-1, y)
					s.SetContent(gutter+x-v.left-1, y, main, append(comb, ch), st)
				}
			} else if r == '\t' {
				for j := 0; j < width; j++ {
					cx := x + j - v.left
					if cx >= 0 && cx < textWidth {
						s.SetContent(gutter+cx, y, ' ', nil, style)
					}
				}
			} else if x-v.left >= 0 && x-v.left+width <= textWidth {
				s.SetContent(gutter+x-v.left, y, ch, nil, style)
			}
		}
		x += width
	}
	status := map[string]string{"NORMAL": "NORMÁLNÍ/NORMAL", "INSERT": "VKLÁDÁNÍ/INSERT", "VISUAL": "VÝBĚR/VISUAL"}[e.mode]
	if e.changed() {
		status += " [+]"
	}
	position := fmt.Sprintf(" %d:%d %d%%", line+1, e.pos-e.start(e.pos)+1, (e.pos+1)*100/(len(e.text)+1))
	name := safeLabel(filepath.Base(e.filename))
	if runewidth.StringWidth(status+position)+1 < w {
		space := w - runewidth.StringWidth(status+position) - 2
		if space > 0 {
			status += " " + runewidth.Truncate(name, space, "…")
		}
		status += position
	}
	if h >= 2 {
		put(s, 0, h-2, w, runewidth.Truncate(status, w, ""), selected)
	}
	msg := safeLabel(e.message)
	if e.prompt != "" {
		msg = safeLabel(e.prompt)
	}
	if h >= 3 {
		put(s, 0, h-1, w, msg, base)
	}
	if e.prompt != "" {
		s.ShowCursor(min(runewidth.StringWidth(msg), w-1), h-1)
	} else {
		s.ShowCursor(gutter+max(0, min(col-v.left, textWidth-1)), max(0, min(line-v.top, height-1)))
	}
	s.Show()
}
func run(e *editor) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return fmt.Errorf("Je nutný interaktivní terminál / an interactive terminal is required")
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(signals)
	original, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), original)
	s, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err = s.Init(); err != nil {
		return err
	}
	defer s.Fini()
	s.EnablePaste()
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-signals:
			s.PostEvent(tcell.NewEventInterrupt(nil))
		case <-done:
		}
	}()
	v := view{}
	pasting := false
	pasteStarted := false
	for !e.quit {
		v.draw(s, e)
		ev := s.PollEvent()
		if ev == nil {
			return fmt.Errorf("Vstup terminálu byl uzavřen / terminal input closed")
		}
		switch ev := ev.(type) {
		case *tcell.EventInterrupt:
			return nil
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventPaste:
			pasting = ev.Start()
			if !pasting && pasteStarted {
				e.key("esc")
				pasteStarted = false
			}
		case *tcell.EventKey:
			k := keyName(ev)
			if k == "" {
				continue
			}
			if pasting {
				if e.prompt != "" {
					if len([]rune(k)) == 1 {
						e.key(k)
					}
					continue
				}
				if e.mode != "INSERT" {
					e.key("i")
					pasteStarted = true
				}
				if k == "esc" {
					continue
				}
			}
			if k == "halfdown" || k == "halfup" {
				if e.prompt == "" {
					_, h := s.Size()
					motion := "j"
					if k == "halfup" {
						motion = "k"
					}
					e.move(motion, max(1, (h-2)/2))
				}
				continue
			}
			oldRegister := string(e.register)
			e.key(k)
			// tcell exposes terminal clipboard output; copying is best effort (OSC 52).
			if (k == "y" || oldRegister != string(e.register)) && len(e.register) > 0 {
				s.SetClipboard([]byte(string(e.register)))
			}
		}
	}
	return nil
}
func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Použití / Usage: nano-vim SOUBOR/FILE")
		os.Exit(2)
	}
	e, err := openEditor(os.Args[1])
	if err == nil {
		err = run(e)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "nano-vim: Chyba / Error:", err)
		os.Exit(1)
	}
}
