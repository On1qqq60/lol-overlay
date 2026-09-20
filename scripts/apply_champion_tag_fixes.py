"""Apply curated champion_tags.json fixes (class + style + by_role).

Run: python scripts/apply_champion_tag_fixes.py
"""
from __future__ import annotations

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
PATH = ROOT / "data" / "champion_tags.json"

# role override: primary + tags
Role = dict  # {"primary": str, "tags": list[str]}


def P(primary: str, *tags: str, by_role: dict[str, Role] | None = None) -> dict:
    out: dict = {"primary": primary, "tags": sorted(set(tags))}
    if by_role:
        out["by_role"] = {
            pos: {"primary": ov["primary"], "tags": sorted(set(ov["tags"]))}
            for pos, ov in by_role.items()
        }
    return out


def role(primary: str, *tags: str) -> Role:
    return {"primary": primary, "tags": list(tags)}


# Full replacement profiles for every champ we discussed (phase 1 + 2).
FIXES: dict[str, dict] = {
    # --- Batch A: fighters wrongly tank ---
    "Gnar": P("fighter", "ad", "ranged", "poke", "engage", "cc_hard", "mobility"),
    "MonkeyKing": P("fighter", "ad", "melee", "engage", "dive", "mobility", "burst", "cc_hard"),
    "Olaf": P("fighter", "ad", "melee", "engage", "sustain", "all_in", "dps"),
    "Renekton": P("fighter", "ad", "melee", "engage", "sustain", "burst", "mobility"),
    "Shyvana": P("fighter", "hybrid", "melee", "dps", "dive", "sustain", "splitpush"),
    "Trundle": P("fighter", "ad", "melee", "sustain", "dps", "splitpush", "engage"),
    "Udyr": P("fighter", "ad", "melee", "engage", "sustain", "dps", "mobility"),
    "Urgot": P("fighter", "ad", "ranged", "juggernaut", "engage", "cc_hard", "all_in"),
    "Volibear": P("fighter", "ad", "melee", "engage", "sustain", "cc_hard", "dive"),
    "Warwick": P("fighter", "ad", "melee", "dive", "sustain", "mobility", "all_in"),
    "XinZhao": P("fighter", "ad", "melee", "engage", "dive", "sustain", "all_in"),
    "Yorick": P("fighter", "ad", "melee", "juggernaut", "splitpush", "sustain"),
    # keep true tanks with better tags
    "DrMundo": P("tank", "ad", "melee", "juggernaut", "sustain", "splitpush"),
    "Rammus": P("tank", "ad", "melee", "engage", "cc_hard", "mobility", "peel"),
    "Skarner": P("tank", "ad", "melee", "engage", "cc_hard", "pick", "sustain"),
    "Sion": P("tank", "ad", "melee", "engage", "cc_hard", "sustain", "juggernaut"),
    "KSante": P(
        "tank",
        "ad",
        "melee",
        "engage",
        "cc_hard",
        "peel",
        "all_in",
        "sustain",
    ),
    "Poppy": P(
        "tank",
        "ad",
        "melee",
        "fighter",
        "engage",
        "anti_dash",
        "cc_hard",
        by_role={
            "JUNGLE": role("tank", "ad", "melee", "engage", "dive", "anti_dash", "cc_hard"),
            "TOP": role("tank", "ad", "melee", "fighter", "engage", "anti_dash", "cc_hard", "sustain"),
            "UTILITY": role("tank", "ad", "melee", "engage", "peel", "anti_dash", "cc_hard"),
        },
    ),
    "Zac": P("tank", "ap", "melee", "engage", "dive", "cc_hard", "sustain", "mobility"),
    # --- Batch B: ADC class ---
    "Corki": P("marksman", "hybrid", "ranged", "dps", "poke", "mobility"),
    "KogMaw": P("marksman", "ad", "ranged", "dps", "poke"),
    "MissFortune": P("marksman", "ad", "ranged", "dps", "poke"),
    "Smolder": P("marksman", "ad", "ranged", "dps", "poke"),
    "Varus": P("marksman", "ad", "ranged", "dps", "poke", "cc_hard", "pick"),
    "Tristana": P("marksman", "ad", "ranged", "dps", "burst", "mobility", "dive"),
    "Quinn": P(
        "marksman",
        "ad",
        "ranged",
        "dps",
        "burst",
        "mobility",
        "pick",
        "splitpush",
        by_role={
            "TOP": role("marksman", "ad", "ranged", "dps", "burst", "mobility", "pick", "splitpush"),
            "BOTTOM": role("marksman", "ad", "ranged", "dps", "burst", "mobility", "pick"),
        },
    ),
    "Akshan": P(
        "marksman",
        "ad",
        "ranged",
        "dps",
        "burst",
        "mobility",
        "pick",
        by_role={
            "BOTTOM": role("marksman", "ad", "ranged", "dps", "burst", "mobility", "pick"),
            "MIDDLE": role("assassin", "ad", "ranged", "burst", "dive", "mobility", "pick"),
        },
    ),
    "Nilah": P("marksman", "ad", "melee", "dps", "dive", "mobility", "sustain"),
    "Kayle": P("marksman", "hybrid", "ranged", "dps", "sustain"),
    "Senna": P(
        "marksman",
        "ad",
        "ranged",
        "dps",
        "poke",
        "healer",
        by_role={
            "BOTTOM": role("marksman", "ad", "ranged", "dps", "poke", "healer"),
            "UTILITY": role("support", "ad", "ranged", "poke", "healer", "peel"),
        },
    ),
    "Teemo": P("mage", "ap", "ranged", "poke", "splitpush"),
    # --- Batch C ---
    "Galio": P(
        "tank",
        "ap",
        "melee",
        "engage",
        "cc_hard",
        "peel",
        "mobility",
        by_role={
            "MIDDLE": role("tank", "ap", "melee", "engage", "cc_hard", "peel", "burst"),
            "TOP": role("tank", "ap", "melee", "engage", "cc_hard", "peel"),
            "JUNGLE": role("tank", "ap", "melee", "engage", "dive", "cc_hard"),
        },
    ),
    "Nunu": P("tank", "ap", "melee", "engage", "cc_hard", "sustain", "mobility"),
    "Singed": P("tank", "ap", "melee", "poke", "cc_hard", "sustain", "splitpush"),
    "Ivern": P(
        "support",
        "ap",
        "ranged",
        "shield",
        "peel",
        "cc_hard",
        by_role={
            "JUNGLE": role("support", "ap", "ranged", "shield", "peel", "cc_hard"),
            "UTILITY": role("support", "ap", "ranged", "shield", "peel", "cc_hard"),
        },
    ),
    "Zilean": P(
        "support",
        "ap",
        "ranged",
        "peel",
        "cc_hard",
        "poke",
        by_role={
            "UTILITY": role("support", "ap", "ranged", "peel", "cc_hard", "poke"),
            "MIDDLE": role("mage", "ap", "ranged", "poke", "burst", "cc_hard"),
        },
    ),
    "Rumble": P(
        "fighter",
        "ap",
        "melee",
        "poke",
        "burst",
        "engage",
        "sustain",
        by_role={
            "TOP": role("fighter", "ap", "melee", "poke", "burst", "engage", "sustain"),
            "MIDDLE": role("fighter", "ap", "melee", "poke", "burst", "engage"),
        },
    ),
    "Lillia": P(
        "fighter",
        "ap",
        "ranged",
        "dps",
        "cc_hard",
        "mobility",
        "engage",
        by_role={
            "JUNGLE": role("fighter", "ap", "ranged", "dps", "cc_hard", "mobility", "engage"),
            "TOP": role("fighter", "ap", "ranged", "dps", "cc_hard", "sustain"),
            "MIDDLE": role("mage", "ap", "ranged", "dps", "cc_hard", "poke"),
        },
    ),
    # --- Batch D: assassin → fighter / mage ---
    "Diana": P(
        "fighter",
        "ap",
        "melee",
        "burst",
        "dive",
        "engage",
        "cc_hard",
        by_role={
            "JUNGLE": role("assassin", "ap", "melee", "burst", "dive", "mobility"),
            "MIDDLE": role("fighter", "ap", "melee", "burst", "dive", "engage", "cc_hard"),
        },
    ),
    "Ambessa": P("fighter", "ad", "melee", "burst", "dive", "engage", "mobility"),
    "Briar": P("fighter", "ad", "melee", "dive", "all_in", "sustain", "mobility"),
    "MasterYi": P("fighter", "ad", "melee", "dps", "dive", "mobility", "true", "splitpush"),
    "Tryndamere": P("fighter", "ad", "melee", "dps", "splitpush", "sustain", "mobility", "all_in"),
    "Viego": P("fighter", "ad", "melee", "dive", "mobility", "dps", "sustain"),
    "Nocturne": P("fighter", "ad", "melee", "burst", "dive", "mobility", "pick"),
    "Kayn": P("fighter", "ad", "melee", "dive", "mobility", "burst", "pick"),
    "Aurora": P("mage", "ap", "ranged", "burst", "mobility", "pick", "dive"),
    "Pyke": P(
        "assassin",
        "ad",
        "melee",
        "engage",
        "dive",
        "pick",
        "cc_hard",
        by_role={
            "UTILITY": role("support", "ad", "melee", "assassin", "engage", "dive", "pick", "cc_hard"),
            "MIDDLE": role("assassin", "ad", "melee", "burst", "dive", "mobility", "pick"),
        },
    ),
    "Jayce": P(
        "fighter",
        "ad",
        "melee",
        "poke",
        "burst",
        "mobility",
        "all_in",
        by_role={
            "TOP": role("fighter", "ad", "melee", "poke", "burst", "all_in", "splitpush"),
            "MIDDLE": role("fighter", "ad", "ranged", "poke", "burst", "mobility"),
        },
    ),
    "Elise": P("assassin", "ap", "ranged", "burst", "dive", "mobility", "pick"),
    "Rell": P("support", "ad", "melee", "tank", "engage", "cc_hard", "peel"),
    # --- Batch E: damage / shield on tanks & supports ---
    "Alistar": P("support", "ad", "tank", "melee", "engage", "peel", "cc_hard", "sustain"),
    "Blitzcrank": P("support", "ad", "tank", "melee", "engage", "cc_hard", "pick"),
    "Braum": P("support", "ad", "tank", "melee", "peel", "shield", "cc_hard", "engage"),
    "Leona": P("support", "ad", "tank", "melee", "engage", "cc_hard", "dive"),
    "Nautilus": P("support", "ap", "tank", "melee", "engage", "cc_hard", "dive", "pick"),
    "TahmKench": P(
        "tank",
        "ap",
        "melee",
        "support",
        "engage",
        "peel",
        "sustain",
        "cc_hard",
    ),
    "Thresh": P("support", "ad", "ranged", "engage", "peel", "cc_hard", "pick", "shield"),
    "Lux": P("mage", "ap", "ranged", "poke", "burst", "cc_hard", "pick", "shield"),
    "Shen": P("tank", "ad", "melee", "peel", "engage", "shield", "sustain", "splitpush"),
    "Kaisa": P("marksman", "hybrid", "ranged", "dps", "burst", "mobility", "dive"),
    "Katarina": P("assassin", "hybrid", "melee", "burst", "dive", "mobility"),
    "Hwei": P("mage", "ap", "ranged", "poke", "burst", "cc_hard"),
    "Nami": P("support", "ap", "ranged", "healer", "peel", "cc_hard", "poke"),
    # --- Batch F: style tags on existing ok primaries ---
    "Ezreal": P("marksman", "ad", "ranged", "dps", "poke", "mobility"),
    "Samira": P("marksman", "ad", "ranged", "dps", "burst", "dive", "mobility", "all_in"),
    "Kindred": P("marksman", "ad", "ranged", "dps", "mobility", "dive"),
    "Vayne": P("marksman", "ad", "ranged", "dps", "true", "mobility", "dive", "splitpush"),
    "Riven": P("fighter", "ad", "melee", "burst", "dive", "mobility", "engage", "all_in"),
    "Kled": P("fighter", "ad", "melee", "engage", "all_in", "mobility", "sustain"),
    "Rakan": P("support", "ap", "melee", "engage", "peel", "shield", "cc_hard", "mobility", "dive"),
    "Leblanc": P("assassin", "ap", "ranged", "burst", "dive", "mobility", "pick"),
    "Gangplank": P("fighter", "ad", "melee", "poke", "burst", "splitpush", "sustain"),
    # --- Phase 2: break mage ap/burst/ranged clones ---
    "Anivia": P("mage", "ap", "ranged", "poke", "cc_hard", "burst"),
    "AurelionSol": P("mage", "ap", "ranged", "dps", "poke", "cc_hard"),
    "Azir": P("mage", "ap", "ranged", "dps", "poke", "engage"),
    "Cassiopeia": P("mage", "ap", "ranged", "dps", "cc_hard", "poke"),
    "Fiddlesticks": P("mage", "ap", "ranged", "burst", "cc_hard", "pick", "engage"),
    "Karthus": P("mage", "ap", "ranged", "dps", "poke"),
    "Lissandra": P("mage", "ap", "ranged", "burst", "cc_hard", "engage", "peel"),
    "Malzahar": P("mage", "ap", "ranged", "poke", "cc_hard", "pick"),
    "Ryze": P("mage", "ap", "ranged", "dps", "poke", "mobility"),
    "TwistedFate": P("mage", "ap", "ranged", "poke", "burst", "pick", "mobility", "cc_hard"),
    "Velkoz": P("mage", "ap", "ranged", "poke", "true", "burst"),
    "Vladimir": P("mage", "ap", "ranged", "sustain", "burst", "dps"),
    # other mages that were thin
    "Heimerdinger": P("mage", "ap", "ranged", "poke", "burst"),
    "Kennen": P("mage", "ap", "ranged", "burst", "engage", "cc_hard", "mobility"),
    "Mel": P("mage", "ap", "ranged", "poke", "burst"),
    "Taliyah": P("mage", "ap", "ranged", "poke", "burst", "mobility", "cc_hard"),
    "Vex": P("mage", "ap", "ranged", "burst", "pick", "anti_dash", "mobility"),
    "Xerath": P("mage", "ap", "ranged", "poke", "burst", "pick"),
    "Zoe": P("mage", "ap", "ranged", "burst", "pick", "poke", "mobility"),
    "Brand": P("mage", "ap", "ranged", "poke", "burst"),
    "Annie": P("mage", "ap", "ranged", "burst", "cc_hard", "pick"),
    "Zyra": P("mage", "ap", "ranged", "poke", "cc_hard", "burst"),
    "Orianna": P("mage", "ap", "ranged", "poke", "burst", "shield", "engage"),
    "Syndra": P("mage", "ap", "ranged", "burst", "poke", "pick", "cc_hard"),
    "Viktor": P("mage", "ap", "ranged", "poke", "dps", "burst"),
    "Veigar": P("mage", "ap", "ranged", "burst", "cc_hard", "pick"),
    "Neeko": P(
        "mage",
        "ap",
        "ranged",
        "burst",
        "cc_hard",
        "pick",
        "engage",
        by_role={
            "MIDDLE": role("mage", "ap", "ranged", "burst", "cc_hard", "pick"),
            "TOP": role("mage", "ap", "ranged", "burst", "cc_hard", "engage"),
            "UTILITY": role("support", "ap", "ranged", "cc_hard", "pick", "peel"),
        },
    ),
    # --- Phase 2: marksman clone breakup ---
    "Aphelios": P("marksman", "ad", "ranged", "dps"),
    "Draven": P("marksman", "ad", "ranged", "dps", "burst"),
    "Graves": P("marksman", "ad", "ranged", "burst", "dps", "mobility", "splitpush"),
    "Jinx": P("marksman", "ad", "ranged", "dps", "splitpush"),
    "Kalista": P("marksman", "ad", "ranged", "dps", "mobility", "engage"),
    "Sivir": P("marksman", "ad", "ranged", "dps", "shield", "mobility"),
    "Xayah": P("marksman", "ad", "ranged", "dps", "peel"),
    "Yunara": P("marksman", "ad", "ranged", "dps"),
    "Zeri": P("marksman", "ad", "ranged", "dps", "mobility"),
    "Jhin": P("marksman", "ad", "ranged", "burst", "poke", "pick"),
    "Ashe": P("marksman", "ad", "ranged", "dps", "poke", "cc_hard", "pick"),
    "Caitlyn": P("marksman", "ad", "ranged", "dps", "poke", "pick"),
    "Lucian": P("marksman", "ad", "ranged", "dps", "burst", "mobility"),
    "Twitch": P("marksman", "ad", "ranged", "dps", "burst", "pick"),
    # --- keep / refresh existing flex by_role champs ---
    "Karma": P(
        "mage",
        "ap",
        "ranged",
        "shield",
        "poke",
        "peel",
        by_role={
            "MIDDLE": role("mage", "ap", "ranged", "poke", "burst", "shield"),
            "TOP": role("mage", "ap", "ranged", "poke", "shield"),
            "UTILITY": role("support", "ap", "ranged", "shield", "peel", "poke"),
        },
    ),
    "Seraphine": P(
        "mage",
        "ap",
        "ranged",
        "poke",
        "shield",
        by_role={
            "BOTTOM": role("mage", "ap", "ranged", "poke", "dps"),
            "MIDDLE": role("mage", "ap", "ranged", "poke", "burst"),
            "UTILITY": role("support", "ap", "ranged", "healer", "shield", "poke"),
        },
    ),
    "Swain": P(
        "mage",
        "ap",
        "ranged",
        "dps",
        "sustain",
        "cc_hard",
        "engage",
        by_role={
            "BOTTOM": role("mage", "ap", "ranged", "dps", "sustain", "poke"),
            "MIDDLE": role("mage", "ap", "ranged", "dps", "sustain", "cc_hard", "engage"),
            "UTILITY": role("support", "ap", "ranged", "cc_hard", "engage", "sustain"),
        },
    ),
    "Sylas": P(
        "mage",
        "ap",
        "melee",
        "burst",
        "dive",
        "mobility",
        "sustain",
        by_role={
            "JUNGLE": role("assassin", "ap", "melee", "burst", "dive"),
            "MIDDLE": role("mage", "ap", "melee", "burst", "dive", "mobility"),
            "TOP": role("fighter", "ap", "melee", "burst", "sustain"),
        },
    ),
    "Gragas": P(
        "fighter",
        "ap",
        "melee",
        "engage",
        "cc_hard",
        "mobility",
        "peel",
        by_role={
            "JUNGLE": role("fighter", "ap", "melee", "engage", "dive", "cc_hard"),
            "MIDDLE": role("mage", "ap", "melee", "burst", "cc_hard"),
            "TOP": role("fighter", "ap", "melee", "engage", "sustain", "cc_hard"),
            "UTILITY": role("support", "ap", "melee", "engage", "peel", "cc_hard"),
        },
    ),
    "Pantheon": P(
        "fighter",
        "ad",
        "melee",
        "engage",
        "all_in",
        by_role={
            "JUNGLE": role("fighter", "ad", "melee", "engage", "dive", "all_in"),
            "MIDDLE": role("assassin", "ad", "melee", "burst", "dive", "pick"),
            "TOP": role("fighter", "ad", "melee", "poke", "all_in", "splitpush"),
            "UTILITY": role("fighter", "ad", "melee", "engage", "peel", "cc_hard"),
        },
    ),
    "Sett": P(
        "fighter",
        "ad",
        "melee",
        "juggernaut",
        "all_in",
        "engage",
        "sustain",
        by_role={
            "MIDDLE": role("fighter", "ad", "melee", "all_in", "burst", "dive"),
            "TOP": role("fighter", "ad", "melee", "juggernaut", "all_in", "sustain"),
            "UTILITY": role("fighter", "ad", "melee", "engage", "peel", "cc_hard", "sustain"),
        },
    ),
    "Mordekaiser": P(
        "fighter",
        "ap",
        "melee",
        "juggernaut",
        "all_in",
        "sustain",
        by_role={
            "JUNGLE": role("fighter", "ap", "melee", "juggernaut", "all_in", "dive"),
            "MIDDLE": role("fighter", "ap", "melee", "juggernaut", "burst", "all_in"),
            "TOP": role("fighter", "ap", "melee", "juggernaut", "all_in", "sustain"),
        },
    ),
    # --- remaining tanks with cleaner tags ---
    "Amumu": P("tank", "ap", "melee", "engage", "cc_hard", "dive", "peel", "sustain"),
    "Malphite": P("tank", "ap", "melee", "engage", "cc_hard", "dive", "peel"),
    "Maokai": P("tank", "ap", "melee", "engage", "cc_hard", "peel", "sustain"),
    "Ornn": P("tank", "ad", "melee", "engage", "cc_hard", "peel", "sustain"),
    "Sejuani": P("tank", "ap", "melee", "engage", "cc_hard", "dive", "peel", "sustain"),
    "Chogath": P("tank", "ap", "melee", "cc_hard", "sustain", "true", "poke", "juggernaut"),
    # supports polish
    "Janna": P("support", "ap", "ranged", "peel", "shield", "cc_hard"),
    "Lulu": P("support", "ap", "ranged", "peel", "shield", "poke"),
    "Morgana": P("support", "ap", "ranged", "peel", "shield", "cc_hard", "poke"),
    "Renata": P("support", "ap", "ranged", "peel", "shield", "cc_hard", "poke"),
    "Bard": P("support", "ap", "ranged", "engage", "peel", "cc_hard", "mobility", "pick"),
    "Soraka": P("support", "ap", "ranged", "healer", "peel"),
    "Sona": P("support", "ap", "ranged", "healer", "shield", "poke"),
    "Yuumi": P("support", "ap", "ranged", "healer", "shield", "mobility"),
    "Milio": P("support", "ap", "ranged", "healer", "shield", "peel"),
    "Taric": P("support", "ad", "melee", "healer", "shield", "peel", "cc_hard", "engage"),
}


def validate_profile(cid: str, p: dict) -> list[str]:
    errs: list[str] = []
    tags = set(p["tags"])
    rng = tags & {"melee", "ranged"}
    if len(rng) != 1:
        errs.append(f"{cid}: range={rng}")
    pure = tags & {"ad", "ap", "hybrid"}
    # true alone doesn't count as damage identity
    if len(pure) > 1:
        errs.append(f"{cid}: multi damage {pure}")
    if len(pure) == 0 and p["primary"] not in ("tank", "support"):
        # still prefer damage; warn
        errs.append(f"{cid}: missing ad/ap/hybrid")
    for pos, ov in (p.get("by_role") or {}).items():
        if pos not in ("TOP", "MIDDLE", "JUNGLE", "BOTTOM", "UTILITY"):
            errs.append(f"{cid}: bad role key {pos}")
        ot = set(ov["tags"])
        if len(ot & {"melee", "ranged"}) != 1:
            errs.append(f"{cid}[{pos}]: range")
    return errs


def main() -> None:
    data = json.loads(PATH.read_text(encoding="utf-8"))
    champs: dict = data["champions"]
    missing = sorted(set(FIXES) - set(champs))
    if missing:
        raise SystemExit(f"unknown champions in FIXES: {missing}")

    for cid, profile in FIXES.items():
        champs[cid] = profile

    errs: list[str] = []
    for cid, p in sorted(champs.items()):
        errs.extend(validate_profile(cid, p))
    # soft: only fail hard range errors
    hard = [e for e in errs if "range" in e or "multi damage" in e or "bad role" in e]
    if hard:
        raise SystemExit("validation failed:\n" + "\n".join(hard))
    if errs:
        print("warnings:")
        for e in errs:
            print(" ", e)

    data["champions"] = {k: champs[k] for k in sorted(champs)}
    data["count"] = len(champs)
    data["source"] = "bootstrap+curated+fixes"
    PATH.write_text(json.dumps(data, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")

    fixes_path = ROOT / "data" / "champion_tag_fixes.json"
    fixes_payload = {
        "source": "curated-fixes",
        "count": len(FIXES),
        "champions": {k: FIXES[k] for k in sorted(FIXES)},
    }
    fixes_path.write_text(
        json.dumps(fixes_payload, indent=2, ensure_ascii=False) + "\n", encoding="utf-8"
    )
    print(f"updated {PATH} applied={len(FIXES)} total={len(champs)}")
    print(f"updated {fixes_path} (bootstrap overlay)")


if __name__ == "__main__":
    main()
