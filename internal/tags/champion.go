package tags

import "strings"

// ChampionProfile is the curated tag set for one champion.
type ChampionProfile struct {
	Primary string                  `json:"primary"`
	Tags    []string                `json:"tags"`
	ByRole  map[string]RoleOverride `json:"by_role,omitempty"`
}

// RoleOverride replaces base tags when Live Client position matches.
type RoleOverride struct {
	Primary string   `json:"primary"`
	Tags    []string `json:"tags"`
}

// Has reports whether tag is primary or in Tags.
func (p ChampionProfile) Has(tag string) bool {
	if p.Primary == tag {
		return true
	}
	for _, t := range p.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// All returns primary + tags without duplicates.
func (p ChampionProfile) All() []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, 1+len(p.Tags))
	add := func(t string) {
		t = strings.TrimSpace(t)
		if t == "" {
			return
		}
		if _, ok := seen[t]; ok {
			return
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	add(p.Primary)
	for _, t := range p.Tags {
		add(t)
	}
	return out
}

// Effective returns role override if present, else base (without by_role map).
func (p ChampionProfile) Effective(position string) ChampionProfile {
	pos := NormalizePosition(position)
	if pos != "" && len(p.ByRole) > 0 {
		if ov, ok := p.ByRole[pos]; ok {
			return ChampionProfile{Primary: ov.Primary, Tags: append([]string(nil), ov.Tags...)}
		}
	}
	return ChampionProfile{Primary: p.Primary, Tags: append([]string(nil), p.Tags...)}
}

// NormalizePosition maps Live Client position strings.
func NormalizePosition(pos string) string {
	switch strings.ToUpper(strings.TrimSpace(pos)) {
	case "TOP":
		return "TOP"
	case "MIDDLE", "MID":
		return "MIDDLE"
	case "JUNGLE", "JNG":
		return "JUNGLE"
	case "BOTTOM", "BOT", "ADC":
		return "BOTTOM"
	case "UTILITY", "SUPPORT", "SUP":
		return "UTILITY"
	default:
		return ""
	}
}

// Catalog maps champion id (e.g. "Ahri") -> profile.
type Catalog map[string]ChampionProfile

// Get returns raw stored profile or a minimal fallback.
func (c Catalog) Get(champID string) ChampionProfile {
	if p, ok := c[champID]; ok {
		return p
	}
	return ChampionProfile{Primary: ClassFighter, Tags: []string{DamageAD, RangeMelee}}
}

// Resolve applies by_role[position] when defined; empty position → base tags.
func (c Catalog) Resolve(champID, position string) ChampionProfile {
	return c.Get(champID).Effective(position)
}

// NaturalPositions returns SR roles where the champion is typically played.
// When by_role is set, those seats plus the class default apply; otherwise
// class heuristics fill in (without stretching burst mages into jungle).
func (p ChampionProfile) NaturalPositions() []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(pos string) {
		pos = NormalizePosition(pos)
		if pos == "" {
			return
		}
		if _, ok := seen[pos]; ok {
			return
		}
		seen[pos] = struct{}{}
		out = append(out, pos)
	}

	if len(p.ByRole) > 0 {
		for pos := range p.ByRole {
			add(pos)
		}
		add(classDefaultPosition(p.Primary))
		return out
	}

	switch p.Primary {
	case ClassMarksman:
		add("BOTTOM")
		add("MIDDLE")
	case ClassSupport:
		add("UTILITY")
	case ClassMage:
		add("MIDDLE")
		if p.Has(StylePoke) || p.Has(StyleHealer) || p.Has(StyleShield) {
			add("UTILITY")
		}
		// Only dive mages (e.g. Diana/Ekko-like), not poke burst (Vel'Koz).
		if p.Has(ThreatDive) {
			add("JUNGLE")
		}
	case ClassAssassin:
		add("MIDDLE")
		add("JUNGLE")
		if p.Has(DamageAD) && !p.Has(DamageAP) {
			add("TOP")
		}
	case ClassFighter:
		add("TOP")
		add("JUNGLE")
	case ClassTank:
		add("TOP")
		add("JUNGLE")
		if p.Has(StyleEngage) || p.Has(StylePeel) {
			add("UTILITY")
		}
	default:
		add("MIDDLE")
		add("TOP")
	}
	return out
}

func classDefaultPosition(primary string) string {
	switch primary {
	case ClassMarksman:
		return "BOTTOM"
	case ClassSupport:
		return "UTILITY"
	case ClassMage, ClassAssassin:
		return "MIDDLE"
	case ClassFighter, ClassTank:
		return "TOP"
	default:
		return "MIDDLE"
	}
}

// IsNaturalRole is true when position is among NaturalPositions (or position empty).
func (p ChampionProfile) IsNaturalRole(position string) bool {
	pos := NormalizePosition(position)
	if pos == "" {
		return true
	}
	for _, n := range p.NaturalPositions() {
		if n == pos {
			return true
		}
	}
	return false
}
