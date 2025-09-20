package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	pokeCache "github.com/demola07/pokedexcli/internal/pokecache"
)

func cleanInput(text string) []string {
	word := strings.Split(text, " ")

	var cleaned []string
	for _, w := range word {
		trimmed := strings.ToLower(strings.TrimSpace(w))
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}

func commandExit(args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(commands map[string]cliCommand) func(args []string) error {
	return func(args []string) error {
		fmt.Println("Welcome to the Pokedex!")
		fmt.Println("Usage:")

		for _, cmd := range commands {
			fmt.Printf("%s: %s\n", cmd.name, cmd.description)
		}
		return nil
	}
}

type pokemon struct {
	Name   string `json:"name"`
	Height int    `json:"height"`
	Weight int    `json:"weight"`
	Stats  []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
	BaseExperience int `json:"base_experience"`
}

type cliCommand struct {
	name        string
	description string
	callback    func(args []string) error
}

type config struct {
	next     *string
	previous *string
	cache    *pokeCache.Cache
	pokedex  map[string]pokemon
}

type locationAreasResponse struct {
	Count    int                `json:"count"`
	Next     *string            `json:"next"`
	Previous *string            `json:"previous"`
	Results  []locationAreaItem `json:"results"`
}

type locationAreaItem struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type exploreResponse struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

// map command
func commandMap(cfg *config) func(args []string) error {
	return func(args []string) error {
		url := "https://pokeapi.co/api/v2/location-area"
		if cfg.next != nil {
			url = *cfg.next
		}

		// try cache first
		if cfg.cache != nil {
			if body, ok := cfg.cache.Get(url); ok {
				fmt.Println("[cache] hit", url)
				var locations locationAreasResponse
				if err := json.Unmarshal(body, &locations); err != nil {
					return err
				}
				for _, loc := range locations.Results {
					fmt.Println(loc.Name)
				}
				cfg.next = locations.Next
				cfg.previous = locations.Previous
				return nil
			}
			fmt.Println("[cache] miss", url)
		}

		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var locations locationAreasResponse
		if err := json.Unmarshal(body, &locations); err != nil {
			return err
		}

		for _, loc := range locations.Results {
			fmt.Println(loc.Name)
		}

		cfg.next = locations.Next
		cfg.previous = locations.Previous
		return nil
	}
}

// map back command
func commandMapb(cfg *config) func(args []string) error {
	return func(args []string) error {
		if cfg.previous == nil {
			fmt.Println("you're on the first page")
			return nil
		}

		url := *cfg.previous

		// try cache first
		if cfg.cache != nil {
			if body, ok := cfg.cache.Get(url); ok {
				fmt.Println("[cache] hit", url)
				var locations locationAreasResponse
				if err := json.Unmarshal(body, &locations); err != nil {
					return err
				}
				for _, loc := range locations.Results {
					fmt.Println(loc.Name)
				}
				cfg.next = locations.Next
				cfg.previous = locations.Previous
				return nil
			}
			fmt.Println("[cache] miss", url)
		}

		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var locations locationAreasResponse
		if err := json.Unmarshal(body, &locations); err != nil {
			return err
		}

		for _, loc := range locations.Results {
			fmt.Println(loc.Name)
		}

		cfg.next = locations.Next
		cfg.previous = locations.Previous
		return nil
	}
}

func commandExplore(cfg *config, cache *pokeCache.Cache) func(args []string) error {
	return func(args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("usage: explore <location-area>")
		}
		areaName := args[0]
		url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", areaName)

		// ✅ Check cache first
		if data, ok := cache.Get(url); ok {
			return printExploreResult(areaName, data)
		}

		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		// ✅ Save response to cache
		cache.Add(url, body)

		return printExploreResult(areaName, body)
	}
}

func printExploreResult(areaName string, data []byte) error {
	var explore exploreResponse
	if err := json.Unmarshal(data, &explore); err != nil {
		return err
	}

	fmt.Printf("Exploring %s...\n", areaName)
	fmt.Println("Found Pokemon:")
	for _, encounter := range explore.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}

	return nil
}

func commandCatch(cfg *config) func(args []string) error {
	return func(args []string) error {
		if len(args) < 1 {
			fmt.Println("Usage: catch <pokemon>")
			return nil
		}
		name := strings.ToLower(args[0])

		// Check if already caught
		if _, ok := cfg.pokedex[name]; ok {
			fmt.Printf("%s is already in your Pokedex!\n", name)
			return nil
		}

		// Fetch from API
		url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", name)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Printf("Could not find Pokémon: %s\n", name)
			return nil
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var poke pokemon
		if err := json.Unmarshal(body, &poke); err != nil {
			return err
		}

		// Catch attempt
		fmt.Printf("Throwing a Pokeball at %s...\n", poke.Name)

		rand.Seed(time.Now().UnixNano())
		chance := rand.Intn(100) // 0–99

		// Use base_experience to scale difficulty
		threshold := 50
		if poke.BaseExperience > 200 {
			threshold = 30
		} else if poke.BaseExperience > 100 {
			threshold = 40
		}

		if chance < threshold {
			fmt.Printf("%s was caught!\n", poke.Name)
			cfg.pokedex[poke.Name] = poke
		} else {
			fmt.Printf("%s escaped!\n", poke.Name)
		}

		return nil
	}
}

func commandInspect(cfg *config) func(args []string) error {
	return func(args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("please provide a pokemon name")
		}

		name := args[0]
		pokemon, ok := cfg.pokedex[name]
		if !ok {
			fmt.Println("you have not caught that pokemon")
			return nil
		}

		// Print details
		fmt.Printf("Name: %s\n", pokemon.Name)
		fmt.Printf("Height: %d\n", pokemon.Height)
		fmt.Printf("Weight: %d\n", pokemon.Weight)

		fmt.Println("Stats:")
		for _, stat := range pokemon.Stats {
			fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
		}

		fmt.Println("Types:")
		for _, t := range pokemon.Types {
			fmt.Printf("  - %s\n", t.Type.Name)
		}

		return nil
	}
}

func commandPokedex(cfg *config) func(args []string) error {
	return func(args []string) error {
		if len(cfg.pokedex) == 0 {
			fmt.Println("Your Pokedex is empty.")
			return nil
		}

		fmt.Println("Your Pokedex:")

		// Collect and sort names so output is consistent
		names := make([]string, 0, len(cfg.pokedex))
		for name := range cfg.pokedex {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			fmt.Printf(" - %s\n", name)
		}

		return nil
	}
}

func main() {

	cache := pokeCache.NewCache(60 * time.Second)
	pokemondex := make(map[string]pokemon)

	cfg := &config{
		cache:   cache,
		pokedex: pokemondex,
	}
	commands := map[string]cliCommand{}

	commands["exit"] = cliCommand{
		name:        "exit",
		description: "Exit the Pokedex",
		callback:    commandExit,
	}
	commands["help"] = cliCommand{
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp(commands),
	}

	commands["map"] = cliCommand{
		name:        "map",
		description: "Display the next 20 location areas",
		callback:    commandMap(cfg),
	}

	commands["mapb"] = cliCommand{
		name:        "mapb",
		description: "Go back to the previous 20 location areas",
		callback:    commandMapb(cfg),
	}

	commands["explore"] = cliCommand{
		name:        "explore",
		description: "Explore a location area for Pokemon",
		callback:    commandExplore(cfg, cache),
	}

	commands["catch"] = cliCommand{
		name:        "catch",
		description: "Attempt to catch a Pokémon by name",
		callback:    commandCatch(cfg),
	}

	commands["inspect"] = cliCommand{
		name:        "inspect",
		description: "View details of a caught Pokémon",
		callback:    commandInspect(cfg),
	}

	commands["pokedex"] = cliCommand{
		name:        "pokedex",
		description: "List all caught Pokémon",
		callback:    commandPokedex(cfg),
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		// print the prompt
		fmt.Print("Pokedex > ")

		// wait for input
		if !scanner.Scan() {
			break
		}

		// get the input text
		input := scanner.Text()

		words := cleanInput(input)
		if len(words) == 0 {
			continue
		}

		commandName := words[0]
		args := words[1:]

		cmd, ok := commands[commandName]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		if err := cmd.callback(args); err != nil {
			fmt.Println("Error:", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}
}
