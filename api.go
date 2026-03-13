package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/oauth2"
)

func (a *App) getUserTokenFromQuery(r *http.Request) (*oauth2.Token, bool) {
	if a.spotify == nil {
		return nil, false
	}
	tgUserID := r.URL.Query().Get("tg_user_id")
	if tgUserID == "" {
		return nil, false
	}
	toks, ok := a.spotifyTokens.Get(tgUserID)
	if !ok {
		return nil, false
	}
	return &oauth2.Token{
		AccessToken:  toks.AccessToken,
		RefreshToken: toks.RefreshToken,
	}, true
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (a *App) handlePlaylists(w http.ResponseWriter, r *http.Request) {
	token, ok := a.getUserTokenFromQuery(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "spotify_not_linked"})
		return
	}
	playlists, err := a.spotify.getUserPlaylists(r.Context(), token)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, playlists)
}

func (a *App) handlePlaylistTracks(w http.ResponseWriter, r *http.Request) {
	token, ok := a.getUserTokenFromQuery(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "spotify_not_linked"})
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing id"})
		return
	}
	tracks, err := a.spotify.getPlaylistTracks(r.Context(), token, id)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, tracks)
}

func (a *App) handleSearchTracks(w http.ResponseWriter, r *http.Request) {
	token, ok := a.getUserTokenFromQuery(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "spotify_not_linked"})
		return
	}
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing q"})
		return
	}
	tracks, err := a.spotify.searchTracks(r.Context(), token, q)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, tracks)
}

func (a *App) handlePlayerCurrent(w http.ResponseWriter, r *http.Request) {
	token, ok := a.getUserTokenFromQuery(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "spotify_not_linked"})
		return
	}
	cur, err := a.spotify.getCurrentlyPlaying(r.Context(), token)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, cur)
}

func (a *App) handlePlayerSimpleAction(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := a.getUserTokenFromQuery(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "spotify_not_linked"})
			return
		}
		if err := a.spotify.playerAction(r.Context(), token, http.MethodPost, path, nil); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
