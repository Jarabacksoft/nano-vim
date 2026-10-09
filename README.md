# nano-vim

Terminálový editor, který zná Vim a hledí si svého. Od Jarabacksoft.

**Čeština je hlavní a řídící jazyk projektu.** Anglický překlad je níže;
při rozporu platí česká verze.

## Česky

### Sestavení a spuštění

Potřebujete Go 1.26.1 nebo novější a unixový terminál (macOS nebo Linux).

```sh
go build -o nano-vim .
./nano-vim 'soubor s mezerami.txt'
```

Pro instalaci zkopírujte binární soubor do adresáře na `PATH`. Předávejte právě
jeden název souboru. Neexistující soubor začne prázdný a vytvoří se při uložení.
Editor začíná v režimu Normal. Jeden proces pracuje s jedním souborem a jedním
bufferem. Nemá konfiguraci, pluginy, projekty, záložky, rozdělená okna, shell ani
služby na pozadí. Závislosti se stahují při sestavení; hotový editor síť nepotřebuje.

### Klávesy a příkazy

Příkazy začínající `:` potvrďte klávesou Enter. Názvy příkazů jsou stejné v obou
jazykových verzích dokumentace.

| Činnost | Klávesy |
| --- | --- |
| Pohyb | `h j k l`, šipky, `w b e`, `0 ^ $`, `gg G` |
| Skok na řádek | `:42`, `42G` nebo `42gg` |
| Zobrazení/skrytí čísel řádků | `:number` |
| Absolutní/relativní čísla | `:relative` |
| Posun o půl obrazovky | Ctrl-D, Ctrl-U |
| Vkládání | `i a I A o O`; Esc vrací do režimu Normal |
| Úpravy při vkládání | Text, Enter, Tab, Backspace, Delete, šipky |
| Výběr | `v` po znacích, `V` po řádcích; pohyb rozšiřuje výběr |
| Práce s výběrem | `d` smazat, `y` zkopírovat, `c` změnit; Esc zrušit |
| Mazání/náhrada | `x`, `r` a znak, `dd`, `dw`, `de`, `d$`, `d0`, `D` |
| Změna textu | `cw`, `cc`, `C` |
| Kopírování/vložení | `yy`, `p` za/pod, `P` před/nad |
| Spojení řádků | `J` |
| Historie | `u`, Ctrl-R; jedno vkládání je jeden krok zpět |
| Opakování změny | `.` včetně vloženého textu |
| Hledání | `/text` dopředu, `?text` dozadu; Enter potvrdit, Esc zrušit |
| Opakování hledání | `n` stejným směrem, `N` opačně; `*`/`#` slovo pod kurzorem |
| Uložení/ukončení | `:w`, `:q`, `:wq`, `:q!` |
| Znovunačtení aktuálního souboru | `:e`; odmítne neuložené změny |

Počty fungují pro pohyby, mazání, kopírování, vložení, náhradu, spojování řádků,
historii, opakování hledání a změn: `3dd`, `2dw`, `4w`, `2p`, `3x`. Počty
operátoru a pohybu se násobí (`2d3w`). Počty vkládání a textových objektů chybí.

Za `d`, `c` nebo `y` lze použít textové objekty `iw`, `aw`, `i"`, `a"`, `i'`,
`a'`, `i(`, `a(`, `i[`, `a[`, `i{`, `a{`. Příklady: `ciw`, `yiw`, `ci"`, `da(`.
Závorky mohou být vnořené a přes více řádků. Uvozovky se párují na aktuálním
řádku; escapované uvozovky se přeskakují. Nerozpoznaný objekt buffer nezmění.
Nejde o parser programovacího jazyka.

### Čísla řádků a přeskakování

Po spuštění se zobrazují absolutní čísla. `:number` čísla skryje nebo zobrazí.
`:relative` přepíná mezi absolutními čísly a vzdáleností od aktuálního řádku;
aktuální řádek vždy ukazuje své absolutní číslo. Relativní `5` nad/pod kurzorem
odpovídá pohybu `5k`/`5j`. Zvolený režim se zachová i při skrytí čísel.
Oba přepínače platí pouze v aktuálním procesu, bez konfiguračního souboru.

`:42`, `42G` a `42gg` přesunou kurzor na začátek řádku 42. Samotné `G` skočí
na poslední řádek, `gg` na první. Číslo za koncem souboru skočí na poslední
řádek; `:0` a čísla mimo rozsah typu integer se odmítnou. Počítané skoky fungují
i při výběru. Čísla se neposouvají vodorovně a v příliš úzkém okně se skryjí.

### Hledání, schránka a stav

Hledání je doslovné, rozlišuje velikost písmen, přechází přes konec/začátek a
prohledává pouze aktuální buffer. Prázdný dotaz opakuje předchozí vzor zvoleným
směrem. Hledání slova používá doslovný vzor.

Interní nepojmenovaný registr funguje vždy. Kopírování do systémové schránky
používá OSC 52 přes tcell, pokud to terminál dovolí. Ze systémové schránky
vkládejte běžnou funkcí terminálu. Bracketed paste vkládá znaky jako text a podle
potřeby vstoupí do režimu Insert.

Stavový řádek ukazuje režim, soubor, pozici kurzoru, přibližný průchod souborem a
`[+]` u neuložených změn. V úzkém okně vynechává méně důležité údaje. Poslední
řádek obsahuje zprávy a vstup hledání či příkazu. Vlastní názvy režimů a
hlášky editoru jsou dvojjazyčné, česky první. Původní chyby systému nebo knihovny
mohou být v jazyce daného prostředí.

### Soubory a terminál

UTF-8 a jednotné LF/CRLF se zachovávají bez přidání závěrečného odřádkování.
Podporované jsou prázdné soubory, tabulátory a názvy s mezerami. Neplatné UTF-8,
binární řídicí znaky, samostatné CR, smíšené LF/CRLF, jiné než běžné soubory a
soubory nad 32 MiB se odmítají. Tento limit platí i při ukládání.

Nové soubory mají práva 0600; existující práva se zachovávají. Existující symlinky
se při otevření vyhodnotí; neplatné symlinky se odmítají. Uložení hardlinků je
odmítnuté. Uložení zapíše a synchronizuje dočasný soubor ve stejném adresáři a
přejmenuje jej na cíl. Chyba zachová rozpracovaný obsah v paměti.

`:q` a `:e` chrání neuložené změny. Uložení kontroluje změnu obsahu, vytvoření,
smazání a změnu konců řádků na disku. Konflikt odmítne; `:q!` výslovně zahodí
úpravy. Změnu mezi poslední kontrolou a přejmenováním nelze zcela vyloučit.
Nahrazení zachovává běžná práva, ale ne všechna rozšířená metadata či ACL.

Editor vyžaduje interaktivní vstupní i výstupní terminál. Změna velikosti překreslí
obrazovku. Normální ukončení, chyby po inicializaci a SIGINT/SIGTERM/SIGHUP obnoví
nastavení terminálu i původní obrazovku. Signály ukončí editor bez uložení;
průběžně používejte `:w`. Ctrl-C funguje jako Esc. SIGKILL ani výpadek stroje
nemohou provést úklid.

### Testování

```sh
go test ./...
go test -race ./...
go vet ./...
go test -fuzz=FuzzKeys -fuzztime=15s
go build -o nano-vim .
python3 tests/terminal_test.py
```

Python test používá standardní knihovnu a skutečné PTY. Ověřuje úpravy, uložení
CRLF, hledání, chráněné ukončení, šipky, výběr, změnu velikosti, vložení textu,
čísla řádků, skoky, signály a obnovu terminálu. Go testy pokrývají příkazy,
textové objekty, historii, Unicode, integritu souborů, neplatné vstupy a vykreslení.

### Známá omezení

Jde o vybranou podmnožinu Vimu. Pozice a úpravy používají jednotlivé Unicode
code pointy; složená emoji a kombinované znaky nejsou nedělitelné jednotky.
Šířky znaků podle knihovny se mohou lišit od konkrétního terminálu. Výběr zahrnuje
oba konce. Schránka závisí na povolení OSC 52; editor schránku nečte ani nespouští
externí pomocníky. Historie ukládá celé kopie bufferu, takže u velkých souborů
spotřebovává paměť. Počítané textové objekty, počty vkládání, automatické odsazení,
zvýraznění syntaxe a regulární výrazy nejsou implementované.

PTY testy byly provedeny na macOS. Linuxové sestavení je ověřené, ale interaktivní
Linux a schránka či zobrazení ve fyzickém terminálu stále vyžadují ruční ověření.

## English

A terminal editor that knows Vim, but minds its own business. By Jarabacksoft.
**Czech is the primary and authoritative project language.** This section is a
translation; the Czech version takes precedence if the two differ.

### Build and run

Requires Go 1.26.1 or newer and a Unix terminal (macOS or Linux).

```sh
go build -o nano-vim .
./nano-vim 'file with spaces.txt'
```

To install, copy the binary to a directory on `PATH`. Supply exactly one filename.
A missing file starts empty and is created on save. Startup uses Normal mode.
One process edits one file and one buffer. There are no configuration files,
plugins, projects, tabs, splits, shells, or background services. Dependencies are
downloaded at build time; the binary needs no network at runtime.

### Keys and commands

Confirm commands beginning with `:` using Enter. Command names are identical in
both documentation languages.

| Action | Keys |
| --- | --- |
| Move | `h j k l`, arrows, `w b e`, `0 ^ $`, `gg G` |
| Jump to line | `:42`, `42G`, or `42gg` |
| Show/hide line numbers | `:number` |
| Absolute/relative numbers | `:relative` |
| Half-screen movement | Ctrl-D, Ctrl-U |
| Insert | `i a I A o O`; Esc returns to Normal |
| Insert editing | Text, Enter, Tab, Backspace, Delete, arrows |
| Selection | `v` character-wise, `V` line-wise; motions extend it |
| Use selection | `d` delete, `y` copy, `c` change; Esc cancels |
| Delete/replace | `x`, `r` then a character, `dd`, `dw`, `de`, `d$`, `d0`, `D` |
| Change | `cw`, `cc`, `C` |
| Copy/paste | `yy`, `p` after/below, `P` before/above |
| Join lines | `J` |
| History | `u`, Ctrl-R; an Insert session is one undo step |
| Repeat change | `.` including inserted text |
| Search | `/text` forward, `?text` backward; Enter confirms, Esc cancels |
| Repeat search | `n` same direction, `N` opposite; `*`/`#` use the word under the cursor |
| Save/exit | `:w`, `:q`, `:wq`, `:q!` |
| Reload current file | `:e`; refuses unsaved changes |

Counts work on motions, deletion, copying, pasting, replacement, joining, history,
search repetition, and repeated changes: `3dd`, `2dw`, `4w`, `2p`, `3x`.
Operator and motion counts multiply (`2d3w`). Insert counts and counted text objects
are absent.

After `d`, `c`, or `y`, use `iw`, `aw`, `i"`, `a"`, `i'`, `a'`, `i(`, `a(`,
`i[`, `a[`, `i{`, `a{`. Examples: `ciw`, `yiw`, `ci"`, `da(`. Brackets may nest
and span lines. Quotes pair within the current line, skipping escaped delimiters.
Unresolved objects leave the buffer intact. This is not a programming-language parser.

### Line numbers and jumping

Absolute numbers appear at startup. `:number` shows or hides the gutter.
`:relative` switches between absolute numbers and distances from the current line;
the current line always shows its absolute number. Relative `5` above/below the
cursor corresponds to `5k`/`5j`. The mode survives hiding the gutter. Both toggles
apply only to the current process, with no configuration file.

`:42`, `42G`, and `42gg` jump to the beginning of line 42. Plain `G` goes to the
last line and `gg` to the first. Out-of-file requests clamp to the last line;
`:0` and integer overflow are rejected. Counted jumps extend Visual selections.
The gutter stays fixed horizontally and disappears in excessively narrow windows.

### Search, clipboard, and status

Search is literal, case-sensitive, wraps around, and examines only this buffer.
An empty query repeats the previous pattern in the entered direction. Word search
uses a literal pattern.

The internal unnamed register always works. Terminal clipboard copying uses OSC 52
through tcell where permitted. Paste from the system clipboard using your terminal.
Bracketed paste inserts characters as text and enters Insert mode when needed.

The status line shows mode, filename, cursor position, approximate progress, and
`[+]` for unsaved changes. Narrow windows omit lower-priority fields. The final
line holds messages and search/command input. Mode labels and editor messages
are bilingual, Czech first. Underlying system or library errors may use the
language of the environment.

### Files and terminal

UTF-8 and uniform LF/CRLF round-trip without adding a final newline. Empty files,
tabs, and filenames with spaces are supported. Invalid UTF-8, binary control
characters, bare CR, mixed LF/CRLF, nonregular files, and files over 32 MiB are
rejected. The limit also applies when saving.

New files use mode 0600; existing permissions are preserved. Existing symlinks are
resolved at open; dangling symlinks are rejected. Hard-linked saves are refused.
Saving writes and syncs a temporary file in the same directory, then renames it over
the target. Failure preserves in-memory work.

`:q` and `:e` protect unsaved edits. Saves check external content changes, creation,
deletion, and line-ending changes. Conflicts are refused; `:q!` explicitly discards
edits. Changes between the final disk check and rename cannot be fully excluded.
Replacement preserves ordinary permissions, but not all extended metadata or ACLs.

Interactive input and output terminals are required. Resizing redraws the screen.
Normal exits, errors after initialization, and SIGINT/SIGTERM/SIGHUP restore terminal
settings and the original screen. Signals exit without saving; use `:w` regularly.
Ctrl-C behaves as Esc. SIGKILL and machine failure cannot run cleanup.

### Testing

```sh
go test ./...
go test -race ./...
go vet ./...
go test -fuzz=FuzzKeys -fuzztime=15s
go build -o nano-vim .
python3 tests/terminal_test.py
```

The Python test uses the standard library and real PTYs to check editing, CRLF saving,
search, protected quit, arrows, selection, resizing, pasted text, line numbers,
jumps, signals, and restoration. Go tests cover commands, text objects, history,
Unicode, file integrity, invalid input, and rendering.

### Known limits

This is a selected Vim subset. Cursor positions and edits use Unicode code points;
complex emoji and combining sequences are not indivisible units. Library width rules
can differ from individual terminals. Selections include both endpoints. Clipboard
support depends on OSC 52 policy; there is no clipboard reader or external helper.
History stores full buffer snapshots, consuming memory on large files. Counted text
objects, insert counts, autoindent, syntax highlighting, and regex search are absent.

PTY tests have run on macOS. Linux builds are verified; interactive Linux and physical
terminal clipboard/display behavior still need manual checks.
