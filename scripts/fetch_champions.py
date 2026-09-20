"""
Выгрузка всех чемпионов LoL из Riot Data Dragon.

Источник:
  1) https://ddragon.leagueoflegends.com/api/versions.json
  2) https://ddragon.leagueoflegends.com/cdn/{version}/data/ru_RU/champion.json
  3) https://ddragon.leagueoflegends.com/cdn/{version}/data/en_US/champion.json

Запуск из корня проекта:
  python scripts/fetch_champions.py
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

import requests

DDRAGON_VERSIONS = "https://ddragon.leagueoflegends.com/api/versions.json"
DDRAGON_CHAMPIONS = "https://ddragon.leagueoflegends.com/cdn/{version}/data/{locale}/champion.json"
ICON_URL = "https://ddragon.leagueoflegends.com/cdn/{version}/img/champion/{image}"

ROOT = Path(__file__).resolve().parents[1]
OUT_DIR = ROOT / "data"
OUT_JSON = OUT_DIR / "champions.json"
OUT_BY_ID = OUT_DIR / "champions_by_id.json"


def fetch_json(url: str) -> dict | list:
    response = requests.get(url, timeout=60)
    response.raise_for_status()
    return response.json()


def normalize_champion(champ_id: str, ru: dict, en: dict | None, version: str) -> dict:
    en = en or {}
    image = (ru.get("image") or {}).get("full") or f"{champ_id}.png"
    info = ru.get("info") or {}
    stats = ru.get("stats") or {}

    return {
        "id": champ_id,
        "key": int(ru.get("key")),
        "name_ru": ru.get("name"),
        "name_en": en.get("name"),
        "title_ru": ru.get("title"),
        "title_en": en.get("title"),
        "blurb_ru": ru.get("blurb"),
        "blurb_en": en.get("blurb"),
        "tags": ru.get("tags") or [],
        "partype": ru.get("partype"),
        "info": {
            "attack": info.get("attack"),
            "defense": info.get("defense"),
            "magic": info.get("magic"),
            "difficulty": info.get("difficulty"),
        },
        "stats": {
            "hp": stats.get("hp"),
            "hpperlevel": stats.get("hpperlevel"),
            "mp": stats.get("mp"),
            "mpperlevel": stats.get("mpperlevel"),
            "movespeed": stats.get("movespeed"),
            "armor": stats.get("armor"),
            "armorperlevel": stats.get("armorperlevel"),
            "spellblock": stats.get("spellblock"),
            "spellblockperlevel": stats.get("spellblockperlevel"),
            "attackrange": stats.get("attackrange"),
            "hpregen": stats.get("hpregen"),
            "hpregenperlevel": stats.get("hpregenperlevel"),
            "mpregen": stats.get("mpregen"),
            "mpregenperlevel": stats.get("mpregenperlevel"),
            "crit": stats.get("crit"),
            "critperlevel": stats.get("critperlevel"),
            "attackdamage": stats.get("attackdamage"),
            "attackdamageperlevel": stats.get("attackdamageperlevel"),
            "attackspeedperlevel": stats.get("attackspeedperlevel"),
            "attackspeed": stats.get("attackspeed"),
        },
        "icon": ICON_URL.format(version=version, image=image),
        "icon_file": image,
    }


def main() -> int:
    print("Fetching Data Dragon versions...")
    versions = fetch_json(DDRAGON_VERSIONS)
    if not versions:
        print("No versions returned", file=sys.stderr)
        return 1

    version = versions[0]
    print(f"Using patch/version: {version}")

    print("Fetching ru_RU champions...")
    ru_payload = fetch_json(DDRAGON_CHAMPIONS.format(version=version, locale="ru_RU"))
    print("Fetching en_US champions...")
    en_payload = fetch_json(DDRAGON_CHAMPIONS.format(version=version, locale="en_US"))

    ru_data: dict = ru_payload.get("data") or {}
    en_data: dict = en_payload.get("data") or {}

    champions = []
    by_id = {}
    for champ_id, ru_champ in sorted(ru_data.items(), key=lambda x: x[0].lower()):
        normalized = normalize_champion(champ_id, ru_champ, en_data.get(champ_id), version)
        champions.append(normalized)
        by_id[champ_id] = normalized

    payload = {
        "source": "ddragon",
        "version": version,
        "locale_primary": "ru_RU",
        "count": len(champions),
        "champions": champions,
    }

    OUT_DIR.mkdir(parents=True, exist_ok=True)
    OUT_JSON.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
    OUT_BY_ID.write_text(json.dumps(by_id, ensure_ascii=False, indent=2), encoding="utf-8")

    print(f"Saved {len(champions)} champions -> {OUT_JSON}")
    print(f"Saved id index -> {OUT_BY_ID}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
