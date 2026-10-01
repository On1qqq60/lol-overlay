package tags

// ItemInfo is the subset of Data Dragon fields needed for tagging.
type ItemInfo struct {
	ID     int
	NameEN string
	Tags   []string
	Stats  map[string]float64
	Gold   ItemGold
	Into   []int
	From   []int
	Depth  *int
}

// ItemGold mirrors purchasable/gold totals.
type ItemGold struct {
	Base        int
	Total       int
	Sell        int
	Purchasable bool
}

// IsLegendary approximates completed non-boot legendary items.
func IsLegendary(it ItemInfo) bool {
	if !it.Gold.Purchasable {
		return false
	}
	for _, t := range it.Tags {
		if t == "Boots" {
			return false
		}
		if t == "Trinket" || t == "Consumable" {
			return false
		}
	}
	if it.Gold.Total < 2000 {
		return false
	}
	if len(it.From) > 0 && len(it.Into) == 0 {
		return true
	}
	return it.Gold.Total >= 2500
}

func containsAny(s string, parts ...string) bool {
	for _, p := range parts {
		if len(p) > 0 && containsFold(s, p) {
			return true
		}
	}
	return false
}

func containsFold(s, sub string) bool {
	return len(sub) > 0 && indexFold(s, sub) >= 0
}

func indexFold(s, sub string) int {
	ls, lsub := len(s), len(sub)
	if lsub > ls {
		return -1
	}
	for i := 0; i+lsub <= ls; i++ {
		if equalFoldASCII(s[i:i+lsub], sub) {
			return i
		}
	}
	return -1
}

func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// ItemCatalog maps item id -> derived tags + metadata.
type ItemCatalog struct {
	ByID map[int]ItemInfo
	Tags map[int][]string
}

// TagsFor returns derived tags for item ids (0 skipped).
func (c *ItemCatalog) TagsFor(itemIDs []int) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, id := range itemIDs {
		if id == 0 {
			continue
		}
		for _, t := range c.Tags[id] {
			if _, ok := seen[t]; ok {
				continue
			}
			seen[t] = struct{}{}
			out = append(out, t)
		}
	}
	return out
}

// LegendaryCount counts legendary items in slots.
func (c *ItemCatalog) LegendaryCount(itemIDs []int) int {
	n := 0
	for _, id := range itemIDs {
		if id == 0 {
			continue
		}
		for _, t := range c.Tags[id] {
			if t == ItemLegendary {
				n++
				break
			}
		}
	}
	return n
}

// TotalLegendary among many players.
func (c *ItemCatalog) TotalLegendary(itemLists ...[]int) int {
	n := 0
	for _, items := range itemLists {
		n += c.LegendaryCount(items)
	}
	return n
}

// CountLegendaryWith reports how many legendaries carry any of the given tags.
func (c *ItemCatalog) CountLegendaryWith(itemIDs []int, want ...string) int {
	wantSet := map[string]struct{}{}
	for _, w := range want {
		wantSet[w] = struct{}{}
	}
	n := 0
	for _, id := range itemIDs {
		if id == 0 {
			continue
		}
		tagList := c.Tags[id]
		isLeg := false
		match := false
		for _, t := range tagList {
			if t == ItemLegendary {
				isLeg = true
			}
			if _, ok := wantSet[t]; ok {
				match = true
			}
		}
		if isLeg && match {
			n++
		}
	}
	return n
}

// ScoreTankiness 0..1 from HP/armor/MR items.
func (c *ItemCatalog) ScoreTankiness(itemIDs []int) float64 {
	var hp, armor, mr float64
	for _, id := range itemIDs {
		it, ok := c.ByID[id]
		if !ok {
			continue
		}
		hp += it.Stats["FlatHPPoolMod"]
		armor += it.Stats["FlatArmorMod"]
		mr += it.Stats["FlatSpellBlockMod"]
	}
	s := clamp01(hp/3000) * 0.45
	s += clamp01(armor/150) * 0.3
	s += clamp01(mr/150) * 0.25
	return s
}

// ScoreADDamage 0..1 from AD/lethality legendaries + stats.
func (c *ItemCatalog) ScoreADDamage(itemIDs []int) float64 {
	var ad, leth float64
	legs := 0
	for _, id := range itemIDs {
		it, ok := c.ByID[id]
		if !ok {
			continue
		}
		ad += it.Stats["FlatPhysicalDamageMod"]
		leth += it.Stats["FlatArmorPenetrationMod"]
		for _, t := range c.Tags[id] {
			if t == ItemLegendary && (hasTag(c.Tags[id], ItemADDamage) || hasTag(c.Tags[id], ItemLethality)) {
				legs++
				break
			}
		}
	}
	s := clamp01(ad/200)*0.4 + clamp01(leth/40)*0.3 + clamp01(float64(legs)/3)*0.3
	return s
}

// ScoreCrit 0..1 from crit items.
func (c *ItemCatalog) ScoreCrit(itemIDs []int) float64 {
	n := 0
	for _, id := range itemIDs {
		if hasTag(c.Tags[id], ItemCrit) {
			n++
		}
	}
	return clamp01(float64(n) / 3)
}

// ScoreMR 0..1 from MR on items (for Void Staff priority).
func (c *ItemCatalog) ScoreMR(itemIDs []int) float64 {
	var mr float64
	for _, id := range itemIDs {
		it, ok := c.ByID[id]
		if !ok {
			continue
		}
		mr += it.Stats["FlatSpellBlockMod"]
	}
	return clamp01(mr / 120)
}

// GoldCost returns total gold for an item id, or 0.
func (c *ItemCatalog) GoldCost(id int) int {
	if it, ok := c.ByID[id]; ok {
		return it.Gold.Total
	}
	return 0
}

// NameEN returns english name or empty.
func (c *ItemCatalog) NameEN(id int) string {
	if it, ok := c.ByID[id]; ok {
		return it.NameEN
	}
	return ""
}

// HasHealCut reports if any item provides grievous wounds.
func (c *ItemCatalog) HasHealCut(itemIDs []int) bool {
	for _, id := range itemIDs {
		if hasTag(c.Tags[id], ItemHealCut) {
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

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
