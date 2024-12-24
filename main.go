package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// create state map with mutex for thread safety
var (
	stateMutex  sync.RWMutex
	playerState = make(map[string]struct {
		status string
		artist string
		title  string
	})
)

func main() {

	// check if player argument is provided
	if len(os.Args) < 2 {
		fmt.Println("Please provide a player name as argument")
		fmt.Println("Example: ./program firefox")
		os.Exit(1)
	}

	// get the player argument
	argument := os.Args[1]

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			playerName := argument
			player := getPlayerName(playerName)

			go func(playerName string) {

				// define content
				icon := getPlayerIcon(playerName)
				newStatus := getStatus(playerName)
				newArtist, newTitle := getArtistTitle(playerName)

				// update state if we got valid output
				stateMutex.Lock()
				if newStatus != "?" || newArtist != "" || newTitle != "" {
					playerState[playerName] = struct {
						status string
						artist string
						title  string
					}{
						status: newStatus,
						artist: newArtist,
						title:  newTitle,
					}
				}

				// get current state
				state, exists := playerState[playerName]
				if !exists {
					state = struct {
						status string
						artist string
						title  string
					}{
						status: newStatus,
						artist: newArtist,
						title:  newTitle,
					}
				}
				stateMutex.Unlock()

				// barf output only if we have data
				if state.artist != "" || state.title != "" {
					fmt.Printf("%s %s %s - %s\n", icon, state.status, state.artist, state.title)
				}
			}(player)
		}
	}
}

// function to get player name
func getPlayerName(playerName string) string {

	// get players
	cmd := exec.Command("playerctl", "-l")
	output, _ := cmd.Output()

	// split output into lines
	players := strings.Split(strings.TrimSpace(string(output)), "\n")

	// find the line containing the player name
	for _, player := range players {
		if strings.Contains(player, playerName) {
			playerName = player
		}
	}

	return playerName
}

func getStatus(player string) string {

	status_icon := "?"

	// get status
	cmd := exec.Command("playerctl", "-p", player, "status")
	output, _ := cmd.Output()
	status := string(output)

	if strings.Contains(status, "Playing") {
		status_icon = ""
	}

	if strings.Contains(status, "Paused") {
		status_icon = ""
	}

	return status_icon
}

// function to get metadata per player
func getArtistTitle(player string) (string, string) {

	// get metadata
	cmd := exec.Command("playerctl", "-p", player, "metadata", "--format", "{{ artist }}|%|{{ title }}")
	output, _ := cmd.Output()

	// convert output to string and split by the delimiter
	// the delimiter |%| is chosen and thought of to be unique
	// as the delimiter should be in a video or song name
	parts := strings.Split(strings.TrimSpace(string(output)), "|%|")
	artist := strings.TrimSpace(parts[0])
	title := strings.TrimSpace(parts[1])
	return artist, title
}

// function to get player icons
func getPlayerIcon(player string) string {

	// default player icon
	player_icon := ""

	// firefox / youtube
	if strings.Contains(player, "firefox") {
		player_icon = "󰗃"
	}

	// spotify
	if player == "spotify" {
		player_icon = ""
	}

	return player_icon
}
