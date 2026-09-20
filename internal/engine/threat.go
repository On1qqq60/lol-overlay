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
	ChampionID    string
	Form          float64
	Econ          float64
	Kit           float64
	DamageThreat  float64
	Weight        float64
	Ignored       bool
	MainTarget    bool
	HealUtility   float64 // separate utility threat (enchanters)
	LaneOpponent  bool
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

		healUtil := 0.0
		// Grievous tracks healers only — shield supports (Morgana) are not heal pressure.
		if profileHas(p.ProfileTags, tags.StyleHealer) || p.Base.Has(tags.StyleHealer) {
			healNow := 0.4*form + 0.6*econ
			if form < 0.05 {
				healNow *= form * 20
			}
			healUtil = kit * healNow * levelFactor
			if p.Player.IsDead {
				healUtil = 0
			}
		}

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
