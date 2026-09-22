package rules

import "lol-build-overlay/internal/tags"

type champSpec struct {
	def    string
	byRole map[string]string
}

func spec(def string, rolePairs ...string) champSpec {
	s := champSpec{def: def}
	if len(rolePairs) > 0 {
		s.byRole = make(map[string]string, len(rolePairs)/2)
		for i := 0; i+1 < len(rolePairs); i += 2 {
			s.byRole[rolePairs[i]] = rolePairs[i+1]
		}
	}
	return s
}

// champFamily is the champion×role core table. Role keys are NormalizePosition values.
var champFamily = map[string]champSpec{
	"Aatrox":       spec(FamFighterEclipse),
	"Ahri":         spec(FamMageBurst),
	"Akali":        spec(FamAssassinLich),
	"Akshan":       spec(FamADCCrit),
	"Alistar":      spec(FamSupportEngage),
	"Ambessa":      spec(FamFighterEclipse),
	"Amumu":        spec(FamTankSunfire, "UTILITY", FamSupportEngage),
	"Anivia":       spec(FamMageRoA),
	"Annie":        spec(FamMageBurst),
	"Aphelios":     spec(FamADCCrit, "UTILITY", FamSupportADC),
	"Ashe":         spec(FamADCCrit),
	"AurelionSol":  spec(FamMagePoke),
	"Aurora":       spec(FamMageBurst),
	"Azir":         spec(FamMageNashor),
	"Bard":         spec(FamSupportBard),
	"Belveth":      spec(FamFighterYi),
	"Blitzcrank":   spec(FamSupportEngage),
	"Brand":        spec(FamSupportMage, "MIDDLE", FamMagePoke),
	"Braum":        spec(FamSupportEngage, "MIDDLE", FamTankSunfire, "TOP", FamTankSunfire),
	"Briar":        spec(FamFighterEclipse),
	"Caitlyn":      spec(FamADCYunTal),
	"Camille":      spec(FamFighterTrinity),
	"Cassiopeia":   spec(FamMageCass),
	"Chogath":      spec(FamTankCho),
	"Corki":        spec(FamADCCorki),
	"Darius":       spec(FamFighterStride),
	"Diana":        spec(FamMageNashor),
	"DrMundo":      spec(FamTankMundo),
	"Draven":       spec(FamADCSamira),
	"Ekko":         spec(FamAssassinLich),
	"Elise":        spec(FamAssassinAP),
	"Evelynn":      spec(FamAssassinLich),
	"Ezreal":       spec(FamADCEzreal),
	"Fiddlesticks": spec(FamMageBurst),
	"Fiora":        spec(FamFighterFiora),
	"Fizz":         spec(FamAssassinAP),
	"Galio":        spec(FamMageBurst, "TOP", FamTankSunfire, "UTILITY", FamSupportEngage),
	"Gangplank":    spec(FamFighterGP),
	"Garen":        spec(FamFighterStride),
	"Gnar":         spec(FamFighterTrinity),
	"Gragas":       spec(FamTankSunfire, "MIDDLE", FamMageRoA, "UTILITY", FamSupportEngage),
	"Graves":       spec(FamADCGraves),
	"Gwen":         spec(FamFighterGwen),
	"Hecarim":      spec(FamFighterTrinity),
	"Heimerdinger": spec(FamMagePoke, "UTILITY", FamSupportMage),
	"Hwei":         spec(FamMagePoke),
	"Illaoi":       spec(FamFighterIllaoi),
	"Irelia":       spec(FamFighterIrelia),
	"Ivern":        spec(FamSupportEnch),
	"Janna":        spec(FamSupportEnch),
	"JarvanIV":     spec(FamFighterEclipse),
	"Jax":          spec(FamFighterTrinity),
	"Jayce":        spec(FamFighterJayce),
	"Jhin":         spec(FamADCJhin),
	"Jinx":         spec(FamADCYunTal),
	"KSante":       spec(FamFighterIceborn, "UTILITY", FamSupportEngage),
	"Kaisa":        spec(FamADCKaisa),
	"Kalista":      spec(FamADCKalista),
	"Karma":        spec(FamSupportKarma, "MIDDLE", FamMagePoke, "TOP", FamMagePoke),
	"Karthus":      spec(FamMagePoke),
	"Kassadin":     spec(FamMageRoA),
	"Katarina":     spec(FamMageKat),
	"Kayle":        spec(FamADCKayle),
	"Kayn":         spec(FamFighterKayn),
	"Kennen":       spec(FamMagePoke),
	"Khazix":       spec(FamAssassinAD),
	"Kindred":      spec(FamADCCrit),
	"Kled":         spec(FamFighterTrinity),
	"KogMaw":       spec(FamADCKog),
	"Leblanc":      spec(FamAssassinAP),
	"LeeSin":       spec(FamFighterEclipse),
	"Leona":        spec(FamSupportEngage),
	"Lillia":       spec(FamMageLillia),
	"Lissandra":    spec(FamMageBurst),
	"Locke":        spec(FamAssassinAD),
	"Lucian":       spec(FamADCLucian),
	"Lulu":         spec(FamSupportEnch),
	"Lux":          spec(FamSupportMage, "MIDDLE", FamMageLuden),
	"Malphite":     spec(FamTankSunfire, "UTILITY", FamSupportEngage),
	"Malzahar":     spec(FamMagePoke),
	"Maokai":       spec(FamTankSunfire, "UTILITY", FamSupportEngage),
	"MasterYi":     spec(FamFighterYi),
	"Mel":          spec(FamMageBurst),
	"Milio":        spec(FamSupportEnch),
	"MissFortune":  spec(FamADCMF),
	"MonkeyKing":   spec(FamFighterTrinity),
	"Mordekaiser":  spec(FamFighterAP),
	"Morgana":      spec(FamSupportMage, "MIDDLE", FamMageBurst),
	"Naafiri":      spec(FamAssassinAD),
	"Nami":         spec(FamSupportEnch),
	"Nasus":        spec(FamTankNasus),
	"Nautilus":     spec(FamSupportEngage),
	"Neeko":        spec(FamMageBurst, "UTILITY", FamSupportMage),
	"Nidalee":      spec(FamAssassinAP),
	"Nilah":        spec(FamADCSamira),
	"Nocturne":     spec(FamFighterEclipse),
	"Nunu":         spec(FamTankSunfire, "UTILITY", FamSupportEngage),
	"Olaf":         spec(FamFighterOlaf),
	"Orianna":      spec(FamMageLuden),
	"Ornn":         spec(FamTankSunfire, "UTILITY", FamSupportEngage),
	"Pantheon":     spec(FamFighterEclipse, "UTILITY", FamSupportPantheon),
	"Poppy":        spec(FamFighterIceborn, "JUNGLE", FamTankSunfire, "UTILITY", FamSupportEngage),
	"Pyke":         spec(FamAssassinAD, "UTILITY", FamSupportPyke),
	"Qiyana":       spec(FamAssassinAD),
	"Quinn":        spec(FamADCCrit),
	"Rakan":        spec(FamSupportRakan),
	"Rammus":       spec(FamTankSunfire, "UTILITY", FamSupportEngage),
	"RekSai":       spec(FamFighterEclipse),
	"Rell":         spec(FamSupportEngage),
	"Renata":       spec(FamSupportRenata),
	"Renekton":     spec(FamFighterTrinity),
	"Rengar":       spec(FamAssassinAD),
	"Riven":        spec(FamFighterEclipse),
	"Rumble":       spec(FamMagePoke),
	"Ryze":         spec(FamMageRoA),
	"Samira":       spec(FamADCSamira),
	"Sejuani":      spec(FamTankSunfire, "UTILITY", FamSupportEngage),
	"Senna":        spec(FamADCSenna, "UTILITY", FamSupportSenna),
	"Seraphine":    spec(FamSupportEnch, "MIDDLE", FamMagePoke, "BOTTOM", FamMagePoke),
	"Sett":         spec(FamFighterStride, "UTILITY", FamSupportEngage),
	"Shaco":        spec(FamAssassinAD),
	"Shen":         spec(FamTankShen, "UTILITY", FamSupportEngage),
	"Shyvana":      spec(FamFighterShyvana),
	"Singed":       spec(FamMageSinged),
	"Sion":         spec(FamTankSunfire, "UTILITY", FamSupportEngage),
	"Sivir":        spec(FamADCCrit),
	"Skarner":      spec(FamFighterIceborn, "UTILITY", FamSupportEngage),
	"Smolder":      spec(FamADCSmolder),
	"Sona":         spec(FamSupportEnch),
	"Soraka":       spec(FamSupportEnch),
	"Swain":        spec(FamMageSwain, "UTILITY", FamSupportMage),
	"Sylas":        spec(FamMageSylas),
	"Syndra":       spec(FamMageBurst),
	"TahmKench":    spec(FamTankSunfire, "UTILITY", FamSupportEngage),
	"Taliyah":      spec(FamMagePoke),
	"Talon":        spec(FamAssassinAD),
	"Taric":        spec(FamSupportEngage),
	"Teemo":        spec(FamMagePoke),
	"Thresh":       spec(FamSupportEngage),
	"Tristana":     spec(FamADCCrit),
	"Trundle":      spec(FamFighterTrinity),
	"Tryndamere":   spec(FamFighterTrynd),
	"TwistedFate":  spec(FamMageTF),
	"Twitch":       spec(FamADCCrit),
	"Udyr":         spec(FamFighterUdyr),
	"Urgot":        spec(FamFighterUrgot),
	"Varus":        spec(FamADCOnHit),
	"Vayne":        spec(FamADCOnHit),
	"Veigar":       spec(FamMageBurst),
	"Velkoz":       spec(FamMageLuden, "UTILITY", FamSupportMage),
	"Vex":          spec(FamMageBurst),
	"Vi":           spec(FamFighterEclipse),
	"Viego":        spec(FamFighterTrinity),
	"Viktor":       spec(FamMageLuden),
	"Vladimir":     spec(FamMageVlad),
	"Volibear":     spec(FamFighterTrinity, "JUNGLE", FamFighterVoli),
	"Warwick":      spec(FamFighterWarwick),
	"Xayah":        spec(FamADCCrit),
	"Xerath":       spec(FamMageLuden, "UTILITY", FamSupportMage),
	"XinZhao":      spec(FamFighterEclipse),
	"Yasuo":        spec(FamFighterCrit),
	"Yone":         spec(FamFighterCrit),
	"Yorick":       spec(FamFighterYorick),
	"Yunara":       spec(FamADCCrit),
	"Yuumi":        spec(FamSupportEnch),
	"Zaahen":       spec(FamFighterCleaver),
	"Zac":          spec(FamTankSunfire, "UTILITY", FamSupportEngage),
	"Zed":          spec(FamAssassinAD),
	"Zeri":         spec(FamADCZeri),
	"Ziggs":        spec(FamMagePoke, "UTILITY", FamSupportMage),
	"Zilean":       spec(FamSupportEnch, "MIDDLE", FamMagePoke),
	"Zoe":          spec(FamMageLuden),
	"Zyra":         spec(FamSupportMage, "MIDDLE", FamMagePoke),
}

func resolveFamily(champ, pos, primary string, allTags []string) (family, label string) {
	if s, ok := champFamily[champ]; ok {
		if pos != "" && s.byRole != nil {
			if f, ok := s.byRole[pos]; ok {
				return f, familyLabel(f)
			}
		}
		if s.def != "" {
			return s.def, familyLabel(s.def)
		}
	}
	if primary == tags.ClassSupport && pos != "UTILITY" {
		fam, lbl := soloLaneSupportFamily(allTags)
		return fam, lbl
	}
	_, lbl := seedByClass(primary, allTags)
	fam := familyFromClassLabel(lbl, primary, allTags)
	return fam, lbl
}

func familyFromClassLabel(label, primary string, allTags []string) string {
	switch label {
	case "mage":
		if hasTag(allTags, tags.StylePoke) && !hasTag(allTags, tags.StyleBurst) {
			return FamMagePoke
		}
		return FamMageBurst
	case "assassin/ap":
		return FamAssassinAP
	case "assassin":
		return FamAssassinAD
	case "marksman":
		return FamADCCrit
	case "fighter/ap":
		return FamFighterAP
	case "fighter":
		if hasTag(allTags, tags.ExtraJuggernaut) {
			return FamFighterCleaver
		}
		return FamFighterTrinity
	case "tank":
		return FamTankSunfire
	case "support":
		if hasTag(allTags, tags.StyleHealer) || hasTag(allTags, tags.StyleShield) {
			return FamSupportEnch
		}
		return FamSupportEngage
	default:
		_ = primary
		return FamMageBurst
	}
}

func soloLaneSupportFamily(allTags []string) (string, string) {
	has := func(t string) bool { return hasTag(allTags, t) }
	if has(tags.ClassTank) || has(tags.StyleEngage) || (has(tags.StylePeel) && has(tags.RangeMelee)) {
		return FamTankSunfire, "tank"
	}
	if has(tags.DamageAP) || has(tags.StyleHealer) || has(tags.StyleShield) || has(tags.StylePoke) {
		if has(tags.StylePoke) && !has(tags.StyleBurst) {
			return FamMagePoke, "mage"
		}
		return FamMageBurst, "mage"
	}
	return FamTankSunfire, "tank"
}

func supportFamilyFor(champ, primary string, allTags []string) string {
	if s, ok := champFamily[champ]; ok {
		if s.byRole != nil {
			if f, ok := s.byRole["UTILITY"]; ok {
				return f
			}
		}
		if isSupportFamily(s.def) {
			return s.def
		}
		// Carry whose default is not a support line (e.g. missing UTILITY override).
		if s.def == FamAssassinAD {
			return FamSupportPyke
		}
		if len(s.def) >= 4 && s.def[:4] == "adc-" {
			return FamSupportADC
		}
	}
	return supportFallbackFamily(primary, allTags)
}

func isSupportFamily(family string) bool {
	switch family {
	case FamSupportEnch, FamSupportEngage, FamSupportMage, FamSupportPyke,
		FamSupportSenna, FamSupportBard, FamSupportRakan, FamSupportRenata,
		FamSupportKarma, FamSupportPantheon, FamSupportADC:
		return true
	default:
		return false
	}
}

// supportFallbackFamily never uses "has AP → Blackfire". Function first: engage/tank,
// then healer, then mage primary, then marksman/assassin carry, else enchanter.
func supportFallbackFamily(primary string, allTags []string) string {
	has := func(t string) bool {
		if primary == t {
			return true
		}
		return hasTag(allTags, t)
	}
	if has(tags.ClassTank) || has(tags.StyleEngage) || (has(tags.StylePeel) && has(tags.RangeMelee)) {
		return FamSupportEngage
	}
	if has(tags.StyleHealer) {
		return FamSupportEnch
	}
	if primary == tags.ClassMage {
		return FamSupportMage
	}
	if primary == tags.ClassAssassin && has(tags.DamageAD) {
		return FamSupportPyke
	}
	if primary == tags.ClassMarksman {
		return FamSupportADC
	}
	if has(tags.StyleShield) && has(tags.StylePeel) {
		return FamSupportEnch
	}
	if has(tags.RangeMelee) {
		return FamSupportEngage
	}
	return FamSupportEnch
}

// ApplyItemGates strips illegal purchases (Cass boots, Vlad mana items, Jhin AS boots).
func ApplyItemGates(champ string, build []Slot) []Slot {
	out := make([]Slot, 0, len(build))
	for _, s := range build {
		if gatedItem(champ, s) {
			continue
		}
		out = append(out, s)
	}
	return out
}

func gatedItem(champ string, s Slot) bool {
	if skipsZhonya(champ) && (s.ItemID == ItemZhonyas || s.ItemID == ItemSeekers) {
		return true
	}
	switch champ {
	case "Cassiopeia":
		return s.Role == "boots" || IsBootsID(s.ItemID)
	case "Yuumi":
		return s.Role == "boots" || IsBootsID(s.ItemID)
	case "Vladimir":
		switch s.ItemID {
		case ItemLostChapter, ItemMalignance, ItemBlackfire, ItemLudens, ItemRodOfAges,
			ItemTear, ItemArchangel, ItemSeraphs, ItemManamune, ItemDoransRing:
			return true
		}
	case "Jhin":
		return s.ItemID == ItemBerserkers || s.ItemID == ItemKraken
	}
	return false
}

func champIsTank(champ string) bool {
	s, ok := champFamily[champ]
	if !ok {
		return false
	}
	return isTankFamily(s.def)
}

func champIsADC(champ string) bool {
	s, ok := champFamily[champ]
	if !ok {
		return false
	}
	return isADCFamily(s.def)
}

func wantsTankGrievous(playerPrimary string, build []Slot) bool {
	if playerPrimary == tags.ClassMarksman {
		return false
	}
	if playerPrimary == tags.ClassTank || champLooksLikeTank(build) {
		return true
	}
	return false
}

// grievousOutranksMR: Thornmail only if AA+heal is a bigger problem than AP burst.
// A support healer into Sylas/Malphite is FoN, not spikes.
func grievousOutranksMR(heal, apBurst, ad float64) bool {
	if heal < 0.12 {
		return false
	}
	return ad > 0.2 && ad+heal > apBurst+0.12
}

func champLooksLikeTank(build []Slot) bool {
	return hasAny(build, ItemHeartsteel, ItemSunfire, ItemIceborn, ItemWarmogs, ItemJakSho, ItemHollowRadiance, ItemThornmail, ItemDeadMans)
}

func junglePetForChamp(champ, family, primary string, allTags []string) (int, string) {
	switch champ {
	case "Ivern":
		return ItemJungleGustwalker, "Gustwalker Hatchling"
	case "Rammus", "Sejuani", "Amumu", "Zac", "Nunu", "Maokai", "Ornn", "Sion",
		"Malphite", "Chogath", "DrMundo", "TahmKench", "Poppy", "Shen":
		return ItemJungleMosstomper, "Mosstomper Seedling"
	}
	if isAPFamily(family) {
		return ItemJungleScorchclaw, "Scorchclaw Pup"
	}
	if isTankFamily(family) || primary == tags.ClassTank {
		return ItemJungleMosstomper, "Mosstomper Seedling"
	}
	_ = allTags
	return ItemJungleGustwalker, "Gustwalker Hatchling"
}
