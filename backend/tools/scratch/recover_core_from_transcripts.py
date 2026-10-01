"""Replay Write/StrReplace from Cursor agent transcripts onto core/ files."""
from __future__ import annotations

import json
import os
from pathlib import Path

ROOT = Path(r"D:\lol-build-overlay")
TRANSCRIPTS = Path(r"C:\Users\kunit\.cursor\projects\d-lol-build-overlay\agent-transcripts")
CORE_PREFIXES = (
    str(ROOT / "core").replace("/", "\\"),
    str(ROOT / "core").replace("\\", "/"),
)

files: dict[str, str] = {}
ops = 0
writes = 0
replaces = 0
replace_miss = 0
log: list[str] = []


def norm(path: str) -> str | None:
    if not path:
        return None
    p = path.replace("/", "\\")
    low = p.lower()
    if "\\core\\" not in low and not low.endswith("\\core"):
        return None
    # Keep Windows path as stored.
    return os.path.normpath(p)


def apply_write(path: str, contents: str) -> None:
    global writes
    n = norm(path)
    if n is None:
        return
    files[n] = contents
    writes += 1


def apply_replace(path: str, old: str, new: str, replace_all: bool) -> None:
    global replaces, replace_miss
    n = norm(path)
    if n is None:
        return
    cur = files.get(n)
    if cur is None:
        replace_miss += 1
        log.append(f"MISS (no file yet) {n}")
        return
    if old not in cur:
        replace_miss += 1
        log.append(f"MISS (old_string) {n}")
        return
    if replace_all:
        files[n] = cur.replace(old, new)
    else:
        files[n] = cur.replace(old, new, 1)
    replaces += 1


def walk_obj(obj) -> None:
    global ops
    if isinstance(obj, list):
        for x in obj:
            walk_obj(x)
        return
    if not isinstance(obj, dict):
        return
    name = obj.get("name") or obj.get("toolName")
    inp = obj.get("input") or obj.get("arguments") or {}
    if not isinstance(inp, dict):
        inp = {}
    if name == "Write":
        ops += 1
        apply_write(str(inp.get("path") or ""), str(inp.get("contents") or ""))
    elif name == "StrReplace":
        ops += 1
        apply_replace(
            str(inp.get("path") or ""),
            str(inp.get("old_string") or ""),
            str(inp.get("new_string") or ""),
            bool(inp.get("replace_all")),
        )
    # Recurse into nested message/content/tool_use.
    for v in obj.values():
        if isinstance(v, (dict, list)):
            walk_obj(v)


def main() -> None:
    jsonls = sorted(TRANSCRIPTS.rglob("*.jsonl"), key=lambda p: p.stat().st_mtime)
    for jp in jsonls:
        with jp.open("r", encoding="utf-8", errors="replace") as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    rec = json.loads(line)
                except json.JSONDecodeError:
                    continue
                walk_obj(rec)

    out_count = 0
    for path, contents in files.items():
        Path(path).parent.mkdir(parents=True, exist_ok=True)
        Path(path).write_text(contents, encoding="utf-8", newline="\n")
        out_count += 1
        print(path, "bytes", len(contents.encode("utf-8")))

    print("---")
    print("jsonl", len(jsonls), "ops", ops, "writes", writes, "replaces", replaces, "miss", replace_miss, "files", out_count)
    for line in log[:40]:
        print(line)
    if len(log) > 40:
        print("... +", len(log) - 40, "more misses")


if __name__ == "__main__":
    main()
