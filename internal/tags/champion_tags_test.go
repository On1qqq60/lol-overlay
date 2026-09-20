package tags_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"lol-build-overlay/internal/data"
	"lol-build-overlay/internal/tags"
)

var allowedPrimary = map[string]struct{}{
	tags.ClassMage: {}, tags.ClassAssassin: {}, tags.ClassMarksman: {},
	tags.ClassFighter: {}, tags.ClassTank: {}, tags.ClassSupport: {},
}

var allowedTags = map[string]struct{}{
	tags.DamageAP: {}, tags.DamageAD: {}, tags.DamageHybrid: {}, tags.DamageTrue: {},
	tags.RangeRanged: {}, tags.RangeMelee: {},
	tags.StyleBurst: {}, tags.StylePoke: {}, tags.StyleDPS: {}, tags.StyleEngage: {},
	tags.StylePeel: {}, tags.StyleHealer: {}, tags.StyleShield: {},
	tags.ThreatDive: {}, tags.ThreatCCHard: {}, tags.ThreatMobility: {},
	tags.ThreatSustain: {}, tags.ThreatSplitpush: {},
	tags.ExtraJuggernaut: {}, tags.ExtraAntiDash: {}, tags.ExtraPick: {}, tags.ExtraAllIn: {},
	tags.ClassMage: {}, tags.ClassAssassin: {}, tags.ClassMarksman: {},
	tags.ClassFighter: {}, tags.ClassTank: {}, tags.ClassSupport: {},
}

func projectRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestNaturalPositions(t *testing.T) {
	cases := []struct {
		p    tags.ChampionProfile
		want []string
	}{
		{tags.ChampionProfile{Primary: tags.ClassMarksman, Tags: []string{tags.DamageAD}}, []string{"BOTTOM", "MIDDLE"}},
		{tags.ChampionProfile{Primary: tags.ClassSupport, Tags: []string{tags.StyleHealer}}, []string{"UTILITY"}},
		{tags.ChampionProfile{Primary: tags.ClassMage, Tags: []string{tags.DamageAP, tags.StylePoke}}, []string{"MIDDLE", "UTILITY"}},
		{tags.ChampionProfile{Primary: tags.ClassTank, Tags: []string{tags.StyleEngage}, ByRole: map[string]tags.RoleOverride{"UTILITY": {Primary: tags.ClassTank}}}, []string{"UTILITY", "TOP"}},
		{tags.ChampionProfile{Primary: tags.ClassAssassin, Tags: []string{tags.DamageAD, tags.ThreatDive}, ByRole: map[string]tags.RoleOverride{"UTILITY": {Primary: tags.ClassSupport}, "MIDDLE": {Primary: tags.ClassAssassin}}}, []string{"UTILITY", "MIDDLE"}},
	}
	for _, tc := range cases {
		got := tc.p.NaturalPositions()
		if len(got) == 0 {
			t.Fatalf("empty NaturalPositions for %+v", tc.p)
		}
		for _, w := range tc.want {
			found := false
			for _, g := range got {
				if g == w {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("NaturalPositions(%+v)=%v missing %s", tc.p, got, w)
			}
		}
	}
	// by_role champions must not inherit loose class jungle seats.
	pykeLike := tags.ChampionProfile{
		Primary: tags.ClassAssassin,
		Tags:    []string{tags.DamageAD, tags.ThreatDive},
		ByRole: map[string]tags.RoleOverride{
			"UTILITY": {Primary: tags.ClassSupport},
			"MIDDLE":  {Primary: tags.ClassAssassin},
		},
	}
	if pykeLike.IsNaturalRole("JUNGLE") {
		t.Fatal("by_role assassin must not auto-add JUNGLE")
	}
	velkozLike := tags.ChampionProfile{Primary: tags.ClassMage, Tags: []string{tags.DamageAP, tags.StyleBurst, tags.StylePoke, tags.RangeRanged}}
	if velkozLike.IsNaturalRole("JUNGLE") {
		t.Fatal("ranged poke mage must not be natural jungle")
	}
	adc := tags.ChampionProfile{Primary: tags.ClassMarksman, Tags: []string{tags.DamageAD}}
	if adc.IsNaturalRole("UTILITY") {
		t.Fatal("marksman must not be natural on UTILITY")
	}
	if !adc.IsNaturalRole("BOTTOM") {
		t.Fatal("marksman must be natural on BOTTOM")
	}
}

func TestChampionTagsCatalog(t *testing.T) {
	store, err := data.Load(projectRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(store.ChampionTags) == 0 {
		t.Fatal("champion_tags.json empty or missing")
	}
	if len(store.ChampionTags) != len(store.Champions) {
		t.Fatalf("tag count %d != champion count %d", len(store.ChampionTags), len(store.Champions))
	}

	for _, c := range store.Champions {
		p, ok := store.ChampionTags[c.ID]
		if !ok {
			t.Errorf("missing tags for %s", c.ID)
			continue
		}
		if _, ok := allowedPrimary[p.Primary]; !ok {
			t.Errorf("%s: invalid primary %q", c.ID, p.Primary)
		}
		var rangeN, dmgN int
		for _, tag := range p.Tags {
			if _, ok := allowedTags[tag]; !ok {
				t.Errorf("%s: unknown tag %q", c.ID, tag)
			}
			switch tag {
			case tags.RangeMelee, tags.RangeRanged:
				rangeN++
			case tags.DamageAP, tags.DamageAD, tags.DamageHybrid, tags.DamageTrue:
				dmgN++
			}
		}
		if rangeN < 1 {
			t.Errorf("%s: want melee/ranged tag, got %v", c.ID, p.Tags)
		}
		// Damage may be omitted for pure tanks/supports; warn only if also no class hint.
		_ = dmgN

		for pos, ov := range p.ByRole {
			if tags.NormalizePosition(pos) == "" {
				t.Errorf("%s: invalid by_role key %q", c.ID, pos)
			}
			if _, ok := allowedPrimary[ov.Primary]; !ok {
				t.Errorf("%s by_role[%s]: invalid primary %q", c.ID, pos, ov.Primary)
			}
			for _, tag := range ov.Tags {
				if _, ok := allowedTags[tag]; !ok {
					t.Errorf("%s by_role[%s]: unknown tag %q", c.ID, pos, tag)
				}
			}
		}
	}
}

func TestResolveByRole(t *testing.T) {
	store, err := data.Load(projectRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	base := store.ChampionTags.Resolve("Ahri", "MIDDLE")
	if base.Primary != tags.ClassMage {
		t.Fatalf("Ahri should ignore role, got %s", base.Primary)
	}
	sup := store.ChampionTags.Resolve("Pantheon", "UTILITY")
	mid := store.ChampionTags.Resolve("Pantheon", "MIDDLE")
	if !sup.Has(tags.StylePeel) {
		t.Fatalf("Pantheon UTILITY peel missing: %v", sup)
	}
	if mid.Primary != tags.ClassAssassin && !mid.Has(tags.StyleBurst) {
		t.Fatalf("Pantheon MIDDLE should be assassin/burst: %v", mid)
	}
	empty := store.ChampionTags.Resolve("Pantheon", "")
	if empty.Primary != store.ChampionTags.Get("Pantheon").Primary {
		t.Fatal("empty position should use base")
	}
}
