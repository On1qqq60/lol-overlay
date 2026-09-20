package tags

// DeriveItemTags classifies a shop item for the engine.
func DeriveItemTags(it ItemInfo) []string {
	out := map[string]struct{}{}
	add := func(t string) { out[t] = struct{}{} }

	hasDDTag := func(name string) bool {
		for _, t := range it.Tags {
			if t == name {
				return true
			}
		}
		return false
	}

	if hasDDTag("Boots") {
		add(ItemBoots)
	}

	ad := it.Stats["FlatPhysicalDamageMod"]
	ap := it.Stats["FlatMagicDamageMod"]
	hp := it.Stats["FlatHPPoolMod"]
	armor := it.Stats["FlatArmorMod"]
	mr := it.Stats["FlatSpellBlockMod"]
	lethality := it.Stats["FlatArmorPenetrationMod"]
	crit := it.Stats["FlatCritChanceMod"]
	as := it.Stats["PercentAttackSpeedMod"]

	if ad >= 25 || (hasDDTag("Damage") && ad > 0) {
		add(ItemADDamage)
	}
	if ap >= 40 || hasDDTag("SpellDamage") {
		add(ItemAPDamage)
	}
	if lethality > 0 || hasDDTag("ArmorPenetration") {
		add(ItemLethality)
		add(ItemADDamage)
	}
	if hp >= 200 || (hasDDTag("Health") && hp > 0) {
		add(ItemTankHP)
	}
	if armor >= 20 || hasDDTag("Armor") {
		add(ItemArmor)
	}
	if mr >= 20 || hasDDTag("SpellBlock") {
		add(ItemMR)
	}
	if crit > 0 || hasDDTag("CriticalStrike") {
		add(ItemCrit)
		add(ItemADDamage)
	}
	if as > 0 || hasDDTag("AttackSpeed") {
		add(ItemAS)
	}
	if hasDDTag("OnHit") {
		add(ItemOnHit)
		add(ItemAS)
	}

	name := it.NameEN
	if containsAny(name,
		"Morellonomicon", "Thornmail", "Chempunk", "Oblivion Orb", "Executioner's",
		"Mortal Reminder", "Chempunk Chainsword",
	) {
		add(ItemHealCut)
	}
	switch it.ID {
	case 3165, 6609, 3033, 3075, 3123, 3916: // Morello, Chainsword, Mortal, Thornmail, Exec, Oblivion
		add(ItemHealCut)
	}

	if containsAny(name,
		"Locket", "Moonstone", "Redemption", "Knight's Vow", "Mikael",
		"Shurelya", "Imperial Mandate", "Staff of Flowing", "Dawncore",
		"Bandleglass", "Kindlegem",
	) || hasDDTag("Aura") && (hasDDTag("Support") || hp > 0 && ap > 0) {
		if containsAny(name, "Locket", "Moonstone", "Knight", "Mikael", "Shurelya", "Imperial", "Flowing", "Dawncore") {
			add(ItemShield)
		}
	}
	switch it.ID {
	case 3190, 6617, 3107, 3109, 3222, 2065, 4005, 6616, 6621:
		add(ItemShield)
	}

	if containsAny(name,
		"Blade of The Ruined", "Blade of the Ruined", "Nashor", "Guinsoo", "Wit's End",
		"Kraken", "Recurve Bow", "Rageblade",
	) {
		add(ItemOnHit)
		add(ItemAS)
	}
	switch it.ID {
	case 3153, 3115, 3124, 3091, 6672, 1043:
		add(ItemOnHit)
		add(ItemAS)
	}

	if containsAny(name, "Infinity Edge", "Zeal", "Collection", "Collector", "Lord Dominik", "Yun Tal", "Navori", "Rapid Firecannon", "Stormrazor") {
		add(ItemCrit)
	}
	switch it.ID {
	case 3031, 3086, 6676, 3036:
		add(ItemCrit)
		add(ItemADDamage)
	}

	if IsLegendary(it) {
		add(ItemLegendary)
	}

	res := make([]string, 0, len(out))
	for t := range out {
		res = append(res, t)
	}
	return res
}
