package main

import (
	"fmt"
	"os"

	"lol-build-overlay/internal/data"
	"lol-build-overlay/internal/engine"
	"lol-build-overlay/internal/liveclient"
	"lol-build-overlay/internal/rules"
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
	fam, label := rules.SeedForContext(rules.SeedContext{
		Primary:    store.ChampionTags.Get(ap.ChampionID).Effective(ap.Position).Primary,
		Tags:       store.ChampionTags.Get(ap.ChampionID).Effective(ap.Position).All(),
		Position:   ap.Position,
		ChampionID: ap.ChampionID,
		Items:      ap.Items,
	})
	fmt.Printf("label=%s nseed=%d\n", label, len(fam))
	for _, s := range fam {
		fmt.Printf("  %s %d %s\n", s.Role, s.ItemID, s.Name)
	}
	rec := engine.Recommend(store, snap)
	fmt.Printf("seedName=%s legendaries:\n", rec.SeedName)
	n := 0
	for _, s := range rec.Build {
		if s.Role != "start" && s.Role != "component" && s.Role != "consumable" && s.Role != "boots" {
			n++
		}
		fmt.Printf("  %s %d %s\n", s.Role, s.ItemID, s.Name)
	}
	fmt.Printf("overlayLegendaries=%d\n", n)
}
