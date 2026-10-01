"""Download Data Dragon item icons next to the overlay executable."""
from __future__ import annotations

import json
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path
from urllib.request import Request, urlopen

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT.parent / "frontend" / "bin" / "assets" / "items"
ITEMS = ROOT / "data" / "items.json"


def main() -> int:
    data = json.loads(ITEMS.read_text(encoding="utf-8"))
    version = data.get("version") or "16.18.1"
    items = data.get("items") or []
    OUT.mkdir(parents=True, exist_ok=True)
    ok = skip = fail = 0

    def dl(item: dict) -> str:
        iid = item.get("id")
        if not isinstance(iid, int):
            return "skip"
        dest = OUT / f"{iid}.png"
        if dest.exists() and dest.stat().st_size > 32:
            return "skip"
        url = f"https://ddragon.leagueoflegends.com/cdn/{version}/img/item/{iid}.png"
        req = Request(url, headers={"User-Agent": "lol-build-overlay"})
        try:
            with urlopen(req, timeout=20) as resp:
                dest.write_bytes(resp.read())
        except Exception:
            return "fail"
        return "ok"

    with ThreadPoolExecutor(max_workers=16) as pool:
        futures = [pool.submit(dl, item) for item in items if isinstance(item, dict)]
        for fut in as_completed(futures):
            status = fut.result()
            if status == "ok":
                ok += 1
            elif status == "fail":
                fail += 1
            else:
                skip += 1

    print(f"icons: wrote {ok}, skipped {skip}, failed {fail} -> {OUT}")
    return 1 if fail and ok == 0 else 0


if __name__ == "__main__":
    raise SystemExit(main())
