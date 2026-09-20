"""DEPRECATED: use scripts/apply_champion_tag_fixes.py + data/champion_tag_fixes.json.

Engine loads data/champion_tags.json. Bootstrap merges champion_tag_fixes.json last.
This script's champion_tag_overrides.json is unused by the Go loader.
"""
from __future__ import annotations

import json
from pathlib import Path

ROOT = Path(r"D:\lol-build-overlay")

# Explicit curated profiles. Tags exclude primary (stored separately).
# Conventions:
# - exactly one of melee/ranged
# - exactly one primary damage: ap | ad | hybrid (true optional extra)
# - style/threat only when kit warrants it

P: dict[str, dict] = {}


def add(cid: str, primary: str, *tags: str) -> None:
    P[cid] = {"primary": primary, "tags": list(tags)}


# --- SUPPORT ---
add("Soraka", "support", "ap", "ranged", "healer", "peel")
add("Sona", "support", "ap", "ranged", "healer", "shield", "poke")
add("Yuumi", "support", "ap", "ranged", "healer", "shield", "mobility")
add("Rakan", "support", "ap", "melee", "engage", "peel", "shield", "mobility", "cc_hard")
add("Janna", "support", "ap", "ranged", "peel", "shield", "cc_hard")
add("Lulu", "support", "ap", "ranged", "peel", "shield", "poke")
add("Nami", "support", "ap", "ranged", "healer", "peel", "cc_hard", "poke")
add("Milio", "support", "ap", "ranged", "healer", "shield", "peel")
add("Karma", "support", "ap", "ranged", "shield", "poke", "peel")
add("Bard", "support", "ap", "ranged", "engage", "peel", "cc_hard", "mobility", "pick")
add("Zilean", "support", "ap", "ranged", "peel", "cc_hard", "poke")
add("Taric", "support", "ad", "melee", "healer", "shield", "peel", "cc_hard", "engage")
add("Thresh", "support", "ap", "ranged", "engage", "peel", "cc_hard", "pick")
add("Blitzcrank", "support", "ad", "melee", "engage", "cc_hard", "pick")
add("Leona", "support", "ad", "melee", "engage", "cc_hard", "tank")
add("Nautilus", "support", "ap", "melee", "engage", "cc_hard", "tank", "pick")
add("Alistar", "support", "ad", "melee", "engage", "cc_hard", "peel", "tank", "sustain")
add("Braum", "support", "ad", "melee", "peel", "cc_hard", "engage", "tank", "shield")
add("Rell", "support", "ad", "melee", "engage", "cc_hard", "tank", "peel")
add("Pyke", "support", "ad", "melee", "assassin", "burst", "dive", "mobility", "pick", "cc_hard")
add("Senna", "support", "ad", "ranged", "dps", "poke", "healer")
add("Seraphine", "support", "ap", "ranged", "poke", "shield", "cc_hard")
add("Renata", "support", "ap", "ranged", "peel", "shield", "cc_hard", "poke")
add("Ivern", "support", "ap", "ranged", "shield", "peel", "cc_hard")
add("Morgana", "support", "ap", "ranged", "peel", "shield", "cc_hard", "poke")

# --- MARKSMAN ---
add("Aphelios", "marksman", "ad", "ranged", "dps")
add("Ashe", "marksman", "ad", "ranged", "dps", "poke", "cc_hard", "pick")
add("Caitlyn", "marksman", "ad", "ranged", "dps", "poke", "pick")
add("Draven", "marksman", "ad", "ranged", "dps", "burst")
add("Ezreal", "marksman", "ad", "ranged", "poke", "mobility", "dps")
add("Jhin", "marksman", "ad", "ranged", "burst", "poke", "pick")
add("Jinx", "marksman", "ad", "ranged", "dps", "splitpush")
add("Kaisa", "marksman", "hybrid", "ranged", "dps", "burst", "mobility", "dive")
add("Kalista", "marksman", "ad", "ranged", "dps", "mobility", "engage")
add("KogMaw", "marksman", "ad", "ranged", "dps", "poke")
add("Lucian", "marksman", "ad", "ranged", "dps", "burst", "mobility")
add("MissFortune", "marksman", "ad", "ranged", "dps", "poke")
add("Nilah", "marksman", "ad", "melee", "dps", "dive", "mobility", "sustain")
add("Quinn", "marksman", "ad", "ranged", "dps", "burst", "mobility", "pick", "splitpush")
add("Samira", "marksman", "ad", "ranged", "dps", "burst", "dive", "mobility", "all_in")
add("Sivir", "marksman", "ad", "ranged", "dps", "shield", "mobility")
add("Smolder", "marksman", "ad", "ranged", "dps", "poke")
add("Tristana", "marksman", "ad", "ranged", "dps", "burst", "mobility", "dive")
add("Twitch", "marksman", "ad", "ranged", "dps", "burst", "pick")
add("Varus", "marksman", "ad", "ranged", "dps", "poke", "cc_hard", "pick")
add("Vayne", "marksman", "ad", "ranged", "dps", "true", "mobility", "dive", "splitpush")
add("Xayah", "marksman", "ad", "ranged", "dps", "peel")
add("Zeri", "marksman", "ad", "ranged", "dps", "mobility")
add("Yunara", "marksman", "ad", "ranged", "dps")
add("Corki", "marksman", "hybrid", "ranged", "dps", "poke", "mobility")
add("Graves", "marksman", "ad", "ranged", "burst", "dps", "mobility", "splitpush")
add("Kindred", "marksman", "ad", "ranged", "dps", "mobility", "dive")
add("Teemo", "marksman", "ap", "ranged", "poke", "splitpush")

# --- ASSASSIN ---
add("Akali", "assassin", "ap", "melee", "burst", "dive", "mobility", "pick")
add("Ekko", "assassin", "ap", "melee", "burst", "dive", "mobility", "pick")
add("Evelynn", "assassin", "ap", "melee", "burst", "dive", "mobility", "pick")
add("Fizz", "assassin", "ap", "melee", "burst", "dive", "mobility", "pick")
add("Kassadin", "assassin", "ap", "melee", "burst", "dive", "mobility", "pick")
add("Katarina", "assassin", "hybrid", "melee", "burst", "dive", "mobility")
add("Leblanc", "assassin", "ap", "ranged", "burst", "dive", "mobility", "pick")
add("Shaco", "assassin", "ad", "melee", "burst", "dive", "mobility", "pick", "splitpush")
add("Talon", "assassin", "ad", "melee", "burst", "dive", "mobility", "pick")
add("Zed", "assassin", "ad", "melee", "burst", "dive", "mobility", "pick")
add("Khazix", "assassin", "ad", "melee", "burst", "dive", "mobility", "pick")
add("Rengar", "assassin", "ad", "melee", "burst", "dive", "mobility", "pick", "all_in")
add("Qiyana", "assassin", "ad", "melee", "burst", "dive", "mobility", "pick", "cc_hard")
add("Naafiri", "assassin", "ad", "melee", "burst", "dive", "mobility", "pick")
add("Nocturne", "assassin", "ad", "melee", "burst", "dive", "mobility", "pick")
add("Akshan", "assassin", "ad", "ranged", "burst", "dive", "mobility", "pick", "dps")
add("Elise", "assassin", "ap", "ranged", "burst", "dive", "mobility", "pick")
add("Nidalee", "assassin", "ap", "ranged", "poke", "burst", "dive", "mobility", "pick")
add("Diana", "assassin", "ap", "melee", "burst", "dive", "mobility", "engage")
add("Sylas", "assassin", "ap", "melee", "burst", "dive", "mobility", "sustain")
add("Locke", "assassin", "ad", "melee", "burst", "dive", "mobility", "pick")

# --- FIGHTER (bruisers / skirmishers / juggernauts) ---
add("Aatrox", "fighter", "ad", "melee", "engage", "sustain", "juggernaut", "all_in")
add("Ambessa", "fighter", "ad", "melee", "engage", "burst", "dive", "mobility")
add("Belveth", "fighter", "ad", "melee", "dps", "dive", "mobility", "sustain", "splitpush")
add("Briar", "fighter", "ad", "melee", "dive", "all_in", "sustain", "mobility")
add("Camille", "fighter", "ad", "melee", "dive", "mobility", "pick", "true", "splitpush")
add("Darius", "fighter", "ad", "melee", "juggernaut", "engage", "sustain", "all_in", "cc_hard")
add("Fiora", "fighter", "ad", "melee", "dps", "true", "mobility", "splitpush", "sustain")
add("Garen", "fighter", "ad", "melee", "juggernaut", "engage", "sustain", "all_in")
add("Gwen", "fighter", "ap", "melee", "dps", "true", "dive", "mobility", "splitpush")
add("Illaoi", "fighter", "ad", "melee", "juggernaut", "sustain", "all_in", "splitpush")
add("Irelia", "fighter", "ad", "melee", "dive", "mobility", "dps", "sustain")
add("Jax", "fighter", "hybrid", "melee", "dps", "dive", "mobility", "splitpush", "sustain")
add("Kayn", "fighter", "ad", "melee", "dive", "mobility", "burst", "pick")  # form varies
add("Kled", "fighter", "ad", "melee", "engage", "all_in", "mobility", "sustain")
add("LeeSin", "fighter", "ad", "melee", "engage", "dive", "mobility", "pick", "peel")
add("MasterYi", "fighter", "ad", "melee", "dps", "dive", "mobility", "true", "splitpush")
add("Mordekaiser", "fighter", "ap", "melee", "juggernaut", "all_in", "sustain")
add("Nasus", "fighter", "ad", "melee", "juggernaut", "sustain", "splitpush", "cc_hard")
add("Olaf", "fighter", "ad", "melee", "engage", "sustain", "all_in", "dps")
add("Pantheon", "fighter", "ad", "melee", "engage", "burst", "dive", "mobility", "pick")
add("Renekton", "fighter", "ad", "melee", "engage", "sustain", "burst", "mobility")
add("Riven", "fighter", "ad", "melee", "burst", "dive", "mobility", "engage", "all_in")
add("Sett", "fighter", "ad", "melee", "juggernaut", "all_in", "engage", "sustain")
add("Trundle", "fighter", "ad", "melee", "sustain", "dps", "splitpush", "engage")
add("Tryndamere", "fighter", "ad", "melee", "dps", "splitpush", "sustain", "mobility", "all_in")
add("Udyr", "fighter", "ad", "melee", "engage", "sustain", "dps", "mobility")
add("Urgot", "fighter", "ad", "ranged", "juggernaut", "engage", "cc_hard", "all_in")
add("Vi", "fighter", "ad", "melee", "engage", "dive", "mobility", "cc_hard", "burst")
add("Viego", "fighter", "ad", "melee", "dive", "mobility", "dps", "sustain")
add("Volibear", "fighter", "ad", "melee", "engage", "sustain", "cc_hard", "dive")
add("Warwick", "fighter", "ad", "melee", "dive", "sustain", "mobility", "all_in")
add("XinZhao", "fighter", "ad", "melee", "engage", "dive", "sustain", "all_in")
add("Yasuo", "fighter", "ad", "melee", "dps", "dive", "mobility", "engage", "all_in")
add("Yone", "fighter", "ad", "melee", "dps", "dive", "mobility", "engage", "all_in")
add("Yorick", "fighter", "ad", "melee", "juggernaut", "splitpush", "sustain")
add("Gangplank", "fighter", "ad", "melee", "poke", "burst", "splitpush", "sustain")
add("Gnar", "fighter", "ad", "ranged", "engage", "cc_hard", "poke", "mobility")
add("Hecarim", "fighter", "ad", "melee", "engage", "dive", "mobility", "sustain")
add("JarvanIV", "fighter", "ad", "melee", "engage", "dive", "mobility", "cc_hard")
add("RekSai", "fighter", "ad", "melee", "engage", "dive", "mobility", "pick")
add("Rumble", "fighter", "ap", "melee", "poke", "burst", "engage", "sustain")
add("Shyvana", "fighter", "hybrid", "melee", "dps", "dive", "sustain", "splitpush")
add("MonkeyKing", "fighter", "ad", "melee", "engage", "dive", "mobility", "burst")
add("Zaahen", "fighter", "ad", "melee", "engage", "juggernaut", "sustain")
add("Lillia", "fighter", "ap", "ranged", "dps", "cc_hard", "mobility", "engage")
add("Jayce", "fighter", "ad", "melee", "poke", "burst", "mobility", "all_in")
add("Gragas", "fighter", "ap", "melee", "engage", "cc_hard", "mobility", "peel")
add("Kayle", "marksman", "hybrid", "melee", "dps", "sustain")

# --- TANK ---
add("Amumu", "tank", "ap", "melee", "engage", "cc_hard", "peel", "sustain")
add("Chogath", "tank", "ap", "melee", "cc_hard", "sustain", "true", "poke")
add("DrMundo", "tank", "ad", "melee", "sustain", "juggernaut", "splitpush")
add("Galio", "tank", "ap", "melee", "engage", "cc_hard", "peel", "mobility")
add("Malphite", "tank", "ap", "melee", "engage", "cc_hard", "peel")
add("Maokai", "tank", "ap", "melee", "engage", "cc_hard", "peel", "sustain")
add("Nunu", "tank", "ap", "melee", "engage", "cc_hard", "sustain", "mobility")
add("Ornn", "tank", "ad", "melee", "engage", "cc_hard", "peel", "sustain")
add("Poppy", "tank", "ad", "melee", "fighter", "engage", "anti_dash", "cc_hard")
add("Rammus", "tank", "ad", "melee", "engage", "cc_hard", "mobility", "peel")
add("Sejuani", "tank", "ap", "melee", "engage", "cc_hard", "peel", "sustain")
add("Shen", "tank", "ad", "melee", "peel", "engage", "shield", "sustain", "splitpush")
add("Singed", "tank", "ap", "melee", "poke", "cc_hard", "sustain", "splitpush")
add("Sion", "tank", "ad", "melee", "engage", "cc_hard", "sustain", "juggernaut")
add("Skarner", "tank", "ad", "melee", "engage", "cc_hard", "pick", "sustain")
add("TahmKench", "tank", "ap", "melee", "support", "engage", "peel", "sustain", "cc_hard")
add("Zac", "tank", "ap", "melee", "engage", "cc_hard", "sustain", "mobility")
add("KSante", "tank", "ad", "melee", "engage", "cc_hard", "peel", "all_in", "sustain")

# --- MAGE ---
add("Ahri", "mage", "ap", "ranged", "burst", "mobility", "pick")
add("Anivia", "mage", "ap", "ranged", "poke", "cc_hard", "burst")
add("Annie", "mage", "ap", "ranged", "burst", "cc_hard", "pick")
add("AurelionSol", "mage", "ap", "ranged", "dps", "poke", "cc_hard")
add("Aurora", "mage", "ap", "ranged", "burst", "mobility", "pick", "dive")
add("Azir", "mage", "ap", "ranged", "dps", "poke", "engage")
add("Brand", "mage", "ap", "ranged", "poke", "burst")
add("Cassiopeia", "mage", "ap", "ranged", "dps", "cc_hard", "poke")
add("Fiddlesticks", "mage", "ap", "ranged", "burst", "cc_hard", "pick", "engage")
add("Heimerdinger", "mage", "ap", "ranged", "poke", "burst")
add("Hwei", "mage", "ap", "ranged", "poke", "burst", "cc_hard")
add("Karthus", "mage", "ap", "ranged", "dps", "poke")
add("Lux", "mage", "ap", "ranged", "poke", "burst", "cc_hard", "pick", "shield")
add("Malzahar", "mage", "ap", "ranged", "poke", "cc_hard", "pick")
add("Neeko", "mage", "ap", "ranged", "burst", "cc_hard", "pick", "engage")
add("Orianna", "mage", "ap", "ranged", "poke", "burst", "shield", "engage")
add("Ryze", "mage", "ap", "ranged", "dps", "poke", "mobility")
add("Swain", "mage", "ap", "ranged", "dps", "sustain", "cc_hard", "engage")
add("Syndra", "mage", "ap", "ranged", "burst", "poke", "pick", "cc_hard")
add("Taliyah", "mage", "ap", "ranged", "poke", "burst", "mobility", "cc_hard")
add("TwistedFate", "mage", "ap", "ranged", "poke", "burst", "pick", "mobility", "cc_hard")
add("Veigar", "mage", "ap", "ranged", "burst", "cc_hard", "pick")
add("Velkoz", "mage", "ap", "ranged", "poke", "true", "burst")
add("Vex", "mage", "ap", "ranged", "burst", "pick", "anti_dash", "mobility")
add("Viktor", "mage", "ap", "ranged", "poke", "dps", "burst")
add("Vladimir", "mage", "ap", "ranged", "sustain", "burst", "dps")
add("Xerath", "mage", "ap", "ranged", "poke", "burst", "pick")
add("Ziggs", "mage", "ap", "ranged", "poke", "burst", "splitpush")
add("Zoe", "mage", "ap", "ranged", "burst", "pick", "poke", "mobility")
add("Zyra", "mage", "ap", "ranged", "poke", "cc_hard", "burst")
add("Mel", "mage", "ap", "ranged", "poke", "burst")
add("Lissandra", "mage", "ap", "ranged", "burst", "cc_hard", "engage", "peel")
add("Kennen", "mage", "ap", "ranged", "burst", "engage", "cc_hard", "mobility")

def main() -> None:
    with open(ROOT / "data" / "champions_by_id.json", encoding="utf-8") as f:
        champs = json.load(f)

    missing = sorted(set(champs) - set(P))
    extra = sorted(set(P) - set(champs))
    print("defined", len(P))
    if missing:
        raise SystemExit(f"missing curated profiles: {missing}")
    if extra:
        raise SystemExit(f"unknown champion ids: {extra}")

    payload = {
        "source": "curated",
        "count": len(P),
        "champions": {k: P[k] for k in sorted(P)},
    }
    out = ROOT / "data" / "champion_tag_overrides.json"
    out.write_text(json.dumps(payload, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    print("wrote", out, "count", len(P))


if __name__ == "__main__":
    main()
