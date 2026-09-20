package tags

// Champion tag layers (primary class + refinements).
const (
	ClassMage     = "mage"
	ClassAssassin = "assassin"
	ClassMarksman = "marksman"
	ClassFighter  = "fighter"
	ClassTank     = "tank"
	ClassSupport  = "support"

	DamageAP     = "ap"
	DamageAD     = "ad"
	DamageHybrid = "hybrid"
	DamageTrue   = "true"

	RangeRanged = "ranged"
	RangeMelee  = "melee"

	StyleBurst  = "burst"
	StylePoke   = "poke"
	StyleDPS    = "dps"
	StyleEngage = "engage"
	StylePeel   = "peel"
	StyleHealer = "healer"
	StyleShield = "shield"

	ThreatDive      = "dive"
	ThreatCCHard    = "cc_hard"
	ThreatMobility  = "mobility"
	ThreatSustain   = "sustain"
	ThreatSplitpush = "splitpush"

	ExtraJuggernaut = "juggernaut"
	ExtraAntiDash  = "anti_dash"
	ExtraPick      = "pick"
	ExtraAllIn     = "all_in"
)

// Item-derived tag groups used by the engine.
const (
	ItemLethality = "lethality"
	ItemADDamage  = "ad_damage"
	ItemAPDamage  = "ap_damage"
	ItemTankHP    = "tank_hp"
	ItemArmor     = "armor"
	ItemMR        = "mr"
	ItemHealCut   = "heal_cut"
	ItemBoots     = "boots"
	ItemLegendary = "legendary"
	ItemCrit      = "crit"
	ItemAS        = "attack_speed"
	ItemOnHit     = "on_hit"
	ItemShield    = "shield"
)
