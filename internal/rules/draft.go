package rules

import "lol-build-overlay/internal/tags"

// Well-known SR item IDs.
const (
	ItemDoransRing     = 1056
	ItemDoransBlade    = 1055
	ItemDoransShield   = 1054
	ItemLostChapter    = 3802
	ItemMalignance     = 3118
	ItemLudens         = 6655
	ItemBlackfire      = 2503
	ItemRodOfAges      = 6657
	ItemMercTreads     = 3111
	ItemSorcs          = 3020
	ItemSteelcaps      = 3047
	ItemBerserkers     = 3006
	ItemBansheeVeil    = 3102
	ItemZhonyas        = 3157
	ItemShadowflame    = 4645
	ItemStormsurge     = 4646
	ItemLiandrys       = 6653
	ItemRabadons       = 3089
	ItemVoidStaff      = 3135
	ItemMorellonomicon = 3165
	ItemOblivionOrb    = 3916
	ItemSeekers        = 2420
	ItemNeedlessly     = 1058
	ItemBlastingWand   = 1026
	ItemNullMagic      = 1033
	ItemFiendishCodex  = 3108
	ItemHauntingGuise  = 3147
	ItemFatedAshes     = 2508
	ItemSerratedDirk   = 3134
	ItemYoumuu         = 3142
	ItemOpportunity    = 6701
	ItemEclipse        = 6692
	ItemInfinityEdge   = 3031
	ItemKraken         = 6672
	ItemBOTRK          = 3153
	ItemTrinity        = 3078
	ItemBlackCleaver   = 3071
	ItemSunderedSky    = 6610
	ItemRanduins       = 3143
	ItemDarkSeal       = 1082
	ItemHextechAlt     = 3145
	ItemAetherWisp     = 3113
	ItemExecutioners   = 3123
	ItemSteraks        = 3053
	ItemIceborn        = 6662
	ItemSpiritVisage   = 3065
)

// Slot is one step in a recommended build path.
type Slot struct {
	ItemID   int
	Name     string
	Priority float64
	Role     string // start, component, core, boots, defensive, offensive, pen, utility
}

// DraftSignals are weighted 0..1 team aggregates from champion tags (pre-items).
type DraftSignals struct {
	Tank    float64
	Healer  float64
	Shield  float64
	APBurst float64
	AD      float64
	Dive    float64
	CCHard  float64
}

// DraftFromProfiles builds equal-weight draft signals from role-resolved profiles.
func DraftFromProfiles(profiles []tags.ChampionProfile) DraftSignals {
	n := len(profiles)
	if n == 0 {
		return DraftSignals{}
	}
	w := 1.0 / float64(n)
	var d DraftSignals
	for _, p := range profiles {
		if p.Primary == tags.ClassTank || p.Has(tags.ClassTank) || p.Has(tags.ExtraJuggernaut) {
			d.Tank += w
		}
		if p.Has(tags.StyleHealer) {
			d.Healer += w
		}
		if p.Has(tags.StyleShield) {
			d.Shield += w
		}
		if p.Has(tags.DamageAP) && p.Has(tags.StyleBurst) {
			d.APBurst += w
		}
		if p.Has(tags.DamageAD) || p.Primary == tags.ClassMarksman || p.Primary == tags.ClassAssassin {
			d.AD += w * 0.7
		}
		if p.Has(tags.ThreatDive) {
			d.Dive += w
		}
		if p.Has(tags.ThreatCCHard) {
			d.CCHard += w
		}
	}
	return d
}

// AdjustDraft applies weighted draft bumps scaled by draftScale (0..1).
// Only bumps items present in the seed; reasons fire only when a bump applied.
func AdjustDraft(seed []Slot, d DraftSignals, draftScale float64, playerPrimary string, allTags []string) ([]Slot, []string) {
	out := cloneSlots(seed)
	var reasons []string
	s := draftScale
	if s <= 0.01 {
		return out, reasons
	}

	ap := usesAPItems(playerPrimary, allTags)

	// Healers → grievous. Do NOT treat shield-only (Morgana) as heal.
	if d.Healer > 0.12 {
		if ap {
			prio := 50 + 25*d.Healer*s
			out = upsert(out, Slot{ItemID: ItemMorellonomicon, Name: "Morellonomicon", Priority: prio, Role: "utility"})
			reasons = append(reasons, "draft: healer → Morello")
		} else {
			prio := 50 + 25*d.Healer*s
			out = upsert(out, Slot{ItemID: ItemExecutioners, Name: "Executioner's Calling", Priority: prio, Role: "utility"})
			reasons = append(reasons, "draft: healer → Executioner's")
		}
	}

	if ap {
		if d.APBurst > 0.15 {
			if bump(&out, ItemBansheeVeil, 20*d.APBurst*s) {
				reasons = append(reasons, "draft: ap burst weight → Banshee ↑")
			}
		}
		if d.Tank > 0.15 {
			if bump(&out, ItemLiandrys, 28*d.Tank*s) {
				reasons = append(reasons, "draft: tank weight → Liandry ↑")
			}
		}
		if d.AD > 0.25 || d.Dive > 0.2 {
			if bump(&out, ItemZhonyas, 18*(d.AD+d.Dive)*s) {
				reasons = append(reasons, "draft: ad/dive weight → Zhonya ↑")
			}
		}
	} else {
		if d.Tank > 0.15 {
			if bump(&out, ItemBlackCleaver, 22*d.Tank*s) {
				reasons = append(reasons, "draft: tank weight → Cleaver ↑")
			}
		}
		if d.AD > 0.25 || d.Dive > 0.2 {
			if bump(&out, ItemSteelcaps, 16*(d.AD+d.Dive)*s) {
				reasons = append(reasons, "draft: ad/dive weight → Steelcaps ↑")
			}
		}
		if d.APBurst > 0.15 {
			if bump(&out, ItemSpiritVisage, 14*d.APBurst*s) {
				reasons = append(reasons, "draft: ap burst → Spirit Visage ↑")
			}
			if bump(&out, ItemForceOfNature, 16*d.APBurst*s) {
				reasons = append(reasons, "draft: ap burst → Force of Nature ↑")
			}
			if bump(&out, ItemMercTreads, 10*d.APBurst*s) {
				reasons = append(reasons, "draft: ap burst → Mercs ↑")
			}
			if bump(&out, ItemSteelcaps, -10*d.APBurst*s) {
				reasons = append(reasons, "draft: ap burst → Steelcaps ↓")
			}
			if bump(&out, ItemRanduins, -8*d.APBurst*s) {
				reasons = append(reasons, "draft: ap burst → Randuin ↓")
			}
		}
		if d.Tank < 0.12 {
			if bump(&out, ItemBOTRK, -18*s) {
				reasons = append(reasons, "draft: low tank → BotRK ↓")
			}
			if bump(&out, ItemRapidFirecannon, 10*s) {
				reasons = append(reasons, "draft: low tank → RFC ↑")
			}
			if bump(&out, ItemCollector, 8*s) {
				reasons = append(reasons, "draft: low tank → Collector ↑")
			}
		} else if d.Tank > 0.2 {
			if bump(&out, ItemBOTRK, 12*d.Tank*s) {
				reasons = append(reasons, "draft: tank → BotRK ↑")
			}
		}
	}

	if d.CCHard > 0.2 {
		if bump(&out, ItemMercTreads, 12*d.CCHard*s) {
			reasons = append(reasons, "draft: cc_hard → Mercs ↑")
		}
	}
	return out, reasons
}

func cloneSlots(in []Slot) []Slot {
	out := make([]Slot, len(in))
	copy(out, in)
	return out
}

// bump adds delta to an existing slot; returns false if item is not in the seed.
func bump(slots *[]Slot, id int, delta float64) bool {
	for i := range *slots {
		if (*slots)[i].ItemID == id {
			(*slots)[i].Priority += delta
			return true
		}
	}
	return false
}

func upsert(slots []Slot, s Slot) []Slot {
	for i := range slots {
		if slots[i].ItemID == s.ItemID {
			if s.Priority > slots[i].Priority {
				slots[i] = s
			}
			return slots
		}
	}
	return append(slots, s)
}
