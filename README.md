# Pokedex CLI

A command-line interface (CLI) application for exploring and interacting with Pokémon information using the PokéAPI.

## Features

- **Explore Location Areas:** List and explore Pokémon in different location areas.
- **Catch Pokémon:** Simulate catching Pokémon and add them to your Pokedex.
- **Inspect Pokémon:** View detailed information about caught Pokémon.
- **Cache:** Uses an in-memory cache to store API responses for faster access.
- **Help:** Displays a list of available commands and their descriptions.

## Installation

1. Clone the repository:

    ```sh
    git clone [https://github.com/yourusername/pokedexcli.git](https://github.com/yourusername/pokedexcli.git)
    cd pokedexcli
    ```

2. Build the project:

    ```sh
    go build
    ```

3. Run the application:

    ```sh
    ./pokedexcli
    ```

## Usage

Once the application is running, you can use the following commands:

- `help`: Displays a help message with a list of available commands.
- `exit`: Exit the Pokedex CLI.
- `map`: Display the next 20 location areas.
- `mapb`: Go back to the previous 20 location areas.
- `explore <location-area>`: Explore a location area for Pokémon.
- `catch <pokemon>`: Attempt to catch a Pokémon by name.
- `inspect <pokemon>`: View details of a caught Pokémon.

## Example

```sh
> help
Welcome to the Pokedex!
Usage:
help: Displays a help message
exit: Exit the Pokedex
map: Display the next 20 location areas
mapb: Go back to the previous 20 location areas
explore: Explore a location area for Pokemon
catch: Attempt to catch a Pokémon by name
inspect: View details of a caught Pokémon

> map
location-area-1
location-area-2
...

> explore location-area-1
Exploring location-area-1...
Found Pokemon:
 - pikachu
 - bulbasaur
 ...

> catch pikachu
Throwing a Pokeball at pikachu...
pikachu was caught!

> inspect pikachu
Name: pikachu
Height: 4
Weight: 60
Stats:
 - speed: 90
 - attack: 55
Types:
 - electric