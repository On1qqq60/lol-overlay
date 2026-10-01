from __future__ import annotations

import json
import re
from pathlib import Path

LIVE = Path(__file__).resolve().parents[1] / "testdata" / "generated" / "live"
text = (LIVE / "recommend_report.md").read_text(encoding="utf-8")
chunks = re.split(r"(?m)^## ", text)[1:]
out = []
for chunk in chunks:
    lines = chunk.splitlines()
    fname = lines[0].strip()
    body = "\n".join(lines[1:])
    am = re.search(r"- active: `([^`]+)` @ `([^`]+)`", body)
    al = re.search(r"- allies: `([^`]*)`", body)
    en = re.search(r"- enemies: `([^`]*)`", body)
    sm = re.search(
        r"Active: (\S+) \((\w+)\) seed=(\S+) gold=([0-9.]+) allyFront=([0-9.]+)",
        body,
    )
    pm = re.search(r"Pressure: ([^\n]+)", body)
    nm = re.search(
        r"Next item: (.+?) \((\d+)\) priority=([0-9.]+) \[([^\]]+)\]", body
    )
    build = []
    inb = False
    reasons = []
    inr = False
    for ln in body.splitlines():
        if ln.startswith("Build order"):
            inb, inr = True, False
            continue
        if ln.startswith("Reasons:"):
            inb, inr = False, True
            continue
        if inb:
            bm = re.match(r"^\s+(\d+)\s+(.+?)\s+\[([^\]]+)\]", ln)
            if bm:
                build.append(
                    {
                        "pri": int(bm.group(1)),
                        "name": bm.group(2).strip(),
                        "role": bm.group(3).strip(),
                    }
                )
            elif ln.strip() == "" or ln.startswith("```"):
                inb = False
        if inr and ln.strip().startswith("•"):
            reasons.append(ln.strip()[1:].strip())
    out.append(
        {
            "file": fname,
            "champ": am.group(1) if am else "?",
            "role": am.group(2) if am else "?",
            "allies": al.group(1) if al else "",
            "enemies": en.group(1) if en else "",
            "seed": sm.group(3) if sm else "",
            "allyFront": float(sm.group(5)) if sm else 0,
            "pressure": pm.group(1) if pm else "",
            "next": f"{nm.group(1)} [{nm.group(4)}]" if nm else "",
            "build": build,
            "reasons": reasons,
        }
    )

(LIVE / "_cases_compact.json").write_text(
    json.dumps(out, ensure_ascii=False, indent=1), encoding="utf-8"
)
txt = []
for i, c in enumerate(out, 1):
    b = " | ".join(f"{x['name']}[{x['role']}]" for x in c["build"])
    txt.append(
        f"{i:04d} {c['champ']:14s} {c['role']:8s} seed={c['seed']:20s} next={c['next']}"
    )
    txt.append(f"     ALL {c['allies']}")
    txt.append(f"     ENM {c['enemies']}")
    txt.append(f"     P {c['pressure']} front={c['allyFront']:.2f}")
    txt.append(f"     {b}")
    if c["reasons"]:
        txt.append("     R " + " // ".join(c["reasons"][:6]))
    txt.append("")
(LIVE / "_cases_review.txt").write_text("\n".join(txt), encoding="utf-8")
print(len(out), "lines", len(txt))
