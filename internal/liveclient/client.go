package liveclient

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"lol-build-overlay/internal/engine"
)

const defaultURL = "https://127.0.0.1:2999/liveclientdata/allgamedata"

// Client fetches Live Client allgamedata.
type Client struct {
	URL        string
	HTTPClient *http.Client
}

// New returns a client that trusts the self-signed Live Client cert.
func New() *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Riot Live Client local cert
	}
	return &Client{
		URL: defaultURL,
		HTTPClient: &http.Client{
			Timeout:   3 * time.Second,
			Transport: tr,
		},
	}
}

// AllGameData is the Riot Live Client payload (subset).
type AllGameData struct {
	ActivePlayer activePlayer `json:"activePlayer"`
	AllPlayers   []allPlayer  `json:"allPlayers"`
}

type activePlayer struct {
	SummonerName string  `json:"summonerName"`
	CurrentGold  float64 `json:"currentGold"`
}

type allPlayer struct {
	ChampionName   string         `json:"championName"`
	SummonerName   string         `json:"summonerName"`
	Team           string         `json:"team"`
	Position       string         `json:"position"`
	Level          int            `json:"level"`
	Scores         scores         `json:"scores"`
	Items          []itemSlot     `json:"items"`
	IsDead         bool           `json:"isDead"`
	RawChampion    string         `json:"rawChampionName"`
	SummonerSpells summonerSpells `json:"summonerSpells"`
}

type summonerSpells struct {
	One spellInfo `json:"summonerSpellOne"`
	Two spellInfo `json:"summonerSpellTwo"`
}

type spellInfo struct {
	DisplayName    string `json:"displayName"`
	RawDisplayName string `json:"rawDisplayName"`
	RawDescription string `json:"rawDescription"`
}

type scores struct {
	Kills   int `json:"kills"`
	Deaths  int `json:"deaths"`
	Assists int `json:"assists"`
}

type itemSlot struct {
	ItemID int `json:"itemID"`
	Slot   int `json:"slot"`
}

// Fetch downloads and normalizes a game snapshot.
func (c *Client) Fetch() (engine.GameSnapshot, error) {
	resp, err := c.HTTPClient.Get(c.URL)
	if err != nil {
		return engine.GameSnapshot{}, fmt.Errorf("live client: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return engine.GameSnapshot{}, fmt.Errorf("live client: status %d", resp.StatusCode)
	}
	var raw AllGameData
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return engine.GameSnapshot{}, fmt.Errorf("live client decode: %w", err)
	}
	return Normalize(raw), nil
}

// LoadFixture reads a saved allgamedata JSON file.
func LoadFixture(path string) (engine.GameSnapshot, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return engine.GameSnapshot{}, err
	}
	var raw AllGameData
	if err := json.Unmarshal(b, &raw); err != nil {
		return engine.GameSnapshot{}, err
	}
	return Normalize(raw), nil
}

// Normalize converts Live Client JSON into engine.GameSnapshot.
func Normalize(raw AllGameData) engine.GameSnapshot {
	activeName := raw.ActivePlayer.SummonerName
	var activeTeam, activeChamp string

	players := make([]engine.PlayerSnapshot, 0, len(raw.AllPlayers))
	for _, p := range raw.AllPlayers {
		champ := pickChampionKey(p.ChampionName, p.RawChampion)
		items := make([]int, 7)
		for _, it := range p.Items {
			slot := it.Slot
			if slot >= 0 && slot < 7 {
				items[slot] = it.ItemID
			} else {
				items = append(items, it.ItemID)
			}
		}
		ps := engine.PlayerSnapshot{
			ChampionID: champ,
			Team:       p.Team,
			Position:   p.Position,
			Level:      p.Level,
			Kills:      p.Scores.Kills,
			Deaths:     p.Scores.Deaths,
			Assists:    p.Scores.Assists,
			Items:      items,
			IsDead:     p.IsDead,
			Summoner:   p.SummonerName,
			SpellOne:   spellKey(p.SummonerSpells.One),
			SpellTwo:   spellKey(p.SummonerSpells.Two),
		}
		players = append(players, ps)
		if activeName != "" && p.SummonerName == activeName {
			activeTeam = p.Team
			activeChamp = champ
		}
	}

	if activeChamp == "" && len(players) > 0 {
		activeChamp = players[0].ChampionID
		activeTeam = players[0].Team
	}

	gold := -1
	if raw.ActivePlayer.CurrentGold > 0 || activeName != "" {
		gold = int(raw.ActivePlayer.CurrentGold)
	}

	return engine.GameSnapshot{
		ActiveChampionID: activeChamp,
		ActiveTeam:       activeTeam,
		CurrentGold:      gold,
		Players:          players,
	}
}

// pickChampionKey prefers English id from rawChampionName when present.
func pickChampionKey(display, raw string) string {
	const prefix = "game_character_displayname_"
	if strings.HasPrefix(raw, prefix) {
		return strings.TrimPrefix(raw, prefix)
	}
	if display != "" {
		return display
	}
	return raw
}

// spellKey extracts SummonerFlash / SummonerSmite from Live Client raw tip keys.
func spellKey(s spellInfo) string {
	for _, raw := range []string{s.RawDisplayName, s.RawDescription} {
		const marker = "SummonerSpell_"
		if i := strings.Index(raw, marker); i >= 0 {
			rest := raw[i+len(marker):]
			if j := strings.IndexByte(rest, '_'); j > 0 {
				return rest[:j]
			}
			return rest
		}
	}
	name := strings.TrimSpace(s.DisplayName)
	switch strings.ToLower(name) {
	case "smite", "кара":
		return "SummonerSmite"
	case "flash", "скачок":
		return "SummonerFlash"
	case "ignite", "воспламенение":
		return "SummonerDot"
	case "teleport", "телепорт":
		return "SummonerTeleport"
	case "heal", "исцеление":
		return "SummonerHeal"
	case "ghost", "призрак":
		return "SummonerHaste"
	case "barrier", "барьер":
		return "SummonerBarrier"
	case "exhaust", "изнурение":
		return "SummonerExhaust"
	case "cleanse", "очищение":
		return "SummonerBoost"
	}
	return name
}
