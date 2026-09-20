package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"lol-build-overlay/internal/data"
	"lol-build-overlay/internal/engine"
	"lol-build-overlay/internal/liveclient"
)

func main() {
	live := flag.Bool("live", false, "fetch from Live Client API")
	fixture := flag.String("fixture", "", "path to allgamedata JSON fixture")
	rootFlag := flag.String("root", "", "project root containing data/")
	jsonOut := flag.Bool("json", false, "print JSON instead of text")
	flag.Parse()

	root := *rootFlag
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			fatal(err)
		}
		root, err = data.FindDataRoot(wd)
		if err != nil {
			fatal(err)
		}
	}

	store, err := data.Load(root)
	if err != nil {
		fatal(err)
	}

	var snap engine.GameSnapshot
	switch {
	case *fixture != "":
		snap, err = liveclient.LoadFixture(*fixture)
	case *live:
		snap, err = liveclient.New().Fetch()
	default:
		fmt.Fprintln(os.Stderr, "usage: recommend -live | -fixture PATH")
		os.Exit(2)
	}
	if err != nil {
		fatal(err)
	}

	rec := engine.Recommend(store, snap)

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(rec)
		return
	}

	printText(snap, rec, store)
}

func printText(snap engine.GameSnapshot, rec engine.Recommendation, store *data.Store) {
	fmt.Printf("Active: %s (%s) seed=%s gold=%d allyFront=%.2f",
		snap.ActiveChampionID, snap.ActiveTeam, rec.SeedName, rec.CurrentGold, rec.AllyFront)
	if rec.Offrole {
		fmt.Print(" [offrole]")
	}
	fmt.Println()
	fmt.Println("Enemies:")
	for i, p := range rec.Profiles {
		w := 0.0
		if i < len(rec.Threats) {
			w = rec.Threats[i].Weight
		}
		flag := ""
		if p.Overridden {
			flag = " [off-meta]"
		}
		pos := p.Player.Position
		if pos == "" {
			pos = "?"
		}
		fmt.Printf("  - %s@%s items=%v weight=%.2f%s\n", p.Player.ChampionID, pos, nonzero(p.Player.Items), w, flag)
	}

	fmt.Printf("\nPressure: tank=%.2f ad=%.2f mr=%.2f heal=%.2f apBurst=%.2f crit=%.2f\n",
		rec.Pressure.Tank, rec.Pressure.AD, rec.Pressure.MR, rec.Pressure.HealUtil, rec.Pressure.APBurst, rec.Pressure.Crit)

	fmt.Println("\nTop threats:")
	for _, t := range rec.TopThreats {
		fmt.Printf("  - %s weight=%.2f form=%.2f econ=%.2f kit=%.2f\n",
			t.ChampionID, t.Weight, t.Form, t.Econ, t.Kit)
	}

	if rec.HasNext {
		name := rec.NextItem.Name
		if it, ok := store.Items.ByID[rec.NextItem.ItemID]; ok && it.NameEN != "" {
			name = it.NameEN
		}
		fmt.Printf("\nNext item: %s (%d) priority=%.0f [%s]\n", name, rec.NextItem.ItemID, rec.NextItem.Priority, rec.NextItem.Role)
	} else {
		fmt.Println("\nNext item: (build complete)")
	}

	fmt.Println("\nBuild order (by priority):")
	for _, s := range rec.Build {
		owned := ""
		for _, id := range rec.OwnedItems {
			if id == s.ItemID {
				owned = " ✓"
				break
			}
		}
		fmt.Printf("  %5.0f  %-24s %s%s\n", s.Priority, s.Name, rolePad(s.Role), owned)
	}

	if len(rec.Reasons) > 0 {
		fmt.Println("\nReasons:")
		for _, r := range rec.Reasons {
			fmt.Printf("  • %s\n", r)
		}
	}
}

func nonzero(items []int) []int {
	var out []int
	for _, id := range items {
		if id != 0 {
			out = append(out, id)
		}
	}
	return out
}

func rolePad(role string) string {
	return fmt.Sprintf("[%s]", role)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "recommend: %v\n", err)
	// Helpful hint when fixture relative path fails
	if strings.Contains(err.Error(), "fixture") {
		fmt.Fprintf(os.Stderr, "cwd hint: %s\n", mustWD())
	}
	os.Exit(1)
}

func mustWD() string {
	wd, err := os.Getwd()
	if err != nil {
		return "?"
	}
	return filepath.Clean(wd)
}
