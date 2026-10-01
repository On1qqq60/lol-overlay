"""Playable mid-game inventories for gen_live_fixtures.py.

Item sequences follow core/internal/rules/templates.go seeds, then a time budget
so a 10-minute snapshot is not a 6-item random stash.
"""

from __future__ import annotations

import random
from typing import Any, Callable

# --- item ids (keep in sync with internal/rules) ---
DORANS_SHIELD = 1054
DORANS_BLADE = 1055
DORANS_RING = 1056
DARK_SEAL = 1082
POTION = 2003
CONTROL_WARD = 2055
WARD = 3340
SWEEPER = 3364
FARSIGHT = 3363
LOST_CHAPTER = 3802
MALIGNANCE = 3118
BLACKFIRE = 2503
ROD_OF_AGES = 6657
MERCS = 3111
SORCS = 3020
STEELCAPS = 3047
BERSERKERS = 3006
IONIAN = 3158
BANSHEE = 3102
ZHONYA = 3157
SHADOWFLAME = 4645
STORMSURGE = 4646
LIANDRY = 6653
RABADON = 3089
VOID_STAFF = 3135
MORELLO = 3165
SEEKERS = 2420
SERRATED = 3134
YOUMUU = 3142
INFINITY = 3031
KRAKEN = 6672
BOTRK = 3153
TRINITY = 3078
CLEAVER = 3071
SUNDERED = 6610
RANDUIN = 3143
STERAK = 3053
ICEBORN = 6662
SPIRIT = 3065
RYLAI = 3116
RIFTMAKER = 4633
COSMIC = 4629
RFC = 3094
COLLECTOR = 6676
LDR = 3036
FORCE_OF_NATURE = 4401
SERYLDA = 6694
AXIOM = 6696
EDGE_OF_NIGHT = 3814
SUNFIRE = 3068
THORNMAIL = 3075
MOONSTONE = 6617
REDEMPTION = 3107
MIKAELS = 3222
LOCKET = 3190
KNIGHTS_VOW = 3109
SCORCHCLAW = 1101
GUSTWALKER = 1102
MOSSTOMPER = 1103
ATLAS = 3865
COMPASS = 3866
BOUNTY = 3867
CELESTIAL = 3869
DREAMMAKER = 3870
ZAZZAK = 3871
BLOODSONG = 3877

Slot = tuple[int, str]
MakeItem = Callable[[int, int, int], dict[str, Any]]

MID_MINUTES_DEFAULT = (10, 12, 14, 15, 16, 17, 18, 20, 22, 25)


def seed_slots(cid: str, position: str, primary: str, tags: set[str]) -> list[Slot]:
    pos = position
    if pos == "UTILITY" and cid not in ("Senna", "Pyke"):
        return support_seed(primary, tags)
    if primary == "support" and pos != "UTILITY":
        slots = solo_lane_support_seed(tags)
        if pos == "JUNGLE":
            return apply_jungle_start(slots, "tank" if "tank" in tags else "mage", tags)
        return slots
    slots, _label = seed_by_class(primary, tags)
    if pos == "JUNGLE":
        slots = apply_jungle_start(slots, primary, tags)
    return slots


def seed_by_class(primary: str, tags: set[str]) -> tuple[list[Slot], str]:
    has = lambda t: t == primary or t in tags
    if primary == "mage":
        if has("poke") and not has("burst"):
            return seed_ap_poke(), "mage"
        return seed_ap_mid(), "mage"
    if primary == "assassin":
        if has("ap") or has("hybrid"):
            return seed_ap_assassin(), "assassin/ap"
        return seed_ad_assassin(), "assassin"
    if primary == "marksman":
        return seed_adc(), "marksman"
    if primary == "fighter":
        if has("ap"):
            if has("juggernaut"):
                return seed_ap_juggernaut(), "fighter/ap"
            return seed_ap_bruiser(), "fighter/ap"
        if has("juggernaut"):
            return seed_ad_juggernaut(), "fighter"
        return seed_ad_bruiser(), "fighter"
    if primary == "tank":
        if has("fighter"):
            return seed_tank_fighter(), "fighter"
        return seed_tank(), "tank"
    if primary == "support":
        if has("healer") or has("shield"):
            return seed_enchanter(), "support"
        return seed_engage_sup(), "support"
    return seed_ap_mid(), "mage"


def support_seed(primary: str, tags: set[str]) -> list[Slot]:
    has = lambda t: t == primary or t in tags
    if primary == "mage" or (has("ap") and has("poke") and not has("healer")):
        return seed_mage_support()
    if has("healer") or has("shield"):
        return seed_enchanter()
    if primary == "mage" or has("ap"):
        return seed_mage_support()
    return seed_engage_sup()


def solo_lane_support_seed(tags: set[str]) -> list[Slot]:
    if "tank" in tags or "engage" in tags or ("peel" in tags and "melee" in tags):
        return seed_tank()
    if "ap" in tags or "healer" in tags or "shield" in tags or "poke" in tags:
        if "poke" in tags and "burst" not in tags:
            return seed_ap_poke()
        return seed_ap_mid()
    return seed_tank()


def apply_jungle_start(slots: list[Slot], primary: str, tags: set[str]) -> list[Slot]:
    pet = jungle_pet_id(primary, tags)
    out = [(pet, "start")]
    for item_id, role in slots:
        if role == "start":
            continue
        out.append((item_id, role))
    return out


def jungle_pet_id(primary: str, tags: set[str]) -> int:
    tank = primary == "tank" or "tank" in tags or "juggernaut" in tags
    ap = primary in ("mage", "support") or (primary == "assassin" and "ap" in tags) or (
        primary == "fighter" and "ap" in tags
    )
    if tank:
        return MOSSTOMPER
    if ap:
        return SCORCHCLAW
    return GUSTWALKER


def seed_ap_mid() -> list[Slot]:
    return [
        (DORANS_RING, "start"),
        (LOST_CHAPTER, "component"),
        (MALIGNANCE, "core"),
        (SORCS, "boots"),
        (MERCS, "boots"),
        (SHADOWFLAME, "offensive"),
        (LIANDRY, "offensive"),
        (RABADON, "offensive"),
        (VOID_STAFF, "pen"),
    ]


def seed_ap_poke() -> list[Slot]:
    return [
        (DORANS_RING, "start"),
        (LOST_CHAPTER, "component"),
        (BLACKFIRE, "core"),
        (SORCS, "boots"),
        (LIANDRY, "offensive"),
        (RABADON, "offensive"),
        (VOID_STAFF, "pen"),
        (SHADOWFLAME, "offensive"),
    ]


def seed_ap_assassin() -> list[Slot]:
    return [
        (DARK_SEAL, "start"),
        (STORMSURGE, "core"),
        (SORCS, "boots"),
        (SHADOWFLAME, "offensive"),
        (RABADON, "offensive"),
        (VOID_STAFF, "pen"),
    ]


def seed_ad_assassin() -> list[Slot]:
    return [
        (DORANS_BLADE, "start"),
        (SERRATED, "component"),
        (YOUMUU, "core"),
        (IONIAN, "boots"),
        (STEELCAPS, "boots"),
        (EDGE_OF_NIGHT, "defensive"),
        (SERYLDA, "pen"),
        (COLLECTOR, "offensive"),
        (AXIOM, "offensive"),
    ]


def seed_adc() -> list[Slot]:
    return [
        (DORANS_BLADE, "start"),
        (BERSERKERS, "boots"),
        (KRAKEN, "core"),
        (INFINITY, "offensive"),
        (RFC, "offensive"),
        (COLLECTOR, "offensive"),
        (LDR, "offensive"),
        (BOTRK, "offensive"),
    ]


def seed_ad_bruiser() -> list[Slot]:
    return [
        (DORANS_BLADE, "start"),
        (TRINITY, "core"),
        (SUNDERED, "offensive"),
        (CLEAVER, "offensive"),
        (STEELCAPS, "boots"),
        (MERCS, "boots"),
    ]


def seed_ad_juggernaut() -> list[Slot]:
    return [
        (DORANS_BLADE, "start"),
        (CLEAVER, "core"),
        (SUNDERED, "offensive"),
        (STERAK, "defensive"),
        (ICEBORN, "defensive"),
        (MERCS, "boots"),
        (STEELCAPS, "boots"),
    ]


def seed_ap_bruiser() -> list[Slot]:
    return [
        (DORANS_RING, "start"),
        (ROD_OF_AGES, "core"),
        (RIFTMAKER, "offensive"),
        (MERCS, "boots"),
    ]


def seed_ap_juggernaut() -> list[Slot]:
    return [
        (DORANS_SHIELD, "start"),
        (RIFTMAKER, "core"),
        (RYLAI, "offensive"),
        (LIANDRY, "offensive"),
        (COSMIC, "offensive"),
        (STEELCAPS, "boots"),
        (MERCS, "boots"),
    ]


def seed_tank() -> list[Slot]:
    return [
        (DORANS_SHIELD, "start"),
        (STEELCAPS, "boots"),
        (MERCS, "boots"),
        (SUNFIRE, "core"),
        (RANDUIN, "defensive"),
        (THORNMAIL, "defensive"),
        (FORCE_OF_NATURE, "defensive"),
        (SPIRIT, "defensive"),
    ]


def seed_tank_fighter() -> list[Slot]:
    return [
        (DORANS_SHIELD, "start"),
        (ICEBORN, "core"),
        (SUNDERED, "offensive"),
        (CLEAVER, "offensive"),
        (STEELCAPS, "boots"),
        (MERCS, "boots"),
        (STERAK, "defensive"),
    ]


def seed_enchanter() -> list[Slot]:
    return [
        (ATLAS, "start"),
        (MOONSTONE, "core"),
        (REDEMPTION, "utility"),
        (MIKAELS, "utility"),
        (IONIAN, "boots"),
        (MORELLO, "utility"),
    ]


def seed_mage_support() -> list[Slot]:
    return [
        (ATLAS, "start"),
        (LOST_CHAPTER, "component"),
        (BLACKFIRE, "core"),
        (SORCS, "boots"),
        (LIANDRY, "offensive"),
        (MORELLO, "utility"),
        (VOID_STAFF, "pen"),
    ]


def seed_engage_sup() -> list[Slot]:
    return [
        (ATLAS, "start"),
        (LOCKET, "core"),
        (KNIGHTS_VOW, "utility"),
        (MERCS, "boots"),
        (THORNMAIL, "defensive"),
    ]


def support_quest_id(minute: float, slots: list[Slot]) -> int:
    roles = {item_id for item_id, _ in slots}
    if minute < 12:
        return COMPASS
    if minute < 16:
        return BOUNTY
    if MOONSTONE in roles:
        return DREAMMAKER
    if BLACKFIRE in roles or LIANDRY in roles:
        return ZAZZAK
    if LOCKET in roles:
        return CELESTIAL
    return BLOODSONG


def upgraded_pet(pet_id: int) -> int:
    if 1101 <= pet_id <= 1103:
        return pet_id + 3
    return pet_id


def phase_plan(minute: float, position: str) -> dict[str, Any]:
    """How many completed legendaries / boots a typical player has at `minute`."""
    m = float(minute)
    if position == "UTILITY":
        if m < 12:
            n_leg, gold, level, cs = 0, 450, 7, 12
        elif m < 16:
            n_leg, gold, level, cs = 1, 700, 9, 18
        elif m < 20:
            n_leg, gold, level, cs = 1, 1100, 10, 22
        elif m < 24:
            n_leg, gold, level, cs = 2, 800, 12, 28
        else:
            n_leg, gold, level, cs = 2, 1200, 13, 32
        return {
            "legendaries": n_leg,
            "boots": True,
            "keep_starter": False,
            "gold": gold,
            "level": level,
            "cs": cs,
        }
    if position == "JUNGLE":
        level = min(18, 3 + int(m * 0.55))
        cs = int(m * 5.2)
    elif position == "BOTTOM":
        level = min(18, 2 + int(m * 0.58))
        cs = int(m * 8.4)
    else:
        level = min(18, 2 + int(m * 0.62))
        cs = int(m * 7.1)

    if m < 12:
        n_leg, gold = 1, 550
        keep = True
    elif m < 15:
        n_leg, gold = 1, 900
        keep = False
    elif m < 18:
        n_leg, gold = 2, 700
        keep = False
    elif m < 22:
        n_leg, gold = 2, 1300
        keep = False
    elif m < 25:
        n_leg, gold = 3, 800
        keep = False
    else:
        n_leg, gold = 3, 1200
        keep = False
    return {
        "legendaries": n_leg,
        "boots": True,
        "keep_starter": keep,
        "gold": gold,
        "level": max(6, min(18, level)),
        "cs": cs,
    }


def trinket_id(position: str, minute: float) -> int:
    if minute < 9:
        return WARD
    if position in ("JUNGLE", "UTILITY"):
        return SWEEPER
    if position == "BOTTOM":
        return FARSIGHT
    return SWEEPER if position == "TOP" else FARSIGHT


def ability_levels(level: int) -> dict[str, int]:
    r = 0
    if level >= 6:
        r = 1
    if level >= 11:
        r = 2
    if level >= 16:
        r = 3
    pts = max(0, level - r)
    q = w = e = 0
    # Q > W > E, cap 5
    for i in range(pts):
        bucket = i % 3
        if bucket == 0 and q < 5:
            q += 1
        elif bucket == 1 and w < 5:
            w += 1
        elif e < 5:
            e += 1
        elif q < 5:
            q += 1
        elif w < 5:
            w += 1
    return {"Q": q, "W": w, "E": e, "R": r}


def scores_for(minute: float, position: str, rng: random.Random) -> dict[str, Any]:
    m = float(minute)
    plan = phase_plan(m, position)
    kills = max(0, int(rng.gauss(m / 8.0, 1.4)))
    deaths = max(0, int(rng.gauss(m / 9.0, 1.2)))
    if position == "UTILITY":
        assists = max(0, int(rng.gauss(m / 3.5, 1.5)))
        kills = max(0, kills - 1)
    else:
        assists = max(0, int(rng.gauss(m / 6.0, 1.3)))
    return {
        "assists": min(20, assists),
        "creepScore": int(plan["cs"]),
        "deaths": min(12, deaths),
        "kills": min(15, kills),
        "wardScore": round(m * (1.8 if position == "UTILITY" else 0.6), 1),
    }


def scaled_stats(champ: dict[str, Any], level: int) -> dict[str, float | str]:
    s = champ.get("stats") or {}
    lv = max(1, level) - 1
    hp = float(s.get("hp") or 600) + float(s.get("hpperlevel") or 100) * lv
    mp = float(s.get("mp") or 0) + float(s.get("mpperlevel") or 0) * lv
    from_partype = champ.get("partype") or ""
    rtype = "MANA"
    if from_partype in ("Энергия", "Energy"):
        rtype = "ENERGY"
    elif from_partype in ("Ярость", "Fury", "Свирепость", "Боевой дух"):
        rtype = "FURY"
    elif from_partype in ("None", "Колодец крови", "Щит"):
        rtype = "NONE"
    return {
        "abilityHaste": 0.0,
        "abilityPower": 0.0 if (champ.get("info") or {}).get("magic", 0) < 6 else 18.0 + 4.0 * lv,
        "armor": float(s.get("armor") or 30) + float(s.get("armorperlevel") or 4) * lv,
        "armorPenetrationFlat": 0.0,
        "armorPenetrationPercent": 1.0,
        "attackDamage": float(s.get("attackdamage") or 50) + float(s.get("attackdamageperlevel") or 3) * lv,
        "attackRange": float(s.get("attackrange") or 125),
        "attackSpeed": float(s.get("attackspeed") or 0.625),
        "bonusArmorPenetrationPercent": 1.0,
        "bonusMagicPenetrationPercent": 1.0,
        "critChance": 0.0,
        "critDamage": 200.0,
        "currentHealth": hp,
        "healShieldPower": 0.0,
        "healthRegenRate": float(s.get("hpregen") or 1.0),
        "lifeSteal": 0.0,
        "magicLethality": 0.0,
        "magicPenetrationFlat": 0.0,
        "magicPenetrationPercent": 1.0,
        "magicResist": float(s.get("spellblock") or 30) + float(s.get("spellblockperlevel") or 1.5) * lv,
        "maxHealth": hp,
        "moveSpeed": float(s.get("movespeed") or 335),
        "omnivamp": 0.0,
        "physicalLethality": 0.0,
        "physicalVamp": 0.0,
        "resourceMax": mp,
        "resourceRegenRate": float(s.get("mpregen") or 0),
        "resourceType": rtype,
        "resourceValue": mp,
        "spellVamp": 0.0,
        "tenacity": 5.0,
    }


# First component toward a legendary (mirrors ComponentsToward, shortened).
NEXT_COMPONENT: dict[int, int] = {
    MALIGNANCE: LOST_CHAPTER,
    BLACKFIRE: LOST_CHAPTER,
    ZHONYA: SEEKERS,
    LIANDRY: 3147,
    SHADOWFLAME: 1058,
    STORMSURGE: 3145,
    RABADON: 1058,
    VOID_STAFF: 1026,
    YOUMUU: SERRATED,
    INFINITY: 1038,
    KRAKEN: 6670,
    TRINITY: 3057,
    SUNDERED: 3133,
    CLEAVER: 3133,
    STERAK: 1037,
    ICEBORN: 3057,
    RIFTMAKER: 3147,
    RYLAI: 1026,
    COSMIC: 3108,
    MOONSTONE: 3067,
    LOCKET: 3067,
    RFC: 6670,
    BOTRK: 1053,
}


def playable_items(
    cid: str,
    position: str,
    minute: float,
    rng: random.Random,
    *,
    primary: str,
    tags: set[str],
    make_item: MakeItem,
) -> tuple[list[dict[str, Any]], int, int, dict[str, Any]]:
    slots = seed_slots(cid, position, primary, tags)
    plan = phase_plan(minute, position)
    legendaries = [(i, r) for i, r in slots if r not in ("start", "component", "boots")]
    boots = [(i, r) for i, r in slots if r == "boots"]
    starts = [(i, r) for i, r in slots if r == "start"]

    owned: list[int] = []
    if position == "JUNGLE":
        pet = starts[0][0] if starts else MOSSTOMPER
        owned.append(upgraded_pet(pet) if minute >= 10 else pet)
    elif position == "UTILITY":
        owned.append(support_quest_id(minute, slots))
    elif plan["keep_starter"] and starts:
        owned.append(starts[0][0])

    if plan["boots"] and boots:
        # First boots in seed is the default (often the "good" one).
        owned.append(boots[0][0])

    n_leg = int(plan["legendaries"])
    for item_id, _role in legendaries[:n_leg]:
        owned.append(item_id)

    next_leg = legendaries[n_leg][0] if n_leg < len(legendaries) else None
    gold = int(plan["gold"]) + rng.randint(-120, 180)
    gold = max(80, gold)
    if next_leg and next_leg in NEXT_COMPONENT and gold >= 800 and len(owned) < 5:
        owned.append(NEXT_COMPONENT[next_leg])
        gold = max(150, gold - 400)

    items: list[dict[str, Any]] = []
    slot = 0
    for item_id in owned:
        items.append(make_item(item_id, slot, 1))
        slot += 1
        if slot >= 6:
            break
    if slot < 6 and minute >= 8:
        items.append(make_item(CONTROL_WARD, slot, 1))
        slot += 1
    items.append(make_item(trinket_id(position, minute), 6, 1))

    jitter = rng.randint(-1, 1)
    level = max(6, min(18, int(plan["level"]) + jitter))
    scores = scores_for(minute, position, rng)
    return items, gold, level, scores
