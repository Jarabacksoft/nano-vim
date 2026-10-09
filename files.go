package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf8"
)

const maxFileSize = 32 << 20

func (e *editor) takeCount() int {
	n, _ := strconv.Atoi(e.count)
	e.count = ""
	if n < 1 {
		n = 1
	}
	return n
}
func readText(name string) (string, string, error) {
	initial, statErr := os.Stat(name)
	if os.IsNotExist(statErr) {
		return "", "\n", nil
	}
	if statErr != nil {
		return "", "", statErr
	}
	if !initial.Mode().IsRegular() {
		return "", "", fmt.Errorf("Cíl není běžný soubor / target is not a regular file")
	}
	f, err := os.Open(name)
	if os.IsNotExist(err) {
		return "", "\n", nil
	}
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return "", "", err
	}
	if !info.Mode().IsRegular() {
		return "", "", fmt.Errorf("Cíl není běžný soubor / target is not a regular file")
	}
	if info.Size() > maxFileSize {
		return "", "", fmt.Errorf("Soubor překračuje limit 32 MiB / file exceeds 32 MiB limit")
	}
	data, err := io.ReadAll(io.LimitReader(f, maxFileSize+1))
	if err != nil {
		return "", "", err
	}
	if len(data) > maxFileSize {
		return "", "", fmt.Errorf("Soubor překračuje limit 32 MiB / file exceeds 32 MiB limit")
	}
	if !utf8.Valid(data) {
		return "", "", fmt.Errorf("Soubor není platné UTF-8 / file is not valid UTF-8")
	}
	s := string(data)
	for _, r := range s {
		if r == 0 || (r < 32 && r != '\n' && r != '\r' && r != '\t') || r == 127 {
			return "", "", fmt.Errorf("Nepodporovaný řídicí znak v textu / unsupported control character in text file")
		}
	}
	ending := "\n"
	if strings.Contains(s, "\r\n") {
		ending = "\r\n"
		s = strings.ReplaceAll(s, "\r\n", "\n")
		if strings.Count(s, "\n") != strings.Count(string(data), "\r\n") {
			return "", "", fmt.Errorf("Smíšené LF a CRLF nejsou podporované / mixed LF and CRLF endings are unsupported")
		}
	}
	if strings.Contains(s, "\r") {
		return "", "", fmt.Errorf("Samostatné CR není podporované / bare CR line endings are unsupported")
	}
	return s, ending, nil
}
func openEditor(name string) (*editor, error) {
	path, err := filepath.Abs(name)
	if err != nil {
		return nil, err
	}
	if resolved, resolveErr := filepath.EvalSymlinks(path); resolveErr == nil {
		path = resolved
	} else if !os.IsNotExist(resolveErr) {
		return nil, resolveErr
	} else if info, lerr := os.Lstat(path); lerr == nil && info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("Neplatný symlink není podporovaný / dangling symlink is unsupported")
	}
	text, ending, err := readText(path)
	if err != nil {
		return nil, err
	}
	e := newEditor(text)
	e.filename = path
	e.ending = ending
	_, statErr := os.Stat(path)
	e.diskExists = statErr == nil
	return e, nil
}
func (e *editor) save() error {
	current, ending, err := readText(e.filename)
	if err != nil {
		return err
	}
	if current != e.saved || (current != "" && ending != e.ending) {
		return fmt.Errorf("Soubor se změnil na disku; uložení odmítnuto / file changed on disk; save refused")
	}
	info, err := os.Lstat(e.filename)
	if e.diskExists != (err == nil) {
		return fmt.Errorf("Soubor na disku vznikl nebo zmizel; uložení odmítnuto / file appeared or disappeared on disk; save refused")
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil && (!info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0) {
		return fmt.Errorf("Cíl již není běžný soubor / target is no longer a regular file")
	}
	if info != nil {
		if stat, ok := info.Sys().(*syscall.Stat_t); ok && stat.Nlink > 1 {
			return fmt.Errorf("Atomické uložení hardlinku není podporované / atomic save of a hard-linked file is unsupported")
		}
	}
	// Check the target's permissions as well as the directory's before replacement.
	if info != nil {
		f, err := os.OpenFile(e.filename, os.O_WRONLY, 0)
		if err != nil {
			return err
		}
		f.Close()
	}
	f, err := os.CreateTemp(filepath.Dir(e.filename), ".nano-vim-save-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if info != nil {
		if err = f.Chmod(info.Mode().Perm()); err != nil {
			f.Close()
			return err
		}
	} else {
		if err = f.Chmod(0600); err != nil {
			f.Close()
			return err
		}
	}
	data := strings.ReplaceAll(string(e.text), "\n", e.ending)
	if len(data) > maxFileSize {
		f.Close()
		return fmt.Errorf("Buffer překračuje limit 32 MiB / buffer exceeds 32 MiB file limit")
	}
	_, err = f.WriteString(data)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	// Recheck immediately before replacing the target.
	latest, latestEnding, err := readText(e.filename)
	if err != nil {
		return err
	}
	if latest != current || latestEnding != ending {
		return fmt.Errorf("Soubor se změnil během ukládání / file changed during save")
	}
	if err = os.Rename(tmp, e.filename); err != nil {
		return err
	}
	e.saved = string(e.text)
	e.diskExists = true
	return nil
}
func (e *editor) command(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd != "" && strings.Trim(cmd, "0123456789") == "" {
		n, err := strconv.Atoi(cmd)
		if err != nil || n < 1 {
			e.message = "Neplatné číslo řádku / Invalid line number"
			return
		}
		e.jumpLine(n)
		return
	}
	switch cmd {
	case "number":
		e.number = !e.number
		if e.number {
			e.message = "Čísla řádků zapnuta / Line numbers on"
		} else {
			e.message = "Čísla řádků vypnuta / Line numbers off"
		}
	case "relative":
		e.relative = !e.relative
		if e.relative {
			e.message = "Relativní čísla zapnuta / Relative line numbers on"
		} else {
			e.message = "Absolutní čísla řádků / Absolute line numbers"
		}

	case "w", "wq":
		if err := e.save(); err != nil {
			e.message = "Uložení selhalo / Save failed: " + err.Error()
			return
		}
		e.message = "Uloženo / Saved"
		if strings.TrimSpace(cmd) == "wq" {
			e.quit = true
		}
	case "q":
		if e.changed() {
			e.message = "Neuložené změny; použijte :w nebo :q! / Unsaved changes; use :w or :q!"
		} else {
			e.quit = true
		}
	case "q!":
		e.quit = true
	case "e":
		if e.changed() {
			e.message = "Neuložené změny; načtení odmítnuto / Unsaved changes; reload refused"
			return
		}
		s, ending, err := readText(e.filename)
		if err != nil {
			e.message = "Načtení selhalo / Reload failed: " + err.Error()
			return
		}
		e.text = []rune(s)
		e.saved = s
		e.ending = ending
		_, statErr := os.Stat(e.filename)
		e.diskExists = statErr == nil
		e.pos = 0
		e.undo = nil
		e.redo = nil
		e.last = nil
		e.message = "Znovu načteno / Reloaded"
	default:
		e.message = "Neznámý příkaz / Unknown command"
	}
}
