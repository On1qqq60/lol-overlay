package rules

// CommitFamily switches the identity family to the branch the player already bought.
// Skip-the-first-core is a valid game; the overlay must continue that tree.
func CommitFamily(family, champ string, owned []int) string {
	if len(owned) == 0 {
		return family
	}
	own := ownedSet(owned)
	has := func(ids ...int) bool {
		for _, id := range ids {
			if _, ok := own[id]; ok {
				return true
			}
		}
		return false
	}

	if has(ItemHeartsteel) {
		switch champ {
		case "Chogath":
			return FamTankCho
		case "DrMundo":
			return FamTankMundo
		default:
			return FamTankSunfire
		}
	}
	if has(ItemSunfire, ItemHollowRadiance) {
		if champ == "Shen" {
			return FamTankShen
		}
		return FamTankSunfire
	}

	if has(ItemInfinityEdge, ItemYunTal) && !has(ItemNashors, ItemRiftmaker, ItemTrinity, ItemSunderedSky) {
		switch champ {
		case "Yasuo", "Yone":
			return FamFighterCrit
		case "Tryndamere":
			return FamFighterTrynd
		case "Gangplank":
			return FamFighterGP
		case "Kayle":
			return FamADCCrit
		default:
			if isADCFamily(family) {
				return FamADCCrit
			}
		}
	}
	if has(ItemKraken) && !has(ItemNashors, ItemRiftmaker, ItemTrinity, ItemSunderedSky, ItemInfinityEdge, ItemYunTal) {
		switch champ {
		case "Graves", "Corki":
			return FamADCCrit
		case "Kaisa":
			// Kraken is Kaisa's first core; stay Nashor/Guinsoo.
		case "Ezreal", "Zeri", "Vayne", "Kalista", "KogMaw", "Varus":
			// Generator (and many live games) put Kraken on every ADC.
			// Do not punch Manamune/Stormrazor/Guinsoo through a Kraken bag.
			if !has(ItemTear, ItemManamune, ItemBOTRK, ItemGuinsoos, ItemStormrazor) {
				return FamADCCrit
			}
		}
	}
	if has(ItemEssenceReaver) && champ == "Gangplank" {
		return FamFighterGP
	}
	if has(ItemNashors) && !has(ItemTrinity, ItemKraken) {
		switch champ {
		case "Gwen":
			return FamFighterGwen
		case "Kayle":
			return FamADCKayle
		case "Volibear":
			return FamFighterVoli
		case "Teemo":
			return FamMageNashor
		}
	}
	if has(ItemYoumuu, ItemOpportunity) && !has(ItemTrinity, ItemEclipse, ItemSunderedSky) {
		switch champ {
		case "Jayce":
			return FamAssassinAD
		case "Kayn":
			return FamAssassinAD
		case "Akshan":
			return FamAssassinAD
		}
	}
	if has(ItemStormsurge) && !has(ItemLichBane, ItemNashors, ItemRodOfAges) {
		if champ == "Shaco" || isAPFamily(family) {
			return FamAssassinAP
		}
	}
	if has(ItemBlackfire) {
		switch champ {
		case "Cassiopeia", "Teemo":
			return FamMagePoke
		case "Singed":
			return FamMageSinged
		default:
			if isSupportFamily(family) {
				return FamSupportMage
			}
			if isAPFamily(family) {
				return FamMagePoke
			}
		}
	}
	if has(ItemMoonstone, ItemRedemption, ItemMikaels, ItemArdent, ItemStaffOfFlowing, ItemDawncore) &&
		!has(ItemLocket, ItemBlackfire, ItemLiandrys) {
		if isSupportFamily(family) || champ == "Ivern" {
			return FamSupportEnch
		}
	}
	if has(ItemLocket, ItemKnightsVow) && !has(ItemMoonstone, ItemBlackfire, ItemYoumuu, ItemEclipse) {
		if isSupportFamily(family) {
			return FamSupportEngage
		}
	}
	if has(ItemLudens) && isAPFamily(family) {
		return FamMageLuden
	}
	if has(ItemMalignance) && (isAPFamily(family) || champ == "Kennen" || champ == "Vladimir") {
		return FamMageBurst
	}
	if has(ItemRodOfAges) {
		switch champ {
		case "Gwen", "Mordekaiser", "Vladimir", "Singed", "Teemo", "Kayle", "Lillia":
			// keep champion identity
		default:
			return FamMageRoA
		}
	}

	if has(ItemStridebreaker) {
		if champ == "Olaf" {
			return FamFighterOlaf
		}
		return FamFighterStride
	}
	if has(ItemSheen) && !has(ItemTrinity, ItemIceborn, ItemLichBane, ItemEssenceReaver) {
		if isADCFamily(family) || isAPFamily(family) || isSupportFamily(family) {
			return family
		}
		switch champ {
		case "Nasus":
			return FamTankNasus
		case "Illaoi":
			return FamFighterIllaoi
		case "KSante", "Poppy":
			return FamFighterIceborn
		default:
			return FamFighterTrinity
		}
	}
	if has(ItemIceborn) && !has(ItemTrinity) {
		if champ == "Nasus" {
			return FamTankNasus
		}
		if champ == "Illaoi" {
			return FamFighterIllaoi
		}
		return FamFighterIceborn
	}
	if has(ItemEclipse) && !has(ItemTrinity, ItemSunderedSky) {
		if isADCFamily(family) {
			return family
		}
		return FamFighterEclipse
	}
	if has(ItemTrinity) || (has(ItemSunderedSky) && !has(ItemEclipse)) {
		switch champ {
		case "Kayle", "Gwen", "Cassiopeia", "Teemo", "Vladimir", "Mordekaiser", "Singed",
			"MasterYi", "Belveth", "Warwick":
			return family
		case "Jayce":
			return FamFighterJayce
		case "Yorick":
			return FamFighterYorick
		default:
			if isADCFamily(family) {
				return family
			}
			return FamFighterTrinity
		}
	}
	return family
}

// DropSkippedFirstCores removes unowned identity cores once the player has
// already committed to a different tree (Trinity vs Eclipse, Sunfire vs Heartsteel).
func DropSkippedFirstCores(seed []Slot, owned []int) []Slot {
	own := ownedSet(owned)
	has := func(id int) bool {
		_, ok := own[id]
		return ok
	}
	first := []int{
		ItemTrinity, ItemEclipse, ItemIceborn, ItemYoumuu, ItemStridebreaker,
		ItemHeartsteel, ItemSunfire, ItemNashors, ItemKraken, ItemYunTal,
		ItemBlackfire, ItemRodOfAges, ItemMalignance, ItemLudens, ItemStormsurge,
		ItemEssenceReaver, ItemBOTRK, ItemRavenousHydra,
		ItemLocket, ItemMoonstone, ItemMandate,
	}
	ownedFirst := false
	for _, id := range first {
		if has(id) {
			ownedFirst = true
			break
		}
	}
	if !ownedFirst && !has(ItemSunderedSky) {
		return seed
	}
	drop := map[int]struct{}{}
	for _, id := range first {
		if !has(id) {
			drop[id] = struct{}{}
		}
	}
	out := make([]Slot, 0, len(seed))
	for _, s := range seed {
		if _, skip := drop[s.ItemID]; skip {
			continue
		}
		out = append(out, s)
	}
	if _, ok := own[ItemCaulfield]; ok {
		if _, hasCleaver := own[ItemBlackCleaver]; !hasCleaver {
			closesER := false
			hasCleaverSlot := false
			for i := range out {
				switch out[i].ItemID {
				case ItemEssenceReaver, ItemTrinity, ItemSunderedSky:
					closesER = true
				case ItemBlackCleaver:
					hasCleaverSlot = true
				}
			}
			if !hasCleaverSlot && !closesER {
				out = append(out, sl(ItemBlackCleaver, "Black Cleaver", 96, "offensive"))
			}
		}
	}
	if _, ok := own[ItemHauntingGuise]; ok {
		if _, has := own[ItemLiandrys]; !has {
			if _, hasRift := own[ItemRiftmaker]; !hasRift {
				hasSlot := false
				for i := range out {
					if out[i].ItemID == ItemLiandrys || out[i].ItemID == ItemRiftmaker {
						hasSlot = true
						break
					}
				}
				if !hasSlot {
					out = append(out, sl(ItemLiandrys, "Liandry's Torment", 97, "offensive"))
				}
			}
		}
	}
	if _, ok := own[1038]; ok {
		if _, hasIE := own[ItemInfinityEdge]; !hasIE {
			hasSlot := false
			for i := range out {
				if out[i].ItemID == ItemInfinityEdge {
					hasSlot = true
					break
				}
			}
			if !hasSlot {
				out = append(out, sl(ItemInfinityEdge, "Infinity Edge", 96, "offensive"))
			}
		}
	}
	if _, ok := own[ItemBlastingWand]; ok {
		closes := false
		for _, id := range []int{ItemShadowflame, ItemRylais, ItemVoidStaff, ItemHorizon} {
			if _, has := own[id]; has {
				closes = true
				break
			}
		}
		if !closes {
			hasSlot := false
			for i := range out {
				switch out[i].ItemID {
				case ItemShadowflame, ItemRylais, ItemVoidStaff, ItemHorizon:
					hasSlot = true
				}
			}
			if !hasSlot {
				out = append(out, sl(ItemShadowflame, "Shadowflame", 97, "offensive"))
			}
		}
	}
	return out
}

// DropOrDemoteTearLine removes Tear/Seraph from champs that do not scale on it.
// Tear-scalers keep the line in the plan but will not next Seraph without Tear.
func DropOrDemoteTearLine(seed []Slot, owned []int, champ, pos string) []Slot {
	own := ownedSet(owned)
	hasTear := false
	for _, id := range []int{ItemTear, ItemArchangel, ItemSeraphs, ItemManamune} {
		if _, ok := own[id]; ok {
			hasTear = true
			break
		}
	}
	if hasTear {
		return seed
	}
	if wantsTearLine(champ, pos) {
		return DemoteTearUpgrades(seed, owned)
	}
	out := make([]Slot, 0, len(seed))
	for _, s := range seed {
		switch s.ItemID {
		case ItemTear, ItemArchangel, ItemSeraphs:
			continue
		}
		out = append(out, s)
	}
	return out
}

func wantsTearLine(champ, pos string) bool {
	switch champ {
	case "Cassiopeia", "Ryze", "Anivia", "Kassadin":
		return true
	default:
		return false
	}
}

// DemoteTearUpgrades keeps Tear itself but will not make Seraph/Archangel the
// next buy unless the player already started the tear line.
func DemoteTearUpgrades(seed []Slot, owned []int) []Slot {
	own := ownedSet(owned)
	hasTear := false
	for _, id := range []int{ItemTear, ItemArchangel, ItemSeraphs, ItemManamune} {
		if _, ok := own[id]; ok {
			hasTear = true
			break
		}
	}
	if hasTear {
		return seed
	}
	out := cloneSlots(seed)
	for i := range out {
		switch out[i].ItemID {
		case ItemArchangel, ItemSeraphs:
			out[i].Priority -= 80
		}
	}
	return out
}

// BoostCrafted raises legendaries whose components are already in inventory,
// so Needlessly → Deathcap wins over an unowned Cosmic Drive.
func BoostCrafted(seed []Slot, owned []int) []Slot {
	if len(owned) == 0 {
		return seed
	}
	own := ownedSet(owned)
	out := cloneSlots(seed)
	for i := range out {
		for _, c := range ComponentsToward[out[i].ItemID] {
			if _, ok := own[c]; ok {
				out[i].Priority += 70
				break
			}
		}
	}
	return out
}
