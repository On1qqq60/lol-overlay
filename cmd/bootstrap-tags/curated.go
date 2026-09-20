package main

import "lol-build-overlay/internal/tags"

func p(primary string, t ...string) tags.ChampionProfile {
	return tags.ChampionProfile{Primary: primary, Tags: t}
}

func role(primary string, t ...string) tags.RoleOverride {
	return tags.RoleOverride{Primary: primary, Tags: t}
}

// curatedOverrides — manual tags + flex by_role. Wins over auto bootstrap.
func curatedOverrides() map[string]tags.ChampionProfile {
	out := map[string]tags.ChampionProfile{
		// --- stable cores ---
		"Ahri":     p(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StyleBurst, tags.ThreatMobility, tags.ExtraPick),
		"Veigar":   p(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StyleBurst, tags.ThreatCCHard, tags.ExtraPick),
		"Syndra":   p(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StyleBurst, tags.ExtraPick),
		"Viktor":   p(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StylePoke, tags.StyleDPS),
		"Orianna":  p(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StylePoke, tags.StyleShield, tags.ThreatCCHard),
		"Annie":    p(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StyleBurst, tags.ThreatCCHard),
		"Lux":      p(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StylePoke, tags.StyleBurst, tags.ThreatCCHard),
		"Xerath":   p(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StylePoke),
		"Ziggs":    p(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StylePoke, tags.ThreatSplitpush),
		"Brand":    p(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StylePoke, tags.StyleBurst),
		"Zyra":     p(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StylePoke, tags.ThreatCCHard),
		"Heimerdinger": p(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StylePoke),

		"Zed":      p(tags.ClassAssassin, tags.DamageAD, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive, tags.ThreatMobility, tags.ExtraPick),
		"Talon":    p(tags.ClassAssassin, tags.DamageAD, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive, tags.ThreatMobility, tags.ExtraPick),
		"Qiyana":   p(tags.ClassAssassin, tags.DamageAD, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive, tags.ThreatMobility, tags.ExtraPick),
		"Katarina": p(tags.ClassAssassin, tags.DamageAP, tags.DamageHybrid, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive, tags.ThreatMobility),
		"Fizz":     p(tags.ClassAssassin, tags.DamageAP, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive, tags.ThreatMobility),
		"Ekko":     p(tags.ClassAssassin, tags.DamageAP, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive, tags.ThreatMobility),
		"Akali":    p(tags.ClassAssassin, tags.DamageAP, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive, tags.ThreatMobility),
		"Diana":    p(tags.ClassAssassin, tags.DamageAP, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive, tags.ThreatCCHard),
		"Leblanc":  p(tags.ClassAssassin, tags.DamageAP, tags.RangeRanged, tags.StyleBurst, tags.ThreatMobility, tags.ExtraPick),
		"Pyke":     p(tags.ClassAssassin, tags.DamageAD, tags.RangeMelee, tags.StyleEngage, tags.ThreatDive, tags.ExtraPick, tags.ThreatCCHard),
		"Rengar":   p(tags.ClassAssassin, tags.DamageAD, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive, tags.ExtraPick),
		"Khazix":   p(tags.ClassAssassin, tags.DamageAD, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive, tags.ThreatMobility),
		"Evelynn":  p(tags.ClassAssassin, tags.DamageAP, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive),
		"Nocturne": p(tags.ClassAssassin, tags.DamageAD, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive),

		"Jinx":   p(tags.ClassMarksman, tags.DamageAD, tags.RangeRanged, tags.StyleDPS, tags.ThreatSplitpush),
		"Ashe":   p(tags.ClassMarksman, tags.DamageAD, tags.RangeRanged, tags.StyleDPS, tags.StylePoke, tags.ThreatCCHard),
		"Caitlyn": p(tags.ClassMarksman, tags.DamageAD, tags.RangeRanged, tags.StyleDPS, tags.StylePoke),
		"Kaisa":  p(tags.ClassMarksman, tags.DamageAD, tags.DamageHybrid, tags.RangeRanged, tags.StyleDPS, tags.ThreatMobility),
		"Jhin":   p(tags.ClassMarksman, tags.DamageAD, tags.RangeRanged, tags.StyleBurst, tags.StylePoke),
		"Vayne":  p(tags.ClassMarksman, tags.DamageAD, tags.DamageTrue, tags.RangeRanged, tags.StyleDPS, tags.ThreatMobility),
		"Twitch": p(tags.ClassMarksman, tags.DamageAD, tags.RangeRanged, tags.StyleDPS),
		"Ezreal": p(tags.ClassMarksman, tags.DamageAD, tags.RangeRanged, tags.StylePoke, tags.ThreatMobility),
		"Lucian": p(tags.ClassMarksman, tags.DamageAD, tags.RangeRanged, tags.StyleDPS, tags.ThreatMobility),
		"Samira": p(tags.ClassMarksman, tags.DamageAD, tags.RangeRanged, tags.StyleBurst, tags.ThreatDive),

		// healers / enchanters
		"Soraka":     p(tags.ClassSupport, tags.DamageAP, tags.RangeRanged, tags.StyleHealer, tags.StylePeel),
		"Sona":       p(tags.ClassSupport, tags.DamageAP, tags.RangeRanged, tags.StyleHealer, tags.StyleShield, tags.StylePoke),
		"Nami":       p(tags.ClassSupport, tags.DamageAP, tags.RangeRanged, tags.StyleHealer, tags.StyleEngage, tags.ThreatCCHard),
		"Yuumi":      p(tags.ClassSupport, tags.DamageAP, tags.RangeRanged, tags.StyleHealer, tags.StyleShield),
		"Milio":      p(tags.ClassSupport, tags.DamageAP, tags.RangeRanged, tags.StyleHealer, tags.StyleShield, tags.StylePeel),
		"Janna":      p(tags.ClassSupport, tags.DamageAP, tags.RangeRanged, tags.StyleShield, tags.StylePeel, tags.ThreatCCHard),
		"Lulu":       p(tags.ClassSupport, tags.DamageAP, tags.RangeRanged, tags.StyleShield, tags.StylePeel, tags.ThreatCCHard),
		"Renata":     p(tags.ClassSupport, tags.DamageAP, tags.RangeRanged, tags.StyleShield, tags.ThreatCCHard, tags.StyleEngage),
		"Taric":      p(tags.ClassSupport, tags.DamageAD, tags.RangeMelee, tags.StyleHealer, tags.StyleShield, tags.StylePeel, tags.ThreatCCHard),
		"Bard":       p(tags.ClassSupport, tags.DamageAP, tags.RangeRanged, tags.StyleEngage, tags.StylePeel, tags.ThreatCCHard),
		"Thresh":     p(tags.ClassSupport, tags.DamageAD, tags.RangeRanged, tags.StyleEngage, tags.StylePeel, tags.ThreatCCHard),
		"Nautilus":   p(tags.ClassSupport, tags.ClassTank, tags.RangeMelee, tags.StyleEngage, tags.ThreatCCHard, tags.ThreatDive),
		"Leona":      p(tags.ClassSupport, tags.ClassTank, tags.RangeMelee, tags.StyleEngage, tags.ThreatCCHard, tags.ThreatDive),
		"Rakan":      p(tags.ClassSupport, tags.DamageAP, tags.RangeMelee, tags.StyleEngage, tags.StyleShield, tags.ThreatCCHard, tags.ThreatMobility),
		"Alistar":    p(tags.ClassSupport, tags.ClassTank, tags.RangeMelee, tags.StyleEngage, tags.StylePeel, tags.ThreatCCHard),
		"Braum":      p(tags.ClassSupport, tags.ClassTank, tags.RangeMelee, tags.StylePeel, tags.StyleShield, tags.ThreatCCHard),
		"Blitzcrank": p(tags.ClassSupport, tags.ClassTank, tags.RangeMelee, tags.StyleEngage, tags.ThreatCCHard, tags.ExtraPick),
		"Morgana":    p(tags.ClassSupport, tags.DamageAP, tags.RangeRanged, tags.StyleShield, tags.ThreatCCHard, tags.StylePeel),

		// juggernauts / tanks
		"Mordekaiser": p(tags.ClassFighter, tags.DamageAP, tags.RangeMelee, tags.ExtraJuggernaut, tags.ExtraAllIn, tags.ThreatSustain),
		"Darius":      p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ExtraJuggernaut, tags.ExtraAllIn, tags.ThreatSustain),
		"Garen":       p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ExtraJuggernaut, tags.ThreatSustain),
		"Nasus":       p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ExtraJuggernaut, tags.ThreatSustain, tags.ThreatSplitpush),
		"Illaoi":      p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ExtraJuggernaut, tags.ThreatSustain, tags.ThreatSplitpush),
		"Sett":        p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ExtraJuggernaut, tags.ExtraAllIn, tags.StyleEngage, tags.ThreatSustain),
		"Sion":        p(tags.ClassTank, tags.DamageAD, tags.RangeMelee, tags.StyleEngage, tags.ThreatCCHard, tags.ThreatSustain, tags.ExtraJuggernaut),
		"Ornn":        p(tags.ClassTank, tags.DamageAD, tags.RangeMelee, tags.StyleEngage, tags.ThreatCCHard, tags.ThreatSustain),
		"Maokai":      p(tags.ClassTank, tags.DamageAP, tags.RangeMelee, tags.StyleEngage, tags.StylePeel, tags.ThreatCCHard),
		"Sejuani":     p(tags.ClassTank, tags.DamageAP, tags.RangeMelee, tags.StyleEngage, tags.ThreatCCHard, tags.ThreatDive),
		"Amumu":       p(tags.ClassTank, tags.DamageAP, tags.RangeMelee, tags.StyleEngage, tags.ThreatCCHard, tags.ThreatDive),
		"Zac":         p(tags.ClassTank, tags.DamageAP, tags.RangeMelee, tags.StyleEngage, tags.ThreatDive, tags.ThreatCCHard, tags.ThreatSustain),
		"TahmKench":   p(tags.ClassTank, tags.ClassSupport, tags.RangeMelee, tags.StyleEngage, tags.StylePeel, tags.ThreatSustain),
		"DrMundo":     p(tags.ClassTank, tags.DamageAD, tags.RangeMelee, tags.ExtraJuggernaut, tags.ThreatSustain, tags.ThreatSplitpush),
		"Chogath":     p(tags.ClassTank, tags.DamageAP, tags.RangeMelee, tags.StyleEngage, tags.ThreatCCHard, tags.ExtraJuggernaut),
		"Malphite":    p(tags.ClassTank, tags.DamageAP, tags.RangeMelee, tags.StyleEngage, tags.ThreatDive, tags.ThreatCCHard),
		"Shen":        p(tags.ClassTank, tags.DamageAD, tags.RangeMelee, tags.StylePeel, tags.StyleEngage, tags.ThreatCCHard),

		// divers / engage junglers
		"LeeSin":   p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ThreatDive, tags.ThreatMobility, tags.StyleEngage),
		"JarvanIV": p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ThreatDive, tags.StyleEngage, tags.ThreatCCHard),
		"Vi":       p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ThreatDive, tags.StyleEngage, tags.ThreatCCHard),
		"Hecarim":  p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ThreatDive, tags.StyleEngage, tags.ThreatMobility),
		"Elise":    p(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.ThreatDive, tags.StyleBurst),
		"RekSai":   p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ThreatDive, tags.ExtraPick),
		"Kayn":     p(tags.ClassAssassin, tags.DamageAD, tags.RangeMelee, tags.ThreatDive, tags.ThreatMobility, tags.StyleBurst),
		"Belveth":  p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ThreatDive, tags.StyleDPS, tags.ThreatMobility),
		"Irelia":   p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ThreatDive, tags.ThreatMobility, tags.StyleDPS),
		"Yasuo":    p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.StyleDPS, tags.ThreatDive, tags.ThreatMobility),
		"Yone":     p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.StyleDPS, tags.ThreatDive, tags.ThreatMobility),
		"Aatrox":   p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ExtraJuggernaut, tags.ExtraAllIn, tags.ThreatSustain),
		"Camille":  p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ThreatDive, tags.ExtraPick, tags.ThreatMobility),
		"Riven":    p(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.StyleBurst, tags.ThreatMobility, tags.ExtraAllIn),
		"Fiora":    p(tags.ClassFighter, tags.DamageAD, tags.DamageTrue, tags.RangeMelee, tags.StyleDPS, tags.ThreatSplitpush),
		"Gwen":     p(tags.ClassFighter, tags.DamageAP, tags.DamageTrue, tags.RangeMelee, tags.StyleDPS, tags.ThreatDive),
	}

	// --- flex by_role ---
	out["Pantheon"] = tags.ChampionProfile{
		Primary: tags.ClassFighter,
		Tags:    []string{tags.DamageAD, tags.RangeMelee, tags.StyleEngage, tags.ExtraAllIn},
		ByRole: map[string]tags.RoleOverride{
			"UTILITY": role(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.StyleEngage, tags.StylePeel, tags.ThreatCCHard),
			"TOP":     role(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.StylePoke, tags.ExtraAllIn, tags.ThreatSplitpush),
			"MIDDLE":  role(tags.ClassAssassin, tags.DamageAD, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive, tags.ExtraPick),
			"JUNGLE":  role(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.StyleEngage, tags.ThreatDive, tags.ExtraAllIn),
		},
	}
	out["Poppy"] = tags.ChampionProfile{
		Primary: tags.ClassTank,
		Tags:    []string{tags.ClassFighter, tags.RangeMelee, tags.StyleEngage, tags.ExtraAntiDash, tags.ThreatCCHard},
		ByRole: map[string]tags.RoleOverride{
			"UTILITY": role(tags.ClassTank, tags.RangeMelee, tags.StyleEngage, tags.StylePeel, tags.ExtraAntiDash, tags.ThreatCCHard),
			"TOP":     role(tags.ClassTank, tags.ClassFighter, tags.RangeMelee, tags.StyleEngage, tags.ExtraAntiDash, tags.ThreatCCHard, tags.ThreatSustain),
			"JUNGLE":  role(tags.ClassTank, tags.RangeMelee, tags.StyleEngage, tags.ThreatDive, tags.ExtraAntiDash, tags.ThreatCCHard),
		},
	}
	out["Sett"] = tags.ChampionProfile{
		Primary: tags.ClassFighter,
		Tags:    []string{tags.DamageAD, tags.RangeMelee, tags.ExtraJuggernaut, tags.ExtraAllIn, tags.StyleEngage, tags.ThreatSustain},
		ByRole: map[string]tags.RoleOverride{
			"UTILITY": role(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.StyleEngage, tags.StylePeel, tags.ThreatCCHard, tags.ThreatSustain),
			"TOP":     role(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ExtraJuggernaut, tags.ExtraAllIn, tags.ThreatSustain),
			"MIDDLE":  role(tags.ClassFighter, tags.DamageAD, tags.RangeMelee, tags.ExtraAllIn, tags.StyleBurst, tags.ThreatDive),
		},
	}
	out["Mordekaiser"] = tags.ChampionProfile{
		Primary: tags.ClassFighter,
		Tags:    []string{tags.DamageAP, tags.RangeMelee, tags.ExtraJuggernaut, tags.ExtraAllIn, tags.ThreatSustain},
		ByRole: map[string]tags.RoleOverride{
			"TOP":    role(tags.ClassFighter, tags.DamageAP, tags.RangeMelee, tags.ExtraJuggernaut, tags.ExtraAllIn, tags.ThreatSustain),
			"MIDDLE": role(tags.ClassFighter, tags.DamageAP, tags.RangeMelee, tags.ExtraJuggernaut, tags.StyleBurst, tags.ExtraAllIn),
			"JUNGLE": role(tags.ClassFighter, tags.DamageAP, tags.RangeMelee, tags.ExtraJuggernaut, tags.ThreatDive, tags.ThreatSustain),
		},
	}
	out["Swain"] = tags.ChampionProfile{
		Primary: tags.ClassMage,
		Tags:    []string{tags.DamageAP, tags.RangeRanged, tags.StyleEngage, tags.ThreatSustain, tags.ThreatCCHard},
		ByRole: map[string]tags.RoleOverride{
			"MIDDLE":  role(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StyleBurst, tags.StyleEngage, tags.ThreatSustain),
			"UTILITY": role(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StyleEngage, tags.ThreatCCHard, tags.ThreatSustain),
			"BOTTOM":  role(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StylePoke, tags.ThreatSustain),
		},
	}
	out["Karma"] = tags.ChampionProfile{
		Primary: tags.ClassMage,
		Tags:    []string{tags.DamageAP, tags.RangeRanged, tags.StyleShield, tags.StylePoke},
		ByRole: map[string]tags.RoleOverride{
			"UTILITY": role(tags.ClassSupport, tags.DamageAP, tags.RangeRanged, tags.StyleShield, tags.StylePeel, tags.StylePoke),
			"MIDDLE":  role(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StylePoke, tags.StyleBurst, tags.StyleShield),
			"TOP":     role(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StylePoke, tags.StyleShield),
		},
	}
	out["Seraphine"] = tags.ChampionProfile{
		Primary: tags.ClassMage,
		Tags:    []string{tags.DamageAP, tags.RangeRanged, tags.StylePoke, tags.StyleShield},
		ByRole: map[string]tags.RoleOverride{
			"UTILITY": role(tags.ClassSupport, tags.DamageAP, tags.RangeRanged, tags.StyleHealer, tags.StyleShield, tags.StylePoke),
			"MIDDLE":  role(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StylePoke, tags.StyleBurst),
			"BOTTOM":  role(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StylePoke, tags.StyleDPS),
		},
	}
	out["Gragas"] = tags.ChampionProfile{
		Primary: tags.ClassFighter,
		Tags:    []string{tags.DamageAP, tags.RangeMelee, tags.StyleEngage, tags.ThreatCCHard},
		ByRole: map[string]tags.RoleOverride{
			"JUNGLE":  role(tags.ClassFighter, tags.DamageAP, tags.RangeMelee, tags.StyleEngage, tags.ThreatDive, tags.ThreatCCHard),
			"TOP":     role(tags.ClassFighter, tags.DamageAP, tags.RangeMelee, tags.StyleEngage, tags.ThreatSustain, tags.ThreatCCHard),
			"MIDDLE":  role(tags.ClassMage, tags.DamageAP, tags.RangeMelee, tags.StyleBurst, tags.ThreatCCHard),
			"UTILITY": role(tags.ClassSupport, tags.DamageAP, tags.RangeMelee, tags.StyleEngage, tags.StylePeel, tags.ThreatCCHard),
		},
	}
	out["Neeko"] = tags.ChampionProfile{
		Primary: tags.ClassMage,
		Tags:    []string{tags.DamageAP, tags.RangeRanged, tags.StyleBurst, tags.ThreatCCHard},
		ByRole: map[string]tags.RoleOverride{
			"MIDDLE":  role(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StyleBurst, tags.ThreatCCHard, tags.ExtraPick),
			"UTILITY": role(tags.ClassSupport, tags.DamageAP, tags.RangeRanged, tags.StyleEngage, tags.ThreatCCHard, tags.ExtraPick),
			"TOP":     role(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StylePoke, tags.ThreatCCHard),
		},
	}
	out["Sylas"] = tags.ChampionProfile{
		Primary: tags.ClassMage,
		Tags:    []string{tags.DamageAP, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive, tags.ThreatMobility},
		ByRole: map[string]tags.RoleOverride{
			"MIDDLE": role(tags.ClassMage, tags.DamageAP, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive, tags.ThreatMobility),
			"JUNGLE": role(tags.ClassAssassin, tags.DamageAP, tags.RangeMelee, tags.StyleBurst, tags.ThreatDive),
			"TOP":    role(tags.ClassFighter, tags.DamageAP, tags.RangeMelee, tags.StyleBurst, tags.ThreatSustain),
		},
	}
	out["Lillia"] = tags.ChampionProfile{
		Primary: tags.ClassMage,
		Tags:    []string{tags.DamageAP, tags.RangeRanged, tags.StyleDPS, tags.ThreatCCHard, tags.ThreatMobility},
		ByRole: map[string]tags.RoleOverride{
			"JUNGLE": role(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StyleDPS, tags.ThreatCCHard, tags.ThreatMobility),
			"TOP":    role(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StyleDPS, tags.ThreatSustain),
			"MIDDLE": role(tags.ClassMage, tags.DamageAP, tags.RangeRanged, tags.StylePoke, tags.StyleDPS),
		},
	}

	return out
}
