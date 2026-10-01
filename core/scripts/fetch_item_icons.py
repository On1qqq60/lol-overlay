"""Download item PNGs for the overlay HUD.

Reads core/data/items.json (Data Dragon URLs) and writes
overlay/assets/items/{id}.png — ImageCache keys files by item id.

Usage (from core/ or repo root):
  python scripts/fetch_item_icons.py
"""

from __future__ import annotations

import json
import sys
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

CORE = Path(__file__).resolve().parents[1]
ITEMS_JSON = CORE / "data" / "items.json"
OUT_DIR = CORE.parent / "overlay" / "assets" / "items"
WORKERS = 16


def download_one(url: str, dest: Path) -> None:
    req = urllib.request.Request(url, headers={"User-Agent": "lol-build-overlay"})
    with urllib.request.urlopen(req, timeout=20) as resp:
        data = resp.read()
    if not data:
        raise RuntimeError("empty body")
    dest.write_bytes(data)


def main() -> int:
    print(f"reading {ITEMS_JSON}", flush=True)
    if not ITEMS_JSON.is_file():
        print(f"missing {ITEMS_JSON}", file=sys.stderr)
        return 1

    payload = json.loads(ITEMS_JSON.read_text(encoding="utf-8"))
    items = payload.get("items") or []
    OUT_DIR.mkdir(parents=True, exist_ok=True)

    jobs: list[tuple[int, str, Path]] = []
    skip = 0
    for it in items:
        iid = it.get("id")
        url = it.get("icon")
        if iid is None or not url:
            continue
        dest = OUT_DIR / f"{iid}.png"
        if dest.is_file() and dest.stat().st_size > 0:
            skip += 1
            continue
        jobs.append((int(iid), str(url), dest))

    print(f"download {len(jobs)} icons, skip {skip} -> {OUT_DIR}", flush=True)
    ok = fail = 0
    with ThreadPoolExecutor(max_workers=WORKERS) as pool:
        futs = {pool.submit(download_one, url, dest): iid for iid, url, dest in jobs}
        for i, fut in enumerate(as_completed(futs), 1):
            iid = futs[fut]
            try:
                fut.result()
                ok += 1
            except Exception as ex:
                print(f"fail {iid}: {ex}", file=sys.stderr, flush=True)
                fail += 1
            if i % 50 == 0 or i == len(futs):
                print(f"  {i}/{len(futs)}", flush=True)

    print(f"item icons: downloaded={ok} skipped={skip} failed={fail}", flush=True)
    return 0 if (ok + skip) > 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
