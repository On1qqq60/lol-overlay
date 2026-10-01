import json
from pathlib import Path

data = json.loads(Path("data/items_by_id.json").read_text(encoding="utf-8"))
needles = [
    "Archangel", "Ardent Censer", "Bami", "Bloodsong", "Caulfield",
    "Celestial Opposition", "Chempunk Chainsword", "Cosmic Drive", "Dawncore",
    "Dead Man's Plate", "Dream Maker", "Essence Reaver", "Forbidden Idol",
    "Frozen Heart", "Giant's Belt", "Guinsoo's Rageblade", "Health Potion",
    "Heartsteel", "Experimental Hexplate", "Hollow Radiance", "Horizon Focus",
    "Jak'Sho", "Kindlegem", "Knight's Vow", "Manamune", "Imperial Mandate",
    "Maw of Malmortius", "Mejai's Soulstealer", "Mikael's Blessing",
    "Moonstone Renewer", "Navori Flickerblade", "Phantom Dancer", "Profane Hydra",
    "Ravenous Hydra", "Redemption", "Hextech Rocketbelt", "Runaan's Hurricane",
    "Runic Compass", "Rylai's Crystal Scepter", "Seraph's Embrace",
    "Immortal Shieldbow", "Spear of Shojin", "Shurelya's Battlesong",
    "Solstice Sleigh", "Staff of Flowing Water", "Statikk Shiv", "Stormrazor",
    "Stridebreaker", "Sunfire Aegis", "Boots of Swiftness", "Tear of the Goddess",
    "Terminus", "Thornmail", "Titanic Hydra", "Umbral Glaive", "Unending Despair",
    "Voltaic Cyclosword", "Warmog's Armor", "Zaz'Zak's", "Locket of the Iron Solari",
    "Bounty of Worlds", "World Atlas",
]
for n in needles:
    found = []
    for v in data.values():
        if not isinstance(v, dict):
            continue
        name = v.get("name_en") or ""
        iid = v.get("id")
        if not isinstance(iid, int) or iid > 20000:
            continue
        if n.lower() in name.lower():
            found.append((iid, name))
    print(f"{n}: {found[:4]}")
