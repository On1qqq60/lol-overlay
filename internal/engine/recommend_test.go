package engine_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"lol-build-overlay/internal/data"
	"lol-build-overlay/internal/engine"
	"lol-build-overlay/internal/liveclient"
	"lol-build-overlay/internal/rules"
	"lol-build-overlay/internal/tags"
)

func projectRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func loadStore(t *testing.T) *data.Store {
	t.Helper()
	store, err := data.Load(projectRoot(t))
	if err != nil {
		t.Fatalf("load store: %v", err)
	}
	if len(store.ChampionTags) == 0 {
		t.Fatal("champion_tags.json empty — run go run ./cmd/bootstrap-tags")
	}
	return store
}

func loadFixture(t *testing.T, name string) engine.GameSnapshot {
	t.Helper()
	path := filepath.Join(projectRoot(t), "testdata", "fixtures", name)
	snap, err := liveclient.LoadFixture(path)
	if err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	return snap
}

func slotPriority(build []rules.Slot, id int) float64 {
	for _, s := range build {
		if s.ItemID == id {
			return s.Priority
		}
	}
	return -999
}

func TestPoppyTankRaisesLiandry(t *testing.T) {
	store := loadStore(t)
	snap := loadFixture(t, "poppy_tank.json")
	rec := engine.Recommend(store, snap)
	liandry := slotPriority(rec.Build, rules.ItemLiandrys)
	if liandry < 45 {
		t.Fatalf("expected Liandry priority high for tank Poppy, got %.1f reasons=%v", liandry, rec.Reasons)
	}
}

func TestPoppyLethalityLowersLiandryRaisesZhonya(t *testing.T) {
	store := loadStore(t)
	snap := loadFixture(t, "poppy_lethality.json")
	rec := engine.Recommend(store, snap)

	var poppy *engine.EnemyProfile
	for i := range rec.Profiles {
		if rec.Profiles[i].Player.ChampionID == "Poppy" {
			poppy = &rec.Profiles[i]
			break
		}
	}
	if poppy == nil {
		t.Fatal("Poppy not in profiles")
	}
	if !poppy.Overridden {
		t.Fatalf("expected Poppy off-meta override, profile=%v live=%v", poppy.ProfileTags, poppy.LiveTags)
	}

	liandry := slotPriority(rec.Build, rules.ItemLiandrys)
	zhonya := slotPriority(rec.Build, rules.ItemZhonyas)
	if liandry >= zhonya {
		t.Fatalf("expected Zhonya (%.1f) > Liandry (%.1f); reasons=%v", zhonya, liandry, rec.Reasons)
	}
	if rec.LiveScale < rec.DraftScale {
		t.Fatalf("expected liveScale >= draftScale mid-game, got live=%.2f draft=%.2f", rec.LiveScale, rec.DraftScale)
	}
}

func TestFeederAssassinIgnored(t *testing.T) {
	store := loadStore(t)
	snap := loadFixture(t, "threat_mix.json")
	rec := engine.Recommend(store, snap)

	var zedWeight, sionWeight float64
	for _, th := range rec.Threats {
		switch th.ChampionID {
		case "Zed":
			zedWeight = th.Weight
			if !th.Ignored {
				t.Fatalf("0/20 Zed should be ignored, weight=%.3f", th.Weight)
			}
		case "Sion":
			sionWeight = th.Weight
		}
	}
	if zedWeight >= 0.08 {
		t.Fatalf("feeder Zed weight too high: %.3f", zedWeight)
	}
	if sionWeight <= zedWeight {
		t.Fatalf("fed Sion (%.3f) should outweigh feeder Zed (%.3f)", sionWeight, zedWeight)
	}
}

func TestFedEnchanterMorello(t *testing.T) {
	store := loadStore(t)
	snap := loadFixture(t, "threat_mix.json")
	rec := engine.Recommend(store, snap)

	found := false
	for _, s := range rec.Build {
		if s.ItemID == rules.ItemMorellonomicon {
			found = true
			if s.Priority < 45 {
				t.Fatalf("Morello priority too low: %.1f", s.Priority)
			}
		}
	}
	if !found {
		t.Fatalf("expected Morellonomicon for fed Soraka; reasons=%v", rec.Reasons)
	}
}

func TestDraftStartTags(t *testing.T) {
	store := loadStore(t)
	snap := loadFixture(t, "draft_start.json")
	rec := engine.Recommend(store, snap)

	if rec.DraftScale < 0.9 {
		t.Fatalf("early game draftScale should be ~1, got %.2f", rec.DraftScale)
	}
	liandry := slotPriority(rec.Build, rules.ItemLiandrys)
	banshee := slotPriority(rec.Build, rules.ItemBansheeVeil)
	if liandry < 42 {
		t.Fatalf("draft Liandry should be boosted, got %.1f", liandry)
	}
	if banshee < 60 {
		t.Fatalf("draft Banshee should be boosted vs Veigar, got %.1f", banshee)
	}
	hasMorello := false
	for _, s := range rec.Build {
		if s.ItemID == rules.ItemMorellonomicon {
			hasMorello = true
		}
	}
	if !hasMorello {
		t.Fatal("draft should include Morello vs Soraka healer")
	}
}

func TestPantheonByRole(t *testing.T) {
	store := loadStore(t)
	snap := loadFixture(t, "pantheon_roles.json")
	rec := engine.Recommend(store, snap)

	var util, mid *engine.EnemyProfile
	for i := range rec.Profiles {
		p := &rec.Profiles[i]
		if p.Player.ChampionID != "Pantheon" {
			continue
		}
		switch p.Player.Position {
		case "UTILITY":
			util = p
		case "MIDDLE":
			mid = p
		}
	}
	if util == nil || mid == nil {
		t.Fatalf("need both Pantheon roles, got profiles=%d", len(rec.Profiles))
	}
	if !util.Base.Has(tags.StylePeel) {
		t.Fatalf("UTILITY Pantheon should have peel, got %v", util.Base)
	}
	if mid.Base.Primary != tags.ClassAssassin && !mid.Base.Has(tags.StyleBurst) {
		t.Fatalf("MIDDLE Pantheon should be assassin/burst, got %v", mid.Base)
	}
}

func TestGoldPicksComponent(t *testing.T) {
	store := loadStore(t)
	snap := loadFixture(t, "draft_start.json") // gold=450, owns Doran's
	rec := engine.Recommend(store, snap)
	if !rec.HasNext {
		t.Fatal("expected next item")
	}
	// 450 gold: Lost Chapter is 900 — should suggest a cheaper component or Lost Chapter as goal.
	// Amplifying Tome etc. — Lost Chapter components; with 450 might get nothing affordable toward LC
	// except maybe nothing — PickAffordable returns full item if can't afford component.
	// Give more targeted check via PickAffordable unit-style:
	owned := map[int]struct{}{1056: {}}
	slot := rules.Slot{ItemID: rules.ItemZhonyas, Name: "Zhonya", Priority: 55, Role: "defensive"}
	picked, ok := rules.PickAffordable(slot, owned, 1600, store.Items.GoldCost, store.Items.NameEN)
	if !ok {
		t.Fatal("expected pick")
	}
	if picked.ItemID != rules.ItemSeekers && picked.ItemID != rules.ItemZhonyas {
		t.Fatalf("with 1600g expected Seekers or Zhonya, got %d %s", picked.ItemID, picked.Name)
	}
	_ = rec
}

func TestFormEconKitMath(t *testing.T) {
	profiles := []engine.EnemyProfile{
		{
			Player: engine.PlayerSnapshot{ChampionID: "Zed", Kills: 0, Deaths: 20, Assists: 0, Items: []int{3142}, Level: 7},
			Kit:    1.2,
		},
		{
			Player: engine.PlayerSnapshot{ChampionID: "Sion", Kills: 20, Deaths: 0, Assists: 5, Items: []int{1, 2, 3, 4, 5, 6}, Level: 16},
			Kit:    0.7,
		},
	}
	legCount := func(items []int) int {
		n := 0
		for _, id := range items {
			if id != 0 {
				n++
			}
		}
		return n
	}
	rows := engine.ComputeThreat(profiles, legCount, 16, "")
	if rows[0].Weight >= rows[1].Weight {
		t.Fatalf("fed weak kit should beat feeder assassin: %.3f vs %.3f", rows[1].Weight, rows[0].Weight)
	}
}

func TestOffMetaThreshold(t *testing.T) {
	store := loadStore(t)
	base := store.ChampionTags.Get("Poppy")
	if base.Primary != tags.ClassTank {
		t.Fatalf("Poppy primary should be tank, got %s", base.Primary)
	}
	items := []int{3142, 3072, 3009}
	profiles := engine.BuildProfiles(
		[]engine.PlayerSnapshot{{ChampionID: "Poppy", Position: "UTILITY", Items: items, Level: 15}},
		store.ChampionTags,
		store.Items,
	)
	if len(profiles) != 1 || !profiles[0].Overridden {
		t.Fatalf("expected override for lethality Poppy, got %+v", profiles)
	}
}

func TestBlendScales(t *testing.T) {
	d0, l0 := engine.BlendScales(0)
	d2, l2 := engine.BlendScales(2)
	if d0 <= d2 {
		t.Fatalf("draft should fade: d0=%.2f d2=%.2f", d0, d2)
	}
	if l2 <= l0 {
		t.Fatalf("live should rise: l0=%.2f l2=%.2f", l0, l2)
	}
}

func TestIllaoiJuggernautNoMorelloFromShield(t *testing.T) {
	store := loadStore(t)
	snap := engine.GameSnapshot{
		ActiveChampionID: "Illaoi",
		ActiveTeam:       "ORDER",
		CurrentGold:      500,
		Players: []engine.PlayerSnapshot{
			{ChampionID: "Illaoi", Team: "ORDER", Position: "TOP", Level: 1, Items: []int{1055, 2003}},
			{ChampionID: "Sivir", Team: "CHAOS", Position: "BOTTOM", Level: 1, Items: []int{}},
			{ChampionID: "Malzahar", Team: "CHAOS", Position: "MIDDLE", Level: 1, Items: []int{}},
			{ChampionID: "MonkeyKing", Team: "CHAOS", Position: "JUNGLE", Level: 1, Items: []int{}},
			{ChampionID: "Maokai", Team: "CHAOS", Position: "TOP", Level: 1, Items: []int{}},
			{ChampionID: "Morgana", Team: "CHAOS", Position: "UTILITY", Level: 1, Items: []int{}},
		},
	}
	rec := engine.Recommend(store, snap)
	if rec.NextItem.ItemID == rules.ItemDoransBlade {
		t.Fatalf("Doran's owned → next should not be Doran's, got %+v", rec.NextItem)
	}
	for _, s := range rec.Build {
		if s.ItemID == rules.ItemMorellonomicon {
			t.Fatalf("AD Illaoi must not get Morello: %v", rec.Build)
		}
		if s.ItemID == rules.ItemTrinity {
			t.Fatalf("juggernaut Illaoi should not default Trinity: %v", rec.Build)
		}
	}
	foundCleaver := false
	for _, s := range rec.Build {
		if s.ItemID == rules.ItemBlackCleaver {
			foundCleaver = true
		}
	}
	if !foundCleaver {
		t.Fatalf("expected Cleaver in juggernaut seed, got %v", rec.Build)
	}
	for _, r := range rec.Reasons {
		if strings.Contains(r, "Morello") || strings.Contains(r, "heal/shield") {
			t.Fatalf("Morgana shield must not trigger Morello reason: %q", r)
		}
		if strings.Contains(r, "Steelcaps") {
			t.Fatalf("pruned boot reason still present: %q (build=%v)", r, rec.Build)
		}
	}
}

func TestJayceADBuildNoAPReasons(t *testing.T) {
	store := loadStore(t)
	snap := engine.GameSnapshot{
		ActiveChampionID: "Jayce",
		ActiveTeam:       "ORDER",
		CurrentGold:      759,
		Players: []engine.PlayerSnapshot{
			{ChampionID: "Jayce", Team: "ORDER", Position: "TOP", Level: 3, Items: []int{2010, 3340}},
			{ChampionID: "Nautilus", Team: "CHAOS", Position: "UTILITY", Level: 2, Items: []int{3865}},
			{ChampionID: "Ashe", Team: "CHAOS", Position: "BOTTOM", Level: 2, Items: []int{}},
			{ChampionID: "Volibear", Team: "CHAOS", Position: "JUNGLE", Level: 1, Items: []int{}},
			{ChampionID: "Sion", Team: "CHAOS", Position: "MIDDLE", Level: 3, Items: []int{1054}},
			{ChampionID: "Irelia", Team: "CHAOS", Position: "TOP", Level: 2, Items: []int{1055}},
		},
	}
	rec := engine.Recommend(store, snap)
	if rec.SeedName != "fighter" {
		t.Fatalf("seed=%s", rec.SeedName)
	}
	if rec.NextItem.ItemID == rules.ItemDoransBlade {
		t.Fatalf("level 3 should skip Doran's, next=%+v", rec.NextItem)
	}
	boots := 0
	for _, s := range rec.Build {
		if s.Role == "boots" {
			boots++
		}
		if s.ItemID == rules.ItemLiandrys || s.ItemID == rules.ItemZhonyas {
			t.Fatalf("AD Jayce build should not list AP item %s", s.Name)
		}
	}
	if boots != 1 {
		t.Fatalf("want exactly 1 boots slot, got %d in %v", boots, rec.Build)
	}
	for _, r := range rec.Reasons {
		if strings.Contains(r, "Liandry") || strings.Contains(r, "Zhonya") || strings.Contains(r, "Banshee") || strings.Contains(r, "Void Staff") {
			t.Fatalf("AD build should not emit AP reason %q", r)
		}
	}
	if rec.Pressure.AD < 0.1 {
		t.Fatalf("Ashe/Irelia should raise kit AD pressure, got %.2f", rec.Pressure.AD)
	}
}

func TestRemapRussianNamesAkaliSeed(t *testing.T) {
	store := loadStore(t)
	snap := engine.GameSnapshot{
		ActiveChampionID: "Акали",
		ActiveTeam:       "ORDER",
		CurrentGold:      500,
		Players: []engine.PlayerSnapshot{
			{ChampionID: "Акали", Team: "ORDER", Position: "MIDDLE", Level: 3, Items: []int{}},
			{ChampionID: "Вейгар", Team: "CHAOS", Position: "MIDDLE", Level: 3, Items: []int{}},
			{ChampionID: "Тимо", Team: "CHAOS", Position: "TOP", Level: 3, Items: []int{}},
			{ChampionID: "Наутилус", Team: "CHAOS", Position: "UTILITY", Level: 3, Items: []int{}},
		},
	}
	rec := engine.Recommend(store, snap)
	if rec.SeedName != "assassin" && rec.SeedName != "assassin/ap" {
		t.Fatalf("Akali seed want assassin or assassin/ap, got %q next=%v reasons=%v", rec.SeedName, rec.NextItem, rec.Reasons)
	}
	active, ok := engine.RemapChampionIDs(store, snap).ActivePlayer()
	if !ok || active.ChampionID != "Akali" {
		t.Fatalf("active remapped to %#v", active)
	}
	if rec.NextItem.ItemID == rules.ItemDoransBlade {
		t.Fatalf("should not recommend Doran's Blade for Akali, got %+v seed=%s", rec.NextItem, rec.SeedName)
	}
	start := rec.Build[0].ItemID
	if start != rules.ItemDarkSeal && start != rules.ItemDoransRing && start != rules.ItemStormsurge {
		t.Fatalf("expected Dark Seal / Stormsurge path start, got %+v", rec.Build[0])
	}
	foundStorm := false
	for _, s := range rec.Build {
		if s.ItemID == rules.ItemStormsurge {
			foundStorm = true
			break
		}
	}
	if !foundStorm {
		t.Fatalf("Akali AP assassin seed should include Stormsurge, build=%v", rec.Build)
	}
	if rec.Pressure.APBurst <= 0 {
		t.Fatalf("Veigar should contribute apBurst, pressure=%+v", rec.Pressure)
	}
}

func TestAllyFrontPresent(t *testing.T) {
	store := loadStore(t)
	snap := loadFixture(t, "poppy_tank.json") // has Malphite ally
	rec := engine.Recommend(store, snap)
	if rec.AllyFront < 0.3 {
		t.Fatalf("expected ally front from Malphite, got %.2f", rec.AllyFront)
	}
}

func TestSupportSoloLaneNotEnchanter(t *testing.T) {
	store := loadStore(t)
	snap := engine.GameSnapshot{
		ActiveChampionID: "Braum",
		ActiveTeam:       "ORDER",
		CurrentGold:      500,
		Players: []engine.PlayerSnapshot{
			{ChampionID: "Braum", Team: "ORDER", Position: "MIDDLE", Level: 1, Items: []int{1056}},
			{ChampionID: "Swain", Team: "CHAOS", Position: "MIDDLE", Level: 1},
			{ChampionID: "Jinx", Team: "CHAOS", Position: "BOTTOM", Level: 1},
			{ChampionID: "LeeSin", Team: "CHAOS", Position: "JUNGLE", Level: 1},
			{ChampionID: "Garen", Team: "CHAOS", Position: "TOP", Level: 1},
			{ChampionID: "Nami", Team: "CHAOS", Position: "UTILITY", Level: 1},
		},
	}
	rec := engine.Recommend(store, snap)
	if rec.SeedName == "support" {
		t.Fatalf("Braum MIDDLE must not keep support/enchanter seed, got %q build=%v", rec.SeedName, rec.Build)
	}
	for _, s := range rec.Build {
		if s.ItemID == 6617 { // Moonstone
			t.Fatalf("solo-lane Braum must not get Moonstone: %v", rec.Build)
		}
		if s.ItemID == rules.ItemWorldAtlas && s.Role == "start" {
			t.Fatalf("solo-lane Braum must not start Atlas: %v", rec.Build)
		}
	}
}

func TestSyncStartMatchesInventory(t *testing.T) {
	store := loadStore(t)
	// Compatible: mage seed Ring + owned Dark Seal → sync to Seal.
	snap := engine.GameSnapshot{
		ActiveChampionID: "Karma",
		ActiveTeam:       "ORDER",
		CurrentGold:      500,
		Players: []engine.PlayerSnapshot{
			{ChampionID: "Karma", Team: "ORDER", Position: "MIDDLE", Level: 1, Items: []int{1082, 2003}},
			{ChampionID: "Sett", Team: "CHAOS", Position: "TOP", Level: 1},
			{ChampionID: "Diana", Team: "CHAOS", Position: "MIDDLE", Level: 1},
			{ChampionID: "Lillia", Team: "CHAOS", Position: "JUNGLE", Level: 1},
			{ChampionID: "Jinx", Team: "CHAOS", Position: "BOTTOM", Level: 1},
			{ChampionID: "Nami", Team: "CHAOS", Position: "UTILITY", Level: 1},
		},
	}
	rec := engine.Recommend(store, snap)
	if len(rec.Build) == 0 || rec.Build[0].Role != "start" {
		t.Fatalf("expected start slot, got %v", rec.Build)
	}
	if rec.Build[0].ItemID != rules.ItemDarkSeal {
		t.Fatalf("compatible AP start must sync to owned Dark Seal, got %+v", rec.Build[0])
	}
}

func TestSyncStartIgnoresWrongDoran(t *testing.T) {
	store := loadStore(t)
	// Mage seed wants Ring; owned Blade must NOT be mirrored.
	snap := engine.GameSnapshot{
		ActiveChampionID: "Karma",
		ActiveTeam:       "ORDER",
		CurrentGold:      500,
		Players: []engine.PlayerSnapshot{
			{ChampionID: "Karma", Team: "ORDER", Position: "TOP", Level: 1, Items: []int{1055, 2003}},
			{ChampionID: "Sett", Team: "CHAOS", Position: "TOP", Level: 1},
			{ChampionID: "Diana", Team: "CHAOS", Position: "MIDDLE", Level: 1},
			{ChampionID: "Lillia", Team: "CHAOS", Position: "JUNGLE", Level: 1},
			{ChampionID: "Jinx", Team: "CHAOS", Position: "BOTTOM", Level: 1},
			{ChampionID: "Nami", Team: "CHAOS", Position: "UTILITY", Level: 1},
		},
	}
	rec := engine.Recommend(store, snap)
	if len(rec.Build) == 0 || rec.Build[0].Role != "start" {
		t.Fatalf("expected start slot, got %v", rec.Build)
	}
	if rec.Build[0].ItemID == rules.ItemDoransBlade {
		t.Fatalf("must not sync invalid Doran's Blade onto mage seed: %+v", rec.Build[0])
	}
	if rec.Build[0].ItemID != rules.ItemDoransRing && rec.Build[0].ItemID != rules.ItemDoransShield {
		t.Fatalf("mage/tank solo start expected Ring or Shield, got %+v", rec.Build[0])
	}
}

func TestLaneADBootsOverTeamAP(t *testing.T) {
	store := loadStore(t)
	snap := engine.GameSnapshot{
		ActiveChampionID: "Zed",
		ActiveTeam:       "ORDER",
		CurrentGold:      500,
		Players: []engine.PlayerSnapshot{
			{ChampionID: "Zed", Team: "ORDER", Position: "TOP", Level: 1, Items: []int{1055}},
			{ChampionID: "Pantheon", Team: "CHAOS", Position: "TOP", Level: 1},
			{ChampionID: "Lillia", Team: "CHAOS", Position: "JUNGLE", Level: 1},
			{ChampionID: "Diana", Team: "CHAOS", Position: "MIDDLE", Level: 1},
			{ChampionID: "Akshan", Team: "CHAOS", Position: "BOTTOM", Level: 1},
			{ChampionID: "Zilean", Team: "CHAOS", Position: "UTILITY", Level: 1},
		},
	}
	rec := engine.Recommend(store, snap)
	var boots rules.Slot
	for _, s := range rec.Build {
		if s.Role == "boots" {
			boots = s
			break
		}
	}
	if boots.ItemID == rules.ItemMercTreads {
		t.Fatalf("vs AD Pantheon lane, Mercs should lose to Steelcaps/Ionians; build=%v reasons=%v", rec.Build, rec.Reasons)
	}
	foundTail := false
	for _, s := range rec.Build {
		if s.ItemID == rules.ItemSeryldas || s.ItemID == rules.ItemCollector || s.ItemID == rules.ItemAxiomArc {
			foundTail = true
		}
	}
	if !foundTail {
		t.Fatalf("lethality build should include Serylda/Collector/Axiom tail, got %v", rec.Build)
	}
}

func TestJungleStartNoDorans(t *testing.T) {
	store := loadStore(t)
	snap := engine.GameSnapshot{
		ActiveChampionID: "Anivia",
		ActiveTeam:       "ORDER",
		CurrentGold:      500,
		Players: []engine.PlayerSnapshot{
			{ChampionID: "Anivia", Team: "ORDER", Position: "JUNGLE", Level: 1, Items: []int{}, SpellOne: "SummonerFlash", SpellTwo: "SummonerSmite"},
			{ChampionID: "Darius", Team: "CHAOS", Position: "TOP", Level: 1},
			{ChampionID: "Poppy", Team: "CHAOS", Position: "JUNGLE", Level: 1},
			{ChampionID: "Galio", Team: "CHAOS", Position: "MIDDLE", Level: 1},
			{ChampionID: "Quinn", Team: "CHAOS", Position: "BOTTOM", Level: 1},
			{ChampionID: "Ivern", Team: "CHAOS", Position: "UTILITY", Level: 1},
		},
	}
	rec := engine.Recommend(store, snap)
	if !strings.Contains(rec.SeedName, "jungle") {
		t.Fatalf("seed want */jungle, got %q", rec.SeedName)
	}
	if !rec.HasNext {
		t.Fatal("expected next item")
	}
	if rec.NextItem.ItemID == rules.ItemDoransRing || rec.NextItem.ItemID == rules.ItemDoransBlade || rec.NextItem.ItemID == rules.ItemDoransShield {
		t.Fatalf("jungle must not start Doran's, next=%+v build=%v", rec.NextItem, rec.Build)
	}
	if !rules.IsJunglePet(rec.NextItem.ItemID) && rec.Build[0].Role == "start" && !rules.IsJunglePet(rec.Build[0].ItemID) {
		t.Fatalf("expected jungle pet start, build0=%+v next=%+v", rec.Build[0], rec.NextItem)
	}
}

func TestUtilityForcesAtlasNotADC(t *testing.T) {
	store := loadStore(t)
	snap := engine.GameSnapshot{
		ActiveChampionID: "Aphelios",
		ActiveTeam:       "ORDER",
		CurrentGold:      500,
		Players: []engine.PlayerSnapshot{
			{ChampionID: "Aphelios", Team: "ORDER", Position: "UTILITY", Level: 1, Items: []int{}},
			{ChampionID: "Karma", Team: "CHAOS", Position: "UTILITY", Level: 1},
			{ChampionID: "Jinx", Team: "CHAOS", Position: "BOTTOM", Level: 1},
			{ChampionID: "Ahri", Team: "CHAOS", Position: "MIDDLE", Level: 1},
			{ChampionID: "LeeSin", Team: "CHAOS", Position: "JUNGLE", Level: 1},
			{ChampionID: "Garen", Team: "CHAOS", Position: "TOP", Level: 1},
		},
	}
	rec := engine.Recommend(store, snap)
	if rec.SeedName != "support" {
		t.Fatalf("Aphelios UTILITY seed want support, got %q", rec.SeedName)
	}
	foundAtlas := false
	for _, s := range rec.Build {
		if s.ItemID == rules.ItemWorldAtlas {
			foundAtlas = true
		}
		if s.ItemID == rules.ItemDoransBlade {
			t.Fatalf("UTILITY ADC should not use Doran's Blade seed: %v", rec.Build)
		}
		if s.ItemID == rules.ItemKraken {
			t.Fatalf("UTILITY ADC should not use crit ADC core: %v", rec.Build)
		}
	}
	if !foundAtlas {
		t.Fatalf("expected World Atlas start, build=%v", rec.Build)
	}
}

func TestLaneOpponentInTopThreats(t *testing.T) {
	store := loadStore(t)
	snap := engine.GameSnapshot{
		ActiveChampionID: "Viktor",
		ActiveTeam:       "ORDER",
		CurrentGold:      500,
		Players: []engine.PlayerSnapshot{
			{ChampionID: "Viktor", Team: "ORDER", Position: "MIDDLE", Level: 1, Items: []int{1056}},
			{ChampionID: "Mordekaiser", Team: "CHAOS", Position: "MIDDLE", Level: 1},
			{ChampionID: "Lillia", Team: "CHAOS", Position: "JUNGLE", Level: 1},
			{ChampionID: "Poppy", Team: "CHAOS", Position: "TOP", Level: 1},
			{ChampionID: "Jinx", Team: "CHAOS", Position: "BOTTOM", Level: 1},
			{ChampionID: "Soraka", Team: "CHAOS", Position: "UTILITY", Level: 1},
		},
	}
	rec := engine.Recommend(store, snap)
	found := false
	for _, th := range rec.TopThreats {
		if th.ChampionID == "Mordekaiser" {
			found = true
			if !th.LaneOpponent {
				t.Fatalf("Mordekaiser should be marked lane opponent: %+v", th)
			}
		}
	}
	if !found {
		t.Fatalf("lane Mordekaiser should be in TopThreats, got %v", rec.TopThreats)
	}
}

func TestLethalitySingleCore(t *testing.T) {
	store := loadStore(t)
	snap := engine.GameSnapshot{
		ActiveChampionID: "Zed",
		ActiveTeam:       "ORDER",
		CurrentGold:      500,
		Players: []engine.PlayerSnapshot{
			{ChampionID: "Zed", Team: "ORDER", Position: "MIDDLE", Level: 1, Items: []int{1055}},
			{ChampionID: "Ahri", Team: "CHAOS", Position: "MIDDLE", Level: 1},
			{ChampionID: "Sejuani", Team: "CHAOS", Position: "JUNGLE", Level: 1},
			{ChampionID: "Sion", Team: "CHAOS", Position: "TOP", Level: 1},
			{ChampionID: "Jinx", Team: "CHAOS", Position: "BOTTOM", Level: 1},
			{ChampionID: "Nautilus", Team: "CHAOS", Position: "UTILITY", Level: 1},
		},
	}
	rec := engine.Recommend(store, snap)
	high := 0
	for _, s := range rec.Build {
		for _, id := range rules.LethalityCoreIDs() {
			if s.ItemID == id && s.Priority >= 60 {
				high++
			}
		}
	}
	if high != 1 {
		t.Fatalf("want exactly 1 high-prio lethality core, got %d in %v", high, rec.Build)
	}
}

func poppyLaneSnap(position, smite string) engine.GameSnapshot {
	self := engine.PlayerSnapshot{
		ChampionID: "Poppy", Team: "ORDER", Position: position, Level: 1,
		Items: []int{1054}, SpellOne: "SummonerFlash", SpellTwo: smite,
	}
	return engine.GameSnapshot{
		ActiveChampionID: "Poppy",
		ActiveTeam:       "ORDER",
		CurrentGold:      500,
		Players: []engine.PlayerSnapshot{
			self,
			{ChampionID: "Darius", Team: "CHAOS", Position: "TOP", Level: 1},
			{ChampionID: "LeeSin", Team: "CHAOS", Position: "JUNGLE", Level: 1},
			{ChampionID: "Ahri", Team: "CHAOS", Position: "MIDDLE", Level: 1},
			{ChampionID: "Jinx", Team: "CHAOS", Position: "BOTTOM", Level: 1},
			{ChampionID: "Nami", Team: "CHAOS", Position: "UTILITY", Level: 1},
		},
	}
}

func TestPoppySoloLaneBruiserSeed(t *testing.T) {
	store := loadStore(t)
	cases := []struct {
		pos, smite, seed string
	}{
		{"TOP", "SummonerTeleport", "fighter"},
		{"JUNGLE", "SummonerSmite", "fighter/jungle"},
	}
	for _, tc := range cases {
		t.Run(tc.pos, func(t *testing.T) {
			rec := engine.Recommend(store, poppyLaneSnap(tc.pos, tc.smite))
			if rec.SeedName != tc.seed {
				t.Fatalf("seed want %q, got %q build=%v", tc.seed, rec.SeedName, rec.Build)
			}
			var iceborn, sundered, cleaver, sunfire rules.Slot
			hasOffensive := false
			for _, s := range rec.Build {
				switch s.ItemID {
				case rules.ItemIceborn:
					iceborn = s
				case rules.ItemSunderedSky:
					sundered = s
				case rules.ItemBlackCleaver:
					cleaver = s
				case 3068:
					sunfire = s
				}
				if s.Role == "core" || s.Role == "offensive" {
					hasOffensive = true
				}
			}
			if !hasOffensive {
				t.Fatalf("Poppy %s must not be defensive-only, build=%v", tc.pos, rec.Build)
			}
			if iceborn.Role != "core" {
				t.Fatalf("Poppy %s want Iceborn core, got %+v", tc.pos, iceborn)
			}
			if sundered.Role != "offensive" {
				t.Fatalf("Poppy %s want Sundered Sky offensive, got %+v", tc.pos, sundered)
			}
			if cleaver.Role != "offensive" {
				t.Fatalf("Poppy %s want Cleaver offensive, got %+v", tc.pos, cleaver)
			}
			if sunfire.ItemID != 0 {
				t.Fatalf("Poppy bruiser should not default Sunfire: %v", rec.Build)
			}
		})
	}
}

func TestPoppySupportStaysEngage(t *testing.T) {
	store := loadStore(t)
	snap := engine.GameSnapshot{
		ActiveChampionID: "Poppy",
		ActiveTeam:       "ORDER",
		CurrentGold:      500,
		Players: []engine.PlayerSnapshot{
			{ChampionID: "Poppy", Team: "ORDER", Position: "UTILITY", Level: 1, Items: []int{3865}},
			{ChampionID: "Jinx", Team: "CHAOS", Position: "BOTTOM", Level: 1},
			{ChampionID: "Ahri", Team: "CHAOS", Position: "MIDDLE", Level: 1},
			{ChampionID: "LeeSin", Team: "CHAOS", Position: "JUNGLE", Level: 1},
			{ChampionID: "Garen", Team: "CHAOS", Position: "TOP", Level: 1},
			{ChampionID: "Nami", Team: "CHAOS", Position: "UTILITY", Level: 1},
		},
	}
	rec := engine.Recommend(store, snap)
	if rec.SeedName != "support" {
		t.Fatalf("Poppy UTILITY want support, got %q", rec.SeedName)
	}
	foundLocket := false
	for _, s := range rec.Build {
		if s.ItemID == 3190 {
			foundLocket = true
		}
		if s.ItemID == rules.ItemSunderedSky || s.ItemID == rules.ItemIceborn {
			t.Fatalf("support Poppy should not use bruiser core: %v", rec.Build)
		}
	}
	if !foundLocket {
		t.Fatalf("support Poppy want Locket, build=%v", rec.Build)
	}
}

func TestMalphiteStaysTankSeed(t *testing.T) {
	store := loadStore(t)
	snap := engine.GameSnapshot{
		ActiveChampionID: "Malphite",
		ActiveTeam:       "ORDER",
		CurrentGold:      500,
		Players: []engine.PlayerSnapshot{
			{ChampionID: "Malphite", Team: "ORDER", Position: "TOP", Level: 1, Items: []int{1054}},
			{ChampionID: "Darius", Team: "CHAOS", Position: "TOP", Level: 1},
			{ChampionID: "LeeSin", Team: "CHAOS", Position: "JUNGLE", Level: 1},
			{ChampionID: "Ahri", Team: "CHAOS", Position: "MIDDLE", Level: 1},
			{ChampionID: "Jinx", Team: "CHAOS", Position: "BOTTOM", Level: 1},
			{ChampionID: "Nami", Team: "CHAOS", Position: "UTILITY", Level: 1},
		},
	}
	rec := engine.Recommend(store, snap)
	if rec.SeedName != "tank" {
		t.Fatalf("Malphite want tank seed, got %q", rec.SeedName)
	}
	if slotPriority(rec.Build, 3068) < 0 {
		t.Fatalf("Malphite want Sunfire core, build=%v", rec.Build)
	}
	if slotPriority(rec.Build, rules.ItemSunderedSky) >= 0 {
		t.Fatalf("true tank should not get Sundered Sky: %v", rec.Build)
	}
}

func TestSmiteParsedFromFixture(t *testing.T) {
	path := filepath.Join(projectRoot(t), "testdata", "generated", "live", "0002_Anivia_JUNGLE.json")
	snap, err := liveclient.LoadFixture(path)
	if err != nil {
		t.Skipf("generated fixture missing: %v", err)
	}
	active, ok := snap.ActivePlayer()
	if !ok {
		t.Fatal("no active")
	}
	if !active.HasSmite() {
		t.Fatalf("Anivia jungle fixture should parse Smite, spells=%q/%q", active.SpellOne, active.SpellTwo)
	}
	store := loadStore(t)
	rec := engine.Recommend(store, snap)
	if !strings.Contains(rec.SeedName, "jungle") {
		t.Fatalf("seed=%q", rec.SeedName)
	}
}
