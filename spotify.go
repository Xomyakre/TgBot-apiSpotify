package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

const (
	spotifyAuthURL  = "https://accounts.spotify.com/authorize"
	spotifyTokenURL = "https://accounts.spotify.com/api/token"
	spotifyAPIBase  = "https://api.spotify.com/v1"
)

type spotifyClient struct {
	oauthConfig *oauth2.Config
	httpClient  *http.Client
}

type spotifyUserProfile struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type spotifyPlaylist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type spotifyTrack struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func newSpotifyClient(cfg Config) *spotifyClient {
	return &spotifyClient{
		oauthConfig: &oauth2.Config{
			ClientID:     cfg.SpotifyClientID,
			ClientSecret: cfg.SpotifySecret,
			RedirectURL:  cfg.SpotifyRedirect,
			Scopes: []string{
				"user-read-email",
				"user-read-private",
				"user-read-playback-state",
				"user-modify-playback-state",
				"playlist-read-private",
				"playlist-read-collaborative",
			},
			Endpoint: oauth2.Endpoint{
				AuthURL:  spotifyAuthURL,
				TokenURL: spotifyTokenURL,
			},
		},
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *spotifyClient) authURL(state string) string {
	return c.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (c *spotifyClient) exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return c.oauthConfig.Exchange(ctx, code)
}

func (c *spotifyClient) apiGet(ctx context.Context, token *oauth2.Token, path string, query url.Values, v interface{}) error {
	base, _ := url.Parse(spotifyAPIBase)
	u, _ := url.Parse(path)
	full := base.ResolveReference(u)
	full.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, full.String(), nil)
	if err != nil {
		return err
	}

	client := c.oauthConfig.Client(ctx, token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("spotify api error: %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

func (c *spotifyClient) getUserProfile(ctx context.Context, token *oauth2.Token) (*spotifyUserProfile, error) {
	var out spotifyUserProfile
	if err := c.apiGet(ctx, token, "/me", url.Values{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type spotifyPlaylistsResponse struct {
	Items []spotifyPlaylist `json:"items"`
}

func (c *spotifyClient) getUserPlaylists(ctx context.Context, token *oauth2.Token) ([]spotifyPlaylist, error) {
	var out spotifyPlaylistsResponse
	if err := c.apiGet(ctx, token, "/me/playlists", url.Values{}, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

type spotifyTracksResponse struct {
	Items []struct {
		Track spotifyTrack `json:"track"`
	} `json:"items"`
}

func (c *spotifyClient) getPlaylistTracks(ctx context.Context, token *oauth2.Token, playlistID string) ([]spotifyTrack, error) {
	var out spotifyTracksResponse
	path := fmt.Sprintf("/playlists/%s/tracks", playlistID)
	if err := c.apiGet(ctx, token, path, url.Values{}, &out); err != nil {
		return nil, err
	}
	tracks := make([]spotifyTrack, 0, len(out.Items))
	for _, it := range out.Items {
		tracks = append(tracks, it.Track)
	}
	return tracks, nil
}

type spotifySearchResponse struct {
	Tracks struct {
		Items []spotifyTrack `json:"items"`
	} `json:"tracks"`
}

func (c *spotifyClient) searchTracks(ctx context.Context, token *oauth2.Token, queryStr string) ([]spotifyTrack, error) {
	params := url.Values{}
	params.Set("q", queryStr)
	params.Set("type", "track")
	params.Set("limit", "20")

	var out spotifySearchResponse
	if err := c.apiGet(ctx, token, "/search", params, &out); err != nil {
		return nil, err
	}
	return out.Tracks.Items, nil
}

type spotifyCurrentlyPlaying struct {
	Item spotifyTrack `json:"item"`
}

func (c *spotifyClient) getCurrentlyPlaying(ctx context.Context, token *oauth2.Token) (*spotifyCurrentlyPlaying, error) {
	var out spotifyCurrentlyPlaying
	if err := c.apiGet(ctx, token, "/me/player/currently-playing", url.Values{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *spotifyClient) playerAction(ctx context.Context, token *oauth2.Token, method, path string, body url.Values) error {
	base, _ := url.Parse(spotifyAPIBase)
	u, _ := url.Parse(path)
	full := base.ResolveReference(u)

	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, full.String(), nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, full.String(), nil)
	}
	if err != nil {
		return err
	}

	client := c.oauthConfig.Client(ctx, token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("spotify player action error: %s", resp.Status)
	}
	return nil
}
