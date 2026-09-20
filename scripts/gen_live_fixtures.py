"""
Generate full Riot Live Client allgamedata JSON fixtures (same shape as
https://127.0.0.1:2999/liveclientdata/allgamedata) and optionally run recommend.

Usage (from repo root):
  python scripts/gen_live_fixtures.py --count 100
  python scripts/gen_live_fixtures.py --count 50 --recommend --report out/report.md
  python scripts/gen_live_fixtures.py --fetch-abilities   # one-time DDragon spell cache
"""

from __future__ import annotations

import argparse
import json
import os
import random
import shutil
import subprocess
import sys
import time
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[1]
DATA = ROOT / "data"
OUT_DIR_DEFAULT = ROOT / "testdata" / "generated" / "live"
ABILITIES_CACHE = DATA / "champion_abilities.json"

POSITIONS = ("TOP", "JUNGLE", "MIDDLE", "BOTTOM", "UTILITY")

# Live Client resourceType strings.
PARTYPE_TO_RESOURCE = {
    "Мана": "MANA",
    "Mana": "MANA",
    "Энергия": "ENERGY",
    "Energy": "ENERGY",
    "Ярость": "FURY",
    "None": "NONE",
    "Колодец крови": "NONE",
    "Щит": "NONE",
    "Ярость ветра": "NONE",
    "Поток": "FLOW",
    "Жар": "HEAT",
    "Свирепость": "FURY",
    "Боевой дух": "FURY",
}

# Starting item ids (SR). Class-aware lane starts; jungle/support by position.
ITEM_WARD = 3340
ITEM_POTION = 2003
ITEM_DORANS_BLADE = 1055
ITEM_DORANS_RING = 1056
ITEM_DORANS_SHIELD = 1054
ITEM_JUNGLE_PET = 1103
ITEM_WORLD_ATLAS = 3865

# Fallback potions-only counts by lane (starter chosen separately).
POTIONS_BY_POS: dict[str, int] = {
    "TOP": 1,
    "JUNGLE": 0,
    "MIDDLE": 2,
    "BOTTOM": 2,
    "UTILITY": 2,
}

# Summoner spells (RU display + raw keys matching Live Client).
SUMMONERS: dict[str, tuple[dict[str, str], dict[str, str]]] = {
    "TOP": (
        {
            "displayName": "Скачок",
            "rawDescription": "GeneratedTip_SummonerSpell_SummonerFlash_Description",
            "rawDisplayName": "GeneratedTip_SummonerSpell_SummonerFlash_DisplayName",
        },
        {
            "displayName": "Телепорт",
            "rawDescription": "GeneratedTip_SummonerSpell_SummonerTeleport_Description",
            "rawDisplayName": "GeneratedTip_SummonerSpell_SummonerTeleport_DisplayName",
        },
    ),
    "JUNGLE": (
        {
            "displayName": "Скачок",
            "rawDescription": "GeneratedTip_SummonerSpell_SummonerFlash_Description",
            "rawDisplayName": "GeneratedTip_SummonerSpell_SummonerFlash_DisplayName",
        },
        {
            "displayName": "Кара",
            "rawDescription": "GeneratedTip_SummonerSpell_SummonerSmite_Description",
            "rawDisplayName": "GeneratedTip_SummonerSpell_SummonerSmite_DisplayName",
        },
    ),
    "MIDDLE": (
        {
            "displayName": "Скачок",
            "rawDescription": "GeneratedTip_SummonerSpell_SummonerFlash_Description",
            "rawDisplayName": "GeneratedTip_SummonerSpell_SummonerFlash_DisplayName",
        },
        {
            "displayName": "Воспламенение",
            "rawDescription": "GeneratedTip_SummonerSpell_SummonerDot_Description",
            "rawDisplayName": "GeneratedTip_SummonerSpell_SummonerDot_DisplayName",
        },
    ),
    "BOTTOM": (
        {
            "displayName": "Барьер",
            "rawDescription": "GeneratedTip_SummonerSpell_SummonerBarrier_Description",
            "rawDisplayName": "GeneratedTip_SummonerSpell_SummonerBarrier_DisplayName",
        },
        {
            "displayName": "Скачок",
            "rawDescription": "GeneratedTip_SummonerSpell_SummonerFlash_Description",
            "rawDisplayName": "GeneratedTip_SummonerSpell_SummonerFlash_DisplayName",
        },
    ),
    "UTILITY": (
        {
            "displayName": "Скачок",
            "rawDescription": "GeneratedTip_SummonerSpell_SummonerFlash_Description",
            "rawDisplayName": "GeneratedTip_SummonerSpell_SummonerFlash_DisplayName",
        },
        {
            "displayName": "Исцеление",
            "rawDescription": "GeneratedTip_SummonerSpell_SummonerHeal_Description",
            "rawDisplayName": "GeneratedTip_SummonerSpell_SummonerHeal_DisplayName",
        },
    ),
}

STAT_RUNES = [
    {"id": 5005, "rawDescription": "perk_tooltip_StatModAttackSpeed"},
    {"id": 5008, "rawDescription": "perk_tooltip_StatModAdaptive"},
    {"id": 5001, "rawDescription": "perk_tooltip_StatModHealthScaling"},
]


def rune(rid: int, name: str, tip_key: str) -> dict[str, Any]:
    return {
        "displayName": name,
        "id": rid,
        "rawDescription": f"perk_tooltip_{tip_key}",
        "rawDisplayName": f"perk_displayname_{tip_key}",
    }


def tree(rid: int, name: str, style_key: str) -> dict[str, Any]:
    return {
        "displayName": name,
        "id": rid,
        "rawDescription": f"perkstyle_tooltip_{style_key}",
        "rawDisplayName": f"perkstyle_displayname_{style_key}",
    }


# Rune page templates keyed by primary class from champion_tags.
RUNE_PAGES: dict[str, dict[str, Any]] = {
    "tank": {
        "keystone": rune(8437, "Хватка нежити", "GraspOfTheUndying"),
        "primary": tree(8400, "Храбрость", "7204"),
        "secondary": tree(8200, "Колдовство", "7202"),
        "general": [
            rune(8437, "Хватка нежити", "GraspOfTheUndying"),
            rune(8446, "Снос", "Demolish"),
            rune(8473, "Железный панцирь", "BonePlating"),
            rune(8242, "Неустрашимость", "Unflinching"),
            rune(8226, "Поток маны", "8226"),
            rune(8210, "Превосходство", "Transcendence"),
        ],
    },
    "fighter": {
        "keystone": rune(8010, "Завоеватель", "Conqueror"),
        "primary": tree(8000, "Точность", "7201"),
        "secondary": tree(8400, "Храбрость", "7204"),
        "general": [
            rune(8010, "Завоеватель", "Conqueror"),
            rune(9111, "Триумф", "Triumph"),
            rune(9104, "Легенда: рвение", "LegendAlacrity"),
            rune(8014, "Удар милосердия", "CoupDeGrace"),
            rune(8473, "Железный панцирь", "BonePlating"),
            rune(8242, "Неустрашимость", "Unflinching"),
        ],
    },
    "assassin": {
        "keystone": rune(8112, "Казнь электричеством", "Electrocute"),
        "primary": tree(8100, "Доминирование", "7200"),
        "secondary": tree(8200, "Колдовство", "7202"),
        "general": [
            rune(8112, "Казнь электричеством", "Electrocute"),
            rune(8139, "Вкус крови", "TasteOfBlood"),
            rune(8140, "Ужасные сувениры", "GrislyMementos"),
            rune(8106, "Абсолютный охотник", "UltimateHunter"),
            rune(8226, "Поток маны", "8226"),
            rune(8210, "Превосходство", "Transcendence"),
        ],
    },
    "mage": {
        "keystone": rune(8229, "Магическая комета", "ArcaneComet"),
        "primary": tree(8200, "Колдовство", "7202"),
        "secondary": tree(8100, "Доминирование", "7200"),
        "general": [
            rune(8229, "Магическая комета", "ArcaneComet"),
            rune(8226, "Поток маны", "8226"),
            rune(8210, "Превосходство", "Transcendence"),
            rune(8237, "Скорбящий урожай", "Scorch"),
            rune(8139, "Вкус крови", "TasteOfBlood"),
            rune(8106, "Абсолютный охотник", "UltimateHunter"),
        ],
    },
    "marksman": {
        "keystone": rune(8008, "Смертельный темп", "LethalTempoTemp"),
        "primary": tree(8000, "Точность", "7201"),
        "secondary": tree(8100, "Доминирование", "7200"),
        "general": [
            rune(8008, "Смертельный темп", "LethalTempoTemp"),
            rune(9111, "Триумф", "Triumph"),
            rune(9104, "Легенда: рвение", "LegendAlacrity"),
            rune(8014, "Удар милосердия", "CoupDeGrace"),
            rune(8139, "Вкус крови", "TasteOfBlood"),
            rune(8135, "Алчность", "RavenousHunter"),
        ],
    },
    "support": {
        "keystone": rune(8465, "Хранитель", "Guardian"),
        "primary": tree(8400, "Храбрость", "7204"),
        "secondary": tree(8200, "Колдовство", "7202"),
        "general": [
            rune(8465, "Хранитель", "Guardian"),
            rune(8463, "Хрупкость", "FontOfLife"),
            rune(8473, "Железный панцирь", "BonePlating"),
            rune(8242, "Неустрашимость", "Unflinching"),
            rune(8210, "Превосходство", "Transcendence"),
            rune(8226, "Поток маны", "8226"),
        ],
    },
}


def load_json(path: Path) -> Any:
    with path.open(encoding="utf-8") as f:
        return json.load(f)


def save_json(path: Path, data: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent="\t")
        f.write("\n")


def fetch_abilities(version: str | None = None) -> dict[str, Any]:
    try:
        import requests
    except ImportError as e:
        raise SystemExit("pip install requests") from e

    if not version:
        versions = requests.get(
            "https://ddragon.leagueoflegends.com/api/versions.json", timeout=60
        ).json()
        version = versions[0]

    print(f"Fetching champion abilities for patch {version}...")
    champs = load_json(DATA / "champions.json")["champions"]
    out: dict[str, Any] = {"source": "ddragon", "version": version, "champions": {}}
    for i, c in enumerate(champs, 1):
        cid = c["id"]
        url = (
            f"https://ddragon.leagueoflegends.com/cdn/{version}/data/ru_RU/champion/{cid}.json"
        )
        resp = requests.get(url, timeout=60)
        resp.raise_for_status()
        data = resp.json()["data"][cid]
        passive = data.get("passive") or {}
        spells = data.get("spells") or []
        # Live Client ability ids look like VexQ / VexP.
        abilities = {
            "Passive": {
                "displayName": passive.get("name") or f"{cid} Passive",
                "id": f"{cid}P",
                "rawDescription": f"GeneratedTip_Passive_{cid}P_Description",
                "rawDisplayName": f"GeneratedTip_Passive_{cid}P_DisplayName",
            }
        }
        for slot, spell in zip(("Q", "W", "E", "R"), spells):
            sid = spell.get("id") or f"{cid}{slot}"
            abilities[slot] = {
                "abilityLevel": 0,
                "displayName": spell.get("name") or sid,
                "id": sid,
                "rawDescription": f"GeneratedTip_Spell_{sid}_Description",
                "rawDisplayName": f"GeneratedTip_Spell_{sid}_DisplayName",
            }
        out["champions"][cid] = abilities
        if i % 20 == 0 or i == len(champs):
            print(f"  {i}/{len(champs)}")
        time.sleep(0.05)

    save_json(ABILITIES_CACHE, out)
    print(f"Wrote {ABILITIES_CACHE}")
    return out


def stub_abilities(cid: str) -> dict[str, Any]:
    out: dict[str, Any] = {
        "Passive": {
            "displayName": f"{cid} Passive",
            "id": f"{cid}P",
            "rawDescription": f"GeneratedTip_Passive_{cid}P_Description",
            "rawDisplayName": f"GeneratedTip_Passive_{cid}P_DisplayName",
        }
    }
    for slot in ("Q", "W", "E", "R"):
        sid = f"{cid}{slot}"
        out[slot] = {
            "abilityLevel": 0,
            "displayName": sid,
            "id": sid,
            "rawDescription": f"GeneratedTip_Spell_{sid}_Description",
            "rawDisplayName": f"GeneratedTip_Spell_{sid}_DisplayName",
        }
    return out


def resource_type(partype: str | None) -> str:
    if not partype:
        return "MANA"
    return PARTYPE_TO_RESOURCE.get(partype, "MANA")


def champion_stats(champ: dict[str, Any]) -> dict[str, float | str]:
    s = champ.get("stats") or {}
    hp = float(s.get("hp") or 600)
    mp = float(s.get("mp") or 0)
    rtype = resource_type(champ.get("partype"))
    return {
        "abilityHaste": 0.0,
        "abilityPower": 0.0 if (champ.get("info") or {}).get("magic", 0) < 6 else 9.0,
        "armor": float(s.get("armor") or 30),
        "armorPenetrationFlat": 0.0,
        "armorPenetrationPercent": 1.0,
        "attackDamage": float(s.get("attackdamage") or 50),
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
        "magicResist": float(s.get("spellblock") or 30),
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


class Catalog:
    def __init__(self) -> None:
        champs_file = load_json(DATA / "champions.json")
        self.version = champs_file.get("version", "16.18.1")
        self.champions = {c["id"]: c for c in champs_file["champions"]}
        self.champ_ids = sorted(self.champions)

        items_file = load_json(DATA / "items.json")
        self.items: dict[int, dict[str, Any]] = {
            int(it["id"]): it for it in items_file["items"]
        }

        tags_file = load_json(DATA / "champion_tags.json")
        self.tags: dict[str, Any] = tags_file.get("champions") or {}

        if ABILITIES_CACHE.exists():
            self.abilities = load_json(ABILITIES_CACHE).get("champions") or {}
        else:
            self.abilities = {}

    def primary(self, cid: str, position: str) -> str:
        profile = self.tags.get(cid) or {}
        by_role = profile.get("by_role") or {}
        if position in by_role:
            return (by_role[position].get("primary") or profile.get("primary") or "fighter").lower()
        return (profile.get("primary") or "fighter").lower()

    def tag_set(self, cid: str, position: str = "") -> set[str]:
        profile = self.tags.get(cid) or {}
        by_role = profile.get("by_role") or {}
        if position and position in by_role:
            ov = by_role[position]
            tags = set(ov.get("tags") or [])
            tags.add((ov.get("primary") or profile.get("primary") or "fighter").lower())
            return tags
        tags = set(profile.get("tags") or [])
        tags.add((profile.get("primary") or "fighter").lower())
        return tags

    def natural_positions(self, cid: str) -> list[str]:
        """Mirror tags.ChampionProfile.NaturalPositions — on-role SR seats."""
        profile = self.tags.get(cid) or {}
        primary = (profile.get("primary") or "fighter").lower()
        tagset = set(profile.get("tags") or [])
        by_role = profile.get("by_role") or {}
        out: list[str] = []
        seen: set[str] = set()

        def add(pos: str) -> None:
            if pos in POSITIONS and pos not in seen:
                seen.add(pos)
                out.append(pos)

        def class_default(p: str) -> str:
            return {
                "marksman": "BOTTOM",
                "support": "UTILITY",
                "mage": "MIDDLE",
                "assassin": "MIDDLE",
                "fighter": "TOP",
                "tank": "TOP",
            }.get(p, "MIDDLE")

        if by_role:
            for pos in by_role:
                add(pos)
            add(class_default(primary))
            return out or ["MIDDLE"]

        if primary == "marksman":
            add("BOTTOM")
            add("MIDDLE")
        elif primary == "support":
            add("UTILITY")
        elif primary == "mage":
            add("MIDDLE")
            if tagset & {"poke", "healer", "shield"}:
                add("UTILITY")
            if "dive" in tagset:
                add("JUNGLE")
        elif primary == "assassin":
            add("MIDDLE")
            add("JUNGLE")
            if "ad" in tagset and "ap" not in tagset:
                add("TOP")
        elif primary == "fighter":
            add("TOP")
            add("JUNGLE")
        elif primary == "tank":
            add("TOP")
            add("JUNGLE")
            if tagset & {"engage", "peel"}:
                add("UTILITY")
        else:
            add("MIDDLE")
            add("TOP")

        return out or ["MIDDLE"]

    def pool_for(self, position: str) -> list[str]:
        return [cid for cid in self.champ_ids if position in self.natural_positions(cid)]

    def lane_start_id(self, cid: str, position: str) -> int:
        if position == "JUNGLE":
            return ITEM_JUNGLE_PET
        if position == "UTILITY":
            return ITEM_WORLD_ATLAS
        primary = self.primary(cid, position)
        tags = self.tag_set(cid, position)
        if primary == "tank" or "tank" in tags:
            return ITEM_DORANS_SHIELD
        if primary == "mage" or (primary == "assassin" and "ap" in tags) or (
            primary == "fighter" and "ap" in tags
        ):
            return ITEM_DORANS_RING
        if primary == "support":
            # Solo-lane support remap: tanky → Shield, else Ring.
            if tags & {"tank", "engage", "peel", "melee"}:
                return ITEM_DORANS_SHIELD
            return ITEM_DORANS_RING
        if primary == "fighter" and "juggernaut" in tags:
            return ITEM_DORANS_SHIELD
        return ITEM_DORANS_BLADE

    def make_item(self, item_id: int, slot: int, count: int = 1) -> dict[str, Any]:
        meta = self.items.get(item_id) or {}
        name = meta.get("name_ru") or meta.get("name_en") or str(item_id)
        gold = (meta.get("gold") or {}).get("total") or 0
        consumable = bool(meta.get("consumed"))
        can_use = consumable or item_id in (ITEM_WARD, 3364, 3363, 2055)
        return {
            "canUse": can_use,
            "consumable": consumable,
            "count": count,
            "displayName": name,
            "itemID": item_id,
            "price": int(gold),
            "rawDescription": f"GeneratedTip_Item_{item_id}_Description",
            "rawDisplayName": f"Item_{item_id}_Name",
            "slot": slot,
        }

    def starting_items(self, cid: str, position: str) -> list[dict[str, Any]]:
        items: list[dict[str, Any]] = []
        slot = 0
        start_id = self.lane_start_id(cid, position)
        if start_id in self.items or start_id == ITEM_WARD:
            items.append(self.make_item(start_id, slot, 1))
            slot += 1
        pots = POTIONS_BY_POS.get(position, 1)
        if pots > 0 and ITEM_POTION in self.items:
            items.append(self.make_item(ITEM_POTION, slot, pots))
            slot += 1
        items.append(self.make_item(ITEM_WARD, 6, 1))
        return items

    def abilities_for(self, cid: str) -> dict[str, Any]:
        return self.abilities.get(cid) or stub_abilities(cid)

    def runes_for(self, cid: str, position: str) -> dict[str, Any]:
        primary = self.primary(cid, position)
        page = RUNE_PAGES.get(primary) or RUNE_PAGES["fighter"]
        return {
            "keystone": page["keystone"],
            "primaryRuneTree": page["primary"],
            "secondaryRuneTree": page["secondary"],
        }

    def full_runes_for(self, cid: str, position: str) -> dict[str, Any]:
        primary = self.primary(cid, position)
        page = RUNE_PAGES.get(primary) or RUNE_PAGES["fighter"]
        return {
            "generalRunes": page["general"],
            "keystone": page["keystone"],
            "primaryRuneTree": page["primary"],
            "secondaryRuneTree": page["secondary"],
            "statRunes": STAT_RUNES,
        }


def bot_name(champ: dict[str, Any]) -> tuple[str, str, str, str]:
    """riotId, riotIdGameName, riotIdTagLine, summonerName."""
    cid = champ["id"]
    name_ru = champ.get("name_ru") or cid
    return f"{cid}#BOT", cid, "BOT", f"{name_ru} (бот)"


def human_ids(index: int) -> tuple[str, str, str, str]:
    tag = f"{10000 + index}"
    game = f"Player{index}"
    full = f"{game}#{tag}"
    return full, game, tag, full


def make_player(
    cat: Catalog,
    champ: dict[str, Any],
    *,
    team: str,
    position: str,
    is_bot: bool,
    is_active: bool,
    level: int,
    gold: float,
    human_index: int,
    with_items: bool,
) -> dict[str, Any]:
    cid = champ["id"]
    name_ru = champ.get("name_ru") or cid
    if is_bot:
        riot_id, game_name, tag, summoner = bot_name(champ)
    else:
        riot_id, game_name, tag, summoner = human_ids(human_index)

    spells = SUMMONERS.get(position, SUMMONERS["MIDDLE"])
    items = cat.starting_items(cid, position) if with_items else []

    player: dict[str, Any] = {
        "championName": name_ru,
        "isBot": is_bot,
        "isDead": False,
        "items": items,
        "level": level,
        "position": position,
        "rawChampionName": f"game_character_displayname_{cid}",
        "rawSkinName": f"game_character_displayname_{cid}",
        "respawnTimer": 0.0,
        "riotId": riot_id,
        "riotIdGameName": game_name,
        "riotIdTagLine": tag,
        "runes": cat.runes_for(cid, position),
        "scores": {
            "assists": 0,
            "creepScore": 0,
            "deaths": 0,
            "kills": 0,
            "wardScore": 0.0,
        },
        "skinID": 0,
        "skinName": name_ru,
        "summonerName": summoner,
        "summonerSpells": {
            "summonerSpellOne": spells[0],
            "summonerSpellTwo": spells[1],
        },
        "team": team,
    }
    if is_active:
        # Used only for matching activePlayer.summonerName
        player["_active"] = True
        player["_gold"] = gold
    return player


def make_allgamedata(
    cat: Catalog,
    *,
    active_id: str,
    active_pos: str,
    ally_ids: dict[str, str],
    enemy_ids: dict[str, str],
    seed: int,
    case_index: int,
    enemy_items: bool,
) -> dict[str, Any]:
    rng = random.Random(seed)
    active = cat.champions[active_id]
    level = 1
    gold = 500.0

    players: list[dict[str, Any]] = []

    # Active human first (matches typical Live Client ordering).
    p_active = make_player(
        cat,
        active,
        team="ORDER",
        position=active_pos,
        is_bot=False,
        is_active=True,
        level=level,
        gold=gold,
        human_index=case_index,
        with_items=True,
    )
    players.append(p_active)

    for pos, cid in ally_ids.items():
        if cid == active_id:
            continue
        players.append(
            make_player(
                cat,
                cat.champions[cid],
                team="ORDER",
                position=pos,
                is_bot=True,
                is_active=False,
                level=level,
                gold=0,
                human_index=0,
                with_items=True,
            )
        )

    for pos, cid in enemy_ids.items():
        players.append(
            make_player(
                cat,
                cat.champions[cid],
                team="CHAOS",
                position=pos,
                is_bot=True,
                is_active=False,
                level=level,
                gold=0,
                human_index=0,
                with_items=enemy_items,
            )
        )

    # Drop helper keys.
    for p in players:
        p.pop("_active", None)
        p.pop("_gold", None)

    abilities = json.loads(json.dumps(cat.abilities_for(active_id)))
    # Ensure abilityLevel keys present for QWER.
    for slot in ("Q", "W", "E", "R"):
        if slot in abilities and "abilityLevel" not in abilities[slot]:
            abilities[slot]["abilityLevel"] = 0

    summoner = p_active["summonerName"]
    riot_id = p_active["riotId"]
    game_name = p_active["riotIdGameName"]
    tag = p_active["riotIdTagLine"]

    return {
        "activePlayer": {
            "abilities": abilities,
            "championStats": champion_stats(active),
            "currentGold": gold,
            "fullRunes": cat.full_runes_for(active_id, active_pos),
            "level": level,
            "riotId": riot_id,
            "riotIdGameName": game_name,
            "riotIdTagLine": tag,
            "summonerName": summoner,
            "teamRelativeColors": True,
        },
        "allPlayers": players,
        "events": {
            "Events": [
                {
                    "EventID": 0,
                    "EventName": "GameStart",
                    "EventTime": round(rng.uniform(0.01, 0.2), 14),
                }
            ]
        },
        "gameData": {
            "gameMode": "PRACTICETOOL",
            "gameTime": round(rng.uniform(20.0, 45.0), 8),
            "mapName": "Map11",
            "mapNumber": 11,
            "mapTerrain": "Default",
        },
    }


def pick_comp(cat: Catalog, rng: random.Random, active_id: str, active_pos: str) -> tuple[dict[str, str], dict[str, str]]:
    used = {active_id}
    allies: dict[str, str] = {active_pos: active_id}
    enemies: dict[str, str] = {}

    def pick_for(pos: str) -> str:
        pool = [cid for cid in cat.pool_for(pos) if cid not in used]
        if not pool:
            # Extremely rare: fall back to any unused champ.
            pool = [cid for cid in cat.champ_ids if cid not in used]
        choice = rng.choice(pool)
        used.add(choice)
        return choice

    for pos in POSITIONS:
        if pos not in allies:
            allies[pos] = pick_for(pos)
        enemies[pos] = pick_for(pos)
    return allies, enemies


def generate_cases(cat: Catalog, count: int, seed: int, enemy_items: bool) -> list[tuple[str, dict[str, Any], dict[str, str]]]:
    """Returns list of (filename, json, meta). Active champ always on a natural role."""
    rng = random.Random(seed)
    cases: list[tuple[str, dict[str, Any], dict[str, str]]] = []

    schedule: list[tuple[str, str]] = []
    for cid in cat.champ_ids:
        for pos in cat.natural_positions(cid):
            schedule.append((cid, pos))
    if not schedule:
        raise SystemExit("no on-role (champ, position) pairs — check champion_tags.json")
    rng.shuffle(schedule)

    for i in range(count):
        cid, pos = schedule[i % len(schedule)]
        case_seed = seed + i * 9973
        crng = random.Random(case_seed)
        allies, enemies = pick_comp(cat, crng, cid, pos)
        payload = make_allgamedata(
            cat,
            active_id=cid,
            active_pos=pos,
            ally_ids=allies,
            enemy_ids=enemies,
            seed=case_seed,
            case_index=i + 1,
            enemy_items=enemy_items,
        )
        fname = f"{i + 1:04d}_{cid}_{pos}.json"
        meta = {
            "active": cid,
            "position": pos,
            "allies": ",".join(f"{p}:{c}" for p, c in sorted(allies.items())),
            "enemies": ",".join(f"{p}:{c}" for p, c in sorted(enemies.items())),
        }
        cases.append((fname, payload, meta))
    return cases


def find_go() -> str:
    env = os.environ.get("GO_BIN") or os.environ.get("GOTOOL")
    if env and Path(env).exists():
        return env
    which = shutil.which("go")
    if which:
        return which
    win = Path(r"C:\Program Files\Go\bin\go.exe")
    if win.exists():
        return str(win)
    raise SystemExit("go not found; set GO_BIN or install Go")


def run_recommend(fixture: Path) -> str:
    go = find_go()
    cmd = [go, "run", "./cmd/recommend", "-fixture", str(fixture)]
    env = os.environ.copy()
    env.setdefault("PYTHONUTF8", "1")
    # Prefer UTF-8 stdout from Go on Windows consoles.
    env["LANG"] = "C.UTF-8"
    env["LC_ALL"] = "C.UTF-8"
    proc = subprocess.run(
        cmd,
        cwd=str(ROOT),
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
        env=env,
    )
    if proc.returncode != 0:
        return f"ERROR:\n{proc.stderr or proc.stdout}"
    return proc.stdout


def write_report(out_dir: Path, cases: list[tuple[str, dict[str, Any], dict[str, str]]], report_path: Path) -> None:
    lines: list[str] = [
        "# Live fixtures recommend report",
        "",
        f"Generated cases: {len(cases)}",
        f"Fixtures dir: `{out_dir}`",
        "",
    ]
    for fname, _payload, meta in cases:
        path = out_dir / fname
        print(f"recommend {fname}...")
        text = run_recommend(path)
        lines.append(f"## {fname}")
        lines.append("")
        lines.append(f"- active: `{meta['active']}` @ `{meta['position']}`")
        lines.append(f"- allies: `{meta['allies']}`")
        lines.append(f"- enemies: `{meta['enemies']}`")
        lines.append("")
        lines.append("```")
        lines.append(text.rstrip())
        lines.append("```")
        lines.append("")
    report_path.parent.mkdir(parents=True, exist_ok=True)
    report_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"Wrote report {report_path}")


def main() -> int:
    ap = argparse.ArgumentParser(description="Generate full Live Client allgamedata fixtures")
    ap.add_argument("--count", type=int, default=100, help="number of games to generate")
    ap.add_argument("--seed", type=int, default=42)
    ap.add_argument("--out", type=Path, default=OUT_DIR_DEFAULT)
    ap.add_argument("--fetch-abilities", action="store_true", help="download DDragon spells cache")
    ap.add_argument("--enemy-items", action="store_true", help="give CHAOS bots starting items too")
    ap.add_argument("--recommend", action="store_true", help="run recommend on each fixture")
    ap.add_argument("--report", type=Path, default=None, help="markdown report path")
    args = ap.parse_args()

    if args.fetch_abilities:
        fetch_abilities()

    cat = Catalog()
    if not cat.abilities:
        print(
            "Note: no data/champion_abilities.json — using stub ability names. "
            "Run with --fetch-abilities for RU spell names.",
            file=sys.stderr,
        )

    cases = generate_cases(cat, args.count, args.seed, args.enemy_items)
    args.out.mkdir(parents=True, exist_ok=True)

    # Clear previous generated fixtures in out dir (only our pattern).
    for old in args.out.glob("*.json"):
        old.unlink()

    index: list[dict[str, str]] = []
    for fname, payload, meta in cases:
        save_json(args.out / fname, payload)
        index.append({"file": fname, **meta})
    save_json(args.out / "index.json", {"count": len(index), "cases": index})
    print(f"Wrote {len(cases)} fixtures -> {args.out}")

    if args.recommend:
        report = args.report or (args.out / "recommend_report.md")
        write_report(args.out, cases, report)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
