#!/usr/bin/env python3
"""Regrese kontroly licencí / License validation regressions."""
import argparse
import contextlib
import importlib.util
import io
from pathlib import Path
import tempfile
import unittest


SCRIPT = Path(__file__).resolve().parents[1] / "scripts" / "package-release.py"
SPEC = importlib.util.spec_from_file_location("package_release", SCRIPT)
PACKAGE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(PACKAGE)


class LicenseTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.directory = Path(self.temporary.name)
        self.parser = argparse.ArgumentParser()

    def write(self, name, text):
        path = self.directory / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(text, encoding="utf-8")

    def verify(self, notices, runtime=False, directory=None):
        PACKAGE.verify_licenses(
            self.parser, directory or self.directory, notices, "example@v1",
            runtime=runtime,
        )

    def rejected(self, notices, runtime=False):
        with contextlib.redirect_stderr(io.StringIO()):
            with self.assertRaises(SystemExit) as error:
                self.verify(notices, runtime=runtime)
        self.assertEqual(error.exception.code, 2)

    def test_version_alone_does_not_cover_license(self):
        self.write("LICENSE", "Original license")
        self.rejected("example@v1")

    def test_all_notices_must_be_present(self):
        self.write("LICENSE.txt", "Original license")
        self.write("NOTICE", "Original attribution")
        self.rejected("Original license")
        self.verify("Original license\nOriginal attribution")

    def test_missing_or_empty_license_is_rejected(self):
        self.rejected("example@v1")
        self.write("LICENSE", "\n")
        self.rejected("example@v1")

    def test_go_requires_license_and_patents(self):
        self.write("PATENTS", "Go patents")
        self.rejected("Go patents", runtime=True)
        self.write("LICENSE", "Go license")
        self.rejected("Go license", runtime=True)
        self.verify("Go license\nGo patents", runtime=True)

    def test_homebrew_license_above_goroot(self):
        self.write("LICENSE", "Go license")
        self.write("libexec/PATENTS", "Go patents")
        self.verify("Go license\nGo patents", runtime=True,
                    directory=self.directory / "libexec")

    def test_go_license_in_goroot(self):
        self.write("LICENSE", "Go license")
        self.write("PATENTS", "Go patents")
        self.verify("Go license\nGo patents", runtime=True)


if __name__ == "__main__":
    unittest.main()
