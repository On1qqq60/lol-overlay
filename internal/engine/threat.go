package engine

import (
	"math"

	"lol-build-overlay/internal/tags"
)

const (
	weightIgnore = 0.08
	weightMain   = 0.25
	laneThreatMult = 2.15
)

// ThreatRow is per-enemy threat math.
type ThreatRow struct {
	ChampionID    string  `json:"championId"`
	Form          float64 `json:"form"`
	Econ          float64 `json:"econ"`
	Kit           float64 `json:"kit"`
	DamageThreat  float64 `json:"damageThreat"`
	Weight        float64 `json:"weight"`
	Ignored       bool    `json:"ignored"`
	MainTarget    bool    `json:"mainTarget"`
	HealUtility   float64 `json:"healUtility"` // healers, sustain kits, vamp items — not shields
	LaneOpponent  bool    `json:"laneOpponent"`
}

// ComputeThreat fills form/econ/kit/weights for enemies.
// activePos boosts the same-position lane opponent into top threats.
func ComputeThreat(profiles []EnemyProfile, itemsLegendaryCount func([]int) int, maxLevel int, activePos string) []ThreatRow {
	if maxLevel < 1 {
		maxLevel = 18
	}
	activePos = tags.NormalizePosition(activePos)
	rows := make([]ThreatRow, len(profiles))
	sum := 0.0

	for i, p := range profiles {
		form := formScore(p.Player.Kills, p.Player.Deaths, p.Player.Assists)
		econ := econScore(itemsLegendaryCount(p.Player.Items))
		kit := p.Kit
		levelFactor := 0.8 + 0.2*(float64(p.Player.Level)/float64(maxLevel))
		// If form ≈ 0 (hard feeder), kit/econ barely matter — matches "0/20 → ignore".
		strengthNow := 0.35*form + 0.65*econ
		if form < 0.1 {
			strengthNow *= form * 10 // form=0 → 0; form=0.1 → unchanged
		}
		lane := false
		laneMult := 1.0
		ep := tags.NormalizePosition(p.Player.Position)
		if activePos != "" && ep != "" && activePos == ep {
			lane = true
			laneMult = laneThreatMult
		}
		threat := kit * strengthNow * levelFactor * laneMult
		if p.Player.IsDead {
			threat = 0
		}

		healUtil := healUtilityFor(p, form, econ, kit, levelFactor)

		rows[i] = ThreatRow{
			ChampionID:   p.Player.ChampionID,
			Form:         form,
			Econ:         econ,
			Kit:          kit,
			DamageThreat: threat,
			HealUtility:  healUtil,
			LaneOpponent: lane,
		}
		sum += threat
	}

	if sum <= 0 {
		// Draft / zero form+econ: weight by kit × lane so lane opponent ranks high.
		wsum := 0.0
		for i, p := range profiles {
			w := rows[i].Kit
			if w <= 0 {
				w = 1
			}
			ep := tags.NormalizePosition(p.Player.Position)
			if activePos != "" && ep != "" && activePos == ep {
				w *= laneThreatMult
				rows[i].LaneOpponent = true
			}
			rows[i].DamageThreat = w
			wsum += w
		}
		if wsum <= 0 {
			wsum = float64(len(rows))
			for i := range rows {
				rows[i].Weight = 1.0 / wsum
			}
		} else {
			for i := range rows {
				rows[i].Weight = rows[i].DamageThreat / wsum
			}
		}
		for i := range rows {
			rows[i].Ignored = rows[i].Weight < weightIgnore
			rows[i].MainTarget = rows[i].Weight > weightMain
		}
		return rows
	}

	for i := range rows {
		w := rows[i].DamageThreat / sum
		rows[i].Weight = w
		rows[i].Ignored = w < weightIgnore
		rows[i].MainTarget = w > weightMain
	}
	return rows
}

// healUtilityFor is grievous pressure: dedicated healers, kit sustain (Sion),
// and vamp/lifesteal items (BotRK / BT). Shield-only kits (Morgana) stay at 0.
func healUtilityFor(p EnemyProfile, form, econ, kit, levelFactor float64) float64 {
	kitMul := kitHealMul(p)
	vamp := vampItemScore(p.Player.Items)
	if kitMul <= 0 && vamp <= 0 {
		return 0
	}
	// Presence floor: even 0/0 Sion with Heartsteel still heals on W.
	// Damage threat still zeros feeders; heal does not.
	healNow := 0.35*form + 0.45*econ + 0.20
	heal := (kitMul*kit + vamp) * healNow * levelFactor
	if p.Player.IsDead {
		return 0
	}
	return heal
}

func kitHealMul(p EnemyProfile) float64 {
	if profileHas(p.ProfileTags, tags.StyleHealer) || p.Base.Has(tags.StyleHealer) {
		return 1
	}
	// Kit sustain only — HP items also stamp ThreatSustain on the live profile
	// (Liandry Morgana) and must not count as grievous pressure.
	if p.Base.Has(tags.ThreatSustain) {
		return 0.55
	}
	return 0
}

func vampItemScore(items []int) float64 {
	s := 0.0
	for _, id := range items {
		switch id {
		case 3153, 3072: // BotRK, Bloodthirster
			s += 0.55
		case 3074, 4633: // Ravenous Hydra, Riftmaker
			s += 0.40
		case 6610: // Sundered Sky
			s += 0.35
		case 6333, 3083: // Death's Dance, Warmogs
			s += 0.25
		case 3065: // Spirit Visage
			s += 0.20
		case 3144, 1053: // Cutlass, Vampiric Scepter
			s += 0.15
		}
	}
	return clamp(s, 0, 1)
}

func formScore(kills, deaths, assists int) float64 {
	kda := (float64(kills) + float64(assists)*0.6) / math.Max(1, float64(deaths))
	return clamp(kda/3, 0, 1)
}

func econScore(legendary int) float64 {
	return clamp(float64(legendary)/4, 0, 1)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MaxLevel among all players in snapshot.
func MaxLevel(players []PlayerSnapshot) int {
	m := 1
	for _, p := range players {
		if p.Level > m {
			m = p.Level
		}
	}
	return m
}
