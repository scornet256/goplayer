package main
import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)


// create waitgroup
var waitGroup = sync.WaitGroup{}

// create state map with mutex for thread safety
var (
  stateMutex sync.RWMutex
  playerState = make(map[string]struct {
    status string
    artist string
    title  string
  })
)


// main function
func main() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
		for {
			select {
			case <-ticker.C:

				// get players
				players := getActivePlayers()
	
				// process each player
				for _, player := range players {

					waitGroup.Add(1)

					go func(playerName string) {
						defer waitGroup.Done()
	
						// define content
						icon                := getPlayerIcon(playerName)
						newStatus           := getStatus(playerName)
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
				waitGroup.Wait()
		}
	}
}


// function to get active players
func getActivePlayers() []string {

	// get players
	cmd := exec.Command("playerctl", "-l")
	player, _ := cmd.Output()

	// return players
	players := strings.Split(strings.TrimSpace(string(player)), "\n")
	return players
}


func getStatus(player string) string {

  status_icon := "?"

  // get status
  cmd := exec.Command("playerctl", "-p", player, "status")
  output, _ := cmd.Output()
  status    := string(output)

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
	parts  := strings.Split(strings.TrimSpace(string(output)), "|%|")
  artist := strings.TrimSpace(parts[0])
  title  := strings.TrimSpace(parts[1])
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
