# Pravidla repozitáře / Repository Guidelines

## Česky — řídící verze

### Jazyk a rozsah

Čeština je hlavní jazyk projektu. Dokumentaci a uživatelské texty pište dvojjazyčně: nejdřív česky, potom anglický překlad. Při rozporu platí čeština. Technické názvy a příkazy zachovávejte. Jeden proces smí upravovat jeden soubor v jednom bufferu. Nepřidávejte konfiguraci, pluginy, projektové hledání, shell, záložky ani rozdělená okna. Výslovně povolené přepínače `:number` a `:relative` platí jen pro aktuální proces.

### Struktura a styl

`editor.go` obsahuje buffer a editační automat; `files.go` načítání, chráněné ukládání a příkazy; `main.go` terminál a vykreslení. Testy jsou v `editor_test.go` a `tests/terminal_test.py`, uživatelská dokumentace v `README.md`. Go formátujte pomocí `gofmt -w`; používá tabulátory. Python používá čtyři mezery a standardní knihovnu. Oddělujte editaci, soubory a terminál; každou závislost zdůvodněte.

### Sestavení a testy

Vyžaduje Go 1.26.1+ a unixový terminál.

- `go build -o nano-vim .` — sestavení.
- `./nano-vim 'soubor.txt'` — spuštění.
- `go test ./...`, `go test -race ./...`, `go vet ./...` — testy a kontroly.
- `go test -fuzz=FuzzKeys -fuzztime=15s` — zátěž parseru.
- `python3 tests/terminal_test.py` — PTY testy po sestavení; vyžadují Python 3.

Používejte Go `testing`, názvy `Test...` a regresní případy. Povinný práh pokrytí není stanoven. Ověřujte Unicode, konce řádků, historii, hranice kurzoru, vykreslení a neúspěšné uložení. Schránku kontrolujte také ručně.

### Commity a PR

Před commitem ověřte kořen repozitáře a oddělte jej od nadřazeného osobního repozitáře. Pište krátké české imperativní titulky, například `Oprav mazání slov`, s anglickým překladem v popisu. PR vysvětlí chování, uvede testy, související issues a omezení; vizuální změny doloží snímkem terminálu. Necommitujte binárky, tajné údaje ani místní cesty. Při změně chování aktualizujte oba jazyky README. Chyba uložení musí zachovat buffer.

## English — translation

Czech is authoritative. Put Czech first and English second in documentation and user-facing text; retain technical identifiers. Keep one process, one file, one buffer. No configuration, plugins, project search, shells, tabs, or splits. The approved `:number` and `:relative` toggles are process-local.

`editor.go` implements editing, `files.go` file handling/commands, and `main.go` terminal rendering. Tests live in `editor_test.go` and `tests/terminal_test.py`; usage is in `README.md`. Use `gofmt` tabs and four-space Python indentation. Justify dependencies. The commands above build, run, test, check races, lint, fuzz, and exercise PTYs with Go 1.26.1+ and Python 3. Use `Test...` regressions; no coverage threshold applies. Check clipboard behavior manually.

Use short Czech imperative commit titles with English body translations. PRs explain behavior, checks, issues, limitations, and include terminal captures for visual changes. Avoid binaries, secrets, and local paths. Update both README languages and preserve the buffer on save failure.
