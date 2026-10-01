import json
import subprocess
import sys

exe = r"D:\lol-build-overlay\overlay\recommend.exe"
root = r"D:\lol-build-overlay\core"
p = subprocess.run([exe, "-json", "-live", "-root", root], capture_output=True, text=True)
t = p.stdout
i = t.find("{")
if i < 0:
    print("stderr", p.stderr)
    print("stdout", t[:500])
    sys.exit(1)
d = json.loads(t[i:])
print("seed", d.get("seedName"), "gold", d.get("currentGold"), "time", round(d.get("gameTime", 0), 1))
print("next", d.get("nextItem"), "goal", d.get("nextGoalId"))
print("owned", d.get("ownedItems"))
print("--- build ---")
for s in d.get("build") or []:
    print(" ", s.get("role"), s.get("itemId"), s.get("name"), "pri", s.get("priority"))
print("--- threats ---")
for th in d.get("threats") or []:
    print(th.get("championId"), "lane", th.get("laneOpponent"), "w", round(th.get("weight") or 0, 3))
print("--- enemies ---")
for p in d.get("profiles") or []:
    pl = p.get("player") or {}
    print(pl.get("championId"), pl.get("position"), pl.get("items"))
