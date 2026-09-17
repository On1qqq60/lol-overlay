"""
Выгрузка всех предметов LoL из Riot Data Dragon.

Источник:
  1) https://ddragon.leagueoflegends.com/api/versions.json
  2) https://ddragon.leagueoflegends.com/cdn/{version}/data/ru_RU/item.json
  3) https://ddragon.leagueoflegends.com/cdn/{version}/data/en_US/item.json

Запуск из корня проекта:
  python scripts/fetch_items.py
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

import requests

DDRAGON_VERSIONS = "https://ddragon.leagueoflegends.com/api/versions.json"
DDRAGON_ITEMS = "https://ddragon.leagueoflegends.com/cdn/{version}/data/{locale}/item.json"
ICON_URL = "https://ddragon.leagueoflegends.com/cdn/{version}/img/item/{image}"

ROOT = Path(__file__).resolve().parents[1]
OUT_DIR = ROOT / "data"
OUT_JSON = OUT_DIR / "items.json"
OUT_BY_ID = OUT_DIR / "items_by_id.json"


def fetch_json(url: str) -> dict | list:
    response = requests.get(url, timeout=60)
    response.raise_for_status()
    return response.json()


def normalize_item(item_id: str, ru: dict, en: dict | None, version: str) -> dict:
    en = en or {}
    gold = ru.get("gold") or {}
    image = (ru.get("image") or {}).get("full") or f"{item_id}.png"

    return {
        "id": int(item_id),
        "id_str": item_id,
        "name_ru": ru.get("name"),
        "name_en": en.get("name"),
        "description_ru": ru.get("description"),
        "plaintext_ru": ru.get("plaintext"),
        "plaintext_en": en.get("plaintext"),
        "tags": ru.get("tags") or [],
        "stats": ru.get("stats") or {},
        "gold": {
            "base": gold.get("base"),
            "total": gold.get("total"),
            "sell": gold.get("sell"),
            "purchasable": gold.get("purchasable"),
        },
        "into": [int(x) for x in (ru.get("into") or [])],
        "from": [int(x) for x in (ru.get("from") or [])],
        "depth": ru.get("depth"),
        "maps": ru.get("maps") or {},
        "in_store": bool(ru.get("inStore", True)),
        "consumed": bool(ru.get("consumed", False)),
        "consume_on_full": bool(ru.get("consumeOnFull", False)),
        "required_champion": ru.get("requiredChampion"),
        "required_ally": ru.get("requiredAlly"),
        "special_recipe": ru.get("specialRecipe"),
        "stacks": ru.get("stacks"),
        "max_stacks": ru.get("maxStacks"),
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

    print("Fetching ru_RU items...")
    ru_payload = fetch_json(DDRAGON_ITEMS.format(version=version, locale="ru_RU"))
    print("Fetching en_US items...")
    en_payload = fetch_json(DDRAGON_ITEMS.format(version=version, locale="en_US"))

    ru_data: dict = ru_payload.get("data") or {}
    en_data: dict = en_payload.get("data") or {}

    items = []
    by_id = {}
    for item_id, ru_item in sorted(ru_data.items(), key=lambda x: int(x[0])):
        normalized = normalize_item(item_id, ru_item, en_data.get(item_id), version)
        items.append(normalized)
        by_id[item_id] = normalized

    # Только то, что реально можно купить в магазине Summoner's Rift
    purchasable_sr = [
        item
        for item in items
        if item["gold"]["purchasable"]
        and item["in_store"]
        and (item["maps"].get("11") is True or item["maps"].get("11") == True)
    ]

    payload = {
        "source": "ddragon",
        "version": version,
        "locale_primary": "ru_RU",
        "count_all": len(items),
        "count_purchasable_sr": len(purchasable_sr),
        "items": items,
        "purchasable_summoners_rift": purchasable_sr,
    }

    OUT_DIR.mkdir(parents=True, exist_ok=True)
    OUT_JSON.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
    OUT_BY_ID.write_text(json.dumps(by_id, ensure_ascii=False, indent=2), encoding="utf-8")

    print(f"Saved {len(items)} items -> {OUT_JSON}")
    print(f"Saved id index -> {OUT_BY_ID}")
    print(f"Purchasable on SR: {len(purchasable_sr)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
