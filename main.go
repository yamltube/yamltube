package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	yt "github.com/mchaynes/yamltube/youtube"
	"google.golang.org/api/youtube/v3"
	"gopkg.in/yaml.v3"
)

const (
	EnvVarAppCreds     = "GOOGLE_APPLICATION_CREDENTIALS"
	EnvVarClientSecret = "GOOGLE_CLIENT_SECRET"

	FileAppCreds     = "./application_credentials.json"
	FileClientSecret = "./client_secret.json"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

type YamlTube struct {
	Playlists []Playlist `yaml:"playlists"`
}

type Playlist struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Visibility  string   `yaml:"visibility"`
	Videos      []string `yaml:"videos"`
}

func run() error {
	ctx := context.Background()
	envOrDefault := func(envVar, def string) string {
		if val, ok := os.LookupEnv(envVar); ok {
			return val
		}
		return def
	}

	endpoint := envOrDefault("YAMLTUBE_ENDPOINT", "https://yamltube.com/google/refresh")
	refreshToken := envOrDefault("YAMLTUBE_REFRESH_TOKEN", "")
	if len(refreshToken) == 0 {
		return fmt.Errorf("YAMLTUBE_REFRESH_TOKEN is empty")
	}

	service, err := yt.New(ctx, endpoint, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to connect to youtube: %w", err)
	}
	b, err := ioutil.ReadFile("tube.yaml")
	if err != nil {
		return err
	}
	var yamltube YamlTube
	if err = yaml.Unmarshal(b, &yamltube); err != nil {
		return err
	}

	playlists, err := service.GetPlaylists(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Got %d playlists\n", len(playlists))

	expectedPlaylists := toMap(yamltube.Playlists, func(p Playlist) string { return strings.ToLower(p.Title) })
	gotPlaylists := toMap(playlists, func(p *youtube.Playlist) string { return strings.ToLower(p.Snippet.Title) })

	for title, ep := range expectedPlaylists {
		playlist, ok := gotPlaylists[title]
		var (
			err error
		)
		// Create playlist if not found
		if !ok {
			playlist, err = service.CreatePlaylist(ctx, ep.Title, ep.Description, ep.Visibility)
			fmt.Printf("Created playlist %q\n", ep.Title)
			if err != nil {
				return err
			}
		}
		// Update playlist if found and description or visibility is different
		if ok && playlist.Snippet.Description != ep.Description || playlist.Status.PrivacyStatus != ep.Visibility {
			_, err = service.UpdatePlaylist(ctx, playlist.Id, ep.Title, ep.Description, ep.Visibility)
			fmt.Printf("Updated %q. Description=%q, Visibility=%q\n", playlist.Snippet.Title, ep.Description, ep.Visibility)
			if err != nil {
				return err
			}
		}
		// Convert video links to video ids
		ids, err := yt.ToVideoIds(ep.Videos)
		if err != nil {
			return err
		}
		result, err := service.SyncPlaylist(ctx, playlist.Id, ids)
		fmt.Printf("Synchronized %q. Inserted %d, Deleted %d.\n", playlist.Snippet.Title, len(result.Inserts), len(result.Deletes))
		if err != nil {
			return err
		}
	}
	return nil
}

func toMap[T any](arr []T, keyExtracter func(t T) string) map[string]T {
	m := make(map[string]T)
	for _, t := range arr {
		m[keyExtracter(t)] = t
	}
	return m
}
