package rules

import "lol-build-overlay/internal/tags"

// Jungle pet / support start IDs.
const (
	ItemJungleScorchclaw  = 1101
	ItemJungleGustwalker  = 1102
	ItemJungleMosstomper  = 1103
	ItemWorldAtlas        = 3865
	ItemRapidFirecannon   = 3094
	ItemCollector         = 6676
	ItemLordDominiks      = 3036
	ItemHubris            = 6697
	ItemForceOfNature     = 4401
	ItemSeryldas          = 6694
	ItemAxiomArc          = 6696
	ItemIonians           = 3158
)

// SeedContext carries role/loadout signals the class template alone cannot see.
type SeedContext struct {
	Primary    string
	Tags       []string
	Position   string
	ChampionID string
	HasSmite   bool
	Items      []int
}

// SeedForPlayer picks a template from class/tags, then applies jungle/support routing.
func SeedForPlayer(primary string, allTags []string) []Slot {
	seed, _ := SeedForContext(SeedContext{Primary: primary, Tags: allTags})
	return seed
}

// SeedForContext returns the build seed and a seed label (e.g. "mage/jungle", "assassin/ap").
func SeedForContext(ctx SeedContext) ([]Slot, string) {
	primary := ctx.Primary
	allTags := ctx.Tags
	pos := tags.NormalizePosition(ctx.Position)

	// UTILITY → support line unless carry-support exception (Senna/Pyke-like).
	if pos == "UTILITY" && !isSupportCarryException(ctx.ChampionID) {
		seed := supportSeedFor(primary, allTags)
		return seed, "support"
	}

	// Solo lane / jungle: never keep pure enchanter/Atlas seed for ClassSupport.
	if primary == tags.ClassSupport && pos != "UTILITY" {
		seed, label := soloLaneSupportSeed(allTags)
		if pos == "JUNGLE" || ctx.HasSmite || HasJunglePet(ctx.Items) {
			seed = applyJungleStart(seed, firstLabelPart(label), allTags)
			return seed, label + "/jungle"
		}
		return seed, label
	}

	seed, label := seedByClass(primary, allTags)

	jungle := pos == "JUNGLE" || ctx.HasSmite || HasJunglePet(ctx.Items)
	if jungle {
		seed = applyJungleStart(seed, primary, allTags)
		label = label + "/jungle"
	}
	return seed, label
}

func firstLabelPart(label string) string {
	for i, c := range label {
		if c == '/' {
			return label[:i]
		}
	}
	return label
}

// soloLaneSupportSeed remaps support champions off UTILITY onto tank or AP lane paths.
func soloLaneSupportSeed(allTags []string) ([]Slot, string) {
	has := func(t string) bool {
		for _, x := range allTags {
			if x == t {
				return true
			}
		}
		return false
	}
	// Engage/tank supports (Braum/Leona/Thresh/Rakan…) → tank solo path.
	if has(tags.ClassTank) || has(tags.StyleEngage) || (has(tags.StylePeel) && has(tags.RangeMelee)) {
		return SeedTank(), "tank"
	}
	// Enchanter/mage supports on solo → AP poke/mid, not Moonstone.
	if has(tags.DamageAP) || has(tags.StyleHealer) || has(tags.StyleShield) || has(tags.StylePoke) {
		if has(tags.StylePoke) && !has(tags.StyleBurst) {
			return SeedAPPoke(), "mage"
		}
		return SeedAPMidBurst(), "mage"
	}
	return SeedTank(), "tank"
}

func isSupportCarryException(champID string) bool {
	switch champID {
	case "Senna", "Pyke":
		return true
	default:
		return false
	}
}

func supportSeedFor(primary string, allTags []string) []Slot {
	has := func(t string) bool {
		if primary == t {
			return true
		}
		for _, x := range allTags {
			if x == t {
				return true
			}
		}
		return false
	}
	// Mage poke supports (Xerath/Zyra/Brand/Vel'Koz) → Atlas + AP poke, not Moonstone.
	if primary == tags.ClassMage || (has(tags.DamageAP) && has(tags.StylePoke) && !has(tags.StyleHealer)) {
		return SeedMageSupport()
	}
	if has(tags.StyleHealer) || has(tags.StyleShield) {
		return SeedEnchanter()
	}
	if primary == tags.ClassMage || has(tags.DamageAP) {
		return SeedMageSupport()
	}
	return SeedSupportEngage()
}

func seedByClass(primary string, allTags []string) ([]Slot, string) {
	has := func(t string) bool {
		if primary == t {
			return true
		}
		for _, x := range allTags {
			if x == t {
				return true
			}
		}
		return false
	}

	switch primary {
	case tags.ClassMage:
		if has(tags.StylePoke) && !has(tags.StyleBurst) {
			return SeedAPPoke(), "mage"
		}
		return SeedAPMidBurst(), "mage"
	case tags.ClassAssassin:
		if has(tags.DamageAP) || has(tags.DamageHybrid) {
			return SeedAPAssassin(), "assassin/ap"
		}
		return SeedADAssassin(), "assassin"
	case tags.ClassMarksman:
		return SeedADC(), "marksman"
	case tags.ClassFighter:
		if has(tags.DamageAP) {
			return SeedAPBruiser(), "fighter/ap"
		}
		if has(tags.ExtraJuggernaut) {
			return SeedADJuggernaut(), "fighter"
		}
		return SeedADBruiser(), "fighter"
	case tags.ClassTank:
		// Poppy-like tank/fighters: Iceborn + damage cores, not Sunfire-only tank.
		if has(tags.ClassFighter) {
			return SeedTankFighter(), "fighter"
		}
		return SeedTank(), "tank"
	case tags.ClassSupport:
		if has(tags.StyleHealer) || has(tags.StyleShield) {
			return SeedEnchanter(), "support"
		}
		return SeedSupportEngage(), "support"
	default:
		return SeedAPMidBurst(), "mage"
	}
}

// HasJunglePet reports SR jungle starter pets.
func HasJunglePet(items []int) bool {
	for _, id := range items {
		if IsJunglePet(id) {
			return true
		}
	}
	return false
}

// IsJunglePet is true for Scorchclaw / Gustwalker / Mosstomper (and upgrades 1104–1107).
func IsJunglePet(id int) bool {
	return id >= 1101 && id <= 1107
}

// IsSupportQuestStart covers Atlas → Compass → Bounty / finished quest items.
func IsSupportQuestStart(id int) bool {
	switch id {
	case ItemWorldAtlas, 3866, 3867, 3869, 3870, 3871, 3876, 3877:
		return true
	default:
		return false
	}
}

// IsLaneStartItem is Doran's / Cull / Dark Seal / pets / Atlas line.
func IsLaneStartItem(id int) bool {
	switch id {
	case ItemDoransBlade, ItemDoransRing, ItemDoransShield, ItemDarkSeal, 1083:
		return true
	}
	return IsJunglePet(id) || IsSupportQuestStart(id)
}

func junglePetFor(primary string, allTags []string) (int, string) {
	ap := primary == tags.ClassMage || primary == tags.ClassSupport ||
		(primary == tags.ClassAssassin && (hasTag(allTags, tags.DamageAP) || hasTag(allTags, tags.DamageHybrid))) ||
		(primary == tags.ClassFighter && hasTag(allTags, tags.DamageAP))
	tank := primary == tags.ClassTank || hasTag(allTags, tags.ClassTank) || hasTag(allTags, tags.ExtraJuggernaut)
	switch {
	case tank:
		return ItemJungleMosstomper, "Mosstomper Seedling"
	case ap:
		return ItemJungleScorchclaw, "Scorchclaw Pup"
	default:
		return ItemJungleGustwalker, "Gustwalker Hatchling"
	}
}

func applyJungleStart(seed []Slot, primary string, allTags []string) []Slot {
	petID, petName := junglePetFor(primary, allTags)
	out := make([]Slot, 0, len(seed)+1)
	replaced := false
	for _, s := range seed {
		if s.Role == "start" {
			if !replaced {
				out = append(out, Slot{ItemID: petID, Name: petName, Priority: 100, Role: "start"})
				replaced = true
			}
			continue
		}
		out = append(out, s)
	}
	if !replaced {
		out = append([]Slot{{ItemID: petID, Name: petName, Priority: 100, Role: "start"}}, out...)
	}
	return out
}

// SeedAPMidBurst — Malignance path (reference from plan).
func SeedAPMidBurst() []Slot {
	return []Slot{
		{ItemID: ItemDoransRing, Name: "Doran's Ring", Priority: 100, Role: "start"},
		{ItemID: ItemLostChapter, Name: "Lost Chapter", Priority: 90, Role: "component"},
		{ItemID: ItemMalignance, Name: "Malignance", Priority: 80, Role: "core"},
		{ItemID: ItemSorcs, Name: "Sorcerer's Shoes", Priority: 72, Role: "boots"},
		{ItemID: ItemMercTreads, Name: "Mercury's Treads", Priority: 68, Role: "boots"},
		{ItemID: ItemBansheeVeil, Name: "Banshee's Veil", Priority: 60, Role: "defensive"},
		{ItemID: ItemZhonyas, Name: "Zhonya's Hourglass", Priority: 55, Role: "defensive"},
		{ItemID: ItemShadowflame, Name: "Shadowflame", Priority: 50, Role: "offensive"},
		{ItemID: ItemLiandrys, Name: "Liandry's Torment", Priority: 45, Role: "offensive"},
		{ItemID: ItemRabadons, Name: "Rabadon's Deathcap", Priority: 40, Role: "offensive"},
		{ItemID: ItemVoidStaff, Name: "Void Staff", Priority: 35, Role: "pen"},
	}
}

// SeedAPPoke — Blackfire / burn path.
func SeedAPPoke() []Slot {
	return []Slot{
		{ItemID: ItemDoransRing, Name: "Doran's Ring", Priority: 100, Role: "start"},
		{ItemID: ItemLostChapter, Name: "Lost Chapter", Priority: 90, Role: "component"},
		{ItemID: ItemBlackfire, Name: "Blackfire Torch", Priority: 80, Role: "core"},
		{ItemID: ItemSorcs, Name: "Sorcerer's Shoes", Priority: 70, Role: "boots"},
		{ItemID: ItemLiandrys, Name: "Liandry's Torment", Priority: 62, Role: "offensive"},
		{ItemID: ItemZhonyas, Name: "Zhonya's Hourglass", Priority: 55, Role: "defensive"},
		{ItemID: ItemBansheeVeil, Name: "Banshee's Veil", Priority: 52, Role: "defensive"},
		{ItemID: ItemRabadons, Name: "Rabadon's Deathcap", Priority: 45, Role: "offensive"},
		{ItemID: ItemVoidStaff, Name: "Void Staff", Priority: 40, Role: "pen"},
		{ItemID: ItemShadowflame, Name: "Shadowflame", Priority: 38, Role: "offensive"},
	}
}

func SeedAPAssassin() []Slot {
	// Energy assassins (Akali/Diana/Ekko…): no Lost Chapter / mana mythics.
	// Burst electrocute path ≈ Stormsurge → Sorcs → Zhonya → Shadowflame → Deathcap → Void|Banshee.
	// Core priorities stay above typical defensive live bumps so unfinished Stormsurge isn't skipped.
	return []Slot{
		{ItemID: ItemDarkSeal, Name: "Dark Seal", Priority: 105, Role: "start"},
		{ItemID: ItemStormsurge, Name: "Stormsurge", Priority: 100, Role: "core"},
		{ItemID: ItemSorcs, Name: "Sorcerer's Shoes", Priority: 86, Role: "boots"},
		{ItemID: ItemZhonyas, Name: "Zhonya's Hourglass", Priority: 74, Role: "defensive"},
		{ItemID: ItemShadowflame, Name: "Shadowflame", Priority: 66, Role: "offensive"},
		{ItemID: ItemRabadons, Name: "Rabadon's Deathcap", Priority: 54, Role: "offensive"},
		{ItemID: ItemVoidStaff, Name: "Void Staff", Priority: 46, Role: "pen"},
		{ItemID: ItemBansheeVeil, Name: "Banshee's Veil", Priority: 44, Role: "defensive"},
	}
}

func SeedADAssassin() []Slot {
	// One lethality core#1 (Youmuu); later pen/finisher slots fill the 4–6 item tail.
	return []Slot{
		{ItemID: ItemDoransBlade, Name: "Doran's Blade", Priority: 100, Role: "start"},
		{ItemID: ItemSerratedDirk, Name: "Serrated Dirk", Priority: 90, Role: "component"},
		{ItemID: ItemYoumuu, Name: "Youmuu's Ghostblade", Priority: 80, Role: "core"},
		{ItemID: ItemIonians, Name: "Ionian Boots of Lucidity", Priority: 62, Role: "boots"},
		{ItemID: ItemSteelcaps, Name: "Plated Steelcaps", Priority: 58, Role: "boots"},
		{ItemID: ItemMercTreads, Name: "Mercury's Treads", Priority: 56, Role: "boots"},
		{ItemID: ItemEdgeOfNight(), Name: "Edge of Night", Priority: 55, Role: "defensive"},
		{ItemID: ItemSeryldas, Name: "Serylda's Grudge", Priority: 50, Role: "pen"},
		{ItemID: ItemCollector, Name: "The Collector", Priority: 46, Role: "offensive"},
		{ItemID: ItemAxiomArc, Name: "Axiom Arc", Priority: 42, Role: "offensive"},
	}
}

func ItemEdgeOfNight() int { return 3814 }

func SeedADC() []Slot {
	return []Slot{
		{ItemID: ItemDoransBlade, Name: "Doran's Blade", Priority: 100, Role: "start"},
		{ItemID: ItemBerserkers, Name: "Berserker's Greaves", Priority: 85, Role: "boots"},
		{ItemID: ItemKraken, Name: "Kraken Slayer", Priority: 78, Role: "core"},
		{ItemID: ItemInfinityEdge, Name: "Infinity Edge", Priority: 70, Role: "offensive"},
		{ItemID: ItemRapidFirecannon, Name: "Rapid Firecannon", Priority: 64, Role: "offensive"},
		{ItemID: ItemCollector, Name: "The Collector", Priority: 60, Role: "offensive"},
		{ItemID: ItemLordDominiks, Name: "Lord Dominik's Regards", Priority: 56, Role: "offensive"},
		{ItemID: ItemBOTRK, Name: "Blade of The Ruined King", Priority: 52, Role: "offensive"},
		{ItemID: ItemSteelcaps, Name: "Plated Steelcaps", Priority: 50, Role: "boots"},
	}
}

func SeedADBruiser() []Slot {
	return []Slot{
		{ItemID: ItemDoransBlade, Name: "Doran's Blade", Priority: 100, Role: "start"},
		{ItemID: ItemTrinity, Name: "Trinity Force", Priority: 80, Role: "core"},
		{ItemID: ItemSunderedSky, Name: "Sundered Sky", Priority: 72, Role: "offensive"},
		{ItemID: ItemBlackCleaver, Name: "Black Cleaver", Priority: 65, Role: "offensive"},
		{ItemID: ItemSteelcaps, Name: "Plated Steelcaps", Priority: 60, Role: "boots"},
		{ItemID: ItemMercTreads, Name: "Mercury's Treads", Priority: 58, Role: "boots"},
		{ItemID: ItemForceOfNature, Name: "Force of Nature", Priority: 50, Role: "defensive"},
		{ItemID: ItemSpiritVisage, Name: "Spirit Visage", Priority: 48, Role: "defensive"},
	}
}

// SeedADJuggernaut — Illaoi/Darius/Morde-adjacent: Cleaver/Sundered/Sterak, not Trinity.
func SeedADJuggernaut() []Slot {
	return []Slot{
		{ItemID: ItemDoransBlade, Name: "Doran's Blade", Priority: 100, Role: "start"},
		{ItemID: ItemBlackCleaver, Name: "Black Cleaver", Priority: 88, Role: "core"},
		{ItemID: ItemSunderedSky, Name: "Sundered Sky", Priority: 82, Role: "offensive"},
		{ItemID: ItemSteraks, Name: "Sterak's Gage", Priority: 74, Role: "defensive"},
		{ItemID: ItemIceborn, Name: "Iceborn Gauntlet", Priority: 68, Role: "defensive"},
		{ItemID: ItemMercTreads, Name: "Mercury's Treads", Priority: 62, Role: "boots"},
		{ItemID: ItemSteelcaps, Name: "Plated Steelcaps", Priority: 58, Role: "boots"},
		{ItemID: ItemSpiritVisage, Name: "Spirit Visage", Priority: 52, Role: "defensive"},
		{ItemID: ItemForceOfNature, Name: "Force of Nature", Priority: 50, Role: "defensive"},
	}
}

func SeedAPBruiser() []Slot {
	return []Slot{
		{ItemID: ItemDoransRing, Name: "Doran's Ring", Priority: 100, Role: "start"},
		{ItemID: ItemRodOfAges, Name: "Rod of Ages", Priority: 80, Role: "core"},
		{ItemID: ItemRiftmaker(), Name: "Riftmaker", Priority: 72, Role: "offensive"},
		{ItemID: ItemZhonyas, Name: "Zhonya's Hourglass", Priority: 65, Role: "defensive"},
		{ItemID: ItemMercTreads, Name: "Mercury's Treads", Priority: 60, Role: "boots"},
	}
}

func ItemRiftmaker() int { return 4633 }

func SeedTank() []Slot {
	return []Slot{
		{ItemID: ItemDoransShield, Name: "Doran's Shield", Priority: 100, Role: "start"},
		{ItemID: ItemSteelcaps, Name: "Plated Steelcaps", Priority: 75, Role: "boots"},
		{ItemID: ItemMercTreads, Name: "Mercury's Treads", Priority: 72, Role: "boots"},
		{ItemID: 3068, Name: "Sunfire Aegis", Priority: 70, Role: "core"},
		{ItemID: ItemRanduins, Name: "Randuin's Omen", Priority: 60, Role: "defensive"},
		{ItemID: 3075, Name: "Thornmail", Priority: 55, Role: "defensive"},
		{ItemID: ItemForceOfNature, Name: "Force of Nature", Priority: 54, Role: "defensive"},
		{ItemID: ItemSpiritVisage, Name: "Spirit Visage", Priority: 52, Role: "defensive"},
	}
}

// SeedTankFighter — Poppy/top-jungle tank-bruisers: sheen core + AD legendaries,
// then the same resist options as SeedTank. Support Poppy stays on the engage line.
func SeedTankFighter() []Slot {
	return []Slot{
		{ItemID: ItemDoransShield, Name: "Doran's Shield", Priority: 100, Role: "start"},
		{ItemID: ItemIceborn, Name: "Iceborn Gauntlet", Priority: 82, Role: "core"},
		{ItemID: ItemSunderedSky, Name: "Sundered Sky", Priority: 74, Role: "offensive"},
		{ItemID: ItemBlackCleaver, Name: "Black Cleaver", Priority: 68, Role: "offensive"},
		{ItemID: ItemSteelcaps, Name: "Plated Steelcaps", Priority: 60, Role: "boots"},
		{ItemID: ItemMercTreads, Name: "Mercury's Treads", Priority: 58, Role: "boots"},
		{ItemID: ItemSteraks, Name: "Sterak's Gage", Priority: 54, Role: "defensive"},
		{ItemID: ItemForceOfNature, Name: "Force of Nature", Priority: 50, Role: "defensive"},
		{ItemID: 3075, Name: "Thornmail", Priority: 48, Role: "defensive"},
		{ItemID: ItemSpiritVisage, Name: "Spirit Visage", Priority: 46, Role: "defensive"},
	}
}

func SeedEnchanter() []Slot {
	return []Slot{
		{ItemID: ItemWorldAtlas, Name: "World Atlas", Priority: 100, Role: "start"},
		{ItemID: 6617, Name: "Moonstone Renewer", Priority: 80, Role: "core"},
		{ItemID: 3107, Name: "Redemption", Priority: 70, Role: "utility"},
		{ItemID: 3222, Name: "Mikael's Blessing", Priority: 62, Role: "utility"},
		{ItemID: ItemMorellonomicon, Name: "Morellonomicon", Priority: 50, Role: "utility"},
	}
}

// SeedMageSupport — Atlas start + AP poke core (Xerath/Zyra/Vel'Koz utility).
func SeedMageSupport() []Slot {
	return []Slot{
		{ItemID: ItemWorldAtlas, Name: "World Atlas", Priority: 100, Role: "start"},
		{ItemID: ItemLostChapter, Name: "Lost Chapter", Priority: 90, Role: "component"},
		{ItemID: ItemBlackfire, Name: "Blackfire Torch", Priority: 80, Role: "core"},
		{ItemID: ItemSorcs, Name: "Sorcerer's Shoes", Priority: 72, Role: "boots"},
		{ItemID: ItemLiandrys, Name: "Liandry's Torment", Priority: 65, Role: "offensive"},
		{ItemID: ItemZhonyas, Name: "Zhonya's Hourglass", Priority: 55, Role: "defensive"},
		{ItemID: ItemMorellonomicon, Name: "Morellonomicon", Priority: 48, Role: "utility"},
		{ItemID: ItemVoidStaff, Name: "Void Staff", Priority: 40, Role: "pen"},
	}
}

func SeedSupportEngage() []Slot {
	return []Slot{
		{ItemID: ItemWorldAtlas, Name: "World Atlas", Priority: 100, Role: "start"},
		{ItemID: 3190, Name: "Locket of the Iron Solari", Priority: 80, Role: "core"},
		{ItemID: 3109, Name: "Knight's Vow", Priority: 70, Role: "utility"},
		{ItemID: ItemMercTreads, Name: "Mercury's Treads", Priority: 65, Role: "boots"},
		{ItemID: 3075, Name: "Thornmail", Priority: 55, Role: "defensive"},
	}
}

// LethalityCoreIDs are mutually exclusive early lethality legendaries.
func LethalityCoreIDs() []int {
	return []int{ItemYoumuu, ItemOpportunity, ItemEclipse, ItemHubris}
}

// CollapseLethalityCores keeps the highest-priority lethality core; drops the rest.
func CollapseLethalityCores(build []Slot) []Slot {
	ids := map[int]struct{}{}
	for _, id := range LethalityCoreIDs() {
		ids[id] = struct{}{}
	}
	bestPri := -1e9
	bestID := 0
	for _, s := range build {
		if _, ok := ids[s.ItemID]; !ok {
			continue
		}
		if s.Priority > bestPri {
			bestPri = s.Priority
			bestID = s.ItemID
		}
	}
	if bestID == 0 {
		return build
	}
	out := make([]Slot, 0, len(build))
	for _, s := range build {
		if _, ok := ids[s.ItemID]; ok && s.ItemID != bestID {
			continue
		}
		if s.ItemID == bestID {
			s.Role = "core"
		}
		out = append(out, s)
	}
	return out
}
