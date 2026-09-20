package engine

import "lol-build-overlay/internal/tags"

// Pressure aggregates weighted team signals for item priority.
type Pressure struct {
	Tank     float64
	AD       float64
	MR       float64
	HealUtil float64
	APBurst  float64
	Crit     float64
	CCHard   float64
	Dive     float64
}

// ComputePressure sums weight[i] * item/profile scores.
func ComputePressure(profiles []EnemyProfile, threats []ThreatRow, items *tags.ItemCatalog) Pressure {
	var p Pressure
	for i, ep := range profiles {
		w := 0.0
		heal := 0.0
		if i < len(threats) {
			w = threats[i].Weight
			heal = threats[i].HealUtility
		}
		p.Tank += w * items.ScoreTankiness(ep.Player.Items)
		adItems := items.ScoreADDamage(ep.Player.Items)
		p.AD += w * adItems
		// Early game: kit AD (Ashe/Irelia) before they buy damage items.
		if adItems < 0.15 && isKitADThreat(ep.Base) {
			p.AD += w * 0.4
		}
		p.MR += w * items.ScoreMR(ep.Player.Items)
		p.Crit += w * items.ScoreCrit(ep.Player.Items)
		p.HealUtil += heal

		if profileHas(ep.ProfileTags, tags.DamageAP) && profileHas(ep.ProfileTags, tags.StyleBurst) {
			p.APBurst += w
		} else if ep.Base.Has(tags.DamageAP) && ep.Base.Has(tags.StyleBurst) && !ep.Overridden {
			p.APBurst += w * 0.8
		}
		if profileHas(ep.ProfileTags, tags.ThreatCCHard) || ep.Base.Has(tags.ThreatCCHard) {
			p.CCHard += w
		}
		if profileHas(ep.ProfileTags, tags.ThreatDive) || ep.Base.Has(tags.ThreatDive) {
			p.Dive += w
		}

		legs := items.LegendaryCount(ep.Player.Items)
		if legs == 0 && (profileHas(ep.ProfileTags, tags.ClassTank) || profileHas(ep.ProfileTags, tags.ExtraJuggernaut)) {
			p.Tank += w * 0.55
		}
	}
	return p
}

func isKitADThreat(base tags.ChampionProfile) bool {
	if base.Has(tags.DamageAD) {
		return true
	}
	if base.Primary == tags.ClassMarksman {
		return true
	}
	if base.Primary == tags.ClassAssassin && !base.Has(tags.DamageAP) {
		return true
	}
	if base.Primary == tags.ClassFighter && !base.Has(tags.DamageAP) {
		return true
	}
	return false
}

// BlendScales: draft fades as enemies complete legendaries; live rises.
func BlendScales(enemyLegendaryTotal int) (draftScale, liveScale float64) {
	switch {
	case enemyLegendaryTotal <= 0:
		return 1.0, 0.35
	case enemyLegendaryTotal == 1:
		return 0.55, 0.75
	case enemyLegendaryTotal == 2:
		return 0.25, 1.0
	default:
		return 0.12, 1.0
	}
}
