package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func keys(e *editor, s string) {
	for _, r := range s {
		if r == 27 {
			e.key("esc")
		} else if r == '\n' {
			e.key("enter")
		} else {
			e.key(string(r))
		}
	}
}
func TestEditing(t *testing.T) {
	cases := []struct{ name, input, commands, want string }{
		{"insert literal motions", "", "ihello 世界\x1b", "hello 世界"},
		{"append", "abc", "A!\x1b", "abc!"},
		{"open below", "a\nb", "ox\x1b", "a\nx\nb"},
		{"open above", "a", "Ox\x1b", "x\na"},
		{"count delete", "a\nb\nc\nd", "3dd", "d"},
		{"combined counts", "a b c d e f g", "2d3w", "g"},
		{"count zero motion", "abcde", "lll2d0", "de"},
		{"count word", "one two three four", "2dw", "three four"},
		{"delete end", "abcdef", "llD", "ab"},
		{"delete beginning", "abcdef", "llld0", "def"},
		{"change word", "a bee", "cwX\x1b", "X bee"},
		{"change line empty", "", "ccX\x1b", "X"},
		{"change line", "abc\ndef", "ccX\x1b", "X\ndef"},
		{"replace", "abc", "rX", "Xbc"},
		{"repeat replace", "abc", "rXl.", "XXc"},
		{"repeat insert", "ab", "iX\x1bl.", "XXab"},
		{"join", "a\n  b", "J", "a b"},
		{"yank paste", "a\nb", "yyp", "a\na\nb"},
		{"paste above", "a\nb", "yyP", "a\na\nb"},
		{"visual delete", "abcde", "vllld", "e"},
		{"visual yank", "abcde", "vly$p", "abcdeab"},
		{"visual first line", "a\nb\nc", "GVggd", ""},
		{"line visual", "a\nb\nc", "Vjd", "c"},
		{"inner word", "one two", "ciwX\x1b", "X two"},
		{"around word", "one two", "daw", "two"},
		{"quotes", "a \"hello\" z", "llci\"X\x1b", "a \"X\" z"},
		{"single quotes", "'hello'", "ci'X\x1b", "'X'"},
		{"nested parentheses", "a(b(c)d)e", "lllda(", "a(bd)e"},
		{"brackets", "[abc]", "ci[X\x1b", "[X]"},
		{"braces", "{abc}", "da{", ""},
		{"unresolved", "abc", "ci\"", "abc"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := newEditor(c.input)
			keys(e, c.commands)
			if string(e.text) != c.want {
				t.Fatalf("got %q, want %q", string(e.text), c.want)
			}
			if e.pos < 0 || e.pos > len(e.text) {
				t.Fatal("invalid cursor")
			}
		})
	}
}
func TestHistory(t *testing.T) {
	e := newEditor("abc")
	keys(e, "iXYZ\x1b")
	e.key("u")
	if string(e.text) != "abc" || e.changed() {
		t.Fatal("insert should undo as one change")
	}
	e.key("redo")
	if string(e.text) != "XYZabc" {
		t.Fatal("redo")
	}
	e.key("u")
	keys(e, "x")
	e.key("redo")
	if string(e.text) != "bc" {
		t.Fatal("new edit must clear redo")
	}
}
func TestSearch(t *testing.T) {
	e := newEditor("word x word y word")
	keys(e, "/word\n")
	if e.pos != 7 {
		t.Fatal(e.pos)
	}
	e.key("n")
	if e.pos != 14 {
		t.Fatal(e.pos)
	}
	e.key("N")
	if e.pos != 7 {
		t.Fatal(e.pos)
	}
	keys(e, "?word\n")
	if e.pos != 0 {
		t.Fatal(e.pos)
	}
	keys(e, "/missing\n")
	if !strings.Contains(e.message, "not found") {
		t.Fatal(e.message)
	}
	keys(e, "/word\x1b")
	if e.pos != 0 {
		t.Fatal("cancel moved cursor")
	}
}
func TestMotions(t *testing.T) {
	e := newEditor("  one two\nthree\n四五")
	keys(e, "^3w")
	if e.pos != len(e.text)-2 {
		t.Fatal(e.pos)
	}
	keys(e, "gg")
	if e.pos != 0 {
		t.Fatal(e.pos)
	}
	keys(e, "G$")
	if e.pos != len(e.text)-1 {
		t.Fatal(e.pos)
	}
	for i := 0; i < 10; i++ {
		e.key("right")
		e.key("down")
	}
	if e.pos >= len(e.text) {
		t.Fatal(e.pos)
	}
}
func TestFiles(t *testing.T) {
	dir := t.TempDir()
	for _, text := range []string{"", "hello", "hello\n世界\n", "hello\r\n世界\r\n"} {
		name := filepath.Join(dir, "file with spaces")
		os.WriteFile(name, []byte(text), 0600)
		e, err := openEditor(name)
		if err != nil {
			t.Fatal(err)
		}
		if err = e.save(); err != nil {
			t.Fatal(err)
		}
		b, _ := os.ReadFile(name)
		if string(b) != text {
			t.Fatalf("round trip %q => %q", text, b)
		}
		st, _ := os.Stat(name)
		if st.Mode().Perm() != 0600 {
			t.Fatal("lost mode")
		}
	}
	name := filepath.Join(dir, "new")
	e, err := openEditor(name)
	if err != nil {
		t.Fatal(err)
	}
	keys(e, "iabc\x1b")
	e.command("q")
	if e.quit {
		t.Fatal("unsafe quit")
	}
	e.command("e")
	if string(e.text) != "abc" {
		t.Fatal("unsafe reload")
	}
	e.command("wq")
	if !e.quit {
		t.Fatal(e.message)
	}
	b, _ := os.ReadFile(name)
	if string(b) != "abc" {
		t.Fatal(string(b))
	}
	e, _ = openEditor(name)
	keys(e, "ix\x1b")
	os.WriteFile(name, []byte("external"), 0600)
	if e.save() == nil {
		t.Fatal("overwrote external changes")
	}
	if !e.changed() {
		t.Fatal("lost buffer")
	}
	e.command("q!")
	if !e.quit {
		t.Fatal("forced quit")
	}
}
func TestRejectUnsupported(t *testing.T) {
	for _, data := range [][]byte{{0xff}, {0}, {27}, []byte("a\r\nb\n"), []byte("a\rb")} {
		p := filepath.Join(t.TempDir(), "bad")
		os.WriteFile(p, data, 0600)
		if _, err := openEditor(p); err == nil {
			t.Fatalf("accepted %q", data)
		}
	}
	if _, err := openEditor(t.TempDir()); err == nil {
		t.Fatal("accepted directory")
	}
}
func TestRenderer(t *testing.T) {
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	e := newEditor("hello\n世界\n" + strings.Repeat("x", 200))
	e.filename = "file"
	v := view{}
	for _, size := range [][2]int{{80, 24}, {15, 5}, {1, 1}, {40, 10}} {
		s.SetSize(size[0], size[1])
		keys(e, "G$")
		v.draw(s, e)
		x, y, visible := s.GetCursor()
		if !visible || x < 0 || x >= size[0] || y < 0 || y >= size[1] {
			t.Fatalf("invalid cursor %d,%d", x, y)
		}
	}
}
func FuzzKeys(f *testing.F) {
	f.Add("abc\ndef", "cc\x1b3dd.u")
	f.Add("", "ccx\x1b")
	f.Fuzz(func(t *testing.T, text, commands string) {
		if len(text) > 1000 || len(commands) > 1000 {
			return
		}
		e := newEditor(text)
		keys(e, commands)
		if e.pos < 0 || e.pos > len(e.text) {
			t.Fatal("invalid cursor")
		}
	})
}

func TestFileFailures(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "readonly")
	if err := os.WriteFile(p, []byte("original"), 0400); err != nil {
		t.Fatal(err)
	}
	e, err := openEditor(p)
	if err != nil {
		t.Fatal(err)
	}
	keys(e, "iX\x1b")
	if e.save() == nil {
		t.Fatal("saved readonly file")
	}
	b, _ := os.ReadFile(p)
	if string(b) != "original" || !e.changed() {
		t.Fatal("failed save lost content")
	}
	p = filepath.Join(dir, "empty")
	os.WriteFile(p, nil, 0600)
	e, _ = openEditor(p)
	os.Remove(p)
	if e.save() == nil {
		t.Fatal("did not detect deletion of empty file")
	}
	p = filepath.Join(dir, "missing")
	e, _ = openEditor(p)
	os.WriteFile(p, nil, 0600)
	if e.save() == nil {
		t.Fatal("did not detect external creation")
	}
	p = filepath.Join(dir, "original")
	os.WriteFile(p, []byte("abc"), 0600)
	link := filepath.Join(dir, "link")
	if err = os.Symlink(p, link); err != nil {
		t.Fatal(err)
	}
	e, err = openEditor(link)
	if err != nil {
		t.Fatal(err)
	}
	keys(e, "iX\x1b")
	if err = e.save(); err != nil {
		t.Fatal(err)
	}
	b, _ = os.ReadFile(p)
	if string(b) != "Xabc" {
		t.Fatal("symlink target not updated")
	}
	hard := filepath.Join(dir, "hard")
	if err = os.Link(p, hard); err != nil {
		t.Fatal(err)
	}
	if e.save() == nil {
		t.Fatal("accepted hard link replacement")
	}
}
func TestStatus(t *testing.T) {
	s := tcell.NewSimulationScreen("UTF-8")
	s.Init()
	defer s.Fini()
	s.SetSize(80, 10)
	e := newEditor("abc")
	e.filename = "example.txt"
	keys(e, "iX")
	v := view{}
	v.draw(s, e)
	var row strings.Builder
	for x := 0; x < 80; x++ {
		r, _, _, _ := s.GetContent(x, 8)
		row.WriteRune(r)
	}
	for _, want := range []string{"INSERT", "example.txt", "1:2", "[+]"} {
		if !strings.Contains(row.String(), want) {
			t.Fatalf("status missing %q: %q", want, row.String())
		}
	}
}
func TestOneMatchSearch(t *testing.T) {
	e := newEditor("word")
	e.search = "word"
	e.find(false)
	if e.message != "" || e.pos != 0 {
		t.Fatal("single match should wrap")
	}
}

func TestCommandBoundary(t *testing.T) {
	e := newEditor("only this buffer")
	e.filename = "current.txt"
	for _, cmd := range []string{"e other.txt", "w other.txt", "set number", "map x y", "!echo unsafe", "split", "tabnew"} {
		e.command(cmd)
		if e.filename != "current.txt" || string(e.text) != "only this buffer" || e.quit || e.message != "Neznámý příkaz / Unknown command" {
			t.Fatalf("unexpected command behavior: %q", cmd)
		}
	}
}

func TestLineJumps(t *testing.T) {
	for _, c := range []struct {
		keys string
		pos  int
	}{
		{"3G", 7}, {"3gg", 7}, {"G1G", 0}, {"G2gg", 3}, {":3\n", 7},
		{"999G", 12}, {":999\n", 12}, {"Ggg", 0},
	} {
		t.Run(c.keys, func(t *testing.T) {
			e := newEditor("αβ\nbee\n四五六七\n")
			keys(e, c.keys)
			if e.pos != c.pos {
				t.Fatalf("got cursor %d, want %d", e.pos, c.pos)
			}
			if e.changed() || len(e.undo) != 0 {
				t.Fatal("jump changed buffer/history")
			}
		})
	}
	e := newEditor("a\nb\nc")
	keys(e, "v3Gd")
	if string(e.text) != "" {
		t.Fatal(string(e.text))
	}
	e = newEditor("a\nb\nc")
	keys(e, "V3ggd")
	if string(e.text) != "" {
		t.Fatal(string(e.text))
	}
	e = newEditor("abc")
	e.pos = 2
	for _, cmd := range []string{"0", "999999999999999999999999999999999999"} {
		e.command(cmd)
		if e.pos != 2 || e.message != "Neplatné číslo řádku / Invalid line number" {
			t.Fatal(cmd, e.pos, e.message)
		}
	}
	e = newEditor("")
	keys(e, "999G999gg:42\n")
	if e.pos != 0 {
		t.Fatal(e.pos)
	}
}
func screenRow(s tcell.Screen, y, width int) string {
	var row strings.Builder
	for x := 0; x < width; x++ {
		r, _, _, _ := s.GetContent(x, y)
		if r == 0 {
			r = ' '
		}
		row.WriteRune(r)
	}
	return row.String()
}
func TestLineNumberGutter(t *testing.T) {
	s := tcell.NewSimulationScreen("UTF-8")
	s.Init()
	defer s.Fini()
	s.SetSize(20, 8)
	e := newEditor("one\n\nthree\nfour\n")
	v := view{}
	v.draw(s, e)
	if got := screenRow(s, 0, 5); got != "1 one" {
		t.Fatal(got)
	}
	if got := screenRow(s, 1, 2); got != "2 " {
		t.Fatal("empty line", got)
	}
	if got := screenRow(s, 4, 2); got != "5 " {
		t.Fatal("trailing line", got)
	}
	e.command("3")
	e.command("relative")
	v.draw(s, e)
	for y, want := range []string{"2 ", "1 ", "3 ", "1 ", "2 "} {
		if got := screenRow(s, y, 2); got != want {
			t.Fatalf("relative row %d: %q", y, got)
		}
	}
	e.command("number")
	v.draw(s, e)
	if got := screenRow(s, 0, 3); got != "one" {
		t.Fatal("hidden gutter", got)
	}
	e.command("number")
	if !e.relative {
		t.Fatal("hiding gutter lost relative mode")
	}
	e.command("relative")
	v.draw(s, e)
	if got := screenRow(s, 0, 2); got != "1 " {
		t.Fatal("absolute", got)
	}
	if e.changed() || len(e.undo) != 0 {
		t.Fatal("display toggle changed buffer/history")
	}
	e = newEditor(strings.Repeat("x", 100) + "\n世界")
	v = view{}
	keys(e, "$")
	v.draw(s, e)
	if got := screenRow(s, 0, 2); got != "1 " {
		t.Fatal("gutter scrolled", got)
	}
	x, y, _ := s.GetCursor()
	if x != 19 || y != 0 {
		t.Fatal(x, y)
	}
	e.command("2")
	v.draw(s, e)
	if got := screenRow(s, 1, 3); got != "2 世" {
		t.Fatal("Unicode offset", got)
	}
	s.SetSize(1, 3)
	v.draw(s, e)
	x, _, _ = s.GetCursor()
	if x != 0 {
		t.Fatal("narrow cursor", x)
	}
	e = newEditor(strings.Repeat("x\n", 120))
	v = view{}
	s.SetSize(20, 5)
	e.command("100")
	v.draw(s, e)
	if got := screenRow(s, 2, 4); got != "100 " {
		t.Fatal("scrolled numbering", got)
	}
	e.command("relative")
	v.draw(s, e)
	if got := screenRow(s, 1, 4); got != "  1 " {
		t.Fatal("scrolled relative numbering", got)
	}
}

func TestCzechPrimaryMessages(t *testing.T) {
	e := newEditor("a\nb")
	for _, c := range []struct{ cmd, prefix string }{{"relative", "Relativní čísla"}, {"number", "Čísla řádků"}, {"0", "Neplatné číslo"}, {"unknown", "Neznámý příkaz"}} {
		e.command(c.cmd)
		if !strings.HasPrefix(e.message, c.prefix) || !strings.Contains(e.message, " / ") {
			t.Fatal(c.cmd, e.message)
		}
	}
	s := tcell.NewSimulationScreen("UTF-8")
	s.Init()
	defer s.Fini()
	s.SetSize(80, 10)
	v := view{}
	v.draw(s, e)
	if !strings.HasPrefix(screenRow(s, 8, 80), "NORMÁLNÍ/NORMAL") {
		t.Fatal("Czech mode label missing")
	}
}
