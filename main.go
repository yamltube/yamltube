package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
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

func run() error {
	ctx := context.Background()

	client := getClient(youtube.YoutubeScope)
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("failed to create youtube service: %w", err)
	}
	err = printPlaylists(service)
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("Successfully saved ./application_credentials.json")
	fmt.Println("Run: ")
	maybePrintCmd := func(env, file string) {
		if _, ok := os.LookupEnv(env); !ok {
			fmt.Printf("    export %s=\"$(cat %s)\"\n", env, file)
		}
	}
	maybePrintCmd(EnvVarAppCreds, FileAppCreds)
	maybePrintCmd(EnvVarClientSecret, FileClientSecret)
	fmt.Println(`    pulumi up`)
	return nil
}

func printPlaylists(service *youtube.Service) error {
	playlistsCall := service.Playlists.List([]string{"snippet,contentDetails"})
	playlistsCall.Mine(true)
	playlistsCall.MaxResults(100)
	playlistsResp, err := playlistsCall.Do()
	if err != nil {
		return fmt.Errorf("failed to get playlists: %w", err)
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	for _, playlist := range playlistsResp.Items {
		link := fmt.Sprintf("https://www.youtube.com/playlist?list=%s", playlist.Id)
		_, _ = fmt.Fprintf(w, "%s\t%s\n", playlist.Snippet.Title, link)
	}
	w.Flush()
	if len(playlistsResp.Items) == 0 {
		fmt.Println("Able to fetch playlists, but none returned")
	}
	return nil
}
