package main

import (
	"log"
	"net/http"
)

type App struct {
	cfg    Config
	router *http.ServeMux

	tgBot         *TelegramBot
	spotify       *spotifyClient
	spotifyTokens *spotifyTokenStore
}

func NewApp(cfg Config) (*App, error) {
	app := &App{
		cfg:           cfg,
		router:        http.NewServeMux(),
		spotifyTokens: newSpotifyTokenStore(),
	}

	if cfg.TelegramBotToken != "" {
		tg, err := NewTelegramBot(cfg.TelegramBotToken, app)
		if err != nil {
			log.Printf("failed to init telegram bot: %v", err)
		} else {
			app.tgBot = tg
		}
	}

	if cfg.SpotifyClientID != "" && cfg.SpotifySecret != "" && cfg.SpotifyRedirect != "" {
		app.spotify = newSpotifyClient(cfg)
	}

	app.registerRoutes()
	return app, nil
}

func (a *App) Router() http.Handler {
	return a.router
}

func (a *App) registerRoutes() {
	// health check
	a.router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Telegram webhook endpoint
	a.router.HandleFunc("/telegram/webhook", a.handleTelegramWebhook)

	// Spotify OAuth
	a.router.HandleFunc("/auth/spotify/login", a.handleSpotifyLogin)
	a.router.HandleFunc("/auth/spotify/callback", a.handleSpotifyCallback)

	// Basic API for mini app
	a.router.HandleFunc("/api/me", a.handleMe)
	a.router.HandleFunc("/api/playlists", a.handlePlaylists)
	a.router.HandleFunc("/api/playlist_tracks", a.handlePlaylistTracks)
	a.router.HandleFunc("/api/search", a.handleSearchTracks)
	a.router.HandleFunc("/api/player/current", a.handlePlayerCurrent)
	a.router.HandleFunc("/api/player/pause", a.handlePlayerSimpleAction("/me/player/pause"))
	a.router.HandleFunc("/api/player/next", a.handlePlayerSimpleAction("/me/player/next"))
	a.router.HandleFunc("/api/player/previous", a.handlePlayerSimpleAction("/me/player/previous"))

	// static mini app
	fs := http.FileServer(http.Dir(a.cfg.WebAppStaticDir))
	a.router.Handle("/static/", http.StripPrefix("/static/", fs))
	a.router.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, a.cfg.WebAppStaticDir+"/index.html")
	})
}
