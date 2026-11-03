# nano-vim licensing

Selected November 3, 2026. [Česky](LICENSING.cs.md).

Own nano-vim source code and project documentation are licensed under **Apache License 2.0**. The holder of own rights is Petr Turek. The complete unchanged text is in [LICENSE](LICENSE); project attribution is in [NOTICE](NOTICE). The English license original is authoritative; the Czech explanation does not replace it. The code license does not grant permission to impersonate the original author or represent yourself under the Jarabacksoft brand.

Dependencies and the Go runtime retain their own terms. Their original license and patent texts are included in [THIRD_PARTY_NOTICES.txt](THIRD_PARTY_NOTICES.txt). Choosing Apache 2.0 does not relicense them. Review and update that file when dependencies or runtime terms change.

For binary redistribution, create a package containing all notices:

```sh
python3 scripts/package-release.py /private/tmp/nano-vim-release
```

The destination directory must be new. The script builds the editor for the current Go environment and includes README, LICENSE, NOTICE, third-party notices and both language versions of this overview. It performs no upload or publication. Before building, it compares original license and notice texts from the module dependencies in use and Go with the included material; a missing or differing text prevents packaging. Dependencies must already be available in the local Go cache; the script does not download them. Plain `go build` still produces only the executable.

For future app projects, select licensing early, for example as a calmer task after the first sustained development session, always before publication/distribution. Choose individually and implement it directly in the project; CZ/EN documentation and a dependency overview are part of this checkpoint.
