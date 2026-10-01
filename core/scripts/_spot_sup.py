import json
from pathlib import Path

d = json.loads(
    Path(r"D:\lol-build-overlay\core\testdata\generated\live\verdicts.json").read_text(
        encoding="utf-8"
    )
)

locket_next = moon_next = 0
er_senna = 0
blood = sleigh = zazzak = dream = celestial = 0
ardent = rylai = fh = 0
void_ench = ldr_tank = 0

print("=== starts (first cycle) ===")
for c in d["cases"]:
    if c["phase"] != "start" or int(c["file"][:4]) > 37:
        continue
    n = (c.get("nextItem") or {}).get("name")
    b = [x["name"] for x in (c.get("build") or [])]
    print(c["file"], "next=", n, b[:5])

print("=== next legendary counts at start ===")
for c in d["cases"]:
    if c["phase"] != "start":
        continue
    n = (c.get("nextItem") or {}).get("name") or ""
    if n == "Locket of the Iron Solari":
        locket_next += 1
    if n == "Moonstone Renewer":
        moon_next += 1
    joined = " | ".join(x["name"] for x in (c.get("build") or []))
    if "Bloodsong" in joined:
        blood += 1
    if "Solstice Sleigh" in joined:
        sleigh += 1
    if "Dream Maker" in joined:
        dream += 1
    if "Celestial Opposition" in joined:
        celestial += 1
    if "Zaz'Zak" in joined:
        zazzak += 1
    if "Ardent" in joined:
        ardent += 1
    if "Rylai" in joined:
        rylai += 1
    if "Frozen Heart" in joined:
        fh += 1

print("start next Locket", locket_next, "Moonstone", moon_next)
print("start blood", blood, "sleigh", sleigh, "dream", dream, "celestial", celestial, "zazzak", zazzak)
print("start ardent", ardent, "rylai", rylai, "frozen", fh)

print("=== Senna / Pyke / Braum / Brand mids ===")
for c in d["cases"]:
    if c["phase"] != "mid":
        continue
    if c["active"] not in ("Senna", "Pyke", "Braum", "Brand", "Lulu", "Yuumi", "Leona", "Pantheon"):
        continue
    n = (c.get("nextItem") or {}).get("name")
    b = [x["name"] for x in (c.get("build") or [])]
    print(c["file"], "next=", n, b)
    if c["active"] == "Senna" and any("Essence Reaver" in x for x in b + [n or ""]):
        er_senna += 1
    if c["active"] in ("Sona", "Yuumi", "Janna", "Lulu") and any("Void Staff" in x for x in b):
        void_ench += 1
    if c["active"] in ("Galio", "Renata") and any("Dominik" in x for x in b):
        ldr_tank += 1

print("senna ER leftover", er_senna, "void on enchanter mid", void_ench, "ldr tank", ldr_tank)

print("=== mid next unaffordable legendaries (gold<800) ===")
bad = 0
for c in d["cases"]:
    if c["phase"] != "mid":
        continue
    gold = c.get("gold") or 0
    n = c.get("nextItem") or {}
    name = n.get("name") or ""
    role = n.get("role") or ""
    if gold and gold < 800 and role != "component" and name in (
        "Locket of the Iron Solari",
        "Moonstone Renewer",
        "Blackfire Torch",
        "Youmuu's Ghostblade",
        "Imperial Mandate",
        "Eclipse",
    ):
        bad += 1
        if bad <= 8:
            print(" bad", c["file"], "gold", gold, "next", name)
print("unaffordable legendary next", bad)
