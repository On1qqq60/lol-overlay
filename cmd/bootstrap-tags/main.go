package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"lol-build-overlay/internal/data"
	"lol-build-overlay/internal/tags"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fatal(err)
	}
	root, err := data.FindDataRoot(wd)
	if err != nil {
		fatal(err)
	}
	store, err := data.Load(root)
	if err != nil {
		fatal(err)
	}

	out := map[string]tags.ChampionProfile{}
	for _, c := range store.Champions {
		out[c.ID] = bootstrapProfile(c)
	}
	for id, p := range curatedOverrides() {
		if _, ok := store.ChampionByID[id]; !ok {
			fmt.Fprintf(os.Stderr, "skip curated unknown champ %s\n", id)
			continue
		}
		out[id] = p
	}
	// JSON curated fixes win over Go curatedOverrides (scripts/apply_champion_tag_fixes.py).
	if fixes, err := loadJSONFixes(filepath.Join(root, "data", "champion_tag_fixes.json")); err != nil {
		fmt.Fprintf(os.Stderr, "champion_tag_fixes.json: %v\n", err)
	} else if len(fixes) > 0 {
		for id, p := range fixes {
			if _, ok := store.ChampionByID[id]; !ok {
				fmt.Fprintf(os.Stderr, "skip fix unknown champ %s\n", id)
				continue
			}
			out[id] = p
		}
		fmt.Printf("applied %d profiles from champion_tag_fixes.json\n", len(fixes))
	}

	ids := make([]string, 0, len(out))
	for id := range out {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	ordered := make(map[string]tags.ChampionProfile, len(ids))
	for _, id := range ids {
		ordered[id] = out[id]
	}

	payload := struct {
		Source    string                          `json:"source"`
		Count     int                             `json:"count"`
		Champions map[string]tags.ChampionProfile `json:"champions"`
	}{
		Source:    "bootstrap+curated",
		Count:     len(ordered),
		Champions: ordered,
	}

	path := filepath.Join(root, "data", "champion_tags.json")
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("wrote %s (%d champions)\n", path, len(ordered))
}

func bootstrapProfile(c data.Champion) tags.ChampionProfile {
	primary := mapDDragonClass(c.Tags)
	tagSet := map[string]struct{}{}
	add := func(t string) {
		if t != "" && t != primary {
			tagSet[t] = struct{}{}
		}
	}

	if c.Stats.AttackRange >= 300 {
		add(tags.RangeRanged)
	} else {
		add(tags.RangeMelee)
	}

	switch {
	case c.Info.Magic >= 7 && c.Info.Attack >= 7:
		add(tags.DamageHybrid)
	case c.Info.Magic >= 6:
		add(tags.DamageAP)
	case c.Info.Attack >= 6:
		add(tags.DamageAD)
	default:
		if primary == tags.ClassMage || primary == tags.ClassSupport {
			add(tags.DamageAP)
		} else {
			add(tags.DamageAD)
		}
	}

	switch primary {
	case tags.ClassMage:
		if c.Info.Difficulty >= 6 {
			add(tags.StyleBurst)
		} else {
			add(tags.StylePoke)
		}
	case tags.ClassAssassin:
		add(tags.StyleBurst)
		add(tags.ThreatDive)
		add(tags.ThreatMobility)
		add(tags.ExtraPick)
	case tags.ClassMarksman:
		add(tags.StyleDPS)
		add(tags.ThreatSplitpush)
	case tags.ClassFighter:
		add(tags.StyleEngage)
		add(tags.ThreatSustain)
		if c.Info.Defense >= 6 {
			add(tags.ExtraJuggernaut)
		}
	case tags.ClassTank:
		add(tags.StyleEngage)
		add(tags.StylePeel)
		add(tags.ThreatCCHard)
		add(tags.ThreatSustain)
	case tags.ClassSupport:
		add(tags.StylePeel)
		if c.Info.Magic >= 6 {
			add(tags.StyleShield)
		} else {
			add(tags.StyleEngage)
			add(tags.ThreatCCHard)
		}
	}

	list := make([]string, 0, len(tagSet))
	for t := range tagSet {
		list = append(list, t)
	}
	sort.Strings(list)
	return tags.ChampionProfile{Primary: primary, Tags: list}
}

func mapDDragonClass(dd []string) string {
	prio := []struct {
		dd  string
		our string
	}{
		{"Assassin", tags.ClassAssassin},
		{"Mage", tags.ClassMage},
		{"Marksman", tags.ClassMarksman},
		{"Tank", tags.ClassTank},
		{"Support", tags.ClassSupport},
		{"Fighter", tags.ClassFighter},
	}
	for _, p := range prio {
		for _, t := range dd {
			if t == p.dd {
				return p.our
			}
		}
	}
	return tags.ClassFighter
}

func loadJSONFixes(path string) (map[string]tags.ChampionProfile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var payload struct {
		Champions map[string]tags.ChampionProfile `json:"champions"`
	}
	if err := json.Unmarshal(b, &payload); err != nil {
		return nil, err
	}
	return payload.Champions, nil
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "bootstrap-tags: %v\n", err)
	os.Exit(1)
}
