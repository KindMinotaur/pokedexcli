package main

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/KindMinotaur/pokedexcli/internal/pokecache"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "Charmander BulbasaUr PIKACHU  ",
			expected: []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			input:    "TeSts  ArE LaMe  ",
			expected: []string{"tests", "are", "lame"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("Cleaning needed...")
			}
		}
	}
}

func TestCommandCatch(t *testing.T) {
	pokedex = make(map[string]Pokemon)

	cache := pokecache.NewCache(5 * time.Minute)
	pikachuJSON := `{"id":25,"name":"pikachu","base_experience":112,"height":4,"weight":60}`
	cache.Add("https://pokeapi.co/api/v2/pokemon/pikachu", []byte(pikachuJSON))

	config := &commandConfig{
		cache: cache,
	}

	err := commandCatch(config, []string{})
	if err == nil {
		t.Errorf("Expected error for no args")
	}

	err = commandCatch(config, []string{"pikachu"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCommandInspect(t *testing.T) {
	pokedex = make(map[string]Pokemon)

	config := &commandConfig{}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := commandInspect(config, []string{})
	if err == nil {
		t.Errorf("Expected error for no args")
	}

	err = commandInspect(config, []string{"pikachu"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	output := string(out)
	if !strings.Contains(output, "you have not caught that pokemon") {
		t.Errorf("Expected 'you have not caught that pokemon', got %s", output)
	}

	pokedex["pikachu"] = Pokemon{
		Name:   "pikachu",
		Height: 4,
		Weight: 60,
		Stats: []struct {
			BaseStat int `json:"base_stat"`
			Effort   int `json:"effort"`
			Stat     struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"stat"`
		}{
			{BaseStat: 35, Stat: struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			}{Name: "hp"}},
			{BaseStat: 55, Stat: struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			}{Name: "attack"}},
		},
		Types: []struct {
			Slot int `json:"slot"`
			Type struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"type"`
		}{
			{Type: struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			}{Name: "electric"}},
		},
	}

	r2, w2, _ := os.Pipe()
	os.Stdout = w2

	err = commandInspect(config, []string{"pikachu"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	w2.Close()
	os.Stdout = old
	out2, _ := io.ReadAll(r2)
	output2 := string(out2)
	if !strings.Contains(output2, "Name: pikachu") {
		t.Errorf("Expected 'Name: pikachu', got %s", output2)
	}
	if !strings.Contains(output2, "Height: 4") {
		t.Errorf("Expected 'Height: 4', got %s", output2)
	}
	if !strings.Contains(output2, "Stats:") {
		t.Errorf("Expected 'Stats:', got %s", output2)
	}
	if !strings.Contains(output2, "-hp: 35") {
		t.Errorf("Expected '-hp: 35', got %s", output2)
	}
	if !strings.Contains(output2, "Types:") {
		t.Errorf("Expected 'Types:', got %s", output2)
	}
	if !strings.Contains(output2, "-electric") {
		t.Errorf("Expected '-electric', got %s", output2)
	}
}
