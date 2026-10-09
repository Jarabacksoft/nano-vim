package main

import (
	"strings"
	"unicode"
)

type snapshot struct {
	text string
	pos  int
}
type editor struct {
	diskExists                bool
	number, relative          bool
	opCount                   int
	text                      []rune
	pos                       int
	mode                      string
	anchor                    int
	lineVisual                bool
	register                  []rune
	lineRegister              bool
	undo, redo                []snapshot
	saved                     string
	filename, ending, message string
	pending, count, prompt    string
	search                    string
	backward                  bool
	quit                      bool
	recording                 []string
	last                      []string
	replay                    bool
	insertStart               snapshot
}

func newEditor(text string) *editor {
	return &editor{text: []rune(text), saved: text, mode: "NORMAL", ending: "\n", number: true}
}
func (e *editor) snap() snapshot { return snapshot{string(e.text), e.pos} }
func (e *editor) changed() bool  { return string(e.text) != e.saved }
func (e *editor) clamp() {
	if e.pos < 0 {
		e.pos = 0
	}
	if e.pos > len(e.text) {
		e.pos = len(e.text)
	}
	if e.mode != "INSERT" && e.pos == e.end(e.pos) && e.pos > e.start(e.pos) {
		e.pos--
	}
}
func (e *editor) start(p int) int {
	for p > 0 && e.text[p-1] != '\n' {
		p--
	}
	return p
}
func (e *editor) end(p int) int {
	for p < len(e.text) && e.text[p] != '\n' {
		p++
	}
	return p
}
func (e *editor) first(p int) int {
	p = e.start(p)
	for p < e.end(p) && (e.text[p] == ' ' || e.text[p] == '\t') {
		p++
	}
	return p
}
func class(r rune) int {
	if unicode.IsSpace(r) {
		return 0
	}
	if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
		return 1
	}
	return 2
}
func (e *editor) word(p int, motion string) int {
	n := len(e.text)
	if n == 0 {
		return 0
	}
	switch motion {
	case "w":
		if p < n {
			c := class(e.text[p])
			for p < n && class(e.text[p]) == c {
				p++
			}
		}
		for p < n && class(e.text[p]) == 0 {
			p++
		}
	case "b":
		if p > 0 {
			p--
		}
		for p > 0 && class(e.text[p]) == 0 {
			p--
		}
		c := class(e.text[p])
		for p > 0 && class(e.text[p-1]) == c {
			p--
		}
	case "e":
		if p < n-1 {
			p++
		}
		for p < n-1 && class(e.text[p]) == 0 {
			p++
		}
		if p < n {
			c := class(e.text[p])
			for p < n-1 && class(e.text[p+1]) == c {
				p++
			}
		}
	}
	return p
}
func (e *editor) move(k string, n int) bool {
	for i := 0; i < n; i++ {
		switch k {
		case "h", "left":
			if e.pos > e.start(e.pos) {
				e.pos--
			}
		case "l", "right":
			if e.pos < e.end(e.pos) {
				e.pos++
			}
		case "j", "down":
			col := e.pos - e.start(e.pos)
			p := e.end(e.pos)
			if p < len(e.text) {
				e.pos = p + 1 + col
				if e.pos > e.end(p+1) {
					e.pos = e.end(p + 1)
				}
			}
		case "k", "up":
			col := e.pos - e.start(e.pos)
			p := e.start(e.pos)
			if p > 0 {
				a := e.start(p - 1)
				e.pos = a + col
				if e.pos > p-1 {
					e.pos = p - 1
				}
			}
		case "w", "b", "e":
			e.pos = e.word(e.pos, k)
		case "0":
			e.pos = e.start(e.pos)
		case "^":
			e.pos = e.first(e.pos)
		case "$":
			e.pos = e.end(e.pos)
			if e.pos > e.start(e.pos) {
				e.pos--
			}
		case "gg":
			e.pos = 0
		case "G":
			e.pos = e.start(len(e.text))
		default:
			return false
		}
	}
	e.clamp()
	return true
}

// jumpLine uses one-based line numbers and clamps to the last buffer line.
func (e *editor) jumpLine(n int) {
	e.pos = 0
	for i, r := range e.text {
		if n <= 1 {
			break
		}
		if r == '\n' {
			e.pos = i + 1
			n--
		}
	}
	e.clamp()
}
func (e *editor) splice(a, b int, s []rune) {
	if a < 0 || b < a || b > len(e.text) {
		return
	}
	t := append([]rune{}, e.text[:a]...)
	t = append(t, s...)
	e.text = append(t, e.text[b:]...)
	e.pos = a
}
func (e *editor) commit(before snapshot) {
	if before.text != string(e.text) {
		e.undo = append(e.undo, before)
		e.redo = nil
		if !e.replay {
			e.last = append([]string{}, e.recording...)
		}
	}
}
func (e *editor) insert() { e.mode = "INSERT"; e.insertStart = e.snap() }
func (e *editor) object(kind rune, outer bool) (int, int, bool) {
	p := e.pos
	if p >= len(e.text) {
		return 0, 0, false
	}
	if kind == 'w' {
		a, b := p, p+1
		c := class(e.text[p])
		for a > 0 && class(e.text[a-1]) == c {
			a--
		}
		for b < len(e.text) && class(e.text[b]) == c {
			b++
		}
		if outer {
			old := b
			for b < len(e.text) && unicode.IsSpace(e.text[b]) {
				b++
			}
			if old == b {
				for a > 0 && unicode.IsSpace(e.text[a-1]) {
					a--
				}
			}
		}
		return a, b, true
	}
	if kind == '"' || kind == '\'' {
		a := -1
		for i := e.start(p); i < e.end(p); i++ {
			if e.text[i] == kind {
				slashes := 0
				for j := i - 1; j >= 0 && e.text[j] == '\\'; j-- {
					slashes++
				}
				if slashes%2 == 1 {
					continue
				}
				if a < 0 {
					a = i
				} else {
					if p >= a && p <= i {
						if outer {
							return a, i + 1, true
						}
						return a + 1, i, true
					}
					a = -1
				}
			}
		}
		return 0, 0, false
	}
	closes := map[rune]rune{'(': ')', '[': ']', '{': '}'}
	close, ok := closes[kind]
	if !ok {
		return 0, 0, false
	}
	stack := []int{}
	bestA, bestB := -1, 0
	for i, r := range e.text {
		if r == kind {
			stack = append(stack, i)
		}
		if r == close && len(stack) > 0 {
			a := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if a <= p && p <= i && a > bestA {
				bestA, bestB = a, i+1
			}
		}
	}
	if bestA < 0 {
		return 0, 0, false
	}
	if !outer {
		bestA++
		bestB--
	}
	return bestA, bestB, true
}
func (e *editor) rangeOp(op string, a, b int, line bool) {
	e.register = append([]rune{}, e.text[a:b]...)
	e.lineRegister = line
	if op == "y" {
		e.mode = "NORMAL"
		return
	}
	if op == "c" {
		e.insertStart = e.snap()
	}
	e.splice(a, b, nil)
	if op == "c" {
		e.mode = "INSERT"
	} else {
		e.mode = "NORMAL"
		e.clamp()
	}
}
func (e *editor) find(back bool) {
	if e.search == "" {
		e.message = "Žádné předchozí hledání / No previous search"
		return
	}
	needle := []rune(e.search)
	n := len(e.text)
	for step := 1; step <= n+1; step++ {
		p := (e.pos + step) % (n + 1)
		if back {
			p = (e.pos - step + n + 1) % (n + 1)
		}
		if p+len(needle) <= n && string(e.text[p:p+len(needle)]) == e.search {
			e.pos = p
			e.clamp()
			e.message = ""
			return
		}
	}
	e.message = "Vzor nenalezen / Pattern not found: " + e.search
}
func (e *editor) key(k string) {
	e.message = ""
	if e.prompt != "" {
		if k == "esc" {
			e.prompt = ""
			return
		}
		if k == "backspace" {
			r := []rune(e.prompt)
			if len(r) > 1 {
				e.prompt = string(r[:len(r)-1])
			}
			return
		}
		if k == "enter" {
			p := e.prompt
			e.prompt = ""
			if p[0] == ':' {
				e.command(p[1:])
			} else {
				if len(p) > 1 {
					e.search = p[1:]
				}
				e.backward = p[0] == '?'
				e.find(e.backward)
			}
			return
		}
		if len([]rune(k)) == 1 {
			e.prompt += k
		}
		return
	}
	if !e.replay {
		e.recording = append(e.recording, k)
	}
	if e.mode == "INSERT" {
		switch k {
		case "esc":
			e.mode = "NORMAL"
			if e.pos > e.start(e.pos) {
				e.pos--
			}
			e.commit(e.insertStart)
			e.recording = nil
		case "backspace":
			if e.pos > 0 {
				p := e.pos
				e.splice(p-1, p, nil)
			}
		case "delete":
			if e.pos < len(e.text) {
				e.splice(e.pos, e.pos+1, nil)
			}
		case "enter":
			p := e.pos
			e.splice(p, p, []rune{'\n'})
			e.pos++
		case "tab":
			p := e.pos
			e.splice(p, p, []rune{'\t'})
			e.pos++
		default:
			if k == "left" || k == "right" || k == "up" || k == "down" {
				e.move(k, 1)
			} else if len([]rune(k)) == 1 {
				p := e.pos
				e.splice(p, p, []rune(k))
				e.pos++
			}
		}
		return
	}
	if k == "esc" {
		e.pending = ""
		e.count = ""
		e.opCount = 0
		e.mode = "NORMAL"
		e.recording = nil
		return
	}
	if e.pending == "r" {
		before := e.snap()
		if len([]rune(k)) == 1 {
			n := e.opCount
			if n < 1 {
				n = 1
			}
			e.opCount = 0
			b := e.pos + n
			if b > e.end(e.pos) {
				b = e.end(e.pos)
			}
			e.splice(e.pos, b, []rune(strings.Repeat(k, b-e.pos)))
			e.commit(before)
		}
		e.pending = ""
		e.recording = nil
		return
	}
	if len(k) == 1 && k[0] >= '0' && k[0] <= '9' && (k != "0" || e.count != "") {
		if len(e.count) < 4 {
			e.count += k
		}
		return
	}
	hadCount := e.count != ""
	n := e.takeCount()
	before := e.snap()
	if k == "G" && hadCount && e.pending == "" {
		e.jumpLine(n)
		if e.mode == "NORMAL" {
			e.recording = nil
		}
		return
	}
	if e.mode == "VISUAL" {
		if e.pending == "g" {
			if e.opCount > 0 {
				n *= e.opCount
			}
			e.opCount = 0
			e.pending = ""
			if k == "g" {
				e.jumpLine(n)
			}
			return
		}
		if k == "g" {
			e.pending = "g"
			e.opCount = n
			return
		}
		if k == "d" || k == "y" || k == "c" {
			a, b := e.anchor, e.pos
			if a > b {
				a, b = b, a
			}
			b++
			if b > len(e.text) {
				b = len(e.text)
			}
			if e.lineVisual {
				a = e.start(a)
				b = e.end(e.posMax(e.anchor, e.pos))
				if b < len(e.text) {
					b++
				}
			}
			e.rangeOp(k, a, b, e.lineVisual)
			if k != "c" {
				e.commit(before)
				e.recording = nil
			}
			return
		}
		e.move(k, n)
		return
	}
	if e.pending != "" {
		pending := e.pending
		e.pending = ""
		if e.opCount > 0 {
			n *= e.opCount
		}
		e.opCount = 0
		if n > 9999 {
			n = 9999
		}
		if pending == "g" {
			if k == "g" {
				e.jumpLine(n)
			}
			e.recording = nil
			return
		}
		op := pending[:1]
		a, b := e.pos, e.pos
		line := false
		ok := true
		if len(pending) == 2 {
			if len([]rune(k)) == 1 {
				a, b, ok = e.object([]rune(k)[0], pending[1] == 'a')
			} else {
				ok = false
			}
		} else if k == "i" || k == "a" {
			e.pending = op + k
			e.opCount = n
			return
		} else if k == op {
			a = e.start(e.pos)
			b = a
			for i := 0; i < n; i++ {
				b = e.end(b)
				if b < len(e.text) {
					b++
				}
			}
			line = true
			if op == "c" {
				if b > a && e.text[b-1] == '\n' {
					b--
				}
				line = false
			}
		} else {
			switch k {
			case "w":
				for i := 0; i < n; i++ {
					if op == "c" && i == n-1 {
						if b < len(e.text) {
							c := class(e.text[b])
							for b < len(e.text) && class(e.text[b]) == c {
								b++
							}
						}
					} else {
						b = e.word(b, "w")
					}
				}
			case "e":
				for i := 0; i < n; i++ {
					b = e.word(b, "e")
				}
				b++
			case "$":
				b = e.end(a)
			case "0":
				a = e.start(a)
			default:
				ok = false
			}
		}
		if ok && a <= b && b <= len(e.text) {
			e.rangeOp(op, a, b, line)
			if op != "c" {
				e.commit(before)
			}
		} else {
			e.message = "Nenalezen textový objekt nebo pohyb / No matching text object or motion"
		}
		if e.mode != "INSERT" {
			e.recording = nil
		}
		return
	}
	if e.move(k, n) {
		e.recording = nil
		return
	}
	switch k {
	case "d", "y", "c", "g", "r":
		e.pending = k
		e.opCount = n
		return
	case "i":
		e.insert()
	case "a":
		if e.pos < e.end(e.pos) {
			e.pos++
		}
		e.insert()
	case "I":
		e.pos = e.first(e.pos)
		e.insert()
	case "A":
		e.pos = e.end(e.pos)
		e.insert()
	case "o", "O":
		e.insert()
		p := e.start(e.pos)
		if k == "o" {
			p = e.end(e.pos)
			e.splice(p, p, []rune{'\n'})
			e.pos = p + 1
		} else {
			e.splice(p, p, []rune{'\n'})
		}
	case "v", "V":
		e.mode = "VISUAL"
		e.anchor = e.pos
		e.lineVisual = k == "V"
	case "x":
		b := e.pos + n
		if b > e.end(e.pos) {
			b = e.end(e.pos)
		}
		e.rangeOp("d", e.pos, b, false)
	case "D", "C":
		op := "d"
		if k == "C" {
			op = "c"
		}
		e.rangeOp(op, e.pos, e.end(e.pos), false)
	case "p", "P":
		if len(e.register) > 0 {
			p := e.pos
			s := []rune(strings.Repeat(string(e.register), n))
			if e.lineRegister {
				p = e.start(p)
				if k == "p" {
					p = e.end(e.pos)
					if p < len(e.text) {
						p++
					} else {
						s = append([]rune{'\n'}, s...)
					}
				}
				if s[len(s)-1] != '\n' && p < len(e.text) {
					s = append(s, '\n')
				}
			} else if k == "p" && p < e.end(p) {
				p++
			}
			e.splice(p, p, s)
		}
	case "J":
		for i := 0; i < n; i++ {
			b := e.end(e.pos)
			if b == len(e.text) {
				break
			}
			c := b + 1
			for c < len(e.text) && (e.text[c] == ' ' || e.text[c] == '\t') {
				c++
			}
			e.splice(b, c, []rune{' '})
		}
	case "u", "redo":
		for i := 0; i < n; i++ {
			if k == "u" && len(e.undo) > 0 {
				e.redo = append(e.redo, e.snap())
				s := e.undo[len(e.undo)-1]
				e.undo = e.undo[:len(e.undo)-1]
				e.text = []rune(s.text)
				e.pos = s.pos
			} else if k == "redo" && len(e.redo) > 0 {
				e.undo = append(e.undo, e.snap())
				s := e.redo[len(e.redo)-1]
				e.redo = e.redo[:len(e.redo)-1]
				e.text = []rune(s.text)
				e.pos = s.pos
			}
		}
		e.clamp()
		e.recording = nil
		return
	case ".":
		last := append([]string{}, e.last...)
		e.replay = true
		for i := 0; i < n; i++ {
			for _, key := range last {
				e.key(key)
			}
		}
		e.replay = false
		e.recording = nil
		return
	case ":", "/", "?":
		e.prompt = k
	case "n", "N":
		back := e.backward
		if k == "N" {
			back = !back
		}
		for i := 0; i < n; i++ {
			e.find(back)
		}
	case "*", "#":
		a, b, ok := e.object('w', false)
		if ok {
			e.search = string(e.text[a:b])
			e.backward = k == "#"
			e.find(e.backward)
		}
	}
	if e.mode != "INSERT" && e.mode != "VISUAL" {
		e.commit(before)
		e.recording = nil
	}
	e.clamp()
}
func (e *editor) posMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
