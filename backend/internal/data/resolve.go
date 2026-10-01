package data

import "strings"

const rawChampPrefix = "game_character_displayname_"

// ResolveChampionID maps Live Client championName / name_ru / name_en / raw id → Data Dragon id.
// Unresolved names are returned unchanged.
func (s *Store) ResolveChampionID(name string) string {
	if s == nil {
		return name
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if _, ok := s.ChampionByID[name]; ok {
		return name
	}
	if id, ok := s.ChampionAlias[name]; ok {
		return id
	}
	if id, ok := s.ChampionAlias[strings.ToLower(name)]; ok {
		return id
	}
	if strings.HasPrefix(name, rawChampPrefix) {
		return s.ResolveChampionID(strings.TrimPrefix(name, rawChampPrefix))
	}
	// Characters/Akali/... style (rare)
	if i := strings.Index(name, "/"); i >= 0 {
		parts := strings.Split(name, "/")
		for _, p := range parts {
			if _, ok := s.ChampionByID[p]; ok {
				return p
			}
			if id, ok := s.ChampionAlias[strings.ToLower(p)]; ok {
				return id
			}
		}
	}
	return name
}

func (s *Store) buildChampionAlias() {
	s.ChampionAlias = make(map[string]string, len(s.Champions)*4)
	add := func(alias, id string) {
		alias = strings.TrimSpace(alias)
		if alias == "" || id == "" {
			return
		}
		s.ChampionAlias[alias] = id
		lower := strings.ToLower(alias)
		if lower != alias {
			s.ChampionAlias[lower] = id
		}
	}
	for _, c := range s.Champions {
		add(c.ID, c.ID)
		add(c.NameEN, c.ID)
		add(c.NameRU, c.ID)
	}
}
