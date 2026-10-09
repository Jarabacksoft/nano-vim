#!/usr/bin/env python3
"""Real PTY checks for the compiled editor; no third-party Python packages."""
import fcntl
import os
from pathlib import Path
import pty
import select
import signal
import struct
import subprocess
import tempfile
import sys
import termios
import time

BINARY = str(Path(__file__).resolve().parents[1] / "nano-vim")

class Session:
    def __init__(self, filename):
        self.master, self.slave = pty.openpty()
        self.before = termios.tcgetattr(self.slave)
        self.resize(80, 24)
        def setup():
            os.setsid()
            fcntl.ioctl(self.slave, termios.TIOCSCTTY, 0)
        report_read, report_write = os.pipe()
        release_read, self.release_write = os.pipe()
        guard = """import os, subprocess, sys
p = subprocess.Popen(sys.argv[1:3])
os.write(int(sys.argv[3]), (str(p.pid)+"\\n").encode())
code = p.wait()
os.write(int(sys.argv[3]), (str(code)+"\\n").encode())
os.read(int(sys.argv[4]), 1)
"""
        self.process = subprocess.Popen(
            [sys.executable, "-c", guard, BINARY, str(filename),
             str(report_write), str(release_read)],
            stdin=self.slave, stdout=self.slave, stderr=self.slave,
            preexec_fn=setup, pass_fds=(report_write, release_read),
            env={**os.environ, "TERM": "xterm-256color"})
        os.close(report_write)
        os.close(release_read)
        self.report = os.fdopen(report_read, "rb", buffering=0)
        self.pid = int(self.report.readline())
        self.returncode = None
        self.output = b""
        self.drain(0.25)
        assert self.poll() is None, self.output

    def resize(self, width, height):
        fcntl.ioctl(self.slave, termios.TIOCSWINSZ,
                    struct.pack("HHHH", height, width, 0, 0))
        if hasattr(self, "process"):
            os.kill(self.pid, signal.SIGWINCH)
            self.drain(0.1)

    def drain(self, duration=0.12):
        deadline = time.monotonic() + duration
        while time.monotonic() < deadline:
            ready, _, _ = select.select([self.master], [], [], 0.02)
            if ready:
                self.output += os.read(self.master, 65536)

    def send(self, text):
        try:
            os.write(self.master, text.encode("utf-8"))
        except OSError:
            raise AssertionError((self.poll(), self.output))
        self.drain()

    def poll(self):
        if self.returncode is None:
            ready, _, _ = select.select([self.report], [], [], 0)
            if ready:
                self.returncode = int(self.report.readline())
        return self.returncode

    def finish(self):
        deadline = time.monotonic() + 3
        while self.poll() is None and time.monotonic() < deadline:
            self.drain(0.05)
        assert self.returncode == 0, (self.returncode, self.output)
        after = termios.tcgetattr(self.slave)
        # macOS sets PENDIN when canonical input is restored; this is a transient
        # kernel flag requesting retyping, rather than an editor setting.
        after[3] &= ~getattr(termios, "PENDIN", 0)
        before = list(self.before)
        before[3] &= ~getattr(termios, "PENDIN", 0)
        assert after == before, ("terminal settings not restored", before, after)
        assert b"\x1b[?1049l" in self.output, "alternate screen not restored"
        os.write(self.release_write, b"x")
        self.process.wait(timeout=3)
        self.report.close()
        os.close(self.release_write)
        os.close(self.master)
        os.close(self.slave)

with tempfile.TemporaryDirectory() as directory:
    filename = Path(directory) / "new file.txt"
    s = Session(filename)
    s.send("ihello 世界")
    s.send("\x1b")
    s.send(":q\r")
    assert s.poll() is None, "unsafe quit allowed"
    assert b"Unsaved" in s.output
    s.resize(15, 5)
    s.resize(100, 30)
    s.send(":wq\r")
    s.finish()
    assert filename.read_text() == "hello 世界"

    filename.write_bytes(b"word x word\r\nsecond\r\nthird\r\n")
    s = Session(filename)
    s.send("/word\r")
    s.send("nN")
    s.send("?word\r")
    s.send("gg")
    s.send("ciwreplacement")
    s.send("\x1b")
    s.send("u")
    s.send("\x12")  # Ctrl-R
    s.send(":wq\r")
    s.finish()
    assert filename.read_bytes() == b"replacement x word\r\nsecond\r\nthird\r\n"

    before = filename.read_bytes()
    s = Session(filename)
    s.send("\x1b[B")  # arrow down
    s.send("Vjd")
    s.send(":q!\r")
    s.finish()
    assert filename.read_bytes() == before

    s = Session(filename)
    s.send("\x1b[200~ijhkw 世界\nsecond\x1b[201~")
    s.send(":wq\r")
    s.finish()
    assert filename.read_bytes().startswith("ijhkw 世界\r\nsecond".encode()), filename.read_bytes()

    filename.write_bytes(b"one\r\ntwo\r\nthree\r\nfour")
    s = Session(filename)
    s.send(":relative\r")
    s.send(":number\r")
    s.send(":number\r")
    s.send(":relative\r")
    s.resize(10, 5)
    s.send(":3\r")
    s.send("Ihit-")
    s.send("\x1b")
    s.send("1G")
    s.send("Istart-")
    s.send("\x1b")
    s.send("4gg")
    s.send("A-end")
    s.send("\x1b")
    s.send(":wq\r")
    s.finish()
    assert filename.read_bytes() == b"start-one\r\ntwo\r\nhit-three\r\nfour-end"

    for sig in (signal.SIGINT, signal.SIGTERM, signal.SIGHUP):
        s = Session(filename)
        os.kill(s.pid, sig)
        s.finish()

    result = subprocess.run([BINARY, str(filename)], capture_output=True)
    assert result.returncode != 0
    assert b"interactive terminal" in result.stderr
print("PTY checks passed: editing, search, CRLF save, protected quit, arrows, visual mode, resize, signals, terminal cleanup")
