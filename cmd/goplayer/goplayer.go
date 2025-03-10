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
			cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(commandTimeout)*time.Second)

			if data, err := getPlayerData(cmdCtx, player); err == nil {
				if lastData == nil ||
					lastData.Artist != data.Artist ||
					lastData.Title != data.Title {
					fmt.Printf("%s - %s\n", data.Artist, data.Title)
				}
				lastData = data
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

	if player == "chromium" {
		actualPlayer, err := findPlayerName(ctx, player)
		if err != nil {
			return nil, fmt.Errorf("player not running")
		}
		execCommand = "chrome"
		player = actualPlayer
	}

	if player == "firefox" {
		actualPlayer, err := findPlayerName(ctx, player)
		if err != nil {
			return nil, fmt.Errorf("player not running")
		}
		execCommand = "zen"
		player = actualPlayer
	}

	pgrepCmd := exec.CommandContext(ctx, "pgrep", execCommand)
	if err := pgrepCmd.Run(); err != nil {
		return nil, fmt.Errorf("player not running")
	}

  // small function to find metadata from player
  metaCmd := func(ctx context.Context, player string, metadata string) (string, error) {
	  cmd := exec.CommandContext(ctx, "playerctl", "-p", player, "metadata", metadata)
    cmdBytes, err := cmd.Output()
    return string(cmdBytes), err
  }

  albumString, err := metaCmd(ctx, player, "album")
  titleString, err := metaCmd(ctx, player, "title")
  artistString, err := metaCmd(ctx, player, "artist")

  if err != nil {
    return nil, fmt.Errorf("player metadata incorrect: %v", err)
  }

	// function to cleanup the string
	cleanString := func(metaString string) string {
		newString := strings.TrimSpace(metaString)
		newString = replaceChars(newString, "&", "&amp;")
		return newString
	}

	album := cleanString(albumString)
	artist := cleanString(artistString)
	title := cleanString(titleString)

	// spotify doesnt show podcast artists,
	// so we have to use album as artist name
	if artist == "" || album != "" {
		artist = album
	}

	if title == "" || artist == "" {
		return nil, fmt.Errorf("empty metadata")
	}

	return &PlayerData{
		Title:  title,
		Artist: artist,
	}, nil
}

func replaceChars(line string,
	oldChar string,
	newChar string) string {

	if strings.Contains(line, oldChar) {
		line = strings.Replace(line, oldChar, newChar, -1)
	}
	return line
}

func findPlayerName(ctx context.Context, playerToFind string) (string, error) {
	cmd := exec.CommandContext(ctx, "playerctl", "-l")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	players := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, player := range players {
		if strings.Contains(player, playerToFind) {
			return player, nil
		}
	}
	return "", fmt.Errorf("%v player not found", playerToFind)
}
