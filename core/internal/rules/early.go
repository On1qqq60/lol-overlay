package rules

const EarlyPhaseSeconds = 300

func IsEarlyPhase(gameTime float64) bool {
	return gameTime < EarlyPhaseSeconds
}

func ownsID(owned map[int]struct{}, id int) bool {
	if owned == nil {
		return false
	}
	_, ok := owned[id]
	return ok
}

func ownsAnyBootsIDs(owned map[int]struct{}) bool {
	if owned == nil {
		return false
	}
	for id := range owned {
		if IsBootsID(id) {
			return true
		}
	}
	return false
}

// EarlyRow is the dynamic first-5-minutes strip: starter, potions, boots, core component.
func EarlyRow(build []Slot, champ string, owned map[int]struct{}) []Slot {
	jungle := false
	for _, s := range build {
		if IsJunglePet(s.ItemID) {
			jungle = true
			break
		}
	}

	var out []Slot
	for _, s := range build {
		if s.Role != "start" {
			continue
		}
		if !ownsID(owned, s.ItemID) {
			out = append(out, s)
		}
		break
	}

	pots := 2
	if jungle {
		pots = 1
	}
	if champ == "Yuumi" {
		pots = 0
	}
	if ownsID(owned, ItemHealthPotion) {
		pots = 0
	}
	for i := 0; i < pots; i++ {
		out = append(out, Slot{ItemID: ItemHealthPotion, Name: "Health Potion", Priority: 95, Role: "consumable"})
	}

	if !ownsAnyBootsIDs(owned) {
		for _, s := range build {
			if s.Role == "boots" {
				out = append(out, s)
				break
			}
		}
	}

	var core, comp Slot
	for _, s := range build {
		if s.Role == "component" && comp.ItemID == 0 {
			comp = s
		}
		if s.Role == "core" && core.ItemID == 0 {
			core = s
		}
	}
	if ownsID(owned, core.ItemID) {
		return out
	}
	if comp.ItemID != 0 && !ownsID(owned, comp.ItemID) {
		out = append(out, comp)
	} else if core.ItemID != 0 && !ownsID(owned, core.ItemID) {
		ids := ComponentsToward[core.ItemID]
		for _, id := range ids {
			if ownsID(owned, id) {
				continue
			}
			out = append(out, Slot{
				ItemID:   id,
				Name:     core.Name + " component",
				Priority: core.Priority,
				Role:     "component",
			})
			break
		}
	}
	return out
}
