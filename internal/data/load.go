package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"lol-build-overlay/internal/tags"
)

// Store holds all static game data used by the engine.
type Store struct {
	Root           string
	Champions      []Champion
	ChampionByID   map[string]Champion
	ChampionAlias  map[string]string // name_ru / name_en / id → id
	Items          *tags.ItemCatalog
	ChampionTags   tags.Catalog
}

type championsFile struct {
	Champions []Champion `json:"champions"`
}

// Champion is a Data Dragon champion row.
type Champion struct {
	ID       string         `json:"id"`
	Key      int            `json:"key"`
	NameRU   string         `json:"name_ru"`
	NameEN   string         `json:"name_en"`
	Tags     []string       `json:"tags"`
	Info     ChampionInfo   `json:"info"`
	Stats    ChampionStats  `json:"stats"`
}

type ChampionInfo struct {
	Attack     int `json:"attack"`
	Defense    int `json:"defense"`
	Magic      int `json:"magic"`
	Difficulty int `json:"difficulty"`
}

type ChampionStats struct {
	AttackRange float64 `json:"attackrange"`
}

type itemsFile struct {
	Items []rawItem `json:"items"`
}

type rawItem struct {
	ID     int                `json:"id"`
	NameEN string             `json:"name_en"`
	Tags   []string           `json:"tags"`
	Stats  map[string]float64 `json:"stats"`
	Gold   struct {
		Base        int  `json:"base"`
		Total       int  `json:"total"`
		Sell        int  `json:"sell"`
		Purchasable bool `json:"purchasable"`
	} `json:"gold"`
	Into  []int `json:"into"`
	From  []int `json:"from"`
	Depth *int  `json:"depth"`
}

type championTagsFile struct {
	Champions map[string]tags.ChampionProfile `json:"champions"`
}

// Load reads data from root/data.
func Load(root string) (*Store, error) {
	s := &Store{Root: root, ChampionByID: map[string]Champion{}}

	chPath := filepath.Join(root, "data", "champions.json")
	var cf championsFile
	if err := readJSON(chPath, &cf); err != nil {
		return nil, fmt.Errorf("champions: %w", err)
	}
	s.Champions = cf.Champions
	for _, c := range cf.Champions {
		s.ChampionByID[c.ID] = c
	}
	s.buildChampionAlias()

	itPath := filepath.Join(root, "data", "items.json")
	var itf itemsFile
	if err := readJSON(itPath, &itf); err != nil {
		return nil, fmt.Errorf("items: %w", err)
	}
	s.Items = buildItemCatalog(itf.Items)

	tagPath := filepath.Join(root, "data", "champion_tags.json")
	if _, err := os.Stat(tagPath); err == nil {
		var tf championTagsFile
		if err := readJSON(tagPath, &tf); err != nil {
			return nil, fmt.Errorf("champion_tags: %w", err)
		}
		s.ChampionTags = tf.Champions
	} else {
		s.ChampionTags = tags.Catalog{}
	}

	return s, nil
}

func buildItemCatalog(raw []rawItem) *tags.ItemCatalog {
	cat := &tags.ItemCatalog{
		ByID: map[int]tags.ItemInfo{},
		Tags: map[int][]string{},
	}
	// Prefer lower IDs (SR base items) when duplicates exist (arena etc.).
	for _, r := range raw {
		if existing, ok := cat.ByID[r.ID]; ok {
			// Keep first (file order: usually base then variants).
			_ = existing
			continue
		}
		info := tags.ItemInfo{
			ID:     r.ID,
			NameEN: r.NameEN,
			Tags:   r.Tags,
			Stats:  r.Stats,
			Gold: tags.ItemGold{
				Base:        r.Gold.Base,
				Total:       r.Gold.Total,
				Sell:        r.Gold.Sell,
				Purchasable: r.Gold.Purchasable,
			},
			Into:  r.Into,
			From:  r.From,
			Depth: r.Depth,
		}
		if info.Stats == nil {
			info.Stats = map[string]float64{}
		}
		cat.ByID[r.ID] = info
		cat.Tags[r.ID] = tags.DeriveItemTags(info)
	}
	return cat
}

func readJSON(path string, dst any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

// FindDataRoot walks up from wd looking for data/champions.json.
func FindDataRoot(start string) (string, error) {
	dir := start
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "data", "champions.json")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("data/champions.json not found from %s", start)
}
