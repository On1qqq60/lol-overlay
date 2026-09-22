package rules

import "lol-build-overlay/internal/tags"

// Swap is one identity-path item replaced by live pressure.
type Swap struct {
	From int    `json:"from"`
	To   int    `json:"to"`
	Why  string `json:"why"`
}

// FillWhy writes a player-facing reason on every slot that still has none.
func FillWhy(slots []Slot, items *tags.ItemCatalog) {
	for i := range slots {
		if slots[i].Why == "" {
			slots[i].Why = ItemWhy(slots[i], items)
		}
	}
}

func setWhy(slots []Slot, id int, why string) {
	if why == "" {
		return
	}
	for i := range slots {
		if slots[i].ItemID == id {
			slots[i].Why = why
			return
		}
	}
}

// ItemWhy is the default "зачем покупать" text for any shop id.
func ItemWhy(s Slot, items *tags.ItemCatalog) string {
	if s.Why != "" {
		return s.Why
	}
	if t, ok := itemWhy[s.ItemID]; ok {
		return t
	}
	if items != nil {
		if w := whyFromTags(s.ItemID, items); w != "" {
			return w
		}
	}
	return roleWhy(s.Role, s.Name)
}

func whyFromTags(id int, items *tags.ItemCatalog) string {
	if items == nil {
		return ""
	}
	list := items.Tags[id]
	has := func(t string) bool {
		for _, x := range list {
			if x == t {
				return true
			}
		}
		return false
	}
	switch {
	case has(tags.ItemHealCut):
		return "Режет лечение врагов"
	case has(tags.ItemBoots):
		return "Сапоги — скорость и защита под матчап"
	case has(tags.ItemLethality):
		return "Леталити: быстрее снимаем хрупких"
	case has(tags.ItemOnHit):
		return "Урон с автоатак — чем чаще бьём, тем сильнее"
	case has(tags.ItemCrit):
		return "Крит: взрывной автоатак-урон"
	case has(tags.ItemShield):
		return "Щиты и поддержка союзников"
	case has(tags.ItemMR) && has(tags.ItemTankHP):
		return "HP и магическое сопротивление"
	case has(tags.ItemArmor) && has(tags.ItemTankHP):
		return "HP и броня против AD"
	case has(tags.ItemMR):
		return "Магическое сопротивление"
	case has(tags.ItemArmor):
		return "Броня против физических ударов"
	case has(tags.ItemAPDamage):
		return "Магический урон"
	case has(tags.ItemADDamage):
		return "Физический урон"
	case has(tags.ItemTankHP):
		return "Запас здоровья"
	case has(tags.ItemAS):
		return "Скорость атаки"
	default:
		return ""
	}
}

func roleWhy(role, name string) string {
	switch role {
	case "start":
		return "Стартовый предмет в линию"
	case "component":
		return "Компонент ключевого предмета"
	case "core":
		return "Ключевой предмет этой сборки"
	case "boots":
		return "Сапоги под матчап"
	case "defensive":
		return "Защита от угрозы в этом составе"
	case "offensive":
		return "Урон"
	case "pen":
		return "Пробивание защит врагов"
	case "utility":
		return "Утилита под вражеский состав"
	case "consumable":
		return "Лечимся в линии"
	default:
		if name != "" {
			return name
		}
		return "Предмет сборки"
	}
}

// DiffSwaps compares the identity 6-path with the live path.
func DiffSwaps(before, after []Slot, items *tags.ItemCatalog) []Swap {
	lost, gained := idSet(pathIDs(before)), idSet(pathIDs(after))
	var lostIDs, gainIDs []int
	for id := range lost {
		if _, ok := gained[id]; !ok {
			lostIDs = append(lostIDs, id)
		}
	}
	for id := range gained {
		if _, ok := lost[id]; !ok {
			gainIDs = append(gainIDs, id)
		}
	}
	usedL := map[int]bool{}
	usedG := map[int]bool{}
	var out []Swap

	pair := func(from, to int) {
		if from == 0 || to == 0 || from == to || usedL[from] || usedG[to] {
			return
		}
		usedL[from] = true
		usedG[to] = true
		out = append(out, Swap{From: from, To: to, Why: swapWhy(from, to, after, items)})
	}

	for _, g := range exclusiveWhyGroups {
		var from, to int
		for _, id := range g {
			if _, ok := lost[id]; ok && !usedL[id] {
				from = id
			}
			if _, ok := gained[id]; ok && !usedG[id] {
				to = id
			}
		}
		pair(from, to)
	}

	roleOf := func(id int, slots []Slot) string {
		for _, s := range slots {
			if s.ItemID == id {
				return s.Role
			}
		}
		return ""
	}
	for _, to := range gainIDs {
		if usedG[to] {
			continue
		}
		want := roleOf(to, after)
		for _, from := range lostIDs {
			if usedL[from] {
				continue
			}
			if roleOf(from, before) == want && want != "" {
				pair(from, to)
				break
			}
		}
	}
	for _, to := range gainIDs {
		if usedG[to] {
			continue
		}
		from := 0
		for _, id := range lostIDs {
			if !usedL[id] {
				from = id
				break
			}
		}
		if from != 0 {
			pair(from, to)
			continue
		}
		out = append(out, Swap{To: to, Why: ItemWhy(slotOf(after, to), items)})
	}
	return out
}

func slotOf(slots []Slot, id int) Slot {
	for _, s := range slots {
		if s.ItemID == id {
			return s
		}
	}
	return Slot{ItemID: id}
}

func swapWhy(from, to int, after []Slot, items *tags.ItemCatalog) string {
	toWhy := ItemWhy(slotOf(after, to), items)
	fromName := ""
	if items != nil {
		fromName = items.NameEN(from)
	}
	if fromName == "" {
		fromName = slotOf(after, from).Name
	}
	if fromName == "" {
		fromName = "предыдущего предмета"
	}
	if toWhy == "" {
		return "Заменили " + fromName + " под этот состав"
	}
	return toWhy + " — вместо " + fromName
}

func pathIDs(slots []Slot) []int {
	var ids []int
	for _, s := range slots {
		switch s.Role {
		case "start", "component", "consumable":
			continue
		}
		if s.ItemID > 0 {
			ids = append(ids, s.ItemID)
		}
	}
	return ids
}

func idSet(ids []int) map[int]struct{} {
	m := map[int]struct{}{}
	for _, id := range ids {
		m[id] = struct{}{}
	}
	return m
}

var exclusiveWhyGroups = [][]int{
	{ItemMercTreads, ItemSteelcaps, ItemSorcs, ItemBerserkers, ItemIonians, ItemSwifties, 3117, 1001, 2422},
	{ItemZhonyas, ItemBansheeVeil},
	{ItemVoidStaff, ItemLordDominiks, ItemMortalReminder, ItemSeryldas, 3137},
	{ItemMalignance, ItemBlackfire, ItemLudens, ItemRodOfAges},
	{ItemMorellonomicon, ItemExecutioners, ItemChempunk, ItemMortalReminder, ItemOblivionOrb, ItemThornmail},
	{ItemThornmail, ItemRanduins},
	{ItemForceOfNature, ItemSpiritVisage, ItemMaw, 2504},
}

// Player-facing defaults. Anything missing falls back to tags / plaintext_ru / role.
var itemWhy = map[int]string{
	ItemDoransRing:     "Стартовое кольцо: мана, урон и восстановление в линии",
	ItemDoransBlade:    "Стартовый клинок: AD и вампиризм в линии",
	ItemDoransShield:   "Стартовый щит: HP и реген, если линия тяжёлая",
	ItemDarkSeal:       "Печать: стопки AP за киллы, дёшево на старте",
	ItemMejais:         "Межай: разгоняем AP со стопок, если уже впереди",
	1083:               "Уничтожитель: вампиризм и урон по миньонам",
	ItemWorldAtlas:     "Атлас: квесты саппорта и золото с линии",
	3866:               "Компас: следующий шаг квестового старта саппорта",
	3867:               "Завершённый квест саппорта — золото и статы",
	ItemHealthPotion:   "Зелье: переживать харасс и трейд в линии",
	2031:               "Пополняемое зелье: лечение на всю раннюю игру",
	2033:               "Вредоносное зелье: урон и лечение за счёт HP",
	ItemJungleScorchclaw: "Питомец: AP-урон с кэмпов и в файтах",
	ItemJungleGustwalker: "Питомец: скорость по карте после кэмпов",
	ItemJungleMosstomper: "Питомец: танк-щит, если нужно выживать",
	1105:               "Питомец травоящера — танк-щит в джангле",
	1106:               "Питомец ветролиса — скорость по карте",
	1107:               "Питомец огневолка — урон с кэмпов",

	ItemLostChapter:    "Глава: мана и перезарядка — собираем кор",
	ItemNeedlessly:     "Большой жезл: сырой AP в кор или второй предмет",
	ItemBlastingWand:   "Разящий жезл: AP-компонент кора или пенетрейшена",
	ItemAmplifyingTome: "Фолиант: дешёвый AP, пока копим на кор",
	ItemFiendishCodex:  "Кодекс: AP и перезарядка в кор",
	ItemHauntingGuise:  "Маска: HP и магическое пробивание",
	ItemFatedAshes:     "Пепел: поджог — в Liandry или Blackfire",
	ItemSeekers:        "Наручи: броня и AP, собираем Zhonya",
	ItemNullMagic:      "Мантия: дешёвый MR, пока копим Banshee",
	ItemOblivionOrb:    "Сфера: ранний грив, пока нет Morello",
	ItemSerratedDirk:   "Кортик: леталити в Youmuu / Eclipse",
	ItemHextechAlt:     "Альтернатор: AP и магпен в Stormsurge",
	ItemAetherWisp:     "Огонёк: AP и скорость — в Lich Bane / Shurelya",
	ItemKindlegem:      "Самоцвет: HP и перезарядка",
	ItemGiantsBelt:     "Пояс: HP в танк-кор или Rylai",
	ItemTear:           "Слеза: копим ману в Manamune / Archangel",
	3803:               "Катализатор: HP и мана в Rod of Ages",
	3057:               "Кинжал: спелл + автоатака в ER / Trinity / Iceborn",
	3044:               "Фазаль: AD и HP в Trinity / Stridebreaker",
	3051:               "Hearthbound: скорость атаки в Trinity",
	3133:               "Caulfield: AD и перезарядка в Cleaver / Essence",
	1038:               "B.F. Sword: сырой AD в IE / Essence / Yun Tal",
	1037:               "Кирка: AD-компонент",
	1036:               "Длинный меч: дешёвый AD",
	1028:               "Рубин: дешёвое HP",
	1029:               "Ткань: дешёвая броня",
	1042:               "Кинжал: скорость атаки",
	1043:               "Изогнутый лук: скорость атаки и онхит",
	1053:               "Вампирский скипетр: вампиризм в BotRK / BT",
	6670:               "Полуденный колчан: AD и скорость атаки в Kraken",
	2015:               "Кирхeis: заряд в Stormrazor / RFC",
	3077:               "Тиамат: клеав в Hydra / Stridebreaker",
	1001:               "Базовые ботинки, пока нет готовых сапог",

	ItemMalignance:     "Кор: ульт сидит на перезарядке и магическом пробивании",
	ItemBlackfire:      "Кор: поджог и перезарядка — давим линию и танков",
	ItemLudens:         "Кор: взрывной маг. урон по хрупким",
	ItemRodOfAges:      "Кор: разгон HP/маны/AP — скалимся в долгую",
	ItemNashors:        "Кор: скорость атаки и AP на автоатаках",
	ItemRiftmaker:      "Кор: омнивамп и истинный урон в длинном файге",
	ItemCosmicDrive:    "Скорость и перезарядка после спеллов",
	ItemStormsurge:     "Кор: взрыв с электрокатом, когда пробиваем squishy",
	ItemLichBane:       "Спелл + автоатака: взорвать после умения",
	ItemArchangel:      "Слеза в щит и AP — мана на всю игру",
	ItemSeraphs:        "Готовая слеза: щит от маны и AP",
	ItemManamune:       "Слеза в AD — мана и урон для спелл-файта",

	ItemShadowflame:    "Магпен по щитам и хрупким — второй предмет после кора",
	ItemRabadons:       "Смертная шапка: множим весь накопленный AP",
	ItemLiandrys:       "Поджог и %HP — нужен, когда враги жирные",
	ItemVoidStaff:      "Магическое пробивание, когда враги собрали MR",
	ItemZhonyas:        "Станза: пережить дайв, AD-ассасина или ключевой ульт",
	ItemBansheeVeil:    "Завеса: съесть один AP-спелл или ключевой контроль",
	ItemMorellonomicon: "Грив: режем лечение врагов",
	ItemRylais:         "Замедление с урона — клеймим цели для команды",

	ItemYoumuu:         "Кор: леталити и скорость — находим и срезаем цель",
	ItemOpportunity:    "Леталити после убийства — добиваем файты",
	ItemEclipse:        "Кор: щит и %HP по одиночной цели",
	ItemUmbral:         "Контроль вардов + леталити",
	ItemVoltaic:        "Электрокют с автоатак после спелла",
	ItemProfane:        "Гидра-клеав с леталити",
	3814:               "Edge of Night: завеса против одного скилла",
	ItemAxiomArc:       "Возврат ульта за такдаун",

	ItemInfinityEdge:   "Крит-множитель — ядро ADC после первого предмета",
	ItemKraken:         "Онхит и истинный урон по танкам / в дуэли",
	ItemBOTRK:          "%HP с автоатак — против жирных и дуэлянтов",
	ItemGuinsoos:       "Разгон атаки и онхит — для гибридных ADC",
	ItemYunTal:         "Крит и кровотечение — кор на ADC",
	ItemStormrazor:     "Заряд движения и крит — кор Zeri/Yasuo-подобных",
	ItemRapidFirecannon: "Дальность автоатаки с заряда — безопасный урон",
	ItemCollector:      "Леталити и казнь — добиваем хрупких",
	ItemEssenceReaver:  "Мана и крит со спеллов — кор спелл-ADC",
	ItemNavori:         "Сброс перезарядки с критов",
	ItemRunaans:        "Молнии в файтах — урон по нескольким",
	ItemPhantomDancer:  "Скорость атаки и призрак сквозь крипов",
	ItemShieldbow:      "Щит и вамп, когда нас фокусят",
	ItemBloodthirster:  "Вампиризм и щит с автоатак",
	ItemLordDominiks:   "Бронепен: враги набрали HP и броню",
	ItemMortalReminder: "Бронепен + грив, если ещё и хилятся",
	ItemTerminus:       "Онхит, который режет и броню, и MR",
	ItemWitsEnd:        "MR и онхит — против AP в дуэли",

	ItemTrinity:        "Кор: спелл + автоатака, скорость и HP",
	ItemSunderedSky:    "Кор: крит-хил с автоатаки после умения",
	ItemBlackCleaver:   "Режем броню стаками — против танков",
	ItemIceborn:        "Замедление и броня после умения",
	ItemSteraks:        "Щит, когда нас пытаются взорвать",
	ItemDeathsDance:    "Откладываем урон — выживаем в файтах",
	ItemMaw:            "Щит от магии — против AP-берста",
	ItemGuardianAngel:  "Второе дыхание, если по тебе фокус",
	ItemShojin:         "Перезарядка и AD для спелл-файта",
	ItemStridebreaker:  "Рывок и замедление — догоняем",
	ItemRavenousHydra:  "Клеав и вамп — волна и файты",
	ItemTitanic:        "HP-клеав: чем толще, тем больнее",
	ItemHexplate:       "Скорость атаки после ульта",

	ItemSunfire:        "Кор танка: аура огня и броня/MR",
	ItemHeartsteel:     "Кор: стопки HP с автоатак по чемпионам",
	ItemWarmogs:        "Реген с большого HP — неубиваемость в сайдлайне",
	ItemFrozenHeart:    "Броня и режем скорость атаки врагов",
	ItemRanduins:       "Антикрит и замедление вокруг нас",
	ItemThornmail:      "Шипы и грив, если нас бьют автоатаками",
	ItemDeadMans:       "Скорость из кустов и удар с разгона",
	ItemForceOfNature:  "MR и скорость — против AP-берста",
	ItemSpiritVisage:   "MR и усиливаем своё лечение",
	ItemKnightsVow:     "Переводим урон с керри на себя",

	ItemMoonstone:      "Кор энчантера: усиливает хилы и щиты",
	ItemLocket:         "Кор энгейджа: командный щит в файтах",
	ItemShurelyas:      "Скорость команде на вход или отход",
	ItemMandate:        "Помечаем цель — команда бьёт больнее",
	ItemRedemption:     "Лечение по области в файтах",
	ItemMikaels:        "Снимаем контроль с союзника",

	ItemSorcs:          "Магпен на сапогах — больше урона со спеллов",
	ItemMercTreads:     "MR и сокращение контроля — против AP/CC",
	ItemSteelcaps:      "Броня и меньше урона с автоатак",
	ItemBerserkers:     "Скорость атаки на сапогах",
	ItemIonians:        "Перезарядка на сапогах — чаще умения и саммоны",
	ItemSwifties:       "Чистое перемещение и меньше замедлений",
	3117:               "Сапоги мобильности — ходим по карте",

	ItemExecutioners:   "Ранний грив, пока нет полного Morello/Mortal",
	ItemChempunk:       "Грив и AD — режем хил в AD-сборке",
	ItemSeryldas:       "Бронепен и замедление со спеллов",
}
