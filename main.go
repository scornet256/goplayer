package main
import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var commandTimeout = 1
var loopTimeout = 2
type PlayerData struct {
	Title  string
	Artist string
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: program <player>")
		fmt.Println("Example: program spotify")
		os.Exit(1)
	}

	player := os.Args[1]
	ctx := context.Background()
	monitorPlayer(ctx, player)
}

func monitorPlayer(ctx context.Context, player string) {
	var lastData *PlayerData
	ticker := time.NewTicker(time.Duration(loopTimeout) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(commandTimeout) * time.Second)
			if data, err := getPlayerData(cmdCtx, player); err == nil {
				lastData = data
				fmt.Printf("%s - %s\n", data.Artist, data.Title)
			} else if err.Error() == "player not running" {
				fmt.Println("")
			} else if lastData != nil {
				fmt.Printf("%s - %s\n", lastData.Artist, lastData.Title)
			}
			cancel()
		}
	}
}

func getPlayerData(ctx context.Context, player string) (*PlayerData, error) {
  var execCommand = player

	// if player is firefox, get the actual player name
	if player == "firefox" {
		actualPlayer, err := findFirefoxPlayer(ctx)
		if err != nil {
			return nil, fmt.Errorf("player not running")
		}
    execCommand = "zen" 
		player = actualPlayer
	}

	// check if player is running
	pgrepCmd := exec.CommandContext(ctx, "pgrep", execCommand)
	if err := pgrepCmd.Run(); err != nil {
		return nil, fmt.Errorf("player not running")
	}

	titleCmd  := exec.CommandContext(ctx, "playerctl", "-p", player, "metadata", "title")
	artistCmd := exec.CommandContext(ctx, "playerctl", "-p", player, "metadata", "artist")

	titleBytes, err := titleCmd.Output()
	if err != nil {
		return nil, err
	}

	artistBytes, err := artistCmd.Output()
	if err != nil {
		return nil, err
	}

	title  := strings.TrimSpace(string(titleBytes))
	artist := strings.TrimSpace(string(artistBytes))

	if title == "" || artist == "" {
		return nil, fmt.Errorf("empty metadata")
	}

	return &PlayerData{
		Title:  title,
		Artist: artist,
	}, nil
}

func findFirefoxPlayer(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "playerctl", "-l")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	players := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, player := range players {
		if strings.Contains(player, "firefox") {
			return player, nil
		}
	}
	return "", fmt.Errorf("firefox player not found")
}
