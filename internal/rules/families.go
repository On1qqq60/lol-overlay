package rules

// Family keys used by the champion×role identity table.
const (
	FamADCCrit         = "adc-crit"
	FamADCOnHit        = "adc-onhit"
	FamADCEzreal       = "adc-ezreal"
	FamADCJhin         = "adc-jhin"
	FamADCZeri         = "adc-zeri"
	FamADCSmolder      = "adc-smolder"
	FamADCCorki        = "adc-corki"
	FamADCGraves       = "adc-graves"
	FamADCKayle        = "adc-kayle"
	FamADCSenna        = "adc-senna"
	FamADCKalista      = "adc-kalista"
	FamADCKaisa        = "adc-kaisa"
	FamADCLucian       = "adc-lucian"
	FamADCKog          = "adc-kog"
	FamADCSamira       = "adc-samira"
	FamADCYunTal       = "adc-yuntal"
	FamADCMF           = "adc-mf"
	FamFighterTrinity  = "fighter-trinity"
	FamFighterEclipse  = "fighter-eclipse"
	FamFighterCleaver  = "fighter-cleaver"
	FamFighterIceborn  = "fighter-iceborn"
	FamFighterCrit     = "fighter-crit"
	FamFighterYi       = "fighter-yi"
	FamFighterTrynd    = "fighter-trynd"
	FamFighterGP       = "fighter-gp"
	FamFighterFiora    = "fighter-fiora"
	FamFighterIrelia   = "fighter-irelia"
	FamFighterOlaf     = "fighter-olaf"
	FamFighterWarwick  = "fighter-warwick"
	FamFighterKayn     = "fighter-kayn"
	FamFighterUdyr     = "fighter-udyr"
	FamFighterShyvana  = "fighter-shyvana"
	FamFighterVoli     = "fighter-voli"
	FamFighterAP       = "fighter-ap"
	FamFighterJayce    = "fighter-jayce"
	FamFighterStride   = "fighter-stride"
	FamFighterUrgot    = "fighter-urgot"
	FamFighterYorick   = "fighter-yorick"
	FamFighterIllaoi   = "fighter-illaoi"
	FamFighterGwen     = "fighter-gwen"
	FamTankShen        = "tank-shen"
	FamMageSinged      = "mage-singed"
	FamMageBurst       = "mage-burst"
	FamMagePoke        = "mage-poke"
	FamMageRoA         = "mage-roa"
	FamMageNashor      = "mage-nashor"
	FamMageVlad        = "mage-vlad"
	FamMageCass        = "mage-cass"
	FamMageKat         = "mage-kat"
	FamMageLillia      = "mage-lillia"
	FamMageSylas       = "mage-sylas"
	FamMageLuden       = "mage-luden"
	FamMageSwain       = "mage-swain"
	FamMageTF          = "mage-tf"
	FamAssassinAD      = "assassin"
	FamAssassinAP      = "assassin-ap"
	FamAssassinLich    = "assassin-lich"
	FamTankSunfire     = "tank"
	FamTankMundo       = "tank-mundo"
	FamTankNasus       = "tank-nasus"
	FamTankCho         = "tank-cho"
	FamSupportEnch     = "support-enchanter"
	FamSupportEngage   = "support-engage"
	FamSupportMage     = "support-mage"
	FamSupportPyke     = "support-pyke"
	FamSupportSenna    = "support-senna"
	FamSupportBard     = "support-bard"
	FamSupportRakan    = "support-rakan"
	FamSupportRenata   = "support-renata"
	FamSupportKarma    = "support-karma"
	FamSupportPantheon = "support-pantheon"
	FamSupportADC      = "support-adc"
)

func sl(id int, name string, pri float64, role string) Slot {
	return Slot{ItemID: id, Name: name, Priority: pri, Role: role}
}

func seedByFamily(family string) []Slot {
	switch family {
	case FamADCCrit:
		return SeedADC()
	case FamADCOnHit:
		return seedADCOnHit()
	case FamADCEzreal:
		return seedEzreal()
	case FamADCJhin:
		return seedJhin()
	case FamADCZeri:
		return seedZeri()
	case FamADCSmolder:
		return seedSmolder()
	case FamADCCorki:
		return seedCorki()
	case FamADCGraves:
		return seedGraves()
	case FamADCKayle:
		return SeedNashorOnHit()
	case FamADCSenna:
		return seedSennaCarry()
	case FamADCKalista:
		return seedKalista()
	case FamADCKaisa:
		return seedKaisa()
	case FamADCLucian:
		return seedLucian()
	case FamADCKog:
		return seedKog()
	case FamADCSamira:
		return seedSamira()
	case FamADCYunTal:
		return seedADCYunTal()
	case FamADCMF:
		return seedMissFortune()
	case FamFighterTrinity:
		return SeedADBruiser()
	case FamFighterEclipse:
		return seedEclipseDiver()
	case FamFighterCleaver:
		return SeedADJuggernaut()
	case FamFighterIceborn:
		return SeedTankFighter()
	case FamFighterCrit:
		return SeedWindCrit()
	case FamFighterYi:
		return seedYi()
	case FamFighterTrynd:
		return seedTrynd()
	case FamFighterGP:
		return seedGangplank()
	case FamFighterFiora:
		return seedFiora()
	case FamFighterIrelia:
		return seedIrelia()
	case FamFighterOlaf:
		return seedOlaf()
	case FamFighterWarwick:
		return seedWarwick()
	case FamFighterKayn:
		return seedKayn()
	case FamFighterUdyr:
		return seedUdyr()
	case FamFighterShyvana:
		return seedShyvana()
	case FamFighterVoli:
		return seedVolibear()
	case FamFighterAP:
		return SeedAPJuggernaut()
	case FamFighterJayce:
		return seedJayce()
	case FamFighterStride:
		return seedStrideJuggernaut()
	case FamFighterUrgot:
		return seedUrgot()
	case FamFighterYorick:
		return seedYorick()
	case FamFighterIllaoi:
		return seedIllaoi()
	case FamFighterGwen:
		return seedGwen()
	case FamTankShen:
		return seedShen()
	case FamMageSinged:
		return seedSinged()
	case FamMageBurst:
		return SeedAPMidBurst()
	case FamMagePoke:
		return SeedAPPoke()
	case FamMageRoA:
		return seedRoAScaler()
	case FamMageNashor:
		return seedAPNashor()
	case FamMageVlad:
		return seedVladimir()
	case FamMageCass:
		return seedCassiopeia()
	case FamMageKat:
		return seedKatarina()
	case FamMageLillia:
		return seedLillia()
	case FamMageSylas:
		return seedSylasJungle()
	case FamMageLuden:
		return seedLudenHorizon()
	case FamMageSwain:
		return seedSwain()
	case FamMageTF:
		return seedTwistedFate()
	case FamAssassinAD:
		return SeedADAssassin()
	case FamAssassinAP:
		return SeedAPAssassin()
	case FamAssassinLich:
		return seedAPLich()
	case FamTankSunfire:
		return SeedTank()
	case FamTankMundo:
		return seedMundo()
	case FamTankCho:
		return seedChoGath()
	case FamTankNasus:
		return seedNasus()
	case FamSupportEnch:
		return SeedEnchanter()
	case FamSupportEngage:
		return SeedSupportEngage()
	case FamSupportMage:
		return SeedMageSupport()
	case FamSupportPyke:
		return seedPykeSupport()
	case FamSupportSenna:
		return seedSennaSupport()
	case FamSupportBard:
		return seedBard()
	case FamSupportRakan:
		return seedRakan()
	case FamSupportRenata:
		return seedRenata()
	case FamSupportKarma:
		return seedKarmaSupport()
	case FamSupportPantheon:
		return seedPantheonSupport()
	case FamSupportADC:
		return seedSupportADC()
	default:
		return SeedAPMidBurst()
	}
}

func familyLabel(family string) string {
	switch family {
	case FamFighterAP:
		return "fighter/ap"
	case FamAssassinAP, FamAssassinLich, FamMageKat:
		return "assassin/ap"
	case FamAssassinAD, FamADCGraves:
		return "assassin"
	case FamTankSunfire, FamTankMundo, FamTankNasus, FamTankCho, FamTankShen:
		return "tank"
	case FamSupportEnch, FamSupportEngage, FamSupportMage, FamSupportPyke,
		FamSupportSenna, FamSupportBard, FamSupportRakan, FamSupportRenata,
		FamSupportKarma, FamSupportPantheon, FamSupportADC:
		return "support"
	case FamMageBurst, FamMagePoke, FamMageRoA, FamMageNashor, FamMageVlad, FamMageCass,
		FamMageLillia, FamMageSylas, FamMageLuden, FamMageSwain, FamMageTF, FamMageSinged:
		return "mage"
	case FamADCCrit, FamADCOnHit, FamADCEzreal, FamADCJhin, FamADCZeri, FamADCSmolder,
		FamADCCorki, FamADCKayle, FamADCSenna, FamADCKalista, FamADCKaisa, FamADCLucian,
		FamADCKog, FamADCSamira, FamADCYunTal, FamADCMF:
		return "marksman"
	default:
		if len(family) >= 6 && family[:6] == "mage-" {
			return "mage"
		}
		if len(family) >= 8 && family[:8] == "fighter-" {
			return "fighter"
		}
		return "fighter"
	}
}

func seedADCOnHit() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemBOTRK, "Blade of The Ruined King", 88, "core"),
		sl(ItemGuinsoos, "Guinsoo's Rageblade", 80, "offensive"),
		sl(ItemBerserkers, "Berserker's Greaves", 70, "boots"),
		sl(ItemTerminus, "Terminus", 64, "offensive"),
		sl(ItemWitsEnd, "Wit's End", 56, "defensive"),
		sl(ItemRunaans, "Runaan's Hurricane", 50, "offensive"),
		sl(ItemKraken, "Kraken Slayer", 44, "offensive"),
		sl(ItemMercTreads, "Mercury's Treads", 40, "boots"),
	}
}

func seedEzreal() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemTear, "Tear of the Goddess", 92, "component"),
		sl(ItemManamune, "Manamune", 88, "core"),
		sl(ItemTrinity, "Trinity Force", 82, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 70, "boots"),
		sl(ItemIceborn, "Iceborn Gauntlet", 64, "defensive"),
		sl(ItemSeryldas, "Serylda's Grudge", 56, "pen"),
		sl(ItemMaw, "Maw of Malmortius", 48, "defensive"),
	}
}

func seedJhin() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemCollector, "The Collector", 88, "core"),
		sl(ItemInfinityEdge, "Infinity Edge", 80, "offensive"),
		sl(ItemRapidFirecannon, "Rapid Firecannon", 72, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 66, "boots"),
		sl(ItemSwifties, "Boots of Swiftness", 60, "boots"),
		sl(ItemLordDominiks, "Lord Dominik's Regards", 54, "pen"),
		sl(ItemYoumuu, "Youmuu's Ghostblade", 46, "offensive"),
	}
}

func seedZeri() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemStormrazor, "Stormrazor", 88, "core"),
		sl(ItemStatikk, "Statikk Shiv", 80, "offensive"),
		sl(ItemRunaans, "Runaan's Hurricane", 70, "offensive"),
		sl(ItemBerserkers, "Berserker's Greaves", 64, "boots"),
		sl(ItemHexplate, "Experimental Hexplate", 56, "offensive"),
		sl(ItemInfinityEdge, "Infinity Edge", 48, "offensive"),
	}
}

func seedSmolder() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemEssenceReaver, "Essence Reaver", 88, "core"),
		sl(ItemRapidFirecannon, "Rapid Firecannon", 80, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 70, "boots"),
		sl(ItemInfinityEdge, "Infinity Edge", 64, "offensive"),
		sl(ItemShojin, "Spear of Shojin", 56, "offensive"),
		sl(ItemLordDominiks, "Lord Dominik's Regards", 48, "pen"),
	}
}

func seedCorki() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemTrinity, "Trinity Force", 88, "core"),
		sl(ItemManamune, "Manamune", 82, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 70, "boots"),
		sl(ItemRapidFirecannon, "Rapid Firecannon", 62, "offensive"),
		sl(ItemInfinityEdge, "Infinity Edge", 54, "offensive"),
		sl(ItemSeryldas, "Serylda's Grudge", 46, "pen"),
	}
}

func seedGraves() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemYoumuu, "Youmuu's Ghostblade", 88, "core"),
		sl(ItemCollector, "The Collector", 78, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 68, "boots"),
		sl(ItemLordDominiks, "Lord Dominik's Regards", 60, "pen"),
		sl(ItemEclipse, "Eclipse", 52, "offensive"),
		sl(ItemDeathsDance, "Death's Dance", 46, "defensive"),
	}
}

func seedKayle() []Slot {
	return []Slot{
		sl(ItemDoransRing, "Doran's Ring", 100, "start"),
		sl(ItemNashors, "Nashor's Tooth", 88, "core"),
		sl(ItemRiftmaker, "Riftmaker", 80, "offensive"),
		sl(ItemBerserkers, "Berserker's Greaves", 70, "boots"),
		sl(ItemRabadons, "Rabadon's Deathcap", 62, "offensive"),
		sl(ItemShadowflame, "Shadowflame", 54, "offensive"),
		sl(ItemVoidStaff, "Void Staff", 46, "pen"),
		sl(ItemZhonyas, "Zhonya's Hourglass", 40, "defensive"),
	}
}

func seedSennaCarry() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemEssenceReaver, "Essence Reaver", 86, "core"),
		sl(ItemRapidFirecannon, "Rapid Firecannon", 78, "offensive"),
		sl(ItemInfinityEdge, "Infinity Edge", 70, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 64, "boots"),
		sl(ItemBlackCleaver, "Black Cleaver", 56, "offensive"),
		sl(ItemCollector, "The Collector", 50, "offensive"),
		sl(ItemLordDominiks, "Lord Dominik's Regards", 44, "pen"),
	}
}

func seedKalista() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemBOTRK, "Blade of The Ruined King", 90, "core"),
		sl(ItemGuinsoos, "Guinsoo's Rageblade", 80, "offensive"),
		sl(ItemBerserkers, "Berserker's Greaves", 70, "boots"),
		sl(ItemRunaans, "Runaan's Hurricane", 62, "offensive"),
		sl(ItemTerminus, "Terminus", 54, "offensive"),
		sl(ItemWitsEnd, "Wit's End", 46, "defensive"),
	}
}

func seedKaisa() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemKraken, "Kraken Slayer", 86, "core"),
		sl(ItemNashors, "Nashor's Tooth", 78, "offensive"),
		sl(ItemBerserkers, "Berserker's Greaves", 70, "boots"),
		sl(ItemGuinsoos, "Guinsoo's Rageblade", 62, "offensive"),
		sl(ItemPhantomDancer, "Phantom Dancer", 54, "offensive"),
		sl(ItemZhonyas, "Zhonya's Hourglass", 44, "defensive"),
	}
}

func seedLucian() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemEssenceReaver, "Essence Reaver", 88, "core"),
		sl(ItemNavori, "Navori Flickerblade", 80, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 70, "boots"),
		sl(ItemInfinityEdge, "Infinity Edge", 62, "offensive"),
		sl(ItemRapidFirecannon, "Rapid Firecannon", 54, "offensive"),
		sl(ItemLordDominiks, "Lord Dominik's Regards", 46, "pen"),
	}
}

func seedKog() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemGuinsoos, "Guinsoo's Rageblade", 90, "core"),
		sl(ItemNashors, "Nashor's Tooth", 82, "offensive"),
		sl(ItemBOTRK, "Blade of The Ruined King", 74, "offensive"),
		sl(ItemBerserkers, "Berserker's Greaves", 68, "boots"),
		sl(ItemRunaans, "Runaan's Hurricane", 58, "offensive"),
		sl(ItemVoidStaff, "Void Staff", 48, "pen"),
	}
}

func seedSamira() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemCollector, "The Collector", 88, "core"),
		sl(ItemInfinityEdge, "Infinity Edge", 80, "offensive"),
		sl(ItemBloodthirster, "Bloodthirster", 70, "offensive"),
		sl(ItemBerserkers, "Berserker's Greaves", 64, "boots"),
		sl(ItemShieldbow, "Immortal Shieldbow", 56, "defensive"),
		sl(ItemLordDominiks, "Lord Dominik's Regards", 48, "pen"),
	}
}

func seedADCYunTal() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemYunTal, "Yun Tal Wildarrows", 88, "core"),
		sl(ItemInfinityEdge, "Infinity Edge", 80, "offensive"),
		sl(ItemRapidFirecannon, "Rapid Firecannon", 70, "offensive"),
		sl(ItemBerserkers, "Berserker's Greaves", 64, "boots"),
		sl(ItemCollector, "The Collector", 56, "offensive"),
		sl(ItemLordDominiks, "Lord Dominik's Regards", 48, "pen"),
	}
}

func seedMissFortune() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemYoumuu, "Youmuu's Ghostblade", 88, "core"),
		sl(ItemCollector, "The Collector", 78, "offensive"),
		sl(ItemInfinityEdge, "Infinity Edge", 70, "offensive"),
		sl(ItemBerserkers, "Berserker's Greaves", 64, "boots"),
		sl(ItemRapidFirecannon, "Rapid Firecannon", 56, "offensive"),
		sl(ItemLordDominiks, "Lord Dominik's Regards", 48, "pen"),
	}
}

func seedEclipseDiver() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemEclipse, "Eclipse", 88, "core"),
		sl(ItemSunderedSky, "Sundered Sky", 80, "offensive"),
		sl(ItemBlackCleaver, "Black Cleaver", 70, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 64, "boots"),
		sl(ItemSteelcaps, "Plated Steelcaps", 58, "boots"),
		sl(ItemShojin, "Spear of Shojin", 52, "offensive"),
		sl(ItemDeathsDance, "Death's Dance", 44, "defensive"),
	}
}

func seedCritFighter() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemBerserkers, "Berserker's Greaves", 72, "boots"),
		sl(ItemYunTal, "Yun Tal Wildarrows", 88, "core"),
		sl(ItemInfinityEdge, "Infinity Edge", 80, "offensive"),
		sl(ItemShieldbow, "Immortal Shieldbow", 70, "defensive"),
		sl(ItemDeathsDance, "Death's Dance", 60, "defensive"),
		sl(ItemLordDominiks, "Lord Dominik's Regards", 50, "pen"),
		sl(ItemSteelcaps, "Plated Steelcaps", 42, "boots"),
	}
}

func seedYi() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemKraken, "Kraken Slayer", 88, "core"),
		sl(ItemBOTRK, "Blade of The Ruined King", 80, "offensive"),
		sl(ItemBerserkers, "Berserker's Greaves", 70, "boots"),
		sl(ItemGuinsoos, "Guinsoo's Rageblade", 62, "offensive"),
		sl(ItemWitsEnd, "Wit's End", 54, "defensive"),
		sl(ItemDeathsDance, "Death's Dance", 46, "defensive"),
	}
}

func seedTrynd() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemKraken, "Kraken Slayer", 88, "core"),
		sl(ItemInfinityEdge, "Infinity Edge", 80, "offensive"),
		sl(ItemBerserkers, "Berserker's Greaves", 70, "boots"),
		sl(ItemPhantomDancer, "Phantom Dancer", 62, "offensive"),
		sl(ItemLordDominiks, "Lord Dominik's Regards", 54, "pen"),
		sl(ItemDeathsDance, "Death's Dance", 46, "defensive"),
	}
}

func seedGangplank() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemTrinity, "Trinity Force", 88, "core"),
		sl(ItemInfinityEdge, "Infinity Edge", 76, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 68, "boots"),
		sl(ItemCollector, "The Collector", 58, "offensive"),
		sl(ItemLordDominiks, "Lord Dominik's Regards", 50, "pen"),
		sl(ItemSeryldas, "Serylda's Grudge", 42, "pen"),
	}
}

func seedFiora() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemRavenousHydra, "Ravenous Hydra", 88, "core"),
		sl(ItemTrinity, "Trinity Force", 78, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 68, "boots"),
		sl(ItemDeathsDance, "Death's Dance", 60, "defensive"),
		sl(ItemSteraks, "Sterak's Gage", 52, "defensive"),
		sl(ItemSteelcaps, "Plated Steelcaps", 46, "boots"),
	}
}

func seedIrelia() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemBOTRK, "Blade of The Ruined King", 90, "core"),
		sl(ItemWitsEnd, "Wit's End", 78, "defensive"),
		sl(ItemTrinity, "Trinity Force", 70, "offensive"),
		sl(ItemMercTreads, "Mercury's Treads", 64, "boots"),
		sl(ItemFrozenHeart, "Frozen Heart", 56, "defensive"),
		sl(ItemDeathsDance, "Death's Dance", 48, "defensive"),
	}
}

func seedOlaf() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemStridebreaker, "Stridebreaker", 88, "core"),
		sl(ItemBOTRK, "Blade of The Ruined King", 80, "offensive"),
		sl(ItemSteraks, "Sterak's Gage", 70, "defensive"),
		sl(ItemMercTreads, "Mercury's Treads", 62, "boots"),
		sl(ItemDeathsDance, "Death's Dance", 54, "defensive"),
		sl(ItemSteelcaps, "Plated Steelcaps", 46, "boots"),
	}
}

func seedWarwick() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemBOTRK, "Blade of The Ruined King", 90, "core"),
		sl(ItemTitanic, "Titanic Hydra", 78, "offensive"),
		sl(ItemWitsEnd, "Wit's End", 68, "defensive"),
		sl(ItemMercTreads, "Mercury's Treads", 60, "boots"),
		sl(ItemSteraks, "Sterak's Gage", 52, "defensive"),
		sl(ItemDeathsDance, "Death's Dance", 44, "defensive"),
	}
}

func seedKayn() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemEclipse, "Eclipse", 88, "core"),
		sl(ItemBlackCleaver, "Black Cleaver", 80, "offensive"),
		sl(ItemSunderedSky, "Sundered Sky", 72, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 64, "boots"),
		sl(ItemShojin, "Spear of Shojin", 56, "offensive"),
		sl(ItemDeathsDance, "Death's Dance", 48, "defensive"),
	}
}

func seedUdyr() []Slot {
	return []Slot{
		sl(ItemDoransRing, "Doran's Ring", 100, "start"),
		sl(ItemLiandrys, "Liandry's Torment", 88, "core"),
		sl(ItemRylais, "Rylai's Crystal Scepter", 78, "offensive"),
		sl(ItemRiftmaker, "Riftmaker", 70, "offensive"),
		sl(ItemMercTreads, "Mercury's Treads", 62, "boots"),
		sl(ItemZhonyas, "Zhonya's Hourglass", 52, "defensive"),
		sl(ItemForceOfNature, "Force of Nature", 44, "defensive"),
	}
}

func seedShyvana() []Slot {
	return []Slot{
		sl(ItemDoransRing, "Doran's Ring", 100, "start"),
		sl(ItemNashors, "Nashor's Tooth", 88, "core"),
		sl(ItemRiftmaker, "Riftmaker", 80, "offensive"),
		sl(ItemRylais, "Rylai's Crystal Scepter", 70, "offensive"),
		sl(ItemSorcs, "Sorcerer's Shoes", 62, "boots"),
		sl(ItemLiandrys, "Liandry's Torment", 54, "offensive"),
		sl(ItemRabadons, "Rabadon's Deathcap", 46, "offensive"),
	}
}

func seedVolibear() []Slot {
	return []Slot{
		sl(ItemDoransShield, "Doran's Shield", 100, "start"),
		sl(ItemNashors, "Nashor's Tooth", 86, "core"),
		sl(ItemRiftmaker, "Riftmaker", 78, "offensive"),
		sl(ItemIceborn, "Iceborn Gauntlet", 70, "defensive"),
		sl(ItemMercTreads, "Mercury's Treads", 62, "boots"),
		sl(ItemSpiritVisage, "Spirit Visage", 54, "defensive"),
		sl(ItemUnendingDespair, "Unending Despair", 46, "defensive"),
	}
}

func ItemUnending() int { return 2502 }

func seedJayce() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemTrinity, "Trinity Force", 88, "core"),
		sl(ItemManamune, "Manamune", 80, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 68, "boots"),
		sl(ItemSeryldas, "Serylda's Grudge", 60, "pen"),
		sl(ItemShojin, "Spear of Shojin", 52, "offensive"),
		sl(ItemLordDominiks, "Lord Dominik's Regards", 44, "pen"),
	}
}

func seedStrideJuggernaut() []Slot {
	return []Slot{
		sl(ItemDoransShield, "Doran's Shield", 100, "start"),
		sl(ItemStridebreaker, "Stridebreaker", 90, "core"),
		sl(ItemBlackCleaver, "Black Cleaver", 80, "offensive"),
		sl(ItemSteraks, "Sterak's Gage", 70, "defensive"),
		sl(ItemSteelcaps, "Plated Steelcaps", 62, "boots"),
		sl(ItemMercTreads, "Mercury's Treads", 56, "boots"),
		sl(ItemDeathsDance, "Death's Dance", 50, "defensive"),
		sl(ItemForceOfNature, "Force of Nature", 40, "defensive"),
	}
}

func seedUrgot() []Slot {
	return []Slot{
		sl(ItemDoransBlade, "Doran's Blade", 100, "start"),
		sl(ItemBlackCleaver, "Black Cleaver", 88, "core"),
		sl(ItemSteraks, "Sterak's Gage", 78, "defensive"),
		sl(ItemTitanic, "Titanic Hydra", 68, "offensive"),
		sl(ItemJakSho, "Jak'Sho, The Protean", 58, "defensive"),
		sl(ItemSteelcaps, "Plated Steelcaps", 52, "boots"),
		sl(ItemThornmail, "Thornmail", 44, "defensive"),
	}
}

func seedYorick() []Slot {
	return []Slot{
		sl(ItemDoransShield, "Doran's Shield", 100, "start"),
		sl(ItemTrinity, "Trinity Force", 88, "core"),
		sl(ItemSunderedSky, "Sundered Sky", 78, "offensive"),
		sl(ItemShojin, "Spear of Shojin", 68, "offensive"),
		sl(ItemSteraks, "Sterak's Gage", 58, "defensive"),
		sl(ItemSteelcaps, "Plated Steelcaps", 50, "boots"),
		sl(ItemMercTreads, "Mercury's Treads", 44, "boots"),
	}
}

func seedIllaoi() []Slot {
	return []Slot{
		sl(ItemDoransShield, "Doran's Shield", 100, "start"),
		sl(ItemIceborn, "Iceborn Gauntlet", 88, "core"),
		sl(ItemBlackCleaver, "Black Cleaver", 78, "offensive"),
		sl(ItemSteraks, "Sterak's Gage", 70, "defensive"),
		sl(ItemSteelcaps, "Plated Steelcaps", 60, "boots"),
		sl(ItemMercTreads, "Mercury's Treads", 54, "boots"),
		sl(ItemSpiritVisage, "Spirit Visage", 46, "defensive"),
	}
}

func seedGwen() []Slot {
	return []Slot{
		sl(ItemDoransShield, "Doran's Shield", 100, "start"),
		sl(ItemNashors, "Nashor's Tooth", 90, "core"),
		sl(ItemRiftmaker, "Riftmaker", 80, "offensive"),
		sl(ItemRylais, "Rylai's Crystal Scepter", 68, "offensive"),
		sl(ItemSorcs, "Sorcerer's Shoes", 60, "boots"),
		sl(ItemZhonyas, "Zhonya's Hourglass", 50, "defensive"),
		sl(ItemRabadons, "Rabadon's Deathcap", 42, "offensive"),
	}
}

func seedShen() []Slot {
	return []Slot{
		sl(ItemDoransShield, "Doran's Shield", 100, "start"),
		sl(ItemSunfire, "Sunfire Aegis", 88, "core"),
		sl(ItemTitanic, "Titanic Hydra", 78, "offensive"),
		sl(ItemJakSho, "Jak'Sho, The Protean", 68, "defensive"),
		sl(ItemThornmail, "Thornmail", 56, "defensive"),
		sl(ItemSteelcaps, "Plated Steelcaps", 50, "boots"),
		sl(ItemMercTreads, "Mercury's Treads", 44, "boots"),
	}
}

func seedSinged() []Slot {
	return []Slot{
		sl(ItemDoransRing, "Doran's Ring", 100, "start"),
		sl(ItemRylais, "Rylai's Crystal Scepter", 90, "core"),
		sl(ItemLiandrys, "Liandry's Torment", 82, "offensive"),
		sl(ItemSwifties, "Boots of Swiftness", 74, "boots"),
		sl(ItemBlackfire, "Blackfire Torch", 66, "offensive"),
		sl(ItemDeadMans, "Dead Man's Plate", 64, "defensive"),
		sl(ItemForceOfNature, "Force of Nature", 48, "defensive"),
	}
}

func seedRoAScaler() []Slot {
	return []Slot{
		sl(ItemDoransRing, "Doran's Ring", 100, "start"),
		sl(ItemTear, "Tear of the Goddess", 92, "component"),
		sl(ItemRodOfAges, "Rod of Ages", 86, "core"),
		sl(ItemSeraphs, "Seraph's Embrace", 74, "offensive"),
		sl(ItemMercTreads, "Mercury's Treads", 64, "boots"),
		sl(ItemRiftmaker, "Riftmaker", 58, "offensive"),
		sl(ItemZhonyas, "Zhonya's Hourglass", 48, "defensive"),
		sl(ItemRabadons, "Rabadon's Deathcap", 40, "offensive"),
	}
}

func seedAPNashor() []Slot {
	return []Slot{
		sl(ItemDoransRing, "Doran's Ring", 100, "start"),
		sl(ItemNashors, "Nashor's Tooth", 90, "core"),
		sl(ItemShadowflame, "Shadowflame", 76, "offensive"),
		sl(ItemSorcs, "Sorcerer's Shoes", 68, "boots"),
		sl(ItemRabadons, "Rabadon's Deathcap", 60, "offensive"),
		sl(ItemZhonyas, "Zhonya's Hourglass", 50, "defensive"),
		sl(ItemVoidStaff, "Void Staff", 42, "pen"),
		sl(ItemLichBane, "Lich Bane", 36, "offensive"),
	}
}

func seedVladimir() []Slot {
	return []Slot{
		sl(ItemDoransShield, "Doran's Shield", 100, "start"),
		sl(ItemCosmicDrive, "Cosmic Drive", 88, "core"),
		sl(ItemRiftmaker, "Riftmaker", 80, "offensive"),
		sl(ItemLiandrys, "Liandry's Torment", 70, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 62, "boots"),
		sl(ItemRabadons, "Rabadon's Deathcap", 54, "offensive"),
		sl(ItemZhonyas, "Zhonya's Hourglass", 46, "defensive"),
	}
}

func seedCassiopeia() []Slot {
	return []Slot{
		sl(ItemDoransRing, "Doran's Ring", 100, "start"),
		sl(ItemTear, "Tear of the Goddess", 92, "component"),
		sl(ItemRodOfAges, "Rod of Ages", 86, "core"),
		sl(ItemArchangel, "Archangel's Staff", 76, "offensive"),
		sl(ItemRylais, "Rylai's Crystal Scepter", 68, "offensive"),
		sl(ItemLiandrys, "Liandry's Torment", 60, "offensive"),
		sl(ItemRabadons, "Rabadon's Deathcap", 50, "offensive"),
		sl(ItemVoidStaff, "Void Staff", 42, "pen"),
	}
}

func seedKatarina() []Slot {
	return []Slot{
		sl(ItemDarkSeal, "Dark Seal", 105, "start"),
		sl(ItemNashors, "Nashor's Tooth", 90, "core"),
		sl(ItemLichBane, "Lich Bane", 78, "offensive"),
		sl(ItemSorcs, "Sorcerer's Shoes", 68, "boots"),
		sl(ItemShadowflame, "Shadowflame", 60, "offensive"),
		sl(ItemZhonyas, "Zhonya's Hourglass", 50, "defensive"),
		sl(ItemRabadons, "Rabadon's Deathcap", 42, "offensive"),
	}
}

func seedLillia() []Slot {
	return []Slot{
		sl(ItemDoransRing, "Doran's Ring", 100, "start"),
		sl(ItemLiandrys, "Liandry's Torment", 90, "core"),
		sl(ItemRylais, "Rylai's Crystal Scepter", 82, "offensive"),
		sl(ItemSorcs, "Sorcerer's Shoes", 68, "boots"),
		sl(ItemRiftmaker, "Riftmaker", 62, "offensive"),
		sl(ItemRabadons, "Rabadon's Deathcap", 52, "offensive"),
		sl(ItemZhonyas, "Zhonya's Hourglass", 42, "defensive"),
	}
}

func seedSylasJungle() []Slot {
	return []Slot{
		sl(ItemDoransRing, "Doran's Ring", 100, "start"),
		sl(ItemRocketbelt, "Hextech Rocketbelt", 90, "core"),
		sl(ItemRiftmaker, "Riftmaker", 78, "offensive"),
		sl(ItemSorcs, "Sorcerer's Shoes", 66, "boots"),
		sl(ItemLichBane, "Lich Bane", 58, "offensive"),
		sl(ItemRabadons, "Rabadon's Deathcap", 50, "offensive"),
		sl(ItemZhonyas, "Zhonya's Hourglass", 42, "defensive"),
	}
}

func seedLudenHorizon() []Slot {
	return []Slot{
		sl(ItemDoransRing, "Doran's Ring", 100, "start"),
		sl(ItemLostChapter, "Lost Chapter", 90, "component"),
		sl(ItemLudens, "Luden's Companion", 88, "core"),
		sl(ItemSorcs, "Sorcerer's Shoes", 70, "boots"),
		sl(ItemHorizon, "Horizon Focus", 64, "offensive"),
		sl(ItemShadowflame, "Shadowflame", 56, "offensive"),
		sl(ItemRabadons, "Rabadon's Deathcap", 48, "offensive"),
		sl(ItemVoidStaff, "Void Staff", 40, "pen"),
	}
}

func seedSwain() []Slot {
	return []Slot{
		sl(ItemDoransRing, "Doran's Ring", 100, "start"),
		sl(ItemLiandrys, "Liandry's Torment", 90, "core"),
		sl(ItemRylais, "Rylai's Crystal Scepter", 82, "offensive"),
		sl(ItemSorcs, "Sorcerer's Shoes", 68, "boots"),
		sl(ItemZhonyas, "Zhonya's Hourglass", 54, "defensive"),
		sl(ItemRabadons, "Rabadon's Deathcap", 46, "offensive"),
		sl(ItemRiftmaker, "Riftmaker", 38, "offensive"),
	}
}

func seedTwistedFate() []Slot {
	return []Slot{
		sl(ItemDoransRing, "Doran's Ring", 100, "start"),
		sl(ItemLichBane, "Lich Bane", 90, "core"),
		sl(ItemRapidFirecannon, "Rapid Firecannon", 78, "offensive"),
		sl(ItemSorcs, "Sorcerer's Shoes", 68, "boots"),
		sl(ItemLudens, "Luden's Companion", 58, "offensive"),
		sl(ItemRabadons, "Rabadon's Deathcap", 50, "offensive"),
		sl(ItemZhonyas, "Zhonya's Hourglass", 42, "defensive"),
	}
}

func seedAPLich() []Slot {
	return []Slot{
		sl(ItemDoransRing, "Doran's Ring", 100, "start"),
		sl(ItemLichBane, "Lich Bane", 88, "core"),
		sl(ItemSorcs, "Sorcerer's Shoes", 80, "boots"),
		sl(ItemShadowflame, "Shadowflame", 72, "offensive"),
		sl(ItemZhonyas, "Zhonya's Hourglass", 64, "defensive"),
		sl(ItemRabadons, "Rabadon's Deathcap", 56, "offensive"),
		sl(ItemVoidStaff, "Void Staff", 48, "pen"),
		sl(ItemBansheeVeil, "Banshee's Veil", 44, "defensive"),
		sl(ItemMercTreads, "Mercury's Treads", 50, "boots"),
	}
}

func seedMundo() []Slot {
	return []Slot{
		sl(ItemDoransShield, "Doran's Shield", 100, "start"),
		sl(ItemHeartsteel, "Heartsteel", 88, "core"),
		sl(ItemWarmogs, "Warmog's Armor", 76, "defensive"),
		sl(ItemSpiritVisage, "Spirit Visage", 68, "defensive"),
		sl(ItemMercTreads, "Mercury's Treads", 60, "boots"),
		sl(ItemTitanic, "Titanic Hydra", 52, "offensive"),
		sl(ItemThornmail, "Thornmail", 44, "defensive"),
		sl(ItemSteelcaps, "Plated Steelcaps", 38, "boots"),
	}
}

func seedChoGath() []Slot {
	return []Slot{
		sl(ItemDoransShield, "Doran's Shield", 100, "start"),
		sl(ItemHeartsteel, "Heartsteel", 90, "core"),
		sl(ItemHollowRadiance, "Hollow Radiance", 80, "core"),
		sl(ItemWarmogs, "Warmog's Armor", 74, "core"),
		sl(ItemGiantsBelt, "Giant's Belt", 88, "component"),
		sl(ItemMercTreads, "Mercury's Treads", 66, "boots"),
		sl(ItemSteelcaps, "Plated Steelcaps", 60, "boots"),
		sl(ItemThornmail, "Thornmail", 64, "defensive"),
		sl(ItemJakSho, "Jak'Sho, The Protean", 56, "defensive"),
		sl(ItemUnendingDespair, "Unending Despair", 40, "defensive"),
	}
}

func seedNasus() []Slot {
	return []Slot{
		sl(ItemDoransShield, "Doran's Shield", 100, "start"),
		sl(ItemIceborn, "Iceborn Gauntlet", 88, "core"),
		sl(ItemFrozenHeart, "Frozen Heart", 76, "defensive"),
		sl(ItemShojin, "Spear of Shojin", 66, "offensive"),
		sl(ItemMercTreads, "Mercury's Treads", 58, "boots"),
		sl(ItemSpiritVisage, "Spirit Visage", 50, "defensive"),
		sl(ItemJakSho, "Jak'Sho, The Protean", 42, "defensive"),
	}
}

func seedPykeSupport() []Slot {
	return []Slot{
		sl(ItemWorldAtlas, "World Atlas", 100, "start"),
		sl(ItemYoumuu, "Youmuu's Ghostblade", 88, "core"),
		sl(ItemOpportunity, "Opportunity", 76, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 68, "boots"),
		sl(ItemUmbral, "Umbral Glaive", 60, "offensive"),
		sl(ItemEdgeOfNight(), "Edge of Night", 52, "defensive"),
		sl(ItemSeryldas, "Serylda's Grudge", 44, "pen"),
	}
}

func seedSennaSupport() []Slot {
	return []Slot{
		sl(ItemWorldAtlas, "World Atlas", 100, "start"),
		sl(ItemMandate, "Imperial Mandate", 86, "core"),
		sl(ItemEclipse, "Eclipse", 76, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 66, "boots"),
		sl(ItemBlackCleaver, "Black Cleaver", 58, "offensive"),
		sl(ItemCollector, "The Collector", 50, "offensive"),
		sl(ItemRapidFirecannon, "Rapid Firecannon", 42, "offensive"),
	}
}

func seedBard() []Slot {
	return []Slot{
		sl(ItemWorldAtlas, "World Atlas", 100, "start"),
		sl(ItemLocket, "Locket of the Iron Solari", 84, "core"),
		sl(ItemDeadMans, "Dead Man's Plate", 74, "defensive"),
		sl(ItemShurelyas, "Shurelya's Battlesong", 66, "utility"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 58, "boots"),
		sl(ItemRedemption, "Redemption", 50, "utility"),
		sl(ItemMikaels, "Mikael's Blessing", 42, "utility"),
	}
}

func seedRakan() []Slot {
	return []Slot{
		sl(ItemWorldAtlas, "World Atlas", 100, "start"),
		sl(ItemLocket, "Locket of the Iron Solari", 86, "core"),
		sl(ItemShurelyas, "Shurelya's Battlesong", 76, "utility"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 66, "boots"),
		sl(ItemRedemption, "Redemption", 58, "utility"),
		sl(ItemKnightsVow, "Knight's Vow", 50, "utility"),
		sl(ItemMikaels, "Mikael's Blessing", 42, "utility"),
	}
}

func seedRenata() []Slot {
	return []Slot{
		sl(ItemWorldAtlas, "World Atlas", 100, "start"),
		sl(ItemMandate, "Imperial Mandate", 86, "core"),
		sl(ItemLocket, "Locket of the Iron Solari", 76, "defensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 66, "boots"),
		sl(ItemRedemption, "Redemption", 58, "utility"),
		sl(ItemShurelyas, "Shurelya's Battlesong", 50, "utility"),
		sl(ItemMikaels, "Mikael's Blessing", 42, "utility"),
	}
}

func seedKarmaSupport() []Slot {
	return []Slot{
		sl(ItemWorldAtlas, "World Atlas", 100, "start"),
		sl(ItemMoonstone, "Moonstone Renewer", 86, "core"),
		sl(ItemMandate, "Imperial Mandate", 76, "utility"),
		sl(ItemShurelyas, "Shurelya's Battlesong", 68, "utility"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 60, "boots"),
		sl(ItemRedemption, "Redemption", 52, "utility"),
		sl(ItemMikaels, "Mikael's Blessing", 44, "utility"),
	}
}

func seedPantheonSupport() []Slot {
	return []Slot{
		sl(ItemWorldAtlas, "World Atlas", 100, "start"),
		sl(ItemEclipse, "Eclipse", 86, "core"),
		sl(ItemUmbral, "Umbral Glaive", 76, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 66, "boots"),
		sl(ItemBlackCleaver, "Black Cleaver", 58, "offensive"),
		sl(ItemSunderedSky, "Sundered Sky", 50, "offensive"),
		sl(ItemKnightsVow, "Knight's Vow", 42, "utility"),
	}
}

func seedSupportADC() []Slot {
	return []Slot{
		sl(ItemWorldAtlas, "World Atlas", 100, "start"),
		sl(ItemCollector, "The Collector", 82, "core"),
		sl(ItemInfinityEdge, "Infinity Edge", 72, "offensive"),
		sl(ItemIonians, "Ionian Boots of Lucidity", 64, "boots"),
		sl(ItemRapidFirecannon, "Rapid Firecannon", 56, "offensive"),
		sl(ItemLordDominiks, "Lord Dominik's Regards", 48, "pen"),
	}
}

func isAPFamily(family string) bool {
	switch family {
	case FamMageBurst, FamMagePoke, FamMageRoA, FamMageNashor, FamMageVlad, FamMageCass,
		FamMageKat, FamMageLillia, FamMageSylas, FamMageLuden, FamMageSwain, FamMageTF,
		FamAssassinAP, FamAssassinLich, FamFighterAP, FamFighterShyvana,
		FamFighterUdyr, FamFighterVoli, FamFighterGwen, FamADCKayle, FamSupportMage,
		FamSupportEnch, FamSupportKarma, FamMageSinged:
		return true
	default:
		return false
	}
}

func isADCFamily(family string) bool {
	switch family {
	case FamADCCrit, FamADCOnHit, FamADCEzreal, FamADCJhin, FamADCZeri, FamADCSmolder,
		FamADCCorki, FamADCKayle, FamADCSenna, FamADCKalista, FamADCKaisa, FamADCLucian,
		FamADCKog, FamADCSamira, FamADCYunTal, FamADCMF, FamADCGraves:
		return true
	default:
		return false
	}
}

func isTankFamily(family string) bool {
	switch family {
	case FamTankSunfire, FamTankMundo, FamTankNasus, FamTankCho, FamTankShen, FamFighterIceborn, FamSupportEngage:
		return true
	default:
		return false
	}
}
