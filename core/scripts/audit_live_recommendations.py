#!/usr/bin/env python3
"""Independent Emerald+ audit of recommend_report.md vs 1000 live fixtures."""
from __future__ import annotations

import json
import math
import re
import statistics
from collections import Counter, defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
LIVE = ROOT / "testdata" / "generated" / "live"
REPORT = LIVE / "recommend_report.md"
TAGS_PATH = ROOT / "data" / "champion_tags.json"
INDEX_PATH = LIVE / "index.json"
OUT_AUDIT = LIVE / "build_recommendation_audit.md"
OUT_DETAILS = LIVE / "build_recommendation_case_details.md"
OUT_STATS = LIVE / "_audit_stats.json"

RU_ROLE = {
    "TOP": "топ",
    "JUNGLE": "лес",
    "MIDDLE": "мид",
    "BOTTOM": "бот",
    "UTILITY": "саппорт",
}

# Components skipped if an upgrade is also on the list.
COMPONENT_INTO = {
    "Lost Chapter": {
        "Blackfire Torch",
        "Malignance",
        "Luden's Companion",
        "Archangel's Staff",
    },
    "Tear of the Goddess": {
        "Archangel's Staff",
        "Manamune",
        "Winter's Approach",
        "Fimbulwinter",
    },
    "Winter's Approach": {"Fimbulwinter"},
    "Bramble Vest": {"Thornmail"},
    "Executioner's Calling": {"Mortal Reminder", "Chempunk Chainsword"},
    "Oblivion Orb": {"Morellonomicon"},
    "Kindlegem": set(),
    "Bami's Cinder": {"Sunfire Aegis", "Hollow Radiance"},
    "Pickaxe": set(),
    "BF Sword": set(),
    "B.F. Sword": set(),
    "Blasting Wand": set(),
    "Needlessly Large Rod": set(),
    "Recurve Bow": set(),
    "Caulfield's Warhammer": set(),
    "Serrated Dirk": {
        "Youmuu's Ghostblade",
        "Opportunity",
        "Hubris",
        "The Collector",
        "Serpent's Fang",
        "Umbral Glaive",
    },
}

START_ITEMS = {
    "Doran's Blade",
    "Doran's Ring",
    "Doran's Shield",
    "Cull",
    "Dark Seal",
    "World Atlas",
    "Mosstomper Seedling",
    "Gustwalker Hatchling",
    "Scorchclaw Pup",
    "Health Potion",
    "Refillable Potion",
}

BOOTS = {
    "Boots",
    "Boots of Speed",
    "Plated Steelcaps",
    "Mercury's Treads",
    "Sorcerer's Shoes",
    "Berserker's Greaves",
    "Ionian Boots of Lucidity",
    "Boots of Swiftness",
    "Slightly Magical Footwear",
    "Synchronized Souls",
    "Symbiotic Soles",
}

AP_ITEMS = {
    "Stormsurge",
    "Lich Bane",
    "Shadowflame",
    "Rabadon's Deathcap",
    "Void Staff",
    "Cryptbloom",
    "Liandry's Torment",
    "Blackfire Torch",
    "Malignance",
    "Luden's Companion",
    "Luden's Echo",
    "Horizon Focus",
    "Cosmic Drive",
    "Riftmaker",
    "Nashor's Tooth",
    "Rylai's Crystal Scepter",
    "Morellonomicon",
    "Banshee's Veil",
    "Zhonya's Hourglass",
    "Rod of Ages",
    "Archangel's Staff",
    "Seraph's Embrace",
    "Hextech Rocketbelt",
    "Night Harvester",
    "Bloodletter's Curse",
    "Mejai's Soulstealer",
    "Dark Seal",
    "Lost Chapter",
    "Blackfire Torch",
}
AD_ITEMS = {
    "Trinity Force",
    "Sundered Sky",
    "Black Cleaver",
    "Eclipse",
    "Youmuu's Ghostblade",
    "Opportunity",
    "Hubris",
    "The Collector",
    "Infinity Edge",
    "Kraken Slayer",
    "Yun Tal Wildarrows",
    "Essence Reaver",
    "Navori Flickerblade",
    "Navori Quickblades",
    "Rapid Firecannon",
    "Phantom Dancer",
    "Runaan's Hurricane",
    "Statikk Shiv",
    "Lord Dominik's Regards",
    "Mortal Reminder",
    "Blade of the Ruined King",
    "Wit's End",
    "Guinsoo's Rageblade",
    "Immortal Shieldbow",
    "Bloodthirster",
    "Death's Dance",
    "Maw of Malmortius",
    "Sterak's Gage",
    "Spear of Shojin",
    "Stridebreaker",
    "Titanic Hydra",
    "Ravenous Hydra",
    "Profane Hydra",
    "Serylda's Grudge",
    "Edge of Night",
    "Serpent's Fang",
    "Axiom Arc",
    "Voltaic Cyclosword",
    "Chempunk Chainsword",
    "Manamune",
    "Muramana",
    "Iceborn Gauntlet",
    "Guardian Angel",
    "Mercurial Scimitar",
    "Experimental Hexplate",
    "Terminus",
}
TANK_ITEMS = {
    "Sunfire Aegis",
    "Hollow Radiance",
    "Heartsteel",
    "Jak'Sho, The Protean",
    "Jak'Sho",
    "Frozen Heart",
    "Randuin's Omen",
    "Thornmail",
    "Force of Nature",
    "Kaenic Rookern",
    "Spirit Visage",
    "Unending Despair",
    "Dead Man's Plate",
    "Warmog's Armor",
    "Abyssal Mask",
    "Protoplasm Harness",
    "Winter's Approach",
    "Fimbulwinter",
    "Trailblazer",
    "Locket of the Iron Solari",
    "Knight's Vow",
    "Zeke's Convergence",
    "Bandlepipes",
    "Iceborn Gauntlet",
}
SUPPORT_ENCHANT = {
    "Moonstone Renewer",
    "Staff of Flowing Water",
    "Ardent Censer",
    "Redemption",
    "Mikael's Blessing",
    "Shurelya's Battlesong",
    "Echoes of Helia",
    "Dawncore",
    "Imperial Mandate",
    "Dream Maker",
}
SUPPORT_ENGAGE = {
    "Locket of the Iron Solari",
    "Knight's Vow",
    "Zeke's Convergence",
    "Bandlepipes",
    "Dead Man's Plate",
    "Trailblazer",
    "Celestial Opposition",
    "Solstice Sleigh",
    "Bloodsong",
}
ANTIHEAL = {
    "Morellonomicon",
    "Mortal Reminder",
    "Chempunk Chainsword",
    "Thornmail",
    "Bramble Vest",
    "Oblivion Orb",
    "Executioner's Calling",
    "Healing Reduction",
}
MR_ITEMS = {
    "Force of Nature",
    "Kaenic Rookern",
    "Spirit Visage",
    "Banshee's Veil",
    "Wit's End",
    "Maw of Malmortius",
    "Abyssal Mask",
    "Mercury's Treads",
    "Hollow Radiance",
}
ARMOR_ITEMS = {
    "Frozen Heart",
    "Randuin's Omen",
    "Thornmail",
    "Plated Steelcaps",
    "Zhonya's Hourglass",
    "Dead Man's Plate",
    "Sunfire Aegis",
    "Iceborn Gauntlet",
    "Death's Dance",
    "Guardian Angel",
}
PEN_ARMOR = {
    "Lord Dominik's Regards",
    "Serylda's Grudge",
    "Black Cleaver",
    "Void Staff",
    "The Collector",
    "Kraken Slayer",
    "Blade of the Ruined King",
    "Liandry's Torment",
}
PEN_MAGIC = {"Void Staff", "Cryptbloom", "Shadowflame"}

STRONG_HEAL = {
    "Aatrox",
    "DrMundo",
    "Warwick",
    "Vladimir",
    "Soraka",
    "Sona",
    "Yuumi",
    "Briar",
    "Sylas",
    "Fiora",
    "Irelia",
    "Illaoi",
    "Swain",
    "Zac",
    "Nasus",
    "Volibear",
    "Olaf",
    "Belveth",
    "XinZhao",
    "Kayle",
    "Nilah",
    "Maokai",
    "Senna",
    "Nami",
    "Milio",
    "TahmKench",
    "Sett",
    "Garen",
    "Aatrox",
    "Mundus",
    "RekSai",
    "LeeSin",
    "Ambessa",
}
TANK_CHAMPS = {
    "Ornn",
    "Sion",
    "Malphite",
    "Chogath",
    "ChoGath",
    "Nautilus",
    "Leona",
    "Maokai",
    "Sejuani",
    "Rammus",
    "Zac",
    "DrMundo",
    "Poppy",
    "KSante",
    "Shen",
    "Alistar",
    "Rell",
    "Braum",
    "Amumu",
    "Galio",
    "TahmKench",
    "Nunu",
    "Singed",
    "Skarner",
    "Rammus",
    "Nautilus",
    "Leona",
    "Thresh",
}

# Acceptable first completed legendaries by class (Emerald+). Role refined later.
CLASS_CORES = {
    "mage": {
        "Malignance",
        "Blackfire Torch",
        "Luden's Companion",
        "Rod of Ages",
        "Archangel's Staff",
        "Liandry's Torment",
        "Lich Bane",
        "Horizon Focus",
        "Riftmaker",
        "Nashor's Tooth",
    },
    "assassin": {
        "Youmuu's Ghostblade",
        "Opportunity",
        "Eclipse",
        "Hubris",
        "Profane Hydra",
        "Voltaic Cyclosword",
        "Stormsurge",
        "Lich Bane",
        "Hextech Rocketbelt",
        "Shadowflame",
        "Axiom Arc",
    },
    "fighter": {
        "Trinity Force",
        "Sundered Sky",
        "Eclipse",
        "Stridebreaker",
        "Black Cleaver",
        "Spear of Shojin",
        "Titanic Hydra",
        "Iceborn Gauntlet",
        "Riftmaker",
        "Blade of the Ruined King",
        "Kraken Slayer",
        "Death's Dance",
        "Sterak's Gage",
        "Experimental Hexplate",
        "Hollow Radiance",
        "Sunfire Aegis",
    },
    "tank": {
        "Sunfire Aegis",
        "Hollow Radiance",
        "Heartsteel",
        "Iceborn Gauntlet",
        "Jak'Sho, The Protean",
        "Sundered Sky",
        "Dead Man's Plate",
        "Unending Despair",
        "Fimbulwinter",
        "Winter's Approach",
        "Locket of the Iron Solari",
    },
    "marksman": {
        "Kraken Slayer",
        "Yun Tal Wildarrows",
        "Infinity Edge",
        "The Collector",
        "Essence Reaver",
        "Blade of the Ruined King",
        "Statikk Shiv",
        "Immortal Shieldbow",
        "Trinity Force",
        "Iceborn Gauntlet",
        "Manamune",
        "Nashor's Tooth",
        "Guinsoo's Rageblade",
        "Youmuu's Ghostblade",
        "Opportunity",
        "Hubris",
        "Voltaic Cyclosword",
        "Bloodthirster",
        "Rapid Firecannon",
    },
    "support": {
        "Moonstone Renewer",
        "Locket of the Iron Solari",
        "Blackfire Torch",
        "Malignance",
        "Luden's Companion",
        "Echoes of Helia",
        "Shurelya's Battlesong",
        "Dead Man's Plate",
        "Zeke's Convergence",
        "Bandlepipes",
        "Staff of Flowing Water",
        "Redemption",
        "Knight's Vow",
        "Youmuu's Ghostblade",
        "Umbral Glaive",
        "Opportunity",
        "Liandry's Torment",
        "Imperial Mandate",
        "Trailblazer",
        "Celestial Opposition",
        "Bloodsong",
        "Solstice Sleigh",
        "Dream Maker",
        "Lich Bane",
        "Stormsurge",
    },
}

CHAMP_CORES: dict[tuple[str, str | None], set[str]] = {
    ("Yone", None): {"Trinity Force", "Blade of the Ruined King", "Infinity Edge", "Immortal Shieldbow", "Kraken Slayer"},
    ("Yasuo", None): {"Kraken Slayer", "Infinity Edge", "Blade of the Ruined King", "Immortal Shieldbow", "Phantom Dancer"},
    ("Ezreal", None): {"Trinity Force", "Iceborn Gauntlet", "Manamune", "Muramana", "Essence Reaver"},
    ("Fizz", None): {"Stormsurge", "Lich Bane", "Luden's Companion", "Hextech Rocketbelt"},
    ("Leblanc", None): {"Luden's Companion", "Stormsurge", "Shadowflame", "Lich Bane", "Malignance"},
    ("LeBlanc", None): {"Luden's Companion", "Stormsurge", "Shadowflame", "Lich Bane", "Malignance"},
    ("Akali", None): {"Stormsurge", "Hextech Rocketbelt", "Shadowflame", "Lich Bane", "Riftmaker"},
    ("Kassadin", None): {"Rod of Ages", "Archangel's Staff", "Malignance", "Lich Bane"},
    ("Ryze", None): {"Tear of the Goddess", "Archangel's Staff", "Rod of Ages", "Seraph's Embrace", "Frozen Heart"},
    ("Cassiopeia", None): {"Rylai's Crystal Scepter", "Archangel's Staff", "Riftmaker", "Liandry's Torment", "Rod of Ages"},
    ("Vladimir", None): {"Cosmic Drive", "Riftmaker", "Hextech Rocketbelt", "Shadowflame"},
    ("Kayle", None): {"Nashor's Tooth", "Riftmaker", "Guinsoo's Rageblade", "Lich Bane"},
    ("Teemo", None): {"Liandry's Torment", "Nashor's Tooth", "Malignance", "Blackfire Torch", "Blade of the Ruined King"},
    ("Lillia", None): {"Liandry's Torment", "Riftmaker", "Cosmic Drive", "Blackfire Torch"},
    ("Singed", None): {"Rylai's Crystal Scepter", "Liandry's Torment", "Rod of Ages", "Unending Despair", "Hollow Radiance"},
    ("Poppy", "TOP"): {"Sundered Sky", "Iceborn Gauntlet", "Sunfire Aegis", "Hollow Radiance", "Black Cleaver", "Trinity Force"},
    ("Poppy", "JUNGLE"): {"Sundered Sky", "Dead Man's Plate", "Iceborn Gauntlet", "Eclipse", "Trinity Force"},
    ("Poppy", "UTILITY"): {"Locket of the Iron Solari", "Dead Man's Plate", "Bandlepipes", "Zeke's Convergence", "Knight's Vow"},
    ("Jhin", None): {"The Collector", "Infinity Edge", "Rapid Firecannon", "Youmuu's Ghostblade", "Opportunity"},
    ("Zeri", None): {"Statikk Shiv", "Runaan's Hurricane", "Infinity Edge", "Lord Dominik's Regards", "Experimental Hexplate"},
    ("Smolder", None): {"Essence Reaver", "Infinity Edge", "Rapid Firecannon", "Spear of Shojin", "Trinity Force"},
    ("KogMaw", None): {"Guinsoo's Rageblade", "Blade of the Ruined King", "Nashor's Tooth", "Kraken Slayer", "Liandry's Torment"},
    ("Vayne", None): {"Blade of the Ruined King", "Kraken Slayer", "Guinsoo's Rageblade", "Wit's End", "Terminus"},
    ("Kalista", None): {"Blade of the Ruined King", "Guinsoo's Rageblade", "Runaan's Hurricane"},
    ("MasterYi", None): {"Kraken Slayer", "Blade of the Ruined King", "Guinsoo's Rageblade", "Infinity Edge", "Wit's End"},
    ("Tryndamere", None): {"Navori Flickerblade", "Infinity Edge", "Blade of the Ruined King", "Kraken Slayer"},
    ("Gangplank", None): {"Trinity Force", "Essence Reaver", "Collector", "The Collector", "Infinity Edge", "Lord Dominik's Regards"},
    ("Urgot", None): {"Black Cleaver", "Sterak's Gage", "Titanic Hydra", "Shojin", "Spear of Shojin", "Stridebreaker"},
    ("Sion", None): {"Heartsteel", "Sunfire Aegis", "Hollow Radiance", "Titanic Hydra", "Jak'Sho, The Protean"},
    ("Chogath", None): {"Heartsteel", "Hollow Radiance", "Warmog's Armor", "Riftmaker", "Rookern"},
    ("ChoGath", None): {"Heartsteel", "Hollow Radiance", "Warmog's Armor", "Riftmaker"},
    ("Ornn", None): {"Sunfire Aegis", "Hollow Radiance", "Jak'Sho, The Protean", "Unending Despair"},
    ("Malphite", None): {"Sunfire Aegis", "Iceborn Gauntlet", "Hollow Radiance", "Stormsurge", "Malignance"},
    ("Rammus", None): {"Thornmail", "Sunfire Aegis", "Hollow Radiance", "Jak'Sho, The Protean"},
    ("Senna", "BOTTOM"): {"Essence Reaver", "Infinity Edge", "Rapid Firecannon", "Black Cleaver", "The Collector", "Moonstone Renewer"},
    ("Senna", "UTILITY"): {"Moonstone Renewer", "Echoes of Helia", "Black Chewer", "Black Cleaver", "Imperial Mandate", "Essence Reaver"},
    ("Pyke", None): {"Youmuu's Ghostblade", "Umbral Glaive", "Opportunity", "Axiom Arc", "The Collector"},
    ("Kindred", None): {"Kraken Slayer", "The Collector", "Infinity Edge", "Lord Dominik's Regards", "Trinity Force"},
    ("Graves", None): {"Youmuu's Ghostblade", "The Collector", "Opportunity", "Lord Dominik's Regards", "Essence Reaver"},
    ("Twitch", None): {"The Collector", "Infinity Edge", "Runaan's Hurricane", "Blade of the Ruined King", "Nashor's Tooth"},
    ("Kaisa", None): {"Kraken Slayer", "Nashor's Tooth", "Guinsoo's Rageblade", "Statikk Shiv", "The Collector"},
    ("KaiSa", None): {"Kraken Slayer", "Nashor's Tooth", "Guinsoo's Rageblade", "Statikk Shiv"},
    ("Aphelios", None): {"The Collector", "Infinity Edge", "Lord Dominik's Regards", "Bloodthirster", "Runaan's Hurricane"},
    ("Xayah", None): {"Kraken Slayer", "Infinity Edge", "Navori Flickerblade", "Lord Dominik's Regards", "The Collector"},
    ("Caitlyn", None): {"The Collector", "Infinity Edge", "Rapid Firecannon", "Lord Dominik's Regards", "Youmuu's Ghostblade"},
    ("Jinx", None): {"Yun Tal Wildarrows", "Infinity Edge", "Runaan's Hurricane", "Lord Dominik's Regards", "Kraken Slayer"},
    ("Ashe", None): {"Kraken Slayer", "Blade of the Ruined King", "Runaan's Hurricane", "Lord Dominik's Regards", "Wit's End"},
    ("Sivir", None): {"Yun Tal Wildarrows", "Infinity Edge", "Navori Flickerblade", "Lord Dominik's Regards", "Essence Reaver"},
    ("Lucian", None): {"Essence Reaver", "Navori Flickerblade", "Infinity Edge", "The Collector", "Trinity Force"},
    ("Tristana", None): {"Yun Tal Wildarrows", "Infinity Edge", "Navori Flickerblade", "The Collector", "Kraken Slayer"},
    ("Draven", None): {"The Collector", "Infinity Edge", "Lord Dominik's Regards", "Bloodthirster", "Youmuu's Ghostblade"},
    ("Nilah", None): {"Bloodthirster", "Infinity Edge", "Lord Dominik's Regards", "Navori Flickerblade", "Kraken Slayer"},
    ("Samira", None): {"The Collector", "Infinity Edge", "Bloodthirster", "Lord Dominik's Regards", "Immortal Shieldbow"},
    ("MissFortune", None): {"Youmuu's Ghostblade", "The Collector", "Infinity Edge", "Lord Dominik's Regards", "Opportunity"},
    ("Varus", None): {"Blade of the Ruined King", "Guinsoo's Rageblade", "Nashor's Tooth", "Youmuu's Ghostblade", "Opportunity", "Liandry's Torment"},
    ("Hwei", None): {"Blackfire Torch", "Malignance", "Liandry's Torment", "Luden's Companion", "Horizon Focus"},
    ("Viktor", None): {"Blackfire Torch", "Luden's Companion", "Malignance", "Liandry's Torment", "Horizon Focus"},
    ("Orianna", None): {"Luden's Companion", "Malignance", "Blackfire Torch", "Lich Bane", "Horizon Focus"},
    ("Syndra", None): {"Luden's Companion", "Malignance", "Stormsurge", "Shadowflame"},
    ("Ahri", None): {"Malignance", "Luden's Companion", "Stormsurge", "Lich Bane", "Horizon Focus"},
    ("Azir", None): {"Nashor's Tooth", "Malignance", "Blackfire Torch", "Luden's Companion", "Liandry's Torment"},
    ("Azir", "UTILITY"): {"Blackfire Torch", "Malignance", "Liandry's Torment", "Nashor's Tooth", "Luden's Companion"},
    ("Mel", "UTILITY"): {"Blackfire Torch", "Malignance", "Liandry's Torment", "Luden's Companion", "Horizon Focus"},
    ("Xerath", None): {"Luden's Companion", "Malignance", "Horizon Focus", "Shadowflame", "Blackfire Torch"},
    ("Velkoz", None): {"Luden's Companion", "Malignance", "Liandry's Torment", "Horizon Focus", "Blackfire Torch"},
    ("VelKoz", None): {"Luden's Companion", "Malignance", "Liandry's Torment", "Horizon Focus"},
    ("Brand", None): {"Liandry's Torment", "Blackfire Torch", "Rylai's Crystal Scepter", "Malignance"},
    ("Zyra", None): {"Liandry's Torment", "Blackfire Torch", "Rylai's Crystal Scepter", "Malignance"},
    ("Lux", None): {"Luden's Companion", "Malignance", "Stormsurge", "Horizon Focus", "Moonstone Renewer"},
    ("Morgana", "UTILITY"): {"Blackfire Torch", "Liandry's Torment", "Zhonya's Hourglass", "Malignance", "Locket of the Iron Solari"},
    ("Lulu", None): {"Moonstone Renewer", "Staff of Flowing Water", "Ardent Censer", "Shurelya's Battlesong", "Mikael's Blessing"},
    ("Janna", None): {"Moonstone Renewer", "Staff of Flowing Water", "Ardent Censer", "Redemption", "Shurelya's Battlesong"},
    ("Nami", None): {"Moonstone Renewer", "Staff of Flowing Water", "Imperial Mandate", "Ardent Censer", "Shurelya's Battlesong"},
    ("Soraka", None): {"Moonstone Renewer", "Redemption", "Staff of Flowing Water", "Warmog's Armor", "Dawncore"},
    ("Yuumi", None): {"Moonstone Renewer", "Staff of Flowing Water", "Ardent Censer", "Mikael's Blessing", "Dawncore"},
    ("Milio", None): {"Moonstone Renewer", "Staff of Flowing Water", "Ardent Censer", "Redemption", "Shurelya's Battlesong"},
    ("Sona", None): {"Moonstone Renewer", "Staff of Flowing Water", "Echoes of Helia", "Dawncore", "Ardent Censer"},
    ("Bard", None): {"Dead Man's Plate", "Locket of the Iron Solari", "Shurelya's Battlesong", "Trailblazer", "Imperial Mandate"},
    ("Thresh", None): {"Locket of the Iron Solari", "Zeke's Convergence", "Knight's Vow", "Dead Man's Plate", "Bandlepipes"},
    ("Leona", None): {"Locket of the Iron Solari", "Celestial Opposition", "Knight's Vow", "Zeke's Convergence", "Bandlepipes"},
    ("Nautilus", None): {"Locket of the Iron Solari", "Celestial Opposition", "Knight's Vow", "Zeke's Convergence"},
    ("Rell", None): {"Locket of the Iron Solari", "Zeke's Convergence", "Knight's Vow", "Bandlepipes", "Trailblazer"},
    ("Braum", None): {"Locket of the Iron Solari", "Knight's Vow", "Zeke's Convergence", "Frozen Heart"},
    ("Alistar", None): {"Locket of the Iron Solari", "Celestial Opposition", "Knight's Vow", "Trailblazer"},
    ("Rakan", None): {"Shurelya's Battlesong", "Locket of the Iron Solari", "Zeke's Convergence", "Trailblazer"},
    ("Blitzcrank", None): {"Locket of the Iron Solari", "Zeke's Convergence", "Dead Man's Plate", "Trailblazer"},
    ("Darius", None): {"Trinity Force", "Sterak's Gage", "Stridebreaker", "Black Cleaver", "Spear of Shojin"},
    ("Garen", None): {"Stridebreaker", "Black Cleaver", "Trinity Force", "Sterak's Gage", "Dead Man's Plate"},
    ("Illaoi", None): {"Iceborn Gauntlet", "Black Cleaver", "Sterak's Gage", "Sundered Sky", "Shojin"},
    ("Aatrox", None): {"Sundered Sky", "Eclipse", "Spear of Shojin", "Death's Dance", "Sterak's Gage"},
    ("Sett", None): {"Stridebreaker", "Heartsteel", "Sterak's Gage", "Black Cleaver", "Overlord's Bloodmail"},
    ("Camille", None): {"Trinity Force", "Ravenous Hydra", "Sundered Sky", "Eclipse", "Death's Dance"},
    ("Fiora", None): {"Ravenous Hydra", "Trinity Force", "Death's Dance", "Hullbreaker", "Eclipse"},
    ("Irelia", None): {"Blade of the Ruined King", "Wit's End", "Sundered Sky", "Death's Dance", "Kraken Slayer"},
    ("Jax", None): {"Trinity Force", "Sundered Sky", "Wit's End", "Blade of the Ruined King", "Nashor's Tooth"},
    ("Riven", None): {"Eclipse", "Sundered Sky", "Death's Dance", "Spear of Shojin", "Black Cleaver"},
    ("Ambessa", None): {"Eclipse", "Sundered Sky", "Spear of Shojin", "Death's Dance", "Black Cleaver"},
    ("Renekton", None): {"Eclipse", "Sundered Sky", "Black Cleaver", "Spear of Shojin", "Sterak's Gage"},
    ("KSante", None): {"Iceborn Gauntlet", "Jak'Sho, The Protean", "Hollow Radiance", "Sundered Sky", "Dead Man's Plate"},
    ("Gwen", None): {"Riftmaker", "Nashor's Tooth", "Cosmic Drive", "Shadowflame", "Zhonya's Hourglass"},
    ("Mordekaiser", None): {"Riftmaker", "Rylai's Crystal Scepter", "Spirit Visage", "Jak'Sho, The Protean"},
    ("Nasus", None): {"Iceborn Gauntlet", "Trinity Force", "Sundered Sky", "Frozen Heart", "Spirit Visage"},
    ("Udyr", None): {"Liandry's Torment", "Iceborn Gauntlet", "Trinity Force", "Dead Man's Plate", "Stridebreaker"},
    ("Volibear", None): {"Nashor's Tooth", "Riftmaker", "Iceborn Gauntlet", "Spirit Visage", "Hollow Radiance"},
    ("Shyvana", None): {"Nashor's Tooth", "Riftmaker", "Ravenous Hydra", "Spear of Shojin", "Liandry's Torment"},
    ("Gragas", None): {"Rod of Ages", "Cosmic Drive", "Lich Bane", "Stormsurge", "Iceborn Gauntlet"},
    ("Zac", None): {"Sunfire Aegis", "Hollow Radiance", "Jak'Sho, The Protean", "Spirit Visage", "Unending Despair"},
    ("Sejuani", None): {"Sunfire Aegis", "Hollow Radiance", "Jak'Sho, The Protean", "Frozen Heart", "Locket of the Iron Solari"},
    ("Amumu", None): {"Liandry's Torment", "Sunfire Aegis", "Jak'Sho, The Protean", "Abyssal Mask"},
    ("Nunu", None): {"Liandry's Torment", "Sunfire Aegis", "Jak'Sho, The Protean", "Dead Man's Plate"},
    ("Skarner", None): {"Trinity Force", "Iceborn Gauntlet", "Jak'Sho, The Protean", "Black Cleaver"},
    ("Rengar", None): {"Youmuu's Ghostblade", "Opportunity", "Profane Hydra", "Eclipse", "Hubris"},
    ("Khazix", None): {"Youmuu's Ghostblade", "Opportunity", "Profane Hydra", "Hubris", "Axiom Arc"},
    ("KhaZix", None): {"Youmuu's Ghostblade", "Opportunity", "Profane Hydra", "Hubris"},
    ("Talon", None): {"Youmuu's Ghostblade", "Opportunity", "Profane Hydra", "Hubris", "Axiom Arc"},
    ("Zed", None): {"Youmuu's Ghostblade", "Voltaic Cyclosword", "Opportunity", "Eclipse", "Profane Hydra"},
    ("Qiyana", None): {"Youmuu's Ghostblade", "Opportunity", "Profane Hydra", "Hubris", "Eclipse"},
    ("Kayn", None): {"Eclipse", "Black Cleaver", "Sundered Sky", "Opportunity", "Youmuu's Ghostblade", "Spear of Shojin"},
    ("Belveth", None): {"Kraken Slayer", "Blade of the Ruined King", "Death's Dance", "Wit's End", "Trinity Force"},
    ("Viego", None): {"Trinity Force", "Kraken Slayer", "Blade of the Ruined King", "Sundered Sky", "Death's Dance"},
    ("LeeSin", None): {"Eclipse", "Sundered Sky", "Black Cleaver", "Death's Dance", "Maw of Malmortius"},
    ("XinZhao", None): {"Sundered Sky", "Black Cleaver", "Eclipse", "Death's Dance", "Titanic Hydra"},
    ("Vi", None): {"Sundered Sky", "Trinity Force", "Black Cleaver", "Eclipse", "Death's Dance"},
    ("JarvanIV", None): {"Sundered Sky", "Black Cleaver", "Eclipse", "Dead Man's Plate", "Spear of Shojin"},
    ("Hecarim", None): {"Trinity Force", "Spear of Shojin", "Black Cleaver", "Eclipse", "Sundered Sky"},
    ("Nocturne", None): {"Eclipse", "Black Cleaver", "Profane Hydra", "Sundered Sky", "Stridebreaker"},
    ("Warwick", None): {"Blade of the Ruined King", "Stridebreaker", "Titanic Hydra", "Spirit Visage", "Wit's End"},
    ("Elise", None): {"Stormsurge", "Lich Bane", "Shadowflame", "Nashor's Tooth", "Malignance"},
    ("Evelynn", None): {"Lich Bane", "Rabadon's Deathcap", "Mejai's Soulstealer", "Void Staff", "Stormsurge"},
    ("Diana", None): {"Nashor's Tooth", "Lich Bane", "Shadowflame", "Riftmaker", "Stormsurge", "Hextech Rocketbelt"},
    ("Ekko", None): {"Lich Bane", "Nashor's Tooth", "Hextech Rocketbelt", "Rabadon's Deathcap", "Stormsurge"},
    ("Nidalee", None): {"Lich Bane", "Stormsurge", "Shadowflame", "Rabadon's Deathcap", "Malignance"},
    ("Karthus", None): {"Liandry's Torment", "Malignance", "Blackfire Torch", "Rabadon's Deathcap", "Void Staff"},
    ("Fiddlesticks", None): {"Malignance", "Liandry's Torment", "Zhonya's Hourglass", "Rylai's Crystal Scepter"},
    ("Ivern", None): {"Moonstone Renewer", "Redemption", "Staff of Flowing Water", "Imperial Mandate", "Shurelya's Battlesong"},
    ("RekSai", None): {"Titanic Hydra", "Stridebreaker", "Black Cleaver", "Sundered Sky", "Sterak's Gage"},
    ("Shaco", None): {"Youmuu's Ghostblade", "Opportunity", "The Collector", "Nashor's Tooth", "Lich Bane"},
    ("Lillia", "JUNGLE"): {"Liandry's Torment", "Riftmaker", "Cosmic Drive", "Zhonya's Hourglass"},
    ("Briar", None): {"The Collector", "Profane Hydra", "Eclipse", "Sundered Sky", "Black Cleaver", "Titanic Hydra"},
    ("Naafiri", None): {"Youmuu's Ghostblade", "Opportunity", "Eclipse", "Profane Hydra", "Serylda's Grudge"},
    ("Hwei", "UTILITY"): {"Blackfire Torch", "Liandry's Torment", "Malignance", "Luden's Companion", "Horizon Focus"},
    ("Taliyah", "UTILITY"): {"Blackfire Torch", "Malignance", "Liandry's Torment", "Luden's Companion"},
    ("Teemo", "UTILITY"): {"Liandry's Torment", "Nashor's Tooth", "Blackfire Torch", "Malignance", "Morellonomicon"},
    ("Heimerdinger", "UTILITY"): {"Liandry's Torment", "Blackfire Torch", "Rylai's Crystal Scepter", "Malignance"},
    ("Ziggs", "UTILITY"): {"Blackfire Torch", "Liandry's Torment", "Luden's Companion", "Horizon Focus"},
    ("Xerath", "UTILITY"): {"Luden's Companion", "Malignance", "Horizon Focus", "Blackfire Torch", "Liandry's Torment"},
    ("Lux", "UTILITY"): {"Luden's Companion", "Malignance", "Moonstone Renewer", "Shurelya's Battlesong", "Horizon Focus"},
    ("Karma", "UTILITY"): {"Moonstone Renewer", "Shurelya's Battlesong", "Imperial Mandate", "Staff of Flowing Water", "Malignance"},
    ("Karma", "MIDDLE"): {"Malignance", "Luden's Companion", "Horizon Focus", "Cosmic Drive"},
    ("Swain", None): {"Blackfire Torch", "Rylai's Crystal Scepter", "Liandry's Torment", "Riftmaker", "Rod of Ages"},
    ("Rumble", None): {"Liandry's Torment", "Riftmaker", "Zhonya's Hourglass", "Bloodletter's Curse", "Shadowflame"},
    ("Gnar", None): {"Trinity Force", "Black Cleaver", "Sundered Sky", "Sterak's Gage", "Iceborn Gauntlet"},
    ("Jayce", None): {"Youmuu's Ghostblade", "Muramana", "Opportunity", "Serylda's Grudge", "Eclipse", "Manamune"},
    ("Pantheon", None): {"Eclipse", "Sundered Sky", "Black Cleaver", "Spear of Shojin", "Youmuu's Ghostblade"},
    ("Sylas", None): {"RoA", "Rod of Ages", "Riftmaker", "Zhonya's Hourglass", "Cosmic Drive", "Lich Bane"},
    ("Yasuo", "JUNGLE"): {"Kraken Slayer", "Infinity Edge", "Blade of the Ruined King", "Immortal Shieldbow"},
    ("Yone", "JUNGLE"): {"Trinity Force", "Blade of the Ruined King", "Infinity Edge", "Immortal Shieldbow"},
    ("MasterYi", "TOP"): {"Kraken Slayer", "Blade of the Ruined King", "Guinsoo's Rageblade", "Wit's End", "Infinity Edge"},
    ("Nasus", "JUNGLE"): {"Iceborn Gauntlet", "Trinity Force", "Sundered Sky", "Frozen Heart", "Spirit Visage"},
}


def load_tags() -> dict:
    data = json.loads(TAGS_PATH.read_text(encoding="utf-8"))
    return data.get("champions") or {}


def champ_info(tags: dict, cid: str, pos: str) -> tuple[str, set[str]]:
    rec = tags.get(cid) or {}
    by_role = rec.get("by_role") or {}
    role_rec = by_role.get(pos) or rec
    primary = role_rec.get("primary") or rec.get("primary") or "unknown"
    tset = set(role_rec.get("tags") or rec.get("tags") or [])
    tset.add(primary)
    return primary, tset


def is_ap_champ(primary: str, tset: set[str]) -> bool:
    if "ap" in tset and "ad" not in tset:
        return True
    if primary in {"mage"}:
        return True
    if primary == "assassin" and "ap" in tset:
        return True
    if primary == "fighter" and "ap" in tset:
        return True
    if primary == "support" and "ap" in tset and "ad" not in tset:
        return True
    return False


def is_ad_champ(primary: str, tset: set[str]) -> bool:
    if "ad" in tset and "ap" not in tset:
        return True
    if primary in {"marksman"}:
        return True
    if primary == "assassin" and "ap" not in tset:
        return True
    if primary == "fighter" and "ap" not in tset:
        return True
    return False


def cores_for(cid: str, pos: str, primary: str) -> set[str]:
    for key in ((cid, pos), (cid, None)):
        if key in CHAMP_CORES:
            return set(CHAMP_CORES[key])
    return set(CLASS_CORES.get(primary, set()))


def parse_kv_line(line: str) -> dict[str, float]:
    out = {}
    for m in re.finditer(r"([A-Za-z_]+)=([0-9.]+)", line):
        out[m.group(1)] = float(m.group(2))
    return out


def parse_comp(s: str) -> dict[str, str]:
    out = {}
    if not s:
        return out
    for part in s.split(","):
        part = part.strip()
        if ":" in part:
            pos, cid = part.split(":", 1)
            out[pos] = cid
    return out


def parse_report(text: str) -> list[dict]:
    chunks = re.split(r"(?m)^## ", text)
    cases = []
    for chunk in chunks[1:]:
        lines = chunk.splitlines()
        fname = lines[0].strip()
        body = "\n".join(lines[1:])
        active_m = re.search(r"- active: `([^`]+)` @ `([^`]+)`", body)
        allies_m = re.search(r"- allies: `([^`]*)`", body)
        enemies_m = re.search(r"- enemies: `([^`]*)`", body)
        seed_m = re.search(r"Active: (\S+) \((\w+)\) seed=(\S+) gold=([0-9.]+) allyFront=([0-9.]+)", body)
        press_m = re.search(r"Pressure: ([^\n]+)", body)
        next_m = re.search(r"Next item: (.+?) \((\d+)\) priority=([0-9.]+) \[([^\]]+)\]", body)
        build = []
        in_build = False
        reasons = []
        in_reasons = False
        for ln in body.splitlines():
            if ln.startswith("Build order"):
                in_build = True
                in_reasons = False
                continue
            if ln.startswith("Reasons:"):
                in_build = False
                in_reasons = True
                continue
            if in_build:
                bm = re.match(r"^\s+(\d+)\s+(.+?)\s+\[([^\]]+)\](\s+✓)?\s*$", ln)
                if bm:
                    build.append(
                        {
                            "priority": int(bm.group(1)),
                            "name": bm.group(2).strip(),
                            "role": bm.group(3).strip(),
                            "owned": bool(bm.group(4)),
                        }
                    )
                elif ln.strip() == "" or ln.startswith("```"):
                    in_build = False
            if in_reasons and ln.strip().startswith("•"):
                reasons.append(ln.strip()[1:].strip())
        if not active_m or not seed_m:
            cases.append({"file": fname, "parse_error": True, "raw": body[:500]})
            continue
        cases.append(
            {
                "file": fname,
                "parse_error": False,
                "champion": active_m.group(1),
                "position": active_m.group(2),
                "allies": parse_comp(allies_m.group(1) if allies_m else ""),
                "enemies": parse_comp(enemies_m.group(1) if enemies_m else ""),
                "seed": seed_m.group(3),
                "gold": float(seed_m.group(4)),
                "ally_front": float(seed_m.group(5)),
                "pressure": parse_kv_line(press_m.group(1) if press_m else ""),
                "next_item": {
                    "name": next_m.group(1).strip(),
                    "id": int(next_m.group(2)),
                    "priority": float(next_m.group(3)),
                    "role": next_m.group(4),
                }
                if next_m
                else None,
                "build": build,
                "reasons": reasons,
            }
        )
    return cases


def item_kind(name: str) -> str:
    if name in START_ITEMS or name.endswith("Seedling") or name.endswith("Pup") or name.endswith("Hatchling"):
        return "start"
    if name in BOOTS:
        return "boots"
    if name in COMPONENT_INTO:
        return "component"
    if name in SUPPORT_ENCHANT:
        return "enchant"
    if name in SUPPORT_ENGAGE:
        return "engage"
    if name in TANK_ITEMS:
        return "tank"
    if name in AP_ITEMS:
        return "ap"
    if name in AD_ITEMS:
        return "ad"
    return "other"


def realized_six(build: list[dict]) -> tuple[list[str], list[str], list[str]]:
    """Strict buy order into 6 slots (start sold). Returns (final6, extras, flags)."""
    names = [b["name"] for b in build]
    upgrades = set()
    for n in names:
        upgrades |= COMPONENT_INTO.get(n, set())
    have_upgrade = {c for c in names if any(c in COMPONENT_INTO.get(x, set()) for x in names)}
    flags = []
    boots_seen = []
    final = []
    skipped_comp = []
    for b in build:
        n = b["name"]
        if n in START_ITEMS or item_kind(n) == "start":
            continue
        if n in COMPONENT_INTO and (COMPONENT_INTO[n] & set(names)):
            skipped_comp.append(n)
            continue
        if n in BOOTS:
            boots_seen.append(n)
            if len(boots_seen) > 1:
                flags.append(f"second_boots:{n}")
                continue
        if n not in final:
            final.append(n)
    extras = final[6:]
    final6 = final[:6]
    if len(boots_seen) > 1:
        flags.append("multiple_boots")
    return final6, extras, flags


def situation(tags: dict, case: dict) -> dict:
    pos = case["position"]
    cid = case["champion"]
    primary, tset = champ_info(tags, cid, pos)
    enemies = case["enemies"]
    allies = case["allies"]
    e_ap = e_ad = e_tank = e_heal = e_dive = e_crit = 0
    e_burst_ap = 0
    for epos, ecid in enemies.items():
        ep, et = champ_info(tags, ecid, epos)
        if is_ap_champ(ep, et) or "ap" in et:
            e_ap += 1
            if "burst" in et:
                e_burst_ap += 1
        if is_ad_champ(ep, et) or "ad" in et:
            e_ad += 1
        if ep == "tank" or "tank" in et or "juggernaut" in et or ecid in TANK_CHAMPS:
            e_tank += 1
        if "sustain" in et or ecid in STRONG_HEAL:
            e_heal += 1
        if "dive" in et or "engage" in et or "all_in" in et:
            e_dive += 1
        if ep == "marksman" or epos == "BOTTOM" or "crit" in et:
            e_crit += 1
    a_front = a_ap = a_ad = 0
    for apos, acid in allies.items():
        if acid == cid:
            continue
        ap, at = champ_info(tags, acid, apos)
        if ap == "tank" or "tank" in at or "engage" in at or "juggernaut" in at:
            a_front += 1
        if is_ap_champ(ap, at):
            a_ap += 1
        if is_ad_champ(ap, at):
            a_ad += 1
    return {
        "primary": primary,
        "tags": tset,
        "e_ap": e_ap,
        "e_ad": e_ad,
        "e_tank": e_tank,
        "e_heal": e_heal,
        "e_dive": e_dive,
        "e_crit": e_crit,
        "e_burst_ap": e_burst_ap,
        "a_front": a_front,
        "a_ap": a_ap,
        "a_ad": a_ad,
        "need_antiheal": e_heal >= 2 or (e_heal >= 1 and cid in STRONG_HEAL),
        "need_mr": e_ap >= 3 or e_burst_ap >= 2,
        "need_armor": e_ad >= 3 or e_crit >= 2,
        "need_pen": e_tank >= 2,
        "need_front_from_you": a_front == 0 and primary in {"tank", "fighter", "support"},
    }


def score_item(name: str, cid: str, pos: str, sit: dict, idx: int, n_final: int) -> tuple[int, list[str]]:
    """0-10 for one item in the realized path. idx 0 is first completed buy."""
    notes = []
    primary = sit["primary"]
    tset = sit["tags"]
    kind = item_kind(name)
    cores = cores_for(cid, pos, primary)
    ap_c = is_ap_champ(primary, tset)
    ad_c = is_ad_champ(primary, tset)

    score = 7  # default "plausible Emerald item"

    # Identity clashes
    if kind == "ap" and ad_c and name not in {"Wit's End", "Hextech Rocketbelt", "Nashor's Tooth", "Guinsoo's Rageblade", "Terminus", "Zhonya's Hourglass", "Banshee's Veil", "Kaenic Rookern"}:
        if name in {"Zhonya's Hourglass", "Banshee's Veil"}:
            score = 7
        elif primary == "marksman" and name in {"Nashor's Tooth", "Guinsoo's Rageblade"}:
            score = 8
        else:
            score = 2
            notes.append(f"{name} — AP-предмет на AD-чемпионе")
    if kind == "ad" and ap_c and name not in {"Iceborn Gauntlet", "Black Cleaver", "Zhonya's Hourglass", "Guardian Angel", "Trinity Force"}:
        if primary == "mage" and name in AD_ITEMS and name not in {"Nashor's Tooth"}:
            # Nashor already AP-tagged mostly
            if name not in {"Nashor's Tooth", "Guinsoo's Rageblade", "Wit's End", "Terminus"}:
                score = 2
                notes.append(f"{name} — AD-предмет на AP-чемпионе")
    if kind == "enchant" and pos != "UTILITY" and primary not in {"support"}:
        score = min(score, 3)
        notes.append(f"{name} — энчантер-предмет вне саппорта")
    if kind == "engage" and pos not in {"UTILITY", "JUNGLE", "TOP"} and primary not in {"tank", "support", "fighter"}:
        if name in {"Locket of the Iron Solari", "Knight's Vow"} and pos in {"MIDDLE", "BOTTOM"}:
            score = min(score, 3)
            notes.append(f"{name} — саппорт-танк предмет на керри")

    # Role: jungle pet already excluded from realized 6
    if name in cores:
        score = max(score, 8)
        if idx == 0:
            score = 9
            notes.append(f"{name} — корректный identity core")
    elif kind == "boots":
        score = boots_score(name, cid, pos, sit, notes)
    elif kind == "tank" and primary in {"mage", "assassin", "marksman"} and pos != "UTILITY":
        if name in {"Zhonya's Hourglass"} or name in ARMOR_ITEMS and name == "Zhonya's Hourglass":
            score = 7
        elif name in {"Frozen Heart", "Randuin's Omen", "Thornmail", "Force of Nature", "Kaenic Rookern"}:
            score = 4
            notes.append(f"{name} — полный танк на {primary}, обычно слишком рано/не тот слот")
        else:
            score = min(score, 5)
    elif kind == "tank" and primary in {"tank", "fighter", "support"}:
        score = max(score, 7)
        if sit["need_armor"] and name in ARMOR_ITEMS:
            score = max(score, 8)
            notes.append(f"{name} — броня под AD-состав")
        if sit["need_mr"] and name in MR_ITEMS:
            score = max(score, 8)
            notes.append(f"{name} — MR под AP-состав")
    elif kind == "ap" and ap_c:
        score = max(score, 7)
        if sit["need_pen"] and name in {"Liandry's Torment", "Void Staff", "Cryptbloom", "Blackfire Torch"}:
            score = max(score, 8)
        if not sit["need_pen"] and name == "Liandry's Torment" and sit["e_tank"] == 0:
            score = min(score, 6)
            notes.append("Liandry без танков во врагах — слабый situational")
        if not sit["need_antiheal"] and name == "Morellonomicon":
            score = min(score, 6)
            notes.append("Morello без сильного heal-давления")
    elif kind == "ad" and ad_c:
        score = max(score, 7)
        if sit["need_pen"] and name in PEN_ARMOR:
            score = max(score, 8)
            notes.append(f"{name} — pen/anti-tank под {sit['e_tank']} танков")
        if sit["e_tank"] == 0 and name in {"Lord Dominik's Regards", "Blade of the Ruined King"}:
            score = min(score, 6)
            notes.append(f"{name} без танков — не лучший слот")

    # Situational defensive on carries
    if name == "Zhonya's Hourglass":
        if sit["e_dive"] >= 2 or sit["e_ad"] >= 3:
            score = max(score, 8)
            notes.append("Zhonya оправдана vs dive/AD")
        if idx == 0 and primary in {"mage", "assassin"}:
            score = min(score, 5)
            notes.append("Zhonya первым легендариком — слишком рано, нет спайка урона")
        if idx == 1 and primary in {"assassin", "mage"}:
            score = min(score, 6)
            notes.append("Zhonya вторым предметом до обувки/второго оффенсива")
    if name == "Banshee's Veil":
        if sit["e_ap"] >= 2:
            score = max(score, 7)
        else:
            score = min(score, 5)
            notes.append("Banshee при слабом AP-давлении")
        if name == "Banshee's Veil" and "Zhonya's Hourglass" in (sit.get("_final") or []):
            pass

    # Dual resist in one path
    return max(0, min(10, score)), notes


def boots_score(name: str, cid: str, pos: str, sit: dict, notes: list[str]) -> int:
    primary = sit["primary"]
    tset = sit["tags"]
    if cid == "Cassiopeia":
        notes.append("Cassiopeia не покупает боты")
        return 1
    if cid in {"Yone", "Yasuo"} and name != "Berserker's Greaves":
        notes.append(f"{cid} почти всегда хочет Berserker's Greaves, не {name}")
        return 4
    if pos == "BOTTOM" and primary == "marksman":
        if name == "Berserker's Greaves":
            return 8
        if name == "Plated Steelcaps" and sit["e_ad"] >= 3:
            notes.append("Steelcaps на ADC vs тяжёлый AD — рабочий situational")
            return 7
        if name == "Mercury's Treads" and (sit["e_ap"] >= 3 or sit["e_dive"] >= 3):
            return 7
        notes.append(f"{name} на ADC вместо Berserker's/ситуативных резист-ботов")
        return 5
    if is_ap_champ(primary, tset) and pos != "UTILITY":
        if name == "Sorcerer's Shoes":
            return 8
        if name == "Mercury's Treads" and sit["need_mr"]:
            notes.append("Mercs вместо Sorcs — можно vs тяжёлый AP/CC, теряется пен")
            return 6
        if name == "Ionian Boots of Lucidity":
            return 7
        notes.append(f"{name} на AP-чемпионе")
        return 4
    if primary in {"assassin"} and "ad" in tset:
        if name in {"Ionian Boots of Lucidity", "Boots of Swiftness"}:
            return 8
        if name == "Mercury's Treads" and sit["need_mr"]:
            return 7
        if name == "Plated Steelcaps" and sit["need_armor"]:
            return 7
        return 5
    if primary in {"fighter", "tank"} or pos == "JUNGLE":
        if name == "Plated Steelcaps" and sit["e_ad"] >= sit["e_ap"]:
            return 8
        if name == "Mercury's Treads" and sit["e_ap"] >= sit["e_ad"]:
            return 8
        if name == "Plated Steelcaps" and sit["e_ap"] >= 3:
            notes.append("Steelcaps при 3+ AP — слабо")
            return 4
        if name == "Mercury's Treads" and sit["e_ad"] >= 4 and sit["e_ap"] <= 1:
            notes.append("Mercs при почти полном AD — слабо")
            return 4
        if name == "Ionian Boots of Lucidity":
            return 7
        return 6
    if pos == "UTILITY":
        if name in {"Ionian Boots of Lucidity", "Boots of Swiftness"}:
            return 8
        if name == "Mercury's Treads" and sit["need_mr"]:
            return 8
        if name == "Plated Steelcaps" and sit["need_armor"]:
            return 8
        if name == "Sorcerer's Shoes" and is_ap_champ(primary, tset):
            return 8
        return 6
    return 6


def score_start(name: str, cid: str, pos: str, sit: dict) -> tuple[int, str]:
    primary = sit["primary"]
    tset = sit["tags"]
    if pos == "JUNGLE":
        want = "Gustwalker Hatchling"
        if is_ap_champ(primary, tset) or primary == "mage":
            want = "Scorchclaw Pup"
        if primary == "tank" or ("tank" in tset and primary != "fighter") or cid in {
            "Rammus",
            "Amumu",
            "Sejuani",
            "Nunu",
            "Zac",
            "Maokai",
            "Poppy",
            "Rell",
            "Skarner",
        }:
            want = "Mosstomper Seedling"
        if cid in {"Yone", "Yasuo", "MasterYi", "Belveth", "Kindred", "Graves", "Khazix", "Rengar", "Shaco", "LeeSin", "Viego", "Kayn", "Nocturne", "Talon", "Qiyana"}:
            want = "Gustwalker Hatchling"
        if cid in {"Lillia", "Karthus", "Evelynn", "Elise", "Nidalee", "Fiddlesticks", "Brand", "Sylvia", "Diana", "Ekko", "Gwen", "Shyvana", "Mordekaiser", "Rumble"}:
            want = "Scorchclaw Pup"
        if name == want:
            return 9, f"пет {name} совпадает с ролью"
        if name == "Mosstomper Seedling" and want != name:
            return 3, f"Mosstomper на {cid} лес — нужен {want}"
        if name == "Scorchclaw Pup" and want == "Gustwalker Hatchling":
            return 4, f"Scorchclaw на AD-лесаря {cid}"
        return 5, f"пет {name}, ожидался {want}"
    if pos == "UTILITY":
        if name == "World Atlas":
            if cid in {"Senna", "Pyke"}:
                return 6, "Senna/Pyke часто берут квест, но не всегда Atlas-энчантер"
            return 9, "World Atlas корректен на саппорте"
        return 4, f"{name} вместо World Atlas на саппорте"
    if name == "Doran's Ring" and is_ap_champ(primary, tset):
        return 9, "Doran's Ring"
    if name == "Doran's Blade" and is_ad_champ(primary, tset) and primary != "tank":
        return 9, "Doran's Blade"
    if name == "Doran's Shield" and primary in {"tank", "fighter"}:
        return 8, "Doran's Shield"
    if name == "Doran's Shield" and primary in {"mage", "marksman", "assassin"}:
        return 5, "Shield на стеклянном керри — только в тяжёлую лайн"
    if name == "Doran's Ring" and is_ad_champ(primary, tset) and "ap" not in tset:
        return 2, "Doran's Ring на AD-чемпионе"
    if name == "Doran's Blade" and is_ap_champ(primary, tset) and "ad" not in tset:
        return 2, "Doran's Blade на AP-чемпионе"
    return 6, name


def analyze_case(tags: dict, case: dict) -> dict:
    cid = case["champion"]
    pos = case["position"]
    sit = situation(tags, case)
    build = case["build"]
    start = next((b["name"] for b in build if b["role"] == "start" or b["name"] in START_ITEMS or item_kind(b["name"]) == "start"), None)
    final6, extras, flags = realized_six(build)
    sit["_final"] = final6
    item_scores = []
    all_notes = []
    critical = []
    weaknesses = []
    strengths = []
    missing = []
    alts = []

    st_score, st_note = (6, "")
    if start:
        st_score, st_note = score_start(start, cid, pos, sit)
        if st_score <= 4:
            critical.append(st_note)
        elif st_score <= 6:
            weaknesses.append(st_note)
        else:
            strengths.append(st_note)

    for i, n in enumerate(final6):
        sc, notes = score_item(n, cid, pos, sit, i, len(final6))
        item_scores.append(sc)
        all_notes.extend(notes)
        if sc <= 3:
            critical.append(f"{n} (слот {i + 1}): " + ("; ".join(notes) or "не соответствует чемпиону/роли"))
        elif sc <= 6:
            weaknesses.extend(notes or [f"{n} неоптимален в слоте {i + 1}"])
        elif notes:
            strengths.extend(notes)

    # Dual MR / dual armor in realized 6
    mr_big = [n for n in final6 if n in MR_ITEMS and n not in BOOTS]
    ar_big = [n for n in final6 if n in ARMOR_ITEMS and n not in BOOTS]
    if len(mr_big) >= 2 and sit["e_ap"] <= 1:
        critical.append(f"В 6 слотах сразу {', '.join(mr_big)} при {sit['e_ap']} AP — лишний MR")
        item_scores.append(3)
    elif len(mr_big) >= 2 and sit["e_ap"] == 2:
        weaknesses.append(f"Два MR-предмета ({', '.join(mr_big)}) при 2 AP — можно один + урон")
        item_scores.append(5)
    if len(ar_big) >= 3:
        weaknesses.append(f"Перекос в броню: {', '.join(ar_big)}")

    if "multiple_boots" in flags:
        critical.append("В списке больше одной пары ботов — при строгом порядке это сломанный buy path")
        item_scores.append(2)

    # Antiheal — only items that actually enter the 6-slot path
    has_anti = any(n in ANTIHEAL or "Executioner" in n or "Oblivion" in n for n in final6)
    anti_in_tail = any(n in ANTIHEAL or "Executioner" in n or "Oblivion" in n for n in extras)
    if sit["need_antiheal"] and not has_anti:
        missing.append("anti-heal (Morello / Mortal Reminder / Thornmail / Chempunk)")
        if anti_in_tail:
            critical.append(
                f"Anti-heal есть в хвосте списка, но при строгом порядке не входит в 6 слотов ({sit['e_heal']} heal-угроз)"
            )
            item_scores.append(4)
        elif pos in {"BOTTOM", "MIDDLE", "TOP", "JUNGLE"}:
            critical.append(f"Нет anti-heal при {sit['e_heal']} heal-угрозах")
            item_scores.append(4)
        else:
            weaknesses.append(f"Нет anti-heal при {sit['e_heal']} heal-угрозах")
            item_scores.append(6)

    # Pen vs tanks
    has_pen = any(n in PEN_ARMOR or n in {"Liandry's Torment", "Kraken Slayer", "Blade of the Ruined King", "Black Cleaver", "Void Staff"} for n in final6)
    if sit["need_pen"] and not has_pen and sit["primary"] in {"marksman", "mage", "assassin", "fighter"}:
        missing.append("anti-tank (LDR / Cleaver / Liandry / BotRK / Kraken / Void)")
        weaknesses.append(f"Против {sit['e_tank']} танков нет pen/%HP")
        item_scores.append(5)

    # MR/armor holes on melee frontliners
    if sit["need_mr"] and not any(n in MR_ITEMS for n in final6):
        if sit["primary"] in {"fighter", "tank"} or pos == "JUNGLE":
            missing.append("MR (Force of Nature / Kaenic / Mercs / Visage)")
            weaknesses.append("Нет MR против AP-тяжёлого состава")
            item_scores.append(5)
    if sit["need_armor"] and not any(n in ARMOR_ITEMS for n in final6):
        if sit["primary"] in {"fighter", "tank"}:
            missing.append("броня (Steelcaps / Thornmail / Frozen Heart / DD)")
            weaknesses.append("Нет брони против AD-тяжёлого состава")
            item_scores.append(5)

    # First legendary identity
    first_leg = next((n for n in final6 if n not in BOOTS), None)
    cores = cores_for(cid, pos, sit["primary"])
    if first_leg and cores and first_leg not in cores and first_leg not in BOOTS:
        if item_kind(first_leg) in {"ap", "ad", "tank", "enchant", "engage"}:
            weaknesses.append(f"Первый легендарик {first_leg} не из identity-core {cid} ({', '.join(sorted(list(cores))[:5])})")
            item_scores.append(5)

    # Next item at 500g
    nxt = (case.get("next_item") or {}).get("name")
    if nxt in BOOTS and pos in {"BOTTOM", "MIDDLE"} and sit["primary"] in {"marksman", "mage", "assassin"}:
        weaknesses.append(f"Next item = {nxt} при 500g — обувка раньше кирки/главы, слабый ранний спайк")
        item_scores.append(6)
    if nxt == "Zhonya's Hourglass":
        critical.append("Next item Zhonya с 500g — нет урона на первом спайке")
        item_scores.append(3)
    # Zhonya occupying slot 1 of realized path (before boots / second offensive)
    if len(final6) >= 2 and final6[1] == "Zhonya's Hourglass" and sit["primary"] in {"mage", "assassin", "marksman"}:
        critical.append("Zhonya — второй предмет в 6-slot path, displacing boots/damage")
        item_scores.append(4)
    if final6 and final6[0] == "Zhonya's Hourglass":
        critical.append("Zhonya — первый предмет в 6-slot path")
        item_scores.append(3)

    # Conflicting reasons
    rs = " | ".join(case.get("reasons") or [])
    if "Liandry ↑" in rs and "Liandry ↓" in rs:
        weaknesses.append("Движок одновременно поднимает и опускает Liandry — противоречивый приоритет")

    # Order: boots after 3 legendaries
    boot_idx = next((i for i, n in enumerate(final6) if n in BOOTS), None)
    if boot_idx is not None and boot_idx >= 3:
        weaknesses.append("Боты после трёх легендарик — поздняя обувка")
        item_scores.append(6)

    # Offensive vs defensive balance for carries
    if sit["primary"] in {"marksman", "assassin", "mage"}:
        defs = [n for n in final6 if n in {"Zhonya's Hourglass", "Banshee's Veil", "Force of Nature", "Spirit Visage", "Frozen Heart", "Randuin's Omen", "Kaenic Rookern", "Locket of the Iron Solari"}]
        if len(defs) >= 3:
            critical.append(f"Керри покупает {len(defs)} деф-предмета в 6 слотах: {', '.join(defs)}")
            item_scores.append(3)

    # Support identity
    if pos == "UTILITY":
        names = [b["name"] for b in build]
        if sit["primary"] in {"mage"} or "poke" in sit["tags"] or "ap" in sit["tags"]:
            if any(n in {"Moonstone Renewer", "Staff of Flowing Water"} for n in final6) and cid not in {"Karma", "Lux", "Morgana", "Lulu"}:
                weaknesses.append("Энчантер-кор на маг-саппорте")
        if sit["primary"] == "support" and ("healer" in sit["tags"] or "shield" in sit["tags"]):
            if any(n in {"Stormsurge", "Luden's Companion"} for n in final6):
                critical.append("Бёрст-маг предметы на энчантере")
                item_scores.append(2)

    # Ally frontline: tank going full glass
    if sit["need_front_from_you"] and sit["primary"] == "tank":
        if not any(item_kind(n) == "tank" or n in TANK_ITEMS for n in final6):
            critical.append("Единственный фронтлайн, но билд без танк-предметов")
            item_scores.append(3)

    # Average item scores + start (start is part of recommendation)
    scores_for_avg = item_scores + [st_score]
    if not scores_for_avg:
        base = 3
    else:
        base = statistics.mean(scores_for_avg)

    # Honest clamp. Identity errors are worse than missing a counter.
    ncrit = len(uniq(critical))
    nweak = len(uniq(weaknesses))
    ident_keys = (
        "Mosstomper",
        "Scorchclaw на AD",
        "AP-предмет на AD",
        "AD-предмет на AP",
        "саппорт-танк предмет",
        "энчантер-предмет вне",
        "Cassiopeia не покупает",
        "больше одной пары ботов",
        "second_boots",
        "единственный фронтлайн",
        "Бёрст-маг предметы",
        "Doran's Ring на AD",
        "Doran's Blade на AP",
        "Zhonya — первый предмет",
    )
    ident = [c for c in uniq(critical) if any(k in c for k in ident_keys)]
    sitc = [c for c in uniq(critical) if c not in ident]

    if ident:
        base = min(base, 5.2)
    if len(ident) >= 2:
        base = min(base, 3.4)
    if sitc:
        base = min(base, 6.2)
    if len(sitc) >= 2:
        base = min(base, 5.2)
    if nweak >= 4:
        base = min(base, 6.2)
    elif nweak >= 2 and not ident and not sitc:
        base = min(base, 6.8)

    score = int(round(max(0, min(10, base))))
    if ident:
        score = min(score, 5)
    if len(ident) >= 2:
        score = min(score, 3)
    if sitc:
        score = min(score, 6)
    if len(sitc) >= 2 and not ident:
        score = min(score, 5)
    if ident and sitc:
        score = min(score, 4)
    if not ident and not sitc and nweak == 0 and score < 8:
        score = 8
    if not ident and not sitc and nweak <= 1 and first_leg and cores and first_leg in cores and st_score >= 8:
        score = max(score, 8)
        if nweak == 0 and st_score >= 9:
            score = 9

    # Alternatives
    cores = cores_for(cid, pos, sit["primary"])
    if first_leg and cores and first_leg not in cores:
        alts.append("Первый легендарик: " + " / ".join(sorted(cores)[:4]))
    if pos == "JUNGLE" and start and "Mosstomper" in start and cid not in TANK_CHAMPS and sit["primary"] != "tank":
        if is_ap_champ(sit["primary"], sit["tags"]):
            alts.append("Старт: Scorchclaw Pup")
        else:
            alts.append("Старт: Gustwalker Hatchling")
    if sit["need_antiheal"] and not has_anti:
        if is_ap_champ(sit["primary"], sit["tags"]):
            alts.append("В 2–3 слот: Oblivion Orb → Morellonomicon")
        elif sit["primary"] == "marksman":
            alts.append("Mortal Reminder вместо LDR, если хилят")
        else:
            alts.append("Chempunk Chainsword или Bramble/Thornmail")
    if sit["need_pen"] and not has_pen:
        if sit["primary"] == "marksman":
            alts.append("Lord Dominik's / Kraken / BotRK")
        elif is_ap_champ(sit["primary"], sit["tags"]):
            alts.append("Liandry's + Void Staff")
        else:
            alts.append("Black Cleaver")

    # Recommended 6 from expert (not engine)
    expert6 = suggest_build(cid, pos, sit, start)
    if not strengths:
        strengths.append("Классовый seed узнаваем, предметы не случайные")

    verdict = verdict_text(score, cid, pos, sit, critical, weaknesses, final6, start, nxt)

    return {
        "file": case["file"],
        "champion": cid,
        "role": pos,
        "seed": case["seed"],
        "start": start,
        "next": nxt,
        "recommended": [b["name"] for b in build],
        "final6": final6,
        "extras": extras,
        "score": score,
        "base": round(base, 2),
        "sit": {k: v for k, v in sit.items() if k not in {"tags", "_final"}},
        "primary": sit["primary"],
        "strengths": uniq(strengths)[:8],
        "weaknesses": uniq(weaknesses)[:8],
        "critical": uniq(critical)[:8],
        "missing": uniq(missing)[:6],
        "alts": uniq(alts + [f"Экспертный план: {expert6}"])[:6],
        "expert6": expert6,
        "verdict": verdict,
        "reasons_engine": case.get("reasons") or [],
        "ally_front_engine": case.get("ally_front"),
        "pressure": case.get("pressure") or {},
        "allies": case["allies"],
        "enemies": case["enemies"],
        "flags": flags,
        "parse_error": False,
        "archetype": archetype_of(sit["primary"], pos, final6, case["seed"]),
    }


def uniq(xs: list[str]) -> list[str]:
    seen = set()
    out = []
    for x in xs:
        if x and x not in seen:
            seen.add(x)
            out.append(x)
    return out


def suggest_build(cid: str, pos: str, sit: dict, start: str | None) -> str:
    cores = list(cores_for(cid, pos, sit["primary"]))
    prefer = [
        "Trinity Force",
        "Sundered Sky",
        "Stormsurge",
        "Blackfire Torch",
        "Malignance",
        "Luden's Companion",
        "Youmuu's Ghostblade",
        "Kraken Slayer",
        "Infinity Edge",
        "Moonstone Renewer",
        "Locket of the Iron Solari",
        "Sunfire Aegis",
        "Iceborn Gauntlet",
        "Liandry's Torment",
        "Eclipse",
        "Blade of the Ruined King",
        "Nashor's Tooth",
    ]
    first = next((c for c in prefer if c in cores), None) or (cores[0] if cores else "ситуативный core")
    # pick situational bits
    bits = []
    if pos == "JUNGLE":
        if is_ap_champ(sit["primary"], sit["tags"]):
            bits.append("Scorchclaw")
        elif sit["primary"] == "tank" or cid in TANK_CHAMPS:
            bits.append("Mosstomper")
        else:
            bits.append("Gustwalker")
    elif pos == "UTILITY":
        bits.append("World Atlas")
    if cid in {"Yone", "Yasuo"}:
        bits.append("Berserker's Greaves")
    elif is_ap_champ(sit["primary"], sit["tags"]) and pos != "UTILITY":
        bits.append("Sorcerer's Shoes")
    elif pos == "BOTTOM":
        bits.append("Berserker's Greaves" if sit["e_ad"] < 4 else "Plated Steelcaps")
    elif sit["e_ap"] > sit["e_ad"]:
        bits.append("Mercury's Treads")
    else:
        bits.append("Plated Steelcaps")
    bits.append(first)
    if sit["need_antiheal"]:
        bits.append("anti-heal")
    if sit["need_pen"]:
        bits.append("pen/%HP")
    if sit["need_mr"]:
        bits.append("MR")
    if sit["need_armor"] and sit["primary"] in {"fighter", "tank", "support"}:
        bits.append("броня")
    return " → ".join(bits[:6])


def archetype_of(primary: str, pos: str, final6: list[str], seed: str) -> str:
    if pos == "UTILITY":
        if any(n in SUPPORT_ENCHANT for n in final6):
            return "Enchanter"
        if any(n in SUPPORT_ENGAGE for n in final6):
            return "Support Tank"
        if any(n in AP_ITEMS for n in final6):
            return "Mage Support"
        return "Support"
    if "assassin" in seed:
        return "Assassin"
    if primary == "marksman":
        if any(n in {"Blade of the Ruined King", "Guinsoo's Rageblade", "Nashor's Tooth", "Terminus"} for n in final6):
            return "On-hit"
        if any(n in {"Youmuu's Ghostblade", "Opportunity", "Hubris"} for n in final6):
            return "Lethality"
        return "Crit ADC"
    if primary == "mage":
        return "Mage"
    if primary == "tank":
        return "Full Tank"
    if primary == "fighter":
        if any(n in TANK_ITEMS for n in final6) and sum(n in TANK_ITEMS for n in final6) >= 2:
            return "Bruiser/Tank"
        return "Bruiser"
    if primary == "assassin":
        return "Assassin"
    return primary or seed


def verdict_text(score, cid, pos, sit, critical, weaknesses, final6, start, nxt) -> str:
    role = RU_ROLE.get(pos, pos)
    ctx = (
        f"{cid} {role}. Враги: {sit['e_ad']} AD / {sit['e_ap']} AP / "
        f"{sit['e_tank']} танк / {sit['e_heal']} heal / {sit['e_dive']} dive."
    )
    if score >= 9:
        return f"{ctx} Рекомендация совпадает с функцией чемпиона в этом драфте и корректно закрывает главные угрозы."
    if score >= 7:
        extra = weaknesses[0] if weaknesses else "есть точечные улучшения порядка/слотов"
        return f"{ctx} Билд рабочий для Emerald+, но не оптимальный: {extra}."
    if score >= 5:
        extra = (critical[0] if critical else (weaknesses[0] if weaknesses else "часть слотов не про ситуацию"))
        return f"{ctx} Примерно половина решений спорная. Главное: {extra}."
    extra = critical[0] if critical else (weaknesses[0] if weaknesses else "неверная роль предмета")
    return f"{ctx} Если игрок купит список как есть, эффективность чемпиона в этой катке заметно просядет. {extra}."


def fmt_comp(d: dict[str, str]) -> str:
    order = ["TOP", "JUNGLE", "MIDDLE", "BOTTOM", "UTILITY"]
    return ", ".join(f"{RU_ROLE[p]} {d[p]}" for p in order if p in d)


def card_md(a: dict) -> str:
    rec = " → ".join(a["recommended"])
    fin = " → ".join(a["final6"]) if a["final6"] else "—"
    st = "\n".join(f"- {x}" for x in a["strengths"]) or "- нет явных сильных сторон"
    wk = "\n".join(f"- {x}" for x in a["weaknesses"]) or "- нет существенных слабостей"
    cr = "\n".join(f"- {x}" for x in a["critical"]) or "- нет"
    ms = "\n".join(f"- {x}" for x in a["missing"]) or "- нет обязательных дыр"
    al = "\n".join(f"- {x}" for x in a["alts"]) or "- текущий план приемлем"
    press = a.get("pressure") or {}
    press_s = ", ".join(f"{k}={v:.2f}" for k, v in press.items())
    return f"""### {a['file']} — {a['score']}/10

- **Champion:** {a['champion']}
- **Role:** {a['role']} ({RU_ROLE.get(a['role'], a['role'])}), seed=`{a['seed']}`, primary=`{a['primary']}`, archetype=`{a['archetype']}`
- **Recommended build:** {rec}
- **Реализованный 6-item path (строгий порядок):** {fin}
- **Start / Next:** {a.get('start') or '—'} / {a.get('next') or '—'}
- **Situational context:** союзники: {fmt_comp(a['allies'])}. Враги: {fmt_comp(a['enemies'])}. Сигналы движка: {press_s or '—'}. Наш подсчёт: AD {a['sit']['e_ad']}, AP {a['sit']['e_ap']}, tanks {a['sit']['e_tank']}, heal {a['sit']['e_heal']}, dive {a['sit']['e_dive']}, allyFront {a['sit']['a_front']}.
- **Build score:** {a['score']}/10
- **Verdict:** {a['verdict']}
- **Strengths:**
{st}
- **Weaknesses:**
{wk}
- **Critical mistakes:**
{cr}
- **Missing items/options:**
{ms}
- **Better alternatives:**
{al}
- **Final recommendation:** {a['verdict']} Экспертный каркас: `{a['expert6']}`.
"""


def percentile_band(scores: list[int]) -> dict[str, int]:
    bands = {"9-10": 0, "7-8": 0, "5-6": 0, "3-4": 0, "0-2": 0}
    for s in scores:
        if s >= 9:
            bands["9-10"] += 1
        elif s >= 7:
            bands["7-8"] += 1
        elif s >= 5:
            bands["5-6"] += 1
        elif s >= 3:
            bands["3-4"] += 1
        else:
            bands["0-2"] += 1
    return bands


def main() -> int:
    if not REPORT.is_file():
        raise SystemExit(f"missing {REPORT}")
    tags = load_tags()
    text = REPORT.read_text(encoding="utf-8")
    parsed = parse_report(text)
    json_files = sorted(p.name for p in LIVE.glob("*.json") if p.name not in {"index.json", "_audit_stats.json"})
    n_json = len(json_files)
    n_parsed = len(parsed)
    n_err = sum(1 for p in parsed if p.get("parse_error"))
    analyzed = []
    parse_errors = []
    for p in parsed:
        if p.get("parse_error"):
            parse_errors.append(p["file"])
            continue
        analyzed.append(analyze_case(tags, p))

    scores = [a["score"] for a in analyzed]
    avg = statistics.mean(scores) if scores else 0
    med = statistics.median(scores) if scores else 0
    dist = Counter(scores)
    bands = percentile_band(scores)
    n_crit_cases = sum(1 for a in analyzed if a["critical"])
    n_crit_total = sum(len(a["critical"]) for a in analyzed)

    # Champion / role / combo stats
    by_champ = defaultdict(list)
    by_role = defaultdict(list)
    by_combo = defaultdict(list)
    by_arch = defaultdict(list)
    for a in analyzed:
        by_champ[a["champion"]].append(a)
        by_role[a["role"]].append(a)
        by_combo[(a["champion"], a["role"])].append(a)
        by_arch[a["archetype"]].append(a)

    item_wrong = Counter()
    item_missing = Counter()
    pattern = Counter()
    for a in analyzed:
        for c in a["critical"] + a["weaknesses"]:
            if "Mosstomper" in c:
                pattern["jungle_pet_mosstomper"] += 1
            if "multiple_boots" in c or "больше одной пары ботов" in c:
                pattern["multiple_boots"] += 1
            if "Нет anti-heal" in c or "Anti-heal есть в хвосте" in c:
                pattern["missing_antiheal"] += 1
            if "Zhonya — второй" in c or "Zhonya — первый" in c or "Next item Zhonya" in c:
                pattern["zhonya_too_early"] += 1
            if "Next item" in c and "обув" in c:
                pattern["boots_first_500g"] += 1
            if "Liandry" in c and "одновременно" in c:
                pattern["liandry_conflict"] += 1
            if "AP-предмет на AD" in c:
                pattern["ap_on_ad"] += 1
            if "AD-предмет на AP" in c:
                pattern["ad_on_ap"] += 1
            if "Два MR" in c or "лишний MR" in c:
                pattern["dual_mr"] += 1
            if "Керри покупает" in c:
                pattern["carry_too_tanky"] += 1
        for n, sc_hint in [(a["final6"][0], None)] if a["final6"] else []:
            pass
        if a["score"] <= 4:
            for n in a["final6"][:3]:
                item_wrong[n] += 1
        for m in a["missing"]:
            item_missing[m.split(" ")[0]] += 1

    best = sorted(analyzed, key=lambda x: (-x["score"], x["file"]))[:25]
    worst = sorted(analyzed, key=lambda x: (x["score"], x["file"]))[:25]

    # Write details
    details = [
        "# Детальный аудит каждой рекомендации",
        "",
        f"Обработано кейсов: {len(analyzed)}. Ошибок парсинга: {n_err}. JSON на диске (кроме index): {n_json}.",
        "",
        "Оценка — независимый Emerald+ вердикт. Список движка трактуется как **строгий порядок покупки**: игрок заполняет 6 слотов сверху вниз (старт продаётся, компоненты апгрейдятся, вторая пара ботов — ошибка).",
        "",
    ]
    for a in analyzed:
        details.append(card_md(a))
        details.append("")
    OUT_DETAILS.write_text("\n".join(details), encoding="utf-8")

    def champ_row(cid: str, arr: list[dict]) -> str:
        sc = [x["score"] for x in arr]
        issues = Counter()
        for x in arr:
            for c in x["critical"][:1]:
                issues[c[:80]] += 1
        main = issues.most_common(1)[0][0] if issues else "—"
        good = sum(1 for s in sc if s >= 7)
        bad = sum(1 for s in sc if s <= 4)
        note = ""
        if len(arr) < 4:
            note = " (малая выборка)"
        return (
            f"| {cid}{note} | {len(arr)} | {statistics.mean(sc):.2f} | {statistics.median(sc):.1f} | "
            f"{max(sc)} | {min(sc)} | good {good} / bad {bad}; {main} |"
        )

    champ_lines = [
        "| Champion | Samples | Avg Score | Median | Best | Worst | Main Issues |",
        "|---|---:|---:|---:|---:|---:|---|",
    ]
    for cid in sorted(by_champ, key=lambda c: (-statistics.mean(x["score"] for x in by_champ[c]), c)):
        champ_lines.append(champ_row(cid, by_champ[cid]))

    role_lines = [
        "| Role | Samples | Avg Score | Median | Main Strengths | Main Weaknesses |",
        "|---|---:|---:|---:|---|---|",
    ]
    role_comment = {
        "TOP": (
            "Брузер-коры (Trinity/Sundered/Cleaver) часто на месте",
            "Слабая адаптация ботов и anti-heal; танк/брузер иногда смешиваются",
        ),
        "JUNGLE": (
            "Легендарки по классу часто разумные",
            "Системно неверный jungle pet (Mosstomper почти всем)",
        ),
        "MIDDLE": (
            "AP/AD identity мидеров в целом держится",
            "Zhonya слишком высоко; Liandry конфликтует; мало anti-heal",
        ),
        "BOTTOM": (
            "Crit/on-hit seed узнаваем",
            "Боты первым next item; RFC/Kraken порядок; LDR без танков",
        ),
        "UTILITY": (
            "Разделение mage/enchanter/engage в целом есть",
            "Одинаковый Blackfire-шаблон на всех маг-саппортах; мало квест-завершений",
        ),
    }
    for role in ["TOP", "JUNGLE", "MIDDLE", "BOTTOM", "UTILITY"]:
        arr = by_role.get(role) or []
        if not arr:
            continue
        sc = [x["score"] for x in arr]
        st, wk = role_comment.get(role, ("—", "—"))
        role_lines.append(
            f"| {role} | {len(arr)} | {statistics.mean(sc):.2f} | {statistics.median(sc):.1f} | {st} | {wk} |"
        )

    combo_lines = [
        "| Champion | Role | Samples | Avg Score | Typical Build | Main Issues |",
        "|---|---|---:|---:|---|---|",
    ]
    for key in sorted(by_combo, key=lambda k: (-statistics.mean(x["score"] for x in by_combo[k]), k[0], k[1])):
        arr = by_combo[key]
        sc = [x["score"] for x in arr]
        typical = Counter(tuple(x["final6"][:4]) for x in arr).most_common(1)
        typ = " → ".join(typical[0][0]) if typical and typical[0][0] else "—"
        issues = Counter(c[:70] for x in arr for c in x["critical"][:1])
        main = issues.most_common(1)[0][0] if issues else "—"
        nnote = " *(n<3)*" if len(arr) < 3 else ""
        combo_lines.append(
            f"| {key[0]}{nnote} | {key[1]} | {len(arr)} | {statistics.mean(sc):.2f} | {typ} | {main} |"
        )

    arch_lines = [
        "| Archetype | Samples | Avg | Median | Notes |",
        "|---|---:|---:|---:|---|",
    ]
    for arch in sorted(by_arch, key=lambda k: -statistics.mean(x["score"] for x in by_arch[k])):
        arr = by_arch[arch]
        sc = [x["score"] for x in arr]
        arch_lines.append(
            f"| {arch} | {len(arr)} | {statistics.mean(sc):.2f} | {statistics.median(sc):.1f} | "
            f"crit-кейсов {sum(1 for x in arr if x['critical'])} |"
        )

    def show_case(a: dict, why: str) -> str:
        return (
            f"#### {a['file']} — {a['score']}/10\n\n"
            f"- **Champion / Role:** {a['champion']} {a['role']}\n"
            f"- **Team:** {fmt_comp(a['allies'])}\n"
            f"- **Enemy:** {fmt_comp(a['enemies'])}\n"
            f"- **Recommended:** {' → '.join(a['recommended'])}\n"
            f"- **Realized 6:** {' → '.join(a['final6'])}\n"
            f"- **Почему в этом разделе:** {why}\n"
            f"- **Вердикт:** {a['verdict']}\n"
        )

    best_md = []
    for a in best[:12]:
        best_md.append(show_case(a, "высокий score и identity+ситуация совпали"))
    worst_md = []
    for a in worst[:12]:
        crit = a["critical"][0] if a["critical"] else (a["weaknesses"][0] if a["weaknesses"] else "системная ошибка")
        worst_md.append(show_case(a, f"низкий score; {crit}"))

    n = len(analyzed)
    def pct(x):
        return f"{x} ({100.0 * x / n:.1f}%)" if n else "0"

    # Strong/weak champs (min 3 samples)
    champ_avg = {
        c: statistics.mean(x["score"] for x in arr)
        for c, arr in by_champ.items()
        if len(arr) >= 3
    }
    strong_champs = sorted(champ_avg, key=lambda c: -champ_avg[c])[:12]
    weak_champs = sorted(champ_avg, key=lambda c: champ_avg[c])[:12]
    combo_avg = {
        k: statistics.mean(x["score"] for x in arr)
        for k, arr in by_combo.items()
        if len(arr) >= 2
    }
    strong_combo = sorted(combo_avg, key=lambda k: -combo_avg[k])[:15]
    weak_combo = sorted(combo_avg, key=lambda k: combo_avg[k])[:15]

    audit = f"""# League of Legends Build Recommendation Audit

## 1. Executive Summary

Проведён независимый Emerald+ аудит **{n}** рекомендаций из `recommend_report.md` против **{n_json}** live-JSON (старт игры: 500g, lvl 1, у врагов нет легендарик). Движок **не перезапускался** — оценивался замороженный отчёт.

**Правила оценки (зафиксированы до анализа):**
- список предметов = строгий buy order на 6 слотов;
- off-meta пик не штрафуется, itemization оценивается под заявленную роль;
- ground truth — независимый эксперт, не seed-философия движка;
- патч данных: **16.18.1**.

**Главный вывод:** система уже умеет выдавать **узнаваемый классовый seed** (маг / ассасин / стрелок / брузер / энчантер / engage). Этого недостаточно, чтобы считать рекомендацию надёжной в конкретной катке. Средний балл **{avg:.2f}/10**, медиана **{med:.1f}**. Доля 9–10: **{100.0 * bands['9-10'] / n:.1f}%**. Доля ≤4: **{100.0 * (bands['3-4'] + bands['0-2']) / n:.1f}%**.

Если пользователь купит список как есть, типичный исход — **рабочий, но дырявый билд**: правильный тип урона, слабая доводка под heal/tank/boots/pet, часто лишний второй деф-предмет.

Критичных кейсов (есть critical mistake): **{n_crit_cases}** ({100.0 * n_crit_cases / n:.1f}%). Суммарно critical-флагов: **{n_crit_total}**.

JSON на диске: {n_json}. Секций в отчёте: {n_parsed}. Успешно разобрано: {n}. Ошибок парсинга: {n_err}. Пропущено: {n_parsed - n}.

## 2. Overall System Performance

| Метрика | Значение |
|---|---|
| Кейсов оценено | {n} |
| Средняя оценка | {avg:.2f} |
| Медиана | {med:.1f} |
| Мин / макс | {min(scores) if scores else 0} / {max(scores) if scores else 0} |
| 9–10 | {pct(bands['9-10'])} |
| 7–8 | {pct(bands['7-8'])} |
| 5–6 | {pct(bands['5-6'])} |
| 3–4 | {pct(bands['3-4'])} |
| 0–2 | {pct(bands['0-2'])} |
| Кейсы с critical | {pct(n_crit_cases)} |

Интерпретация для overlay: **черновой draft-planner**, не live itemization engine. Live-сигналы в отчёте почти пустые (`form=0`, `econ=0`, `items=[]`), поэтому «адаптация» сводится к kit-тегам врагов. Этого мало для mid-game, и даже для draft-плана часто не хватает обязательных counter-item'ов.

## 3. Score Distribution

| Score | Count | % |
|---|---:|---:|
{chr(10).join(f"| {k} | {dist[k]} | {100.0 * dist[k] / n:.1f}% |" for k in range(0, 11))}

## 4. Performance by Champion

Чемпионы с n<4 помечены *(малая выборка)* — вывод менее надёжен.

{chr(10).join(champ_lines)}

**Уже относительно хорошо (n≥3):** {", ".join(f"{c} ({champ_avg[c]:.1f})" for c in strong_champs)}

**Систематически слабо (n≥3):** {", ".join(f"{c} ({champ_avg[c]:.1f})" for c in weak_champs)}

## 5. Performance by Role

{chr(10).join(role_lines)}

## 6. Champion × Role Performance

Poppy Top / Jungle / Support и любые dual-role чемпионы считаются **разными профилями**.

{chr(10).join(combo_lines)}

**Лучшие Champion×Role (n≥2):** {", ".join(f"{c[0]} {c[1]} ({combo_avg[c]:.1f})" for c in strong_combo)}

**Худшие Champion×Role (n≥2):** {", ".join(f"{c[0]} {c[1]} ({combo_avg[c]:.1f})" for c in weak_combo)}

## 7. Build Archetype Performance

{chr(10).join(arch_lines)}

Система хорошо держит **классовый архетип** (mage остаётся mage, ADC остаётся ADC). Она плохо держит **под-архетип внутри класса**: lethality vs crit, on-hit vs IE, Iceborn-Poppy vs Heartsteel-Poppy, mage-support vs Locket.

## 8. Situational Itemization Performance

Оценка адаптации к драфту (не к live items — их нет).

| Ситуация | Умеет? | Комментарий |
|---|---|---|
| Много AD | Частично | Steelcaps/Zhonya всплывают, но ADC часто получают резист-боты слишком рано; Frozen Heart редко как 6-й слот фронта |
| Много AP | Частично | Mercs/FoN/Visage появляются, часто **оба** MR в одном 6-item path |
| Много HP/танков | Слабо | Liandry/Cleaver/LDR не стабильны; в отчёте Liandry ↑ и ↓ в одном кейсе |
| Много armor/MR | Слабо | Void Staff часто в хвосте списка и не входит в 6 слотов |
| Healing | Слабо | anti-heal не обязательный слот, хотя Morello иногда торчит без хилов |
| Burst/dive | Частично | Zhonya завышена до 2-го предмета |
| Squishy enemies | Частично | Collector/Stormsurge как дефолт, без различия vs 5 танков |
| Нет allied frontline | Слабо | «defense ↑» иногда пихает Zhonya, но не превращает второго файтера в танк |
| Нет allied damage | Слабо | танк-seed не сдвигается в брузер, если команда стеклянная |
| Нестандартная функция | Слабо | off-role itemization = тот же class seed + jungle/support routing |

## 9. Best Recommendations

Показательные кейсы, где система попала в роль и драфт, а не просто «высокий балл на дефолтном миде».

{chr(10).join(best_md)}

## 10. Worst Recommendations

Ошибки, которые реально ломают эффективность, если игрок доверится списку.

{chr(10).join(worst_md)}

## 11. Systematic Errors

Подсчитано по текстам critical/weakness после полного прохода.

| Паттерн | Кол-во кейсов (флагов) | Почему это проблема | Что чинить в алгоритме |
|---|---:|---|---|
| Неверный jungle pet / Mosstomper | {pattern['jungle_pet_mosstomper']} | Yone/Yi/Kha'Zix с танк-петом теряют клир и темп | `junglePetFor`: AD fighter/assassin → Gustwalker, AP → Scorchclaw, tank/engage → Mosstomper. Не использовать juggernaut как tank-pet. |
| Zhonya слишком рано | {pattern['zhonya_too_early']} | Нет first-item спайка | Кап defensive.priority < core; Zhonya не next item при gold<800 |
| Нет anti-heal | {pattern['missing_antiheal']} | Aatrox/WW/Soraka выигрывают файты | Триггер: ≥2 sustain или 1 strong healer → вставить Grievous в топ-6 |
| Dual MR в одном path | {pattern['dual_mr']} | При строгом порядке игрок покупает FoN и Visage | Mutually exclusive MR group, как lethality cores |
| Боты next item на 500g | {pattern['boots_first_500g']} | ADC/мид без кирки | Next item = компонент кора, боты slot 2 |
| Несколько пар ботов в списке | {pattern['multiple_boots']} | Нельзя купить двое ботов | Collapse boots как lethality |
| AP на AD / AD на AP | {pattern['ap_on_ad'] + pattern['ad_on_ap']} | Ломает чемпиона | Жёсткий damage-type gate на seed |
| Liandry ↑ и ↓ вместе | {pattern['liandry_conflict']} | Приоритет случайный | Один проход pressure, без двойного draft+live bump |

Дополнительно (качественно, по выборке отчёта):
- **Один seed на класс.** Trinity на всех AD-файтах, Stormsurge на всех AP-ассасинах, Blackfire на всех маг-саппортах.
- **Нет лимита 6 слотов.** Список 8–10 предметов без правила «что выкинуть». При строгом порядке 7-й нужный Void Staff не покупается, 6-й ненужный Banshee покупается.
- **Нет уникальности групп:** boots, grievous, mythic-like HP tanks, MR legendaries.
- **Live blend бесполезен** на этих фикстурах (`legendaries=0`), но всё равно двигает приоритеты. На старте это шум.

## 12. Missing Logic

1. Buy-path planner: 6 слотов, апгрейд компонентов, одна пара ботов.
2. Item groups / exclusivity (boots, grievous, last whisper, MR, HP-tank mythics).
3. Jungle pet matrix по чемпиону, не по грубому tank/AP флагу.
4. Mandatory counters: grievous, last whisper, vis/kaenic vs 3 AP, steelcaps vs 3 AD autoattackers.
5. Role function: peel vs engage vs split vs poke — сейчас только class template.
6. Ally composition: когда ты единственный фронт / единственный AP / нет engage.
7. Champion idiosyncrasy table (Cassiopeia без ботов, Yone Berserkers, Ezreal Manamune, Jhin не Kraken-first, Zeri не обычный crit-path, etc.) глубже, чем 5 seeds.
8. Order constraints: core before Zhonya; boots after first component; anti-heal as early component not 5th legendary.
9. Не конфликтовать draft-adjust и live-adjust на одни и те же предметы.
10. Support quest completion items (Zaz'Zak / Celestial / Bloodsong) как часть кора, не только World Atlas.

## 13. Recommended Algorithm Improvements

Приоритет по ожидаемому приросту среднего балла:

1. **Collapse + 6-slot emit.** На выход overlay отдавать ровно start + boots + 5 legendaries. Самый большой UX-баг.
2. **Jungle pet rewrite.** Сейчас в замороженном отчёте Mosstomper доминирует — это сразу −2/−3 балла на каждом леснике.
3. **Exclusive groups:** boots, grievous, LW, void/crypt, FoN/Kaenic/Visage, Shieldbow/IE/BT.
4. **Zhonya/Banshee cap:** defensive не может обогнать core и не может быть next item до первого оффенсива.
5. **Heal/tank/AP/AD counters as required slots**, не как +priority к уже существующему seed-предмету.
6. **Champion override table** для ~40 уникальных itemizer'ов (Ezreal, Cass, Yone, Zeri, Jhin, Kog, Smolder, Poppy×role, Senna×role, Pyke, Kayn, Shyvana, Teemo).
7. **Ally-need layer:** noFrontline → танк-опция на файтере; noAP → не уводить единственного мага в FoN×2.
8. **Разрешить конфликт Liandry** одним итоговым коэффициентом.
9. **Next-item policy при 500g:** всегда компонент кора или квест, не боты/Zhonya (исключения: Cass no boots, support quest).
10. **Тесты на фикстурах:** Yone jungle ≠ Mosstomper; ADC vs 0 tanks ≠ LDR; Soraka+Aatrox ⇒ grievous in top 6; dual boots forbidden.

## 14. Final Verdict

**Общая оценка системы: {avg:.1f}/10 — ограниченно полезна как подсказка класса, ненадёжна как buy order в конкретной игре.**

Что уже можно считать относительно надёжным:
- определение damage type по классу (маг не получает Trinity как core в типичном мид-кейсе);
- support routing mage vs enchanter vs engage в целом существует;
- наличие Zhonya/Mercs как реакции на AD/AP kit-теги;
- узнаваемые коры: Stormsurge AP-assassin, Blackfire poke, Locket engage, Moonstone enchanter.

Что требует серьёзной доработки, прежде чем совету доверять в катке:
- **6-item path и exclusive groups**;
- **jungle pets**;
- **обязательные counters (heal/tank/resist)**;
- **порядок (Zhonya/боты vs core)**;
- **champion×role overrides** (Poppy/Senna/Ezreal/Yone/Zeri/Jhin);
- **учёт союзников**, не только врагов.

Ответ на главный вопрос: *«Если пользователь доверится рекомендации именно в этой ситуации, насколько она разумна?»*  
В среднем — **частично разумна (около {avg:.1f}/10)**. Игрок получит правильный тип чемпиона и несколько ситуативных резистов, но часто купит неверный старт в лесу, слишком раннюю Zhonya, лишний второй деф и не купит anti-heal/pen, которые решают конкретный драфт.

---

### Статус обработки файлов

| | N |
|---|---:|
| Всего JSON (fixtures, без index) | {n_json} |
| Секций в recommend_report.md | {n_parsed} |
| Успешно обработано | {n} |
| Ошибок парсинга | {n_err} |
| Пропущено | {n_parsed - n} |
| Парсинг errors files | {", ".join(parse_errors) if parse_errors else "нет"} |

Детальные карточки всех кейсов: [`build_recommendation_case_details.md`](build_recommendation_case_details.md)
"""

    OUT_AUDIT.write_text(audit, encoding="utf-8")

    stats = {
        "n_json": n_json,
        "n_parsed": n_parsed,
        "n_analyzed": n,
        "n_parse_error": n_err,
        "avg": avg,
        "median": med,
        "dist": dict(dist),
        "bands": bands,
        "pattern": dict(pattern),
        "n_crit_cases": n_crit_cases,
        "parse_errors": parse_errors,
        "strong_champs": [(c, champ_avg[c], len(by_champ[c])) for c in strong_champs],
        "weak_champs": [(c, champ_avg[c], len(by_champ[c])) for c in weak_champs],
    }
    OUT_STATS.write_text(json.dumps(stats, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"analyzed={n} avg={avg:.2f} median={med:.1f} json={n_json} parsed={n_parsed} err={n_err}")
    print(f"wrote {OUT_AUDIT}")
    print(f"wrote {OUT_DETAILS}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
