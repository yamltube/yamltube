package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
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
	/* FETCH PLAYLIST */
	if err := getPlaylistAndItems(service); err != nil {
		return err
	}

	return nil
}

func getPlaylistAndItems(service *youtube.Service) error {
	playlistsCall := service.Playlists.List([]string{"snippet,contentDetails"})
	playlistsCall.Mine(true)
	playlistsCall.MaxResults(100)
	playlistsResp, err := playlistsCall.Do()
	if err != nil {
		return fmt.Errorf("failed to get playlists: %w", err)
	}
	mustPrettyPrint(playlistsResp.MarshalJSON)

	return nil
}

func mustPrettyPrint(f func() ([]byte, error)) {
	data, err := f()
	if err != nil {
		panic(fmt.Errorf("failed to marshal data to json: %w", err))
	}
	buf := bytes.NewBuffer([]byte{})
	json.Indent(buf, data, " ", " ")
	fmt.Println(buf.String())
}
