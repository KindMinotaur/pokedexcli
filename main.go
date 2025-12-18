package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/KindMinotaur/pokedexcli/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*commandConfig, []string) error
}

type commandConfig struct {
	nextURL     string
	previousURL string
	cache       *pokecache.Cache
}

type LocationAreaList struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type LocationDetails struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	GameIndex            int    `json:"game_index"`
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	Location struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Names []struct {
		Name     string `json:"name"`
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
			MaxChance        int `json:"max_chance"`
			EncounterDetails []struct {
				MinLevel        int   `json:"min_level"`
				MaxLevel        int   `json:"max_level"`
				ConditionValues []any `json:"condition_values"`
				Chance          int   `json:"chance"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
			} `json:"encounter_details"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

var commands map[string]cliCommand

func init() {
	commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Displays the next 20 pages of the Pokedex map",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous 20 pages of the Pokedex map",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Explore a specific location area in the Pokedex map",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Catch a Pokemon by name",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a caught Pokemon by name",
			callback:    commandInspect,
		},
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	config := &commandConfig{
		cache: pokecache.NewCache(5 * time.Minute),
	}

	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		text := scanner.Text()
		cleanText := cleanInput(text)

		if len(cleanText) == 0 {
			continue
		}

		firstWord := cleanText[0]

		if cmd, exists := commands[firstWord]; exists {
			cmd.callback(config, cleanText[1:])
		} else {
			fmt.Println("Unknown command")
		}
	}
}

func commandExit(config *commandConfig, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(config *commandConfig, args []string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for name, cmd := range commands {
		fmt.Printf("%s: %s\n", name, cmd.description)
	}
	return nil
}

func commandMap(config *commandConfig, args []string) error {
	url := "https://pokeapi.co/api/v2/location-area/"
	if config.nextURL != "" {
		url = config.nextURL
	}
	if data, ok := config.cache.Get(url); ok {
		var list LocationAreaList
		if err := json.Unmarshal(data, &list); err != nil {
			return err
		}
		config.nextURL = list.Next
		config.previousURL = list.Previous
		for _, result := range list.Results {
			fmt.Println(result.Name)
		}
		return nil
	}

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if res.StatusCode > 299 {
		return fmt.Errorf("Response failed with status code: %d and body: %s", res.StatusCode, body)
	}
	if err != nil {
		return err
	}

	var list LocationAreaList
	err = json.Unmarshal(body, &list)
	if err != nil {
		return err
	}

	config.nextURL = list.Next
	config.previousURL = list.Previous

	for _, result := range list.Results {
		fmt.Println(result.Name)
	}

	config.cache.Add(url, body)

	return nil
}

func commandMapb(config *commandConfig, args []string) error {
	url := "https://pokeapi.co/api/v2/location-area/"
	if config.previousURL == "" {
		fmt.Println("you're on the first page")
		return nil
	}
	if data, ok := config.cache.Get(url); ok {
		var list LocationAreaList
		if err := json.Unmarshal(data, &list); err != nil {
			return err
		}
		config.nextURL = list.Next
		config.previousURL = list.Previous
		for _, result := range list.Results {
			fmt.Println(result.Name)
		}
		return nil
	}
	res, err := http.Get(config.previousURL)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		return fmt.Errorf("Response failed with status code: %d and body: %s", res.StatusCode, body)
	}
	if err != nil {
		return err
	}

	var list LocationAreaList
	err = json.Unmarshal(body, &list)
	if err != nil {
		return err
	}

	config.nextURL = list.Next
	config.previousURL = list.Previous

	for _, result := range list.Results {
		fmt.Println(result.Name)
	}

	config.cache.Add(url, body)

	return nil
}

func commandExplore(config *commandConfig, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("explore command requires an area name")
	}
	areaName := args[0]
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", areaName)
	if data, ok := config.cache.Get(url); ok {
		var location LocationDetails
		if err := json.Unmarshal(data, &location); err != nil {
			return err
		}
		for _, result := range location.PokemonEncounters {
			fmt.Println(result.Pokemon.Name)
		}
		return nil
	}
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	var location LocationDetails
	err = json.Unmarshal(body, &location)
	if err != nil {
		return err
	}
	for _, result := range location.PokemonEncounters {
		fmt.Println(result.Pokemon.Name)
	}
	return nil
}

func commandCatch(config *commandConfig, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("catch command requires a pokemon name")
	}
	pokemonName := args[0]
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)
	return nil
}

func commandInspect(config *commandConfig, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("inspect command requires a pokemon name")
	}
	pokemonName := args[0]
	fmt.Printf("Inspecting %s...\n", pokemonName)
	return nil
}
