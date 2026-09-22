package engine

import (
	"fmt"
	"sort"
	"strings"

	"lol-build-overlay/internal/data"
	"lol-build-overlay/internal/rules"
	"lol-build-overlay/internal/tags"
)

// Recommendation is the engine output for one snapshot.
type Recommendation struct {
	Build       []rules.Slot   `json:"build"`
	NextItem    rules.Slot     `json:"nextItem"`
	HasNext     bool           `json:"hasNext"`
	TopThreats  []ThreatRow    `json:"topThreats"`
	Pressure    Pressure       `json:"pressure"`
	Reasons     []string       `json:"reasons"`
	Profiles    []EnemyProfile `json:"profiles"`
	Threats     []ThreatRow    `json:"threats"`
	OwnedItems  []int          `json:"ownedItems"`
	DraftScale  float64        `json:"draftScale"`
	LiveScale   float64        `json:"liveScale"`
	AllyFront   float64        `json:"allyFront"`
	CurrentGold int            `json:"currentGold"`
	SeedName    string         `json:"seedName"`
	Offrole     bool           `json:"offrole"` // active champ outside NaturalPositions for their tags
}

// Recommend runs seed → draft/live adjust → next unfinished item.
func Recommend(store *data.Store, snap GameSnapshot) Recommendation {
	snap = RemapChampionIDs(store, snap)

	active, ok := snap.ActivePlayer()
	playerPrimary := tags.ClassMage
	var playerTags []string
	offrole := false
	if ok {
		raw := store.ChampionTags.Get(active.ChampionID)
		pp := raw.Effective(active.Position)
		playerPrimary = pp.Primary
		playerTags = pp.All()
		offrole = !raw.IsNaturalRole(active.Position)
	}

	enemies := snap.Enemies()
	profiles := BuildProfiles(enemies, store.ChampionTags, store.Items)

	draftProfiles := make([]tags.ChampionProfile, 0, len(enemies))
	for _, e := range enemies {
		draftProfiles = append(draftProfiles, store.ChampionTags.Resolve(e.ChampionID, e.Position))
	}
	draftSig := rules.DraftFromProfiles(draftProfiles)

	enemyLegs := 0
	for _, e := range enemies {
		enemyLegs += store.Items.LegendaryCount(e.Items)
	}
	draftScale, liveScale := BlendScales(enemyLegs)

	var seed []rules.Slot
	seedName := playerPrimary
	if ok {
		seed, seedName = rules.SeedForContext(rules.SeedContext{
			Primary:    playerPrimary,
			Tags:       playerTags,
			Position:   active.Position,
			ChampionID: active.ChampionID,
			HasSmite:   active.HasSmite(),
			Items:      active.Items,
		})
		seed = rules.SyncStartToInventory(seed, active.Items)
	} else {
		seed, seedName = rules.SeedForContext(rules.SeedContext{
			Primary: playerPrimary,
			Tags:    playerTags,
		})
	}
	build, reasons := rules.AdjustDraft(seed, draftSig, draftScale, playerPrimary, playerTags)
	reasons = append(reasons, fmt.Sprintf("blend: draft×%.2f live×%.2f (enemy legendaries=%d)", draftScale, liveScale, enemyLegs))

	maxLvl := MaxLevel(snap.Players)
	activePos := ""
	if ok {
		activePos = active.Position
	}
	threats := ComputeThreat(profiles, store.Items.LegendaryCount, maxLvl, activePos)
	pressure := ComputePressure(profiles, threats, store.Items)

	var liveReasons []string
	build, liveReasons = rules.AdjustByPressure(build, rules.PressureSignals{
		Tank:     pressure.Tank,
		AD:       pressure.AD,
		MR:       pressure.MR,
		HealUtil: pressure.HealUtil,
		APBurst:  pressure.APBurst,
		Crit:     pressure.Crit,
		CCHard:   pressure.CCHard,
		Dive:     pressure.Dive,
	}, liveScale, playerPrimary, playerTags)
	reasons = append(reasons, liveReasons...)

	// Lane opponent boots beat diffuse team draft when same-position enemy exists.
	if ok {
		laneEnemy := tags.ChampionProfile{}
		for _, e := range enemies {
			if tags.NormalizePosition(e.Position) != "" &&
				tags.NormalizePosition(e.Position) == tags.NormalizePosition(active.Position) {
				laneEnemy = store.ChampionTags.Resolve(e.ChampionID, e.Position)
				break
			}
		}
		var laneBootReasons []string
		build, laneBootReasons = rules.AdjustLaneBoots(build, laneEnemy, draftScale)
		reasons = append(reasons, laneBootReasons...)
	}

	selfID := ""
	if ok {
		selfID = active.ChampionID
	}
	allyFront := AllyFrontScore(snap.Allies(), selfID, store.ChampionTags)
	var allyReasons []string
	build, allyReasons = rules.AdjustAllyFront(build, allyFront, playerPrimary, playerTags)
	reasons = append(reasons, allyReasons...)

	ownedIDs := []int{}
	if ok {
		ownedIDs = active.Items
	}
	build = rules.FinalizeBuild(build, selfID, draftSig, rules.PressureSignals{
		Tank:     pressure.Tank,
		AD:       pressure.AD,
		MR:       pressure.MR,
		HealUtil: pressure.HealUtil,
		APBurst:  pressure.APBurst,
		Crit:     pressure.Crit,
		CCHard:   pressure.CCHard,
		Dive:     pressure.Dive,
	}, liveScale, ownedIDs)
	rules.FillWhy(build, store.Items)
	reasons = pruneBootReasons(reasons, build)

	if offrole {
		reasons = append(reasons, fmt.Sprintf("offrole: %s on %s", selfID, activePos))
	}

	sort.SliceStable(build, func(i, j int) bool {
		return build[i].Priority > build[j].Priority
	})

	owned := map[int]struct{}{}
	if ok {
		for _, id := range active.Items {
			if id != 0 {
				owned[id] = struct{}{}
			}
		}
	}

	pastStart := false
	gold := snap.CurrentGold
	if ok {
		legs := store.Items.LegendaryCount(active.Items)
		seedStart := 0
		ownsStart := false
		for _, s := range build {
			if s.Role != "start" {
				continue
			}
			seedStart = s.ItemID
			if _, has := owned[s.ItemID]; has {
				ownsStart = true
			}
			break
		}
		// Wrong-family Doran (e.g. Blade on mage) does not count as owning start.
		if !ownsStart && seedStart != 0 {
			for id := range owned {
				if rules.IsLaneStartItem(id) && rules.CompatibleStarts(seedStart, id) {
					ownsStart = true
					break
				}
			}
		}
		// Skip starters once owned, or once laning is underway.
		pastStart = legs > 0 || ownsStart || active.Level >= 2
	}
	next, hasNext := nextPurchase(build, owned, pastStart, gold, store.Items)
	build = dropFinishedComponents(build, owned)

	top := topThreats(threats, 2)

	ownedList := []int{}
	if ok {
		ownedList = active.Items
	}

	return Recommendation{
		Build:       build,
		NextItem:    next,
		HasNext:     hasNext,
		TopThreats:  top,
		Pressure:    pressure,
		Reasons:     reasons,
		Profiles:    profiles,
		Threats:     threats,
		OwnedItems:  ownedList,
		DraftScale:  draftScale,
		LiveScale:   liveScale,
		AllyFront:   allyFront,
		CurrentGold: gold,
		SeedName:    seedName,
		Offrole:     offrole,
	}
}

func nextPurchase(build []rules.Slot, owned map[int]struct{}, pastStart bool, gold int, items *tags.ItemCatalog) (rules.Slot, bool) {
	costOf := items.GoldCost
	nameOf := items.NameEN

	for _, s := range build {
		if _, ok := owned[s.ItemID]; ok {
			continue
		}
		if pastStart && s.Role == "start" {
			continue
		}
		if s.Role == "component" && ownsUpgradeOf(owned, s.ItemID) {
			continue
		}
		if ownsUpgradeOf(owned, s.ItemID) {
			continue
		}
		// Skip alternate boots if another boot owned.
		if s.Role == "boots" && ownsAnyBoots(owned, items) {
			continue
		}
		picked, ok := rules.PickAffordable(s, owned, gold, costOf, nameOf)
		if ok {
			return picked, true
		}
	}
	return rules.Slot{}, false
}

func dropFinishedComponents(build []rules.Slot, owned map[int]struct{}) []rules.Slot {
	out := make([]rules.Slot, 0, len(build))
	for _, s := range build {
		if s.Role == "component" && (ownsUpgradeOf(owned, s.ItemID) || hasOwned(owned, s.ItemID)) {
			continue
		}
		out = append(out, s)
	}
	return out
}

func hasOwned(owned map[int]struct{}, id int) bool {
	_, ok := owned[id]
	return ok
}

func ownsAnyBoots(owned map[int]struct{}, items *tags.ItemCatalog) bool {
	for id := range owned {
		if hasTag(items.Tags[id], tags.ItemBoots) {
			return true
		}
	}
	return false
}

func hasTag(list []string, t string) bool {
	for _, x := range list {
		if x == t {
			return true
		}
	}
	return false
}

func ownsUpgradeOf(owned map[int]struct{}, componentID int) bool {
	upgrades := map[int][]int{
		rules.ItemLostChapter:    {rules.ItemMalignance, rules.ItemLudens, rules.ItemBlackfire, rules.ItemRodOfAges},
		rules.ItemRecurveBow:      {rules.ItemNashors, rules.ItemBOTRK},
		rules.ItemAmplifyingTome: {rules.ItemNashors},
		rules.ItemOblivionOrb:    {rules.ItemMorellonomicon},
		rules.ItemSeekers:        {rules.ItemZhonyas},
		rules.ItemSerratedDirk:   {rules.ItemYoumuu, rules.ItemOpportunity, rules.ItemEclipse, 6698, 6699},
		rules.ItemHauntingGuise:  {rules.ItemLiandrys},
		rules.ItemFatedAshes:     {rules.ItemLiandrys, rules.ItemBlackfire},
	}
	for _, up := range upgrades[componentID] {
		if _, ok := owned[up]; ok {
			return true
		}
	}
	// Also: if recommending a component that is listed under ComponentsToward of an owned legendary
	for leg, comps := range rules.ComponentsToward {
		if _, ok := owned[leg]; !ok {
			continue
		}
		for _, c := range comps {
			if c == componentID {
				return true
			}
		}
	}
	return false
}

func topThreats(rows []ThreatRow, n int) []ThreatRow {
	cp := append([]ThreatRow(nil), rows...)
	sort.SliceStable(cp, func(i, j int) bool {
		return cp[i].Weight > cp[j].Weight
	})
	if len(cp) > n {
		cp = cp[:n]
	}
	return cp
}

func pruneBootReasons(reasons []string, build []rules.Slot) []string {
	has := func(id int) bool {
		for _, s := range build {
			if s.ItemID == id {
				return true
			}
		}
		return false
	}
	out := reasons[:0:0]
	for _, r := range reasons {
		if strings.Contains(r, "Steelcaps") && !has(rules.ItemSteelcaps) {
			continue
		}
		if strings.Contains(r, "Mercs") && !has(rules.ItemMercTreads) {
			continue
		}
		if strings.Contains(r, "Ionians") && !has(rules.ItemIonians) {
			continue
		}
		out = append(out, r)
	}
	return out
}
