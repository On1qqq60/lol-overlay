import json
from pathlib import Path

d = json.loads(
    Path(r"D:\lol-build-overlay\core\testdata\generated\live\verdicts.json").read_text(
        encoding="utf-8"
    )
)

print("=== remaining starts ===")
for c in d["cases"]:
    if c["phase"] != "start":
        continue
    if c["active"] in (
        "Jinx",
        "Caitlyn",
        "Lucian",
        "Samira",
        "Kaisa",
        "Zeri",
        "Vayne",
        "Nilah",
        "Smolder",
        "Senna",
    ):
        if int(c["file"][:4]) <= 25:
            n = (c.get("nextItem") or {}).get("name")
            b = [x["name"] for x in (c.get("build") or [])]
            print(c["file"], "next=", n, b[:6])

print("=== Kaisa mid full ===")
for c in d["cases"]:
    if c["active"] == "Kaisa" and c["phase"] == "mid":
        n = (c.get("nextItem") or {}).get("name")
        b = [x["name"] for x in (c.get("build") or [])]
        print(c["file"], "next=", n, b)

print("=== FoN / Randuin ===")
n_fon = n_rand = 0
for c in d["cases"]:
    b = [x["name"] for x in (c.get("build") or [])]
    joined = " | ".join(b)
    if "Force of Nature" in joined:
        n_fon += 1
        print("fon", c["file"], c["active"], b)
    if "Randuin" in joined:
        n_rand += 1
        print("rand", c["file"], c["active"], b)
print("fon", n_fon, "randuin", n_rand)
print(
    "maw sixth",
    sum(
        1
        for c in d["cases"]
        if (c.get("build") or []) and "Maw" in c["build"][-1]["name"]
    ),
)
