package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/mchaynes/yamltube/youtube"
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
	Playlists []Playlist	
}

type Playlist struct {
	Title string `yaml:"title"`
	Description string `yaml:"description"`
	Visibility string `yaml:"visibility"`
	Videos []string `yaml:"videos"`
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

	service, err := youtube.New(ctx, endpoint, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to connect to youtube: %w", err)
	}
	playlists, err := service.GetPlaylists(ctx)
	b, err := ioutil.ReadFile("tube.yaml")
	if err != nil {
		return err
	}
	var yamltube YamlTube
	if err = yaml.Unmarshal(b, &yamltube); err != nil {
		return err
	}
	
	return nil
}

func toPlaylistMap(playlists []Playlist) map[string]Playlist {
	playlistMap := make(map[string]Playlist)
	for _, playlist := range playlists {
		playlistMap[strings.ToLower(playlist.Title)] = playlist
	}
	return playlistMap
}
