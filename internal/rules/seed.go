package rules

import "lol-build-overlay/internal/tags"

// PressureSignals are live team aggregates used to tweak priorities.
type PressureSignals struct {
	Tank     float64
	AD       float64
	MR       float64
	HealUtil float64
	APBurst  float64
	Crit     float64
	CCHard   float64
	Dive     float64
}

func isAPBuild(primary string) bool {
	switch primary {
	case tags.ClassMage, tags.ClassSupport:
		return true
	default:
		return false
	}
}

func isAPAssassin(primary string, allTags []string) bool {
	if primary != tags.ClassAssassin {
		return false
	}
	for _, t := range allTags {
		if t == tags.DamageAP || t == tags.DamageHybrid {
			return true
		}
	}
	return false
}

func usesAPItems(primary string, allTags []string) bool {
	return isAPBuild(primary) || isAPAssassin(primary, allTags) ||
		(primary == tags.ClassFighter && hasTag(allTags, tags.DamageAP))
}

func hasTag(all []string, want string) bool {
	for _, t := range all {
		if t == want {
			return true
		}
	}
	return false
}

// AdjustByPressure mutates slot priorities from live pressure, scaled by liveScale.
// playerPrimary/allTags gate AP-only vs AD-only adjustments so reasons match the seed.
func AdjustByPressure(seed []Slot, p PressureSignals, liveScale float64, playerPrimary string, allTags []string) ([]Slot, []string) {
	out := cloneSlots(seed)
	var reasons []string
	s := liveScale
	if s <= 0.01 {
		return out, reasons
	}

	ap := usesAPItems(playerPrimary, allTags)

	if ap {
		if p.Tank < 0.15 {
			if bump(&out, ItemLiandrys, -22*s) {
				reasons = append(reasons, "live: low tank_pressure → Liandry ↓")
			}
		} else if p.Tank > 0.35 {
			if bump(&out, ItemLiandrys, 18*s) {
				reasons = append(reasons, "live: high tank_pressure → Liandry ↑")
			}
		}
		if p.MR < 0.12 {
			if bump(&out, ItemVoidStaff, -16*s) {
				reasons = append(reasons, "live: low enemy MR → Void Staff ↓")
			}
		} else if p.MR > 0.3 {
			if bump(&out, ItemVoidStaff, 16*s) {
				reasons = append(reasons, "live: high enemy MR → Void Staff ↑")
			}
		}
		if p.AD > 0.3 || p.Dive > 0.25 {
			if bump(&out, ItemZhonyas, 20*s) {
				bump(&out, ItemShadowflame, -5*s)
				reasons = append(reasons, "live: ad/dive pressure → Zhonya ↑")
			}
		}
		if p.CCHard > 0.15 && p.APBurst > 0.1 {
			if bump(&out, ItemZhonyas, 14*s) {
				reasons = append(reasons, "live: ap+cc cage threat → Zhonya ↑")
			}
		}
		if p.APBurst > 0.25 {
			if bump(&out, ItemBansheeVeil, 12*s) {
				reasons = append(reasons, "live: ap burst → Banshee ↑")
			}
		}
		if p.Tank < 0.15 && p.AD > 0.2 {
			if bump(&out, ItemShadowflame, 12*s) {
				reasons = append(reasons, "live: squishy AD → Shadowflame ↑")
			}
		}
	} else {
		// AD / bruiser / ADC path
		if p.Tank > 0.2 {
			if bump(&out, ItemBlackCleaver, 16*s) {
				reasons = append(reasons, "live: tank pressure → Cleaver ↑")
			}
			if bump(&out, ItemBOTRK, 18*s) {
				reasons = append(reasons, "live: tank pressure → BotRK ↑")
			}
			if bump(&out, ItemLordDominiks, 12*s) {
				reasons = append(reasons, "live: tank pressure → LDR ↑")
			}
		} else if p.Tank < 0.15 {
			if bump(&out, ItemBOTRK, -22*s) {
				reasons = append(reasons, "live: low tank_pressure → BotRK ↓")
			}
			if bump(&out, ItemRapidFirecannon, 14*s) {
				reasons = append(reasons, "live: low tank_pressure → RFC ↑")
			}
			if bump(&out, ItemCollector, 12*s) {
				reasons = append(reasons, "live: low tank_pressure → Collector ↑")
			}
		}
		if p.AD > 0.25 || p.Dive > 0.2 {
			if bump(&out, ItemSteelcaps, 14*s) {
				reasons = append(reasons, "live: ad/dive pressure → Steelcaps ↑")
			}
		}
		if p.Crit > 0.25 {
			out = upsert(out, Slot{ItemID: ItemRanduins, Name: "Randuin's Omen", Priority: 52 + 20*p.Crit*s, Role: "defensive"})
			reasons = append(reasons, "live: crit pressure → Randuin option")
		}
		if p.APBurst > 0.2 {
			if bump(&out, ItemMercTreads, 16*s) {
				reasons = append(reasons, "live: ap burst → Mercs ↑")
			}
			if bump(&out, ItemForceOfNature, 18*s) {
				reasons = append(reasons, "live: ap burst → Force of Nature ↑")
			}
			if bump(&out, ItemSpiritVisage, 14*s) {
				reasons = append(reasons, "live: ap burst → Spirit Visage ↑")
			}
			if bump(&out, ItemSteelcaps, -12*s) {
				reasons = append(reasons, "live: ap burst → Steelcaps ↓")
			}
			if bump(&out, ItemRanduins, -10*s) {
				reasons = append(reasons, "live: ap burst → Randuin ↓")
			}
		}
	}

	if p.HealUtil > 0.12 {
		if ap {
			out = upsert(out, Slot{ItemID: ItemMorellonomicon, Name: "Morellonomicon", Priority: 58 + 20*p.HealUtil*s, Role: "utility"})
			reasons = append(reasons, "live: heal_utility → Morellonomicon")
		} else {
			out = upsert(out, Slot{ItemID: ItemExecutioners, Name: "Executioner's Calling", Priority: 58 + 20*p.HealUtil*s, Role: "utility"})
			reasons = append(reasons, "live: heal_utility → Executioner's")
		}
	}

	if p.CCHard > 0.25 {
		if bump(&out, ItemMercTreads, 10*s) {
			reasons = append(reasons, "live: cc_hard → Mercs ↑")
		}
	}

	return out, reasons
}

// AdjustAllyFront: with ally frontline, lean damage; without, lean defense.
func AdjustAllyFront(seed []Slot, allyFront float64, playerPrimary string, allTags []string) ([]Slot, []string) {
	out := cloneSlots(seed)
	var reasons []string
	ap := usesAPItems(playerPrimary, allTags)

	if allyFront >= 0.45 {
		if ap {
			n := 0
			if bump(&out, ItemShadowflame, 10) {
				n++
			}
			if bump(&out, ItemRabadons, 8) {
				n++
			}
			if bump(&out, ItemLiandrys, 4) {
				n++
			}
			bump(&out, ItemBansheeVeil, -6)
			bump(&out, ItemZhonyas, -4)
			if n > 0 {
				reasons = append(reasons, "ally: frontline present → damage ↑, defense ↓")
			}
		} else {
			n := 0
			if bump(&out, ItemBlackCleaver, 8) {
				n++
			}
			if bump(&out, ItemSunderedSky, 6) {
				n++
			}
			if bump(&out, ItemTrinity, 4) {
				n++
			}
			if n > 0 {
				reasons = append(reasons, "ally: frontline present → AD damage ↑")
			}
		}
	} else if allyFront < 0.2 {
		if ap {
			n := 0
			if bump(&out, ItemZhonyas, 8) {
				n++
			}
			if bump(&out, ItemBansheeVeil, 6) {
				n++
			}
			bump(&out, ItemShadowflame, -4)
			if n > 0 {
				reasons = append(reasons, "ally: little frontline → defense ↑")
			}
		} else {
			n := 0
			if bump(&out, ItemSteelcaps, 6) {
				n++
			}
			if bump(&out, ItemMercTreads, 4) {
				n++
			}
			if n > 0 {
				reasons = append(reasons, "ally: little frontline → armor/MR boots ↑")
			}
		}
	}
	return out, reasons
}

// CollapseBoots keeps a single boots slot (highest priority).
func CollapseBoots(build []Slot) []Slot {
	bestPri := -1e9
	found := false
	for _, s := range build {
		if s.Role == "boots" && (!found || s.Priority > bestPri) {
			bestPri = s.Priority
			found = true
		}
	}
	if !found {
		return build
	}
	out := make([]Slot, 0, len(build))
	kept := false
	for _, s := range build {
		if s.Role != "boots" {
			out = append(out, s)
			continue
		}
		if !kept && s.Priority == bestPri {
			out = append(out, s)
			kept = true
		}
	}
	return out
}

// ProtectCorePriority keeps unfinished core/component above defensive/utility bumps
// so Zhonya/Morello never outrank Rod of Ages / Stormsurge / etc. early.
func ProtectCorePriority(build []Slot) []Slot {
	maxCore := -1e9
	hasCore := false
	startPri := -1e9
	for _, s := range build {
		if s.Role == "core" || s.Role == "component" {
			if s.Priority > maxCore {
				maxCore = s.Priority
				hasCore = true
			}
		}
		if s.Role == "start" && s.Priority > startPri {
			startPri = s.Priority
		}
	}
	out := cloneSlots(build)
	if hasCore {
		ceiling := maxCore - 1
		for i := range out {
			switch out[i].Role {
			case "defensive", "utility":
				if out[i].Priority > ceiling {
					out[i].Priority = ceiling
				}
			}
		}
	}
	// Keep boots under the start slot so Doran's/Atlas stay first in the list.
	if startPri > 0 {
		bootCap := startPri - 1
		for i := range out {
			if out[i].Role == "boots" && out[i].Priority >= startPri {
				out[i].Priority = bootCap
			}
		}
	}
	return out
}
