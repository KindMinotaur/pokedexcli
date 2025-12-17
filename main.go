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
	callback    func(*commandConfig) error
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
			cmd.callback(config)
		} else {
			fmt.Println("Unknown command")
		}
	}
}

func commandExit(config *commandConfig) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(config *commandConfig) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for name, cmd := range commands {
		fmt.Printf("%s: %s\n", name, cmd.description)
	}
	return nil
}

func commandMap(config *commandConfig) error {
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
	// Proceed with API fetch if not cached
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

	// Optionally add to cache after fetching
	config.cache.Add(url, body)

	return nil
}

func commandMapb(config *commandConfig) error {
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
