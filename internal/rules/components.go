package rules

// ComponentsToward maps legendary/core item → preferred buy-order components.
var ComponentsToward = map[int][]int{
	ItemMalignance:     {ItemLostChapter, ItemFiendishCodex},
	ItemLudens:         {ItemLostChapter, ItemNeedlessly},
	ItemBlackfire:      {ItemLostChapter, ItemFatedAshes},
	ItemRodOfAges:      {3803, ItemNeedlessly}, // Catalyst
	ItemZhonyas:        {ItemSeekers, ItemNeedlessly},
	ItemBansheeVeil:    {ItemNullMagic, ItemNeedlessly},
	ItemLiandrys:       {ItemHauntingGuise, ItemFatedAshes},
	ItemShadowflame:    {ItemNeedlessly, ItemBlastingWand},
	ItemStormsurge:     {ItemHextechAlt, ItemAetherWisp},
	ItemRabadons:       {ItemNeedlessly, ItemNeedlessly},
	ItemVoidStaff:      {ItemBlastingWand},
	ItemMorellonomicon: {ItemOblivionOrb, ItemBlastingWand},
	ItemYoumuu:         {ItemSerratedDirk},
	ItemOpportunity:    {ItemSerratedDirk},
	ItemEclipse:        {ItemSerratedDirk},
	ItemInfinityEdge:   {1038}, // BF-like if present; else skipped by gold lookup
	ItemKraken:         {6670}, // Noonquiver
	ItemBOTRK:          {1053, 1043},
	ItemTrinity:        {3057, 3044, 3051},
	ItemSunderedSky:    {3133},
	ItemBlackCleaver:   {3133, 3067},
	ItemSteraks:        {1037, 1028}, // pickaxe / ruby-ish; gold picker soft
	ItemIceborn:        {3057, 1028},
	ItemExecutioners:   {},
}

// PickAffordable returns the best purchase for a target slot given gold.
// gold < 0 means "ignore gold" (fixtures / unknown) → return full item.
func PickAffordable(target Slot, owned map[int]struct{}, gold int, costOf func(int) int, nameOf func(int) string) (Slot, bool) {
	if _, ok := owned[target.ItemID]; ok {
		return Slot{}, false
	}
	if gold < 0 {
		return target, true
	}

	fullCost := costOf(target.ItemID)
	if fullCost > 0 && gold >= fullCost {
		return target, true
	}

	comps := ComponentsToward[target.ItemID]
	var best Slot
	bestCost := -1
	for _, cid := range comps {
		if _, ok := owned[cid]; ok {
			continue
		}
		cc := costOf(cid)
		if cc <= 0 {
			continue
		}
		if gold >= cc && cc > bestCost {
			bestCost = cc
			name := nameOf(cid)
			if name == "" {
				name = target.Name + " component"
			}
			best = Slot{ItemID: cid, Name: name, Priority: target.Priority, Role: "component"}
		}
	}
	if bestCost >= 0 {
		return best, true
	}

	// Cannot afford anything — still surface the goal item as next objective.
	return target, true
}
