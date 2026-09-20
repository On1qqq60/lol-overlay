package data_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"lol-build-overlay/internal/data"
)

func TestResolveChampionID_RUAndRaw(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	store, err := data.Load(root)
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]string{
		"Akali":                             "Akali",
		"Акали":                             "Akali",
		"akali":                             "Akali",
		"game_character_displayname_Akali":  "Akali",
		"Тимо":                              "Teemo",
		"Вейгар":                            "Veigar",
		"Гекарим":                           "Hecarim",
		"Наутилус":                          "Nautilus",
		"Тристана":                          "Tristana",
		"MonkeyKing":                        "MonkeyKing",
	}
	for in, want := range cases {
		got := store.ResolveChampionID(in)
		if got != want {
			t.Errorf("ResolveChampionID(%q)=%q want %q", in, got, want)
		}
	}
}
