package rules

// Extra SR item IDs used by champion×role families / commit / emit.
const (
	ItemArchangel            = 3003
	ItemArdent               = 3504
	ItemBamis                = 6660
	ItemBloodsong            = 3877
	ItemBountyOfWorlds       = 3867
	ItemCaulfield            = 3133
	ItemCelestialOpposition  = 3869
	ItemChempunk             = 6609
	ItemCosmicDrive          = 4629
	ItemDawncore             = 6621
	ItemDeadMans             = 3742
	ItemDreamMaker           = 3870
	ItemEssenceReaver        = 3508
	ItemForbiddenIdol        = 3114
	ItemFrozenHeart          = 3110
	ItemGiantsBelt           = 1011
	ItemGuinsoos             = 3124
	ItemHealthPotion         = 2003
	ItemHeartsteel           = 3084
	ItemHexplate             = 3073
	ItemHollowRadiance       = 6664
	ItemHorizon              = 4628
	ItemJakSho               = 6665
	ItemKindlegem            = 3067
	ItemKnightsVow           = 3109
	ItemLocket               = 3190
	ItemManamune             = 3004
	ItemMandate              = 4005
	ItemMaw                  = 3156
	ItemMejais               = 3041
	ItemMikaels              = 3222
	ItemMoonstone            = 6617
	ItemNavori               = 6675
	ItemPhantomDancer        = 3046
	ItemProfane              = 6698
	ItemRavenousHydra        = 3074
	ItemRedemption           = 3107
	ItemRocketbelt           = 3152
	ItemRunaans              = 3085
	ItemRunicCompass         = 3866
	ItemRylais               = 3116
	ItemSeraphs              = 3040
	ItemShieldbow            = 6673
	ItemShojin               = 3161
	ItemShurelyas            = 2065
	ItemSolsticeSleigh       = 3876
	ItemStaffOfFlowing       = 6616
	ItemStatikk              = 3087
	ItemStormrazor           = 3095
	ItemStridebreaker        = 6631
	ItemSunfire              = 3068
	ItemSwifties             = 3009
	ItemTear                 = 3070
	ItemTerminus             = 3302
	ItemThornmail            = 3075
	ItemTitanic              = 3748
	ItemUmbral               = 3179
	ItemUnendingDespair      = 2502
	ItemVoltaic              = 6699
	ItemWarmogs              = 3083
	ItemRiftmaker            = 4633
	ItemZazzak               = 3871
)

func IsBootsID(id int) bool {
	switch id {
	case 1001, ItemSorcs, ItemMercTreads, ItemSteelcaps, ItemBerserkers,
		ItemIonians, ItemSwifties, 3117:
		return true
	default:
		return false
	}
}

func IsFinishedBootsID(id int) bool {
	return IsBootsID(id) && id != 1001
}
