package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type spotifyTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]spotifyUserTokens
}

type spotifyUserTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func newSpotifyTokenStore() *spotifyTokenStore {
	return &spotifyTokenStore{
		tokens: make(map[string]spotifyUserTokens),
	}
}

func (s *spotifyTokenStore) Set(tgUserID string, tok spotifyUserTokens) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[tgUserID] = tok
}

func (s *spotifyTokenStore) Get(tgUserID string) (spotifyUserTokens, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tokens[tgUserID]
	return t, ok
}

func (a *App) handleSpotifyLogin(w http.ResponseWriter, r *http.Request) {
	tgUserID := r.URL.Query().Get("tg_user_id")
	if tgUserID == "" {
		http.Error(w, "missing tg_user_id", http.StatusBadRequest)
		return
	}

	state := tgUserID
	url := a.spotify.authURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func (a *App) handleSpotifyCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Error(w, "missing code or state", http.StatusBadRequest)
		return
	}

	tok, err := a.spotify.exchange(context.Background(), code)
	if err != nil {
		log.Printf("spotify token exchange error: %v", err)
		http.Error(w, "spotify auth failed", http.StatusBadGateway)
		return
	}

	a.spotifyTokens.Set(state, spotifyUserTokens{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<html><body><script>window.close();</script>Авторизация Spotify выполнена, можно вернуться в Telegram.</body></html>`))
}

func (a *App) handleMe(w http.ResponseWriter, r *http.Request) {
	tgUserID := r.URL.Query().Get("tg_user_id")
	if tgUserID == "" {
		http.Error(w, "missing tg_user_id", http.StatusBadRequest)
		return
	}

	_, ok := a.spotifyTokens.Get(tgUserID)

	resp := map[string]any{
		"telegram_user_id": tgUserID,
		"spotify_linked":   ok,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
