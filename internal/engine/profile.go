package engine

import (
	"strings"

	"lol-build-overlay/internal/data"
	"lol-build-overlay/internal/tags"
)

// PlayerSnapshot is one participant from Live Client (or fixture).
type PlayerSnapshot struct {
	ChampionID string
	Team       string // "ORDER" / "CHAOS"
	Position   string // TOP MIDDLE JUNGLE BOTTOM UTILITY (may be empty)
	Level      int
	Kills      int
	Deaths     int
	Assists    int
	Items      []int
	IsDead     bool
	Summoner   string
	// SpellOne/SpellTwo are Live Client keys, e.g. "SummonerFlash", "SummonerSmite".
	SpellOne string
	SpellTwo string
}

// HasSmite reports jungle Smite on either summoner spell slot.
func (p PlayerSnapshot) HasSmite() bool {
	return isSmiteSpell(p.SpellOne) || isSmiteSpell(p.SpellTwo)
}

func isSmiteSpell(key string) bool {
	k := strings.ToLower(strings.TrimSpace(key))
	return strings.Contains(k, "smite") || strings.Contains(k, "summonersmite")
}

// RemapChampionIDs resolves localized / raw Live Client names to Data Dragon ids.
func RemapChampionIDs(store *data.Store, snap GameSnapshot) GameSnapshot {
	if store == nil {
		return snap
	}
	out := snap
	out.Players = make([]PlayerSnapshot, len(snap.Players))
	copy(out.Players, snap.Players)
	for i := range out.Players {
		out.Players[i].ChampionID = store.ResolveChampionID(out.Players[i].ChampionID)
	}
	out.ActiveChampionID = store.ResolveChampionID(snap.ActiveChampionID)
	return out
}

// GameSnapshot is a normalized match state.
type GameSnapshot struct {
	ActiveChampionID string
	ActiveTeam       string
	CurrentGold      int // -1 if unknown
	Players          []PlayerSnapshot
}

// Enemies returns opposing team players.
func (g GameSnapshot) Enemies() []PlayerSnapshot {
	var out []PlayerSnapshot
	for _, p := range g.Players {
		if p.Team != "" && p.Team != g.ActiveTeam {
			out = append(out, p)
		}
	}
	return out
}

// Allies returns same-team players including self.
func (g GameSnapshot) Allies() []PlayerSnapshot {
	var out []PlayerSnapshot
	for _, p := range g.Players {
		if p.Team == g.ActiveTeam {
			out = append(out, p)
		}
	}
	return out
}

// ActivePlayer finds the local player.
func (g GameSnapshot) ActivePlayer() (PlayerSnapshot, bool) {
	for _, p := range g.Players {
		if p.ChampionID == g.ActiveChampionID && p.Team == g.ActiveTeam {
			return p, true
		}
	}
	for _, p := range g.Players {
		if p.ChampionID == g.ActiveChampionID {
			return p, true
		}
	}
	// Last resort: unique ORDER/CHAOS human often first in list matching team.
	for _, p := range g.Players {
		if p.Team == g.ActiveTeam {
			return p, true
		}
	}
	return PlayerSnapshot{}, false
}

// EnemyProfile is merged base∪live (or live override) for one enemy.
type EnemyProfile struct {
	Player      PlayerSnapshot
	Base        tags.ChampionProfile
	LiveTags    []string
	ProfileTags []string
	Overridden  bool
	Kit         float64
}

// BuildProfiles merges role-resolved champion tags with live item tags.
func BuildProfiles(enemies []PlayerSnapshot, champTags tags.Catalog, items *tags.ItemCatalog) []EnemyProfile {
	out := make([]EnemyProfile, 0, len(enemies))
	for _, e := range enemies {
		base := champTags.Resolve(e.ChampionID, e.Position)
		live := items.TagsFor(e.Items)
		prof, overridden := mergeProfile(base, live, items, e.Items)
		out = append(out, EnemyProfile{
			Player:      e,
			Base:        base,
			LiveTags:    live,
			ProfileTags: prof,
			Overridden:  overridden,
			Kit:         kitScore(base, prof),
		})
	}
	return out
}

// Off-meta threshold: ≥2 legendary lethality/AD damage → strip tank/juggernaut.
func mergeProfile(base tags.ChampionProfile, live []string, items *tags.ItemCatalog, itemIDs []int) (profile []string, overridden bool) {
	lethADLegs := items.CountLegendaryWith(itemIDs, tags.ItemLethality, tags.ItemADDamage)
	tankish := base.Has(tags.ClassTank) || base.Has(tags.ExtraJuggernaut) || base.Primary == tags.ClassTank

	if tankish && lethADLegs >= 2 {
		seen := map[string]struct{}{}
		add := func(t string) {
			if t == "" {
				return
			}
			if _, ok := seen[t]; ok {
				return
			}
			seen[t] = struct{}{}
			profile = append(profile, t)
		}
		add(tags.ClassAssassin)
		add(tags.DamageAD)
		add(tags.StyleBurst)
		for _, t := range live {
			switch t {
			case tags.ItemLethality, tags.ItemADDamage, tags.ItemCrit:
				add(tags.DamageAD)
				add(tags.StyleBurst)
			case tags.ItemAPDamage:
				add(tags.DamageAP)
			case tags.ItemOnHit, tags.ItemAS:
				add(tags.StyleDPS)
			}
		}
		return profile, true
	}

	seen := map[string]struct{}{}
	add := func(t string) {
		if t == "" {
			return
		}
		if _, ok := seen[t]; ok {
			return
		}
		seen[t] = struct{}{}
		profile = append(profile, t)
	}
	for _, t := range base.All() {
		add(t)
	}
	for _, t := range live {
		switch t {
		case tags.ItemLethality:
			add(tags.DamageAD)
			add(tags.StyleBurst)
		case tags.ItemADDamage, tags.ItemCrit:
			add(tags.DamageAD)
		case tags.ItemAPDamage:
			add(tags.DamageAP)
		case tags.ItemTankHP, tags.ItemArmor, tags.ItemMR:
			add(tags.ClassTank)
			add(tags.ThreatSustain)
		case tags.ItemShield:
			add(tags.StyleShield)
		case tags.ItemOnHit, tags.ItemAS:
			add(tags.StyleDPS)
		}
	}
	return profile, false
}

func kitScore(base tags.ChampionProfile, profile []string) float64 {
	has := func(t string) bool {
		if base.Primary == t || base.Has(t) {
			return true
		}
		for _, p := range profile {
			if p == t {
				return true
			}
		}
		return false
	}

	switch {
	case has(tags.ClassAssassin) || has(tags.ClassMarksman):
		return 1.15
	case has(tags.ClassMage) && has(tags.StyleBurst):
		return 1.15
	case has(tags.ClassMage):
		return 1.05
	case has(tags.ExtraJuggernaut) || (has(tags.ClassFighter) && !has(tags.ClassTank)):
		return 1.0
	case has(tags.ClassTank):
		return 0.55
	case has(tags.ClassSupport) && (has(tags.StyleHealer) || has(tags.StyleShield)):
		return 0.6
	case has(tags.ClassSupport):
		return 0.65
	default:
		return 1.0
	}
}

func profileHas(profile []string, tag string) bool {
	for _, t := range profile {
		if t == tag {
			return true
		}
	}
	return false
}

// AllyFrontScore 0..1 — how much frontline the ally team provides.
func AllyFrontScore(allies []PlayerSnapshot, selfID string, champTags tags.Catalog) float64 {
	if len(allies) <= 1 {
		return 0
	}
	score := 0.0
	count := 0
	for _, a := range allies {
		if a.ChampionID == selfID {
			continue
		}
		count++
		p := champTags.Resolve(a.ChampionID, a.Position)
		switch {
		case p.Primary == tags.ClassTank || p.Has(tags.ClassTank):
			score += 1.0
		case p.Has(tags.ExtraJuggernaut):
			score += 0.85
		case p.Has(tags.StyleEngage) && (p.Primary == tags.ClassFighter || p.Has(tags.ThreatDive)):
			score += 0.55
		case p.Has(tags.StylePeel) && p.Primary == tags.ClassSupport:
			score += 0.35
		}
	}
	if count == 0 {
		return 0
	}
	return clamp01(score / float64(count) * 1.2)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
