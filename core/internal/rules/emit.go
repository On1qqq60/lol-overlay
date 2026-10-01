package rules

import "sort"

// FinalizeBuild collapses exclusives, pins inventory, emits a buy-order path,
// and promotes mandatory counters into that path.
func FinalizeBuild(build []Slot, champ string, d DraftSignals, p PressureSignals, liveScale float64, owned []int) []Slot {
	own := ownedSet(owned)
	build = ApplyItemGates(champ, build)
	build = commitInventory(build, own)
	build = CollapseBoots(build)
	build = CollapseLethalityCores(build)
	build = collapseExclusiveGroups(build, d, p, liveScale, own)
	build = ProtectCorePriority(build)
	build = ensureCounters(build, champ, d, p, liveScale)
	build = emitPath(build, champ)
	build = ApplyItemGates(champ, build)
	return build
}

func ownedSet(owned []int) map[int]struct{} {
	m := map[int]struct{}{}
	for _, id := range owned {
		if id != 0 {
			m[id] = struct{}{}
		}
	}
	return m
}

// commitInventory locks boots and exclusive cores to what the player already bought
// (or started crafting), so live ticks cannot swap Mercs↔Sorcs or Zhonya↔Banshee.
func commitInventory(build []Slot, own map[int]struct{}) []Slot {
	if len(own) == 0 {
		return build
	}
	out := cloneSlots(build)

	ownedBoot := 0
	for id := range own {
		if IsFinishedBootsID(id) {
			ownedBoot = id
			break
		}
	}
	if ownedBoot != 0 {
		kept := false
		filtered := out[:0]
		for _, s := range out {
			if s.Role != "boots" && !IsBootsID(s.ItemID) {
				filtered = append(filtered, s)
				continue
			}
			if s.ItemID == ownedBoot {
				s.Role = "boots"
				s.Priority = 90
				filtered = append(filtered, s)
				kept = true
			}
		}
		if !kept {
			filtered = append(filtered, Slot{ItemID: ownedBoot, Name: bootName(ownedBoot), Priority: 90, Role: "boots"})
		}
		out = filtered
	}

	if committed(own, ItemZhonyas, ItemSeekers) {
		out = dropIDs(out, ItemBansheeVeil)
		out = bumpOrInsert(out, Slot{ItemID: ItemZhonyas, Name: "Zhonya's Hourglass", Priority: 72, Role: "defensive"})
	} else if _, has := own[ItemBansheeVeil]; has {
		out = dropIDs(out, ItemZhonyas)
		out = bumpOrInsert(out, Slot{ItemID: ItemBansheeVeil, Name: "Banshee's Veil", Priority: 72, Role: "defensive"})
	}

	return out
}

func committed(own map[int]struct{}, legendary, component int) bool {
	if _, ok := own[legendary]; ok {
		return true
	}
	_, ok := own[component]
	return ok
}

func dropIDs(build []Slot, ids ...int) []Slot {
	drop := map[int]struct{}{}
	for _, id := range ids {
		drop[id] = struct{}{}
	}
	out := make([]Slot, 0, len(build))
	for _, s := range build {
		if _, ok := drop[s.ItemID]; ok {
			continue
		}
		out = append(out, s)
	}
	return out
}

func bumpOrInsert(build []Slot, s Slot) []Slot {
	for i := range build {
		if build[i].ItemID == s.ItemID {
			if s.Priority > build[i].Priority {
				build[i].Priority = s.Priority
			}
			return build
		}
	}
	return append(build, s)
}

func bootName(id int) string {
	switch id {
	case ItemSorcs:
		return "Sorcerer's Shoes"
	case ItemMercTreads:
		return "Mercury's Treads"
	case ItemSteelcaps:
		return "Plated Steelcaps"
	case ItemBerserkers:
		return "Berserker's Greaves"
	case ItemIonians:
		return "Ionian Boots of Lucidity"
	case ItemSwifties:
		return "Boots of Swiftness"
	default:
		return "Boots"
	}
}

func collapseExclusiveGroups(build []Slot, d DraftSignals, p PressureSignals, liveScale float64, own map[int]struct{}) []Slot {
	apBurst := d.APBurst
	if liveScale > 0.01 && p.APBurst > apBurst {
		apBurst = p.APBurst
	}
	mrKeep := 1
	if apBurst >= 0.3 {
		mrKeep = 2
	}
	build = keepTopInGroup(build, []int{
		ItemForceOfNature, ItemSpiritVisage, ItemBansheeVeil, ItemMaw, 2504,
	}, mrKeep, own)
	build = keepTopInGroup(build, []int{
		ItemLordDominiks, ItemMortalReminder, ItemSeryldas, ItemVoidStaff, 3137,
	}, 1, own)
	build = keepTopInGroup(build, []int{
		ItemTrinity, ItemIceborn, ItemEclipse, ItemStridebreaker, ItemLichBane, ItemEssenceReaver,
	}, 1, own)
	build = keepTopInGroup(build, []int{
		ItemMalignance, ItemBlackfire, ItemLudens, ItemRodOfAges,
	}, 1, own)
	build = keepTopInGroup(build, []int{ItemLocket, ItemMoonstone, ItemBlackfire}, 1, own)
	build = keepTopInGroup(build, []int{
		ItemCelestialOpposition, ItemDreamMaker, ItemZazzak, ItemSolsticeSleigh, ItemBloodsong,
	}, 1, own)
	// One mage stasis/veil. Inventory commit already forced the winner; otherwise
	// pick by draft (stable for the whole game), not live item ticks.
	build = keepTopInGroup(build, []int{ItemZhonyas, ItemBansheeVeil}, 1, own)
	_ = d
	return build
}

func keepTopInGroup(build []Slot, ids []int, keep int, own map[int]struct{}) []Slot {
	if keep <= 0 {
		return build
	}
	set := map[int]struct{}{}
	for _, id := range ids {
		set[id] = struct{}{}
	}
	type ranked struct {
		idx int
		pri float64
	}
	var hits []ranked
	for i, s := range build {
		if _, ok := set[s.ItemID]; ok {
			pri := s.Priority
			if _, has := own[s.ItemID]; has {
				pri += 1000
			}
			hits = append(hits, ranked{i, pri})
		}
	}
	if len(hits) <= keep {
		return build
	}
	chosen := map[int]struct{}{}
	for k := 0; k < keep; k++ {
		best := -1
		bestPri := -1e9
		for _, h := range hits {
			if _, ok := chosen[h.idx]; ok {
				continue
			}
			if h.pri > bestPri {
				bestPri = h.pri
				best = h.idx
			}
		}
		if best >= 0 {
			chosen[best] = struct{}{}
		}
	}
	out := make([]Slot, 0, len(build))
	for i, s := range build {
		if _, ok := set[s.ItemID]; ok {
			if _, keepIdx := chosen[i]; !keepIdx {
				continue
			}
		}
		out = append(out, s)
	}
	return out
}

// IdentityPath is the champion default 6-slot shopping list before draft/live swaps.
func IdentityPath(seed []Slot, champ string) []Slot {
	return emitPath(cloneSlots(seed), champ)
}
// shopComponentIDs must never occupy the 6-row. Grievous on that row is a full item.
func shopComponentID(id int) bool {
	switch id {
	case ItemExecutioners, ItemOblivionOrb, ItemLostChapter, ItemSerratedDirk,
		ItemSeekers, ItemGiantsBelt, ItemBamis, ItemTear, ItemKindlegem,
		ItemHauntingGuise, ItemFatedAshes, ItemNeedlessly, ItemBlastingWand,
		ItemForbiddenIdol, ItemAmplifyingTome:
		return true
	default:
		return false
	}
}

// Identity (core/offensive/pen) fills first so Zhonya cannot eject Liandry.
// Buy order: start → components → core → boots → other identity → flex.
func emitPath(build []Slot, champ string) []Slot {
	var start, boots, comps, cores, identity, flex []Slot
	for _, s := range build {
		if shopComponentID(s.ItemID) && s.Role != "component" && s.Role != "start" {
			continue
		}
		switch s.Role {
		case "start":
			if len(start) == 0 {
				start = append(start, s)
			}
		case "boots":
			if len(boots) == 0 && !gatedItem(champ, s) {
				boots = append(boots, s)
			}
		case "component":
			comps = append(comps, s)
		case "core":
			cores = append(cores, s)
		case "offensive", "pen":
			identity = append(identity, s)
		default:
			flex = append(flex, s)
		}
	}
	sortSlotsByPriority(cores)
	sortSlotsByPriority(identity)
	sortSlotsByPriority(flex)
	// Pin 3 identity slots (core + top offensives). Last 2 compete by priority so
	// Morello/Banshee can appear without letting Zhonya jump ahead of Liandry.
	pinned := append([]Slot{}, cores...)
	if len(pinned) > 3 {
		pinned = pinned[:3]
	}
	for _, s := range identity {
		if len(pinned) >= 3 {
			break
		}
		if s.Role == "offensive" {
			pinned = append(pinned, s)
		}
	}
	pinnedIDs := map[int]struct{}{}
	for _, s := range pinned {
		pinnedIDs[s.ItemID] = struct{}{}
	}
	// Heartsteel-style 3-core tanks: keep one seed defensive so live counters
	// cannot auction both leftover slots (Randuin/LDR vs Jak'Sho).
	if len(cores) >= 3 {
		for _, s := range flex {
			if s.Role != "defensive" {
				continue
			}
			if _, ok := pinnedIDs[s.ItemID]; ok {
				continue
			}
			pinned = append(pinned, s)
			pinnedIDs[s.ItemID] = struct{}{}
			break
		}
	}
	var rest []Slot
	for _, s := range append(append([]Slot{}, identity...), flex...) {
		if _, ok := pinnedIDs[s.ItemID]; ok {
			continue
		}
		rest = append(rest, s)
	}
	sortSlotsByPriority(rest)
	legendaries := pinned
	for _, s := range rest {
		if len(legendaries) >= 5 {
			break
		}
		legendaries = append(legendaries, s)
	}
	var keptCores, keptRest []Slot
	for _, s := range legendaries {
		if s.Role == "core" {
			keptCores = append(keptCores, s)
		} else {
			keptRest = append(keptRest, s)
		}
	}
	out := make([]Slot, 0, 8)
	out = append(out, start...)
	out = append(out, comps...)
	out = append(out, keptCores...)
	out = append(out, boots...)
	out = append(out, keptRest...)
	return out
}

func sortSlotsByPriority(s []Slot) {
	sort.SliceStable(s, func(i, j int) bool {
		return s[i].Priority > s[j].Priority
	})
}

func ensureCounters(build []Slot, champ string, d DraftSignals, p PressureSignals, liveScale float64) []Slot {
	healer := d.Healer > 0.12 || (liveScale > 0.01 && p.HealUtil > 0.12)
	tank := d.Tank > 0.2 || (liveScale > 0.01 && p.Tank > 0.2)
	apBurst := d.APBurst
	if liveScale > 0.01 && p.APBurst > apBurst {
		apBurst = p.APBurst
	}
	needMR := apBurst > 0.25

	ap := false
	for _, s := range build {
		switch s.ItemID {
		case ItemMalignance, ItemBlackfire, ItemLudens, ItemRodOfAges, ItemLiandrys,
			ItemNashors, ItemRiftmaker, ItemRabadons, ItemMoonstone, ItemShadowflame:
			ap = true
		}
	}

	heal := d.Healer
	if liveScale > 0.01 && p.HealUtil > heal {
		heal = p.HealUtil
	}
	ad := d.AD
	if liveScale > 0.01 && p.AD > ad {
		ad = p.AD
	}

	if healer && !hasAny(build, ItemMorellonomicon, ItemExecutioners, ItemChempunk, ItemMortalReminder, ItemThornmail, ItemOblivionOrb) {
		if champIsTank(champ) {
			if grievousOutranksMR(heal, apBurst, ad) {
				build = replaceTail(build, Slot{ItemID: ItemThornmail, Name: "Thornmail", Priority: 51, Role: "defensive", Why: "Автоатаки и хил — шипы и грив"})
			}
		} else if ap && !hasAny(build, ItemForceOfNature, ItemSpiritVisage) {
			build = replaceTail(build, Slot{ItemID: ItemMorellonomicon, Name: "Morellonomicon", Priority: 51, Role: "utility", Why: "У врагов хил — режем восстановление"})
		} else if !ap {
			build = replaceTail(build, adGrievousSlot(build, 51))
		}
	}
	if tank && !champIsTank(champ) && !supportSkipsPen(build) && !hasAny(build, ItemLordDominiks, ItemMortalReminder, ItemSeryldas, ItemVoidStaff, ItemBlackCleaver, ItemLiandrys) {
		if ap {
			build = replaceTail(build, Slot{ItemID: ItemVoidStaff, Name: "Void Staff", Priority: 49, Role: "pen", Why: "Враги жирные / в MR — нужно магическое пробивание"})
		} else {
			build = replaceTail(build, Slot{ItemID: ItemLordDominiks, Name: "Lord Dominik's Regards", Priority: 49, Role: "pen", Why: "Враги жирные — режем броню и HP"})
		}
	}
	if needMR && !hasAny(build, ItemForceOfNature, ItemSpiritVisage, ItemBansheeVeil, ItemMaw, ItemMercTreads, ItemGuardianAngel) {
		if ap {
			build = replaceTail(build, Slot{ItemID: ItemBansheeVeil, Name: "Banshee's Veil", Priority: 47, Role: "defensive", Why: "AP-берст врагов — завеса на один скилл"})
		} else if champIsADC(champ) {
			build = replaceTail(build, Slot{ItemID: ItemMaw, Name: "Maw of Malmortius", Priority: 47, Role: "defensive", Why: "AP-берст — щит Maw, не танковый FoN"})
		} else {
			build = replaceTail(build, Slot{ItemID: ItemForceOfNature, Name: "Force of Nature", Priority: 47, Role: "defensive", Why: "AP-берст врагов — MR и скорость"})
		}
	}
	return ApplyItemGates(champ, build)
}

func supportSkipsPen(build []Slot) bool {
	if hasAny(build, ItemBlackfire, ItemLiandrys, ItemLudens, ItemMalignance) {
		return false
	}
	return hasAny(build,
		ItemWorldAtlas, ItemRunicCompass, ItemBountyOfWorlds,
		ItemMoonstone, ItemLocket, ItemRedemption, ItemMikaels,
		ItemKnightsVow, ItemArdent, ItemMandate, ItemShurelyas)
}

func hasAny(build []Slot, ids ...int) bool {
	set := map[int]struct{}{}
	for _, id := range ids {
		set[id] = struct{}{}
	}
	for _, s := range build {
		if _, ok := set[s.ItemID]; ok {
			return true
		}
	}
	return false
}

func adGrievousSlot(build []Slot, prio float64) Slot {
	if hasAny(build, ItemInfinityEdge, ItemKraken, ItemCollector, ItemYunTal, ItemRapidFirecannon) {
		return Slot{ItemID: ItemMortalReminder, Name: "Mortal Reminder", Priority: prio, Role: "pen", Why: "Хил и крит — грив с бронепеном"}
	}
	return Slot{ItemID: ItemChempunk, Name: "Chempunk Chainsword", Priority: prio, Role: "utility", Why: "Враги хилятся — полный грив в AD-сборке"}
}

func replaceTail(build []Slot, s Slot) []Slot {
	for i := len(build) - 1; i >= 0; i-- {
		role := build[i].Role
		if role == "start" || role == "boots" || role == "core" || role == "component" {
			continue
		}
		switch build[i].ItemID {
		case ItemDeadMans, ItemForceOfNature, ItemSpiritVisage:
			continue
		}
		build[i] = s
		return build
	}
	return build
}

// SkipEarlyBoots is true when gold is too low to buy boots before an unfinished core.
func SkipEarlyBoots(gold int, s Slot, build []Slot, owned map[int]struct{}) bool {
	if s.Role != "boots" || gold < 0 || gold >= 800 {
		return false
	}
	for _, x := range build {
		if IsSupportQuestStart(x.ItemID) {
			return false
		}
	}
	for _, x := range build {
		if x.Role != "core" && x.Role != "component" {
			continue
		}
		if _, ok := owned[x.ItemID]; ok {
			continue
		}
		if ownsUpgradeOfRules(owned, x.ItemID) {
			continue
		}
		return true
	}
	return false
}

func ownsUpgradeOfRules(owned map[int]struct{}, componentID int) bool {
	for _, up := range ComponentsToward[componentID] {
		if _, ok := owned[up]; ok {
			return true
		}
	}
	for leg, comps := range ComponentsToward {
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
