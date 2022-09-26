package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YouTube struct {
	service *youtube.Service
}

type PlaylistInsert struct {
	VideoId  string
	Position int64
}

type PlaylistItemDelete struct {
	ItemId string
}

type PlaylistDiffResult struct {
	Inserts []PlaylistInsert
	Deletes []PlaylistItemDelete
}

type source struct {
	mu           sync.Mutex
	endpoint     string
	refreshToken string
	tok          oauth2.Token
}

func (s *source) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.tok.AccessToken) == 0 || s.tok.Expiry.After(time.Now()) {
		resp, err := http.Get(fmt.Sprintf("%s?refresh_token=%s", s.endpoint, s.refreshToken))
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		// yamltube response struct
		type token struct {
			AccessToken string `json:"access_token"`
			ExpiresIn   int    `json:"expires_in"`
			TokenType   string `json:"token_type"`
		}
		var tok token
		if err = json.Unmarshal(b, &tok); err != nil {
			return nil, err
		}
		s.tok = oauth2.Token{
			AccessToken: tok.AccessToken,
			TokenType:   tok.TokenType,
			Expiry:      time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second),
		}
	}
	return &s.tok, nil
}

func New(ctx context.Context, endpoint, refreshToken string) (*YouTube, error) {

	sauce := source{
		endpoint:     endpoint,
		refreshToken: refreshToken,
	}
	service, err := youtube.NewService(ctx, option.WithTokenSource(&sauce))
	if err != nil {
		return nil, fmt.Errorf("failed to create youtube service: %w", err)
	}
	return &YouTube{
		service: service,
	}, nil
}

func ToVideoIds(urlsOrIds []string) ([]string, error) {
	var ids []string
	for i, urlOrId := range urlsOrIds {
		id, err := ToVideoId(urlOrId)
		if err != nil {
			return nil, fmt.Errorf("invalid entry at %d: %w", i, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func ToVideoId(urlOrId string) (string, error) {
	if len(urlOrId) == 0 {
		return "", fmt.Errorf("urlOrId must be set")
	}
	// if its not prefixed with http, then assume its an ID
	if !strings.HasPrefix(urlOrId, "http") {
		return urlOrId, nil
	}

	URL, err := url.Parse(urlOrId)
	if err != nil {
		return "", fmt.Errorf("invalid url %q: %w", urlOrId, err)
	}
	values := URL.Query()
	v := values.Get("v")
	if len(v) == 0 {
		return "", fmt.Errorf("url %q does not contain ?v=<videoId> param", urlOrId)
	}
	return v, nil
}

func (y *YouTube) DiffPlaylist(wantIds []string, gotItems []*youtube.PlaylistItem) PlaylistDiffResult {
	diff := PlaylistDiffResult{}
	// if what we got is longer than what we want, remove everything that's after
	// what we want
	if len(gotItems) > len(wantIds) {
		for i := len(wantIds); i < len(gotItems); i++ {
			diff.Deletes = append(diff.Deletes, PlaylistItemDelete{
				ItemId: gotItems[i].Id,
			})
		}
	}
	for i, wantVideoId := range wantIds {
		// check if we're out of bounds
		if len(gotItems) > i {
			gotItem := gotItems[i]
			// happy case, nothing to do
			if wantVideoId == gotItem.ContentDetails.VideoId {
				continue
			}
			// delete the playlistItem since it doesnt match
			diff.Deletes = append(diff.Deletes, PlaylistItemDelete{
				ItemId: gotItem.Id,
			})

			// we fall through to the insert here because
			// we didn't find the playlistItem we were looking for,
			// so we stil need to add it to the playlist
		}
		diff.Inserts = append(diff.Inserts, PlaylistInsert{
			VideoId:  wantVideoId,
			Position: int64(i),
		})
	}
	return diff
}

func (y *YouTube) CreatePlaylist(ctx context.Context, title, desc, visibility string) (*youtube.Playlist, error) {
	playlist, err := y.service.Playlists.Insert([]string{"snippet,status"}, &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       title,
			Description: desc,
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: visibility,
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create playlist: %w", err)
	}
	return playlist, nil
}

func (y *YouTube) DeletePlaylist(ctx context.Context, id string) error {
	return y.service.Playlists.Delete(id).Context(ctx).Do()
}

func (y *YouTube) UpdatePlaylist(ctx context.Context, id, title, desc, visibility string) (*youtube.Playlist, error) {
	return y.service.Playlists.Update([]string{"id,snippet,status"}, &youtube.Playlist{
		Id: id,
		Snippet: &youtube.PlaylistSnippet{
			Title:       title,
			Description: desc,
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: visibility,
		},
	}).Context(ctx).Do()
}

func (y *YouTube) GetPlaylist(ctx context.Context, id string) (*youtube.Playlist, error) {
	list, err := y.service.Playlists.List([]string{"id,snippet,status"}).
		Id(id).
		Context(ctx).
		Do()
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("playlist %q not found", id)
	}
	return list.Items[0], nil
}

func (y *YouTube) SyncPlaylist(ctx context.Context, playlistId string, wantIds []string) (*PlaylistDiffResult, error) {
	items, err := y.GetPlaylistItems(ctx, playlistId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch playlist for sync: %w", err)
	}

	diff := y.DiffPlaylist(wantIds, items)

	for _, item := range diff.Deletes {
		err := y.service.PlaylistItems.Delete(item.ItemId).
			Context(ctx).
			Do()
		if err != nil {
			return nil, fmt.Errorf("failed to delete playlistItem %q: %w", item, err)
		}
	}

	for _, video := range diff.Inserts {
		_, err := y.service.PlaylistItems.Insert([]string{"id,snippet"}, &youtube.PlaylistItem{
			Snippet: &youtube.PlaylistItemSnippet{
				PlaylistId: playlistId,
				Position:   video.Position,
				ResourceId: &youtube.ResourceId{
					Kind:    "youtube#video",
					VideoId: video.VideoId,
				},
			},
		}).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to insert playlistItem: %w", err)
		}
	}
	return &diff, nil
}

func (y *YouTube) GetPlaylistItems(ctx context.Context, playlistId string) ([]*youtube.PlaylistItem, error) {
	var pageToken string
	var items []*youtube.PlaylistItem
	firstLoop := true
	for len(pageToken) > 0 || firstLoop {
		listResp, err := y.service.PlaylistItems.List([]string{"snippet,contentDetails"}).
			PlaylistId(playlistId).
			Context(ctx).
			MaxResults(50).
			PageToken(pageToken).
			Do()

		if err != nil {
			return nil, fmt.Errorf("failed to fetch items: %w", err)
		}
		pageToken = listResp.NextPageToken
		items = append(items, listResp.Items...)
		firstLoop = false
	}
	return items, nil
}

func (y *YouTube) GetPlaylists(ctx context.Context) ([]*youtube.Playlist, error) {
	var playlists []*youtube.Playlist
	var pageToken string
	firstPage := true
	for len(pageToken) > 0 || firstPage {
		resp, err := y.service.Playlists.List([]string{"id,snippet,status"}).
			MaxResults(50).
			Context(ctx).
			Do()
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, resp.Items...)
	}
	return playlists, nil
}
