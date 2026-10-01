package main

import (
	"fmt"
	"os"

	"lol-build-overlay/internal/data"
	"lol-build-overlay/internal/engine"
	"lol-build-overlay/internal/liveclient"
	"lol-build-overlay/internal/rules"
	"lol-build-overlay/internal/tags"
)

func main() {
	store, err := data.Load(`D:\lol-build-overlay\core`)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	snap, err := liveclient.New().Fetch()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	snap = engine.RemapChampionIDs(store, snap)
	ap, ok := snap.ActivePlayer()
	fmt.Printf("ok=%v activeID=%s team=%s player=%s pos=%s items=%v\n",
		ok, snap.ActiveChampionID, snap.ActiveTeam, ap.ChampionID, ap.Position, ap.Items)
	pp := store.ChampionTags.Get(ap.ChampionID).Effective(ap.Position)
	seed, label := rules.SeedForContext(rules.SeedContext{
		Primary:    pp.Primary,
		Tags:       pp.All(),
		Position:   ap.Position,
		ChampionID: ap.ChampionID,
		Items:      ap.Items,
	})
	fmt.Printf("primary=%s label=%s nseed=%d class=%s\n", pp.Primary, label, len(seed), tags.ClassTank)
	for _, s := range seed {
		fmt.Printf("  seed %s %d %s\n", s.Role, s.ItemID, s.Name)
	}
	rec := engine.Recommend(store, snap)
	fmt.Printf("seedName=%s\n", rec.SeedName)
	n := 0
	for _, s := range rec.Build {
		if s.Role != "start" && s.Role != "component" && s.Role != "consumable" && s.Role != "boots" {
			n++
		}
		fmt.Printf("  rec %s %d %s\n", s.Role, s.ItemID, s.Name)
	}
	fmt.Printf("overlayLegendaries=%d\n", n)
}
