package rules

import "lol-build-overlay/internal/tags"

// SyncStartToInventory replaces the seed start with the owned starter only when
// that starter is in the same family as the seed (Blade≠Ring≠Shield≠Atlas≠pet).
// Invalid fixture/inventory Dorans are ignored so the recommendation stays correct.
func SyncStartToInventory(build []Slot, owned []int) []Slot {
	ownedStart := 0
	for _, id := range owned {
		if id != 0 && IsLaneStartItem(id) {
			ownedStart = id
			break
		}
	}
	if ownedStart == 0 {
		return build
	}

	seedStart := 0
	for _, s := range build {
		if s.Role == "start" {
			seedStart = s.ItemID
			break
		}
	}
	if seedStart != 0 && !CompatibleStarts(seedStart, ownedStart) {
		return build
	}

	out := cloneSlots(build)
	for i := range out {
		if out[i].Role != "start" {
			continue
		}
		if out[i].ItemID == ownedStart {
			return out
		}
		out[i].ItemID = ownedStart
		out[i].Name = startItemName(ownedStart)
		out[i].Priority = 100
		return out
	}
	return append([]Slot{{ItemID: ownedStart, Name: startItemName(ownedStart), Priority: 100, Role: "start"}}, out...)
}

// CompatibleStarts is true when owned starter matches the seed family.
func CompatibleStarts(seedID, ownedID int) bool {
	if seedID == ownedID {
		return true
	}
	if IsJunglePet(seedID) && IsJunglePet(ownedID) {
		return true
	}
	if IsSupportQuestStart(seedID) && IsSupportQuestStart(ownedID) {
		return true
	}
	if startFamily(seedID) != "" && startFamily(seedID) == startFamily(ownedID) {
		return true
	}
	return false
}

func startFamily(id int) string {
	switch id {
	case ItemDoransBlade, 1083: // Cull
		return "ad"
	case ItemDoransRing, ItemDarkSeal:
		return "ap"
	case ItemDoransShield:
		return "tank"
	}
	if IsJunglePet(id) {
		return "jungle"
	}
	if IsSupportQuestStart(id) {
		return "support"
	}
	return ""
}

func startItemName(id int) string {
	switch id {
	case ItemDoransBlade:
		return "Doran's Blade"
	case ItemDoransRing:
		return "Doran's Ring"
	case ItemDoransShield:
		return "Doran's Shield"
	case ItemDarkSeal:
		return "Dark Seal"
	case ItemWorldAtlas, 3866, 3867:
		return "World Atlas"
	case ItemJungleScorchclaw:
		return "Scorchclaw Pup"
	case ItemJungleGustwalker:
		return "Gustwalker Hatchling"
	case ItemJungleMosstomper:
		return "Mosstomper Seedling"
	default:
		if IsJunglePet(id) {
			return "Jungle pet"
		}
		if IsSupportQuestStart(id) {
			return "World Atlas"
		}
		return "Starter"
	}
}

// AdjustLaneBoots prefers boots vs the same-lane opponent's damage type over
// diffuse team-wide draft signals (e.g. AD laner → Steelcaps even if team has AP).
func AdjustLaneBoots(build []Slot, lane tags.ChampionProfile, scale float64) ([]Slot, []string) {
	if scale <= 0.01 {
		return build, nil
	}
	if lane.Primary == "" && len(lane.Tags) == 0 {
		return build, nil
	}
	out := cloneSlots(build)
	var reasons []string

	apLane := lane.Has(tags.DamageAP) || lane.Primary == tags.ClassMage ||
		(lane.Primary == tags.ClassAssassin && lane.Has(tags.DamageAP)) ||
		(lane.Primary == tags.ClassSupport && lane.Has(tags.DamageAP))
	adLane := lane.Has(tags.DamageAD) || lane.Primary == tags.ClassMarksman ||
		lane.Primary == tags.ClassFighter ||
		(lane.Primary == tags.ClassAssassin && !lane.Has(tags.DamageAP)) ||
		(lane.Primary == tags.ClassTank && !lane.Has(tags.DamageAP))

	switch {
	case adLane && !apLane:
		if bump(&out, ItemSteelcaps, 22*scale) {
			reasons = append(reasons, "lane: AD opponent → Steelcaps ↑")
		}
		if bump(&out, ItemIonians, 10*scale) {
			reasons = append(reasons, "lane: AD opponent → Ionians ↑")
		}
		if bump(&out, ItemMercTreads, -16*scale) {
			reasons = append(reasons, "lane: AD opponent → Mercs ↓")
		}
	case apLane && !adLane:
		if bump(&out, ItemMercTreads, 20*scale) {
			reasons = append(reasons, "lane: AP opponent → Mercs ↑")
		}
		if bump(&out, ItemSteelcaps, -14*scale) {
			reasons = append(reasons, "lane: AP opponent → Steelcaps ↓")
		}
	case apLane && adLane:
		if bump(&out, ItemMercTreads, 8*scale) {
			reasons = append(reasons, "lane: hybrid threat → Mercs ↑")
		}
	}
	return out, reasons
}
