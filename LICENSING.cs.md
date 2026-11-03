# Licence nano-vim

Zvoleno 3. listopadu 2026. [English](LICENSING.en.md).

Vlastní zdrojový kód nano-vim a jeho projektová dokumentace jsou licencované pod **Apache License 2.0**. Nositelem vlastních práv je Petr Turek. Úplný nezměněný text je v [LICENSE](LICENSE); oznámení projektu v [NOTICE](NOTICE). Anglický licenční originál je autoritativní; toto české vysvětlení jej nenahrazuje. Licence kódu neuděluje právo vystupovat pod značkou Jarabacksoft nebo se vydávat za původního autora.

Použité knihovny a běhové prostředí Go zachovávají své vlastní podmínky. Jejich původní licenční a patentové texty jsou v [THIRD_PARTY_NOTICES.txt](THIRD_PARTY_NOTICES.txt). Výběr Apache 2.0 je nepřelicencuje. Při změně závislostí nebo podmínek běhového prostředí tento soubor znovu zkontrolujte a aktualizujte.

Pro redistribuci binárky vytvořte balíček se všemi oznámeními:

```sh
python3 scripts/package-release.py /private/tmp/nano-vim-release
```

Cílový adresář musí být nový. Skript sestaví editor pro aktuální prostředí Go a přiloží README, LICENSE, NOTICE, oznámení třetích stran a obě jazykové verze tohoto přehledu. Neprovádí upload ani publikaci. Před sestavením porovná původní licenční a oznamovací texty používaných modulových závislostí a Go s přiloženými podklady; při chybějícím nebo odlišném textu balíček nevytvoří. Závislosti musí být předem dostupné v místní Go cache, skript je nestahuje. Samotné `go build` nadále vytvoří pouze binárku.

U dalších aplikačních projektů zařaďte výběr licence co nejdříve, například jako klidnější krok po prvním delším bloku práce, vždy před zveřejněním/distribucí. Licence se vybírá individuálně a zavádí přímo v projektu; CZ/EN dokumentace a přehled závislostí jsou součástí tohoto checkpointu.
