#!/usr/bin/env python3
"""Sestav lokální balíček s licencemi / Build a local bundle with licenses."""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile


FILES = (
    "LICENSE", "NOTICE", "THIRD_PARTY_NOTICES.txt", "README.md",
    "LICENSING.cs.md", "LICENSING.en.md",
)


def verify_licenses(parser, directory, notices, identity, runtime=False):
    if runtime:
        sources = [directory / name for name in ("LICENSE", "PATENTS")]
        # Homebrew keeps LICENSE above its libexec GOROOT.
        if directory.name == "libexec" and not sources[0].is_file():
            sources[0] = directory.parent / "LICENSE"
    else:
        sources = sorted(
            path for path in directory.iterdir()
            if path.is_file() and path.name.upper().split(".")[0]
            in {"LICENSE", "LICENCE", "COPYING", "NOTICE", "PATENTS", "COPYRIGHT"}
        )
        if not any(path.name.upper().split(".")[0]
                   in {"LICENSE", "LICENCE", "COPYING"} for path in sources):
            parser.error(f"Chybí licence / Missing license: {identity}")
    for source in sources:
        if not source.is_file():
            parser.error(f"Chybí licence / Missing license: {identity}: {source.name}")
        original = source.read_text(encoding="utf-8").strip()
        if not original or original not in notices:
            parser.error(
                f"Aktualizujte licenční text / Update license text: {identity}: {source.name}"
            )


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("destination", type=Path)
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[1]
    destination = args.destination.resolve()
    if destination.exists():
        parser.error("Cíl již existuje / Destination already exists")
    for name in FILES:
        if not (root / name).is_file():
            parser.error(f"Chybí licenční soubor / Missing licensing file: {name}")
    notices = (root / "THIRD_PARTY_NOTICES.txt").read_text(encoding="utf-8")
    env = {**os.environ, "GOPROXY": "off", "GOSUMDB": "off"}
    result = subprocess.run(
        ["go", "list", "-deps", "-json", "."], cwd=root, env=env,
        check=True, capture_output=True, text=True,
    ).stdout
    decoder = json.JSONDecoder()
    checked = set()
    while result.strip():
        package, end = decoder.raw_decode(result.lstrip())
        result = result.lstrip()[end:]
        module = package.get("Module")
        if not module or module.get("Main"):
            continue
        identity = f"{module['Path']}@{module['Version']}"
        if module.get("Replace") or identity not in notices:
            parser.error(f"Aktualizujte licence závislostí / Update dependency notices: {identity}")
        if identity not in checked:
            verify_licenses(parser, Path(module["Dir"]), notices, identity)
            checked.add(identity)
    goroot = Path(subprocess.run(
        ["go", "env", "GOROOT"], cwd=root, env=env,
        check=True, capture_output=True, text=True,
    ).stdout.strip())
    verify_licenses(parser, goroot, notices, "Go", runtime=True)
    with tempfile.TemporaryDirectory(prefix="nano-vim-package-") as temporary:
        bundle = Path(temporary) / "nano-vim-release"
        bundle.mkdir()
        subprocess.run(
            ["go", "build", "-trimpath", "-o", str(bundle / "nano-vim"), "."],
            cwd=root, env=env, check=True,
        )
        for name in FILES:
            shutil.copy2(root / name, bundle / name)
        destination.parent.mkdir(parents=True, exist_ok=True)
        shutil.copytree(bundle, destination)
    print(f"Balíček vytvořen / Package created: {destination}")


if __name__ == "__main__":
    try:
        main()
    except subprocess.CalledProcessError as error:
        if error.stderr:
            print(error.stderr, file=sys.stderr, end="")
        raise SystemExit(error.returncode)
