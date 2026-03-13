package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	Addr             string
	TelegramBotToken string
	SpotifyClientID  string
	SpotifySecret    string
	SpotifyRedirect  string
	AppBaseURL       string
	WebAppStaticDir  string
	SessionSecret    string
}

func loadConfig() Config {
	cfg := Config{
		Addr:             getenv("APP_ADDR", ":8080"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		SpotifyClientID:  os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifySecret:    os.Getenv("SPOTIFY_CLIENT_SECRET"),
		SpotifyRedirect:  os.Getenv("SPOTIFY_REDIRECT_URL"),
		AppBaseURL:       os.Getenv("APP_BASE_URL"),
		WebAppStaticDir:  getenv("WEBAPP_STATIC_DIR", "static"),
		SessionSecret:    getenv("SESSION_SECRET", "dev-secret-change-me"),
	}

	return cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	cfg := loadConfig()

	if cfg.TelegramBotToken == "" {
		log.Println("WARNING: TELEGRAM_BOT_TOKEN is not set")
	}
	if cfg.SpotifyClientID == "" || cfg.SpotifySecret == "" || cfg.SpotifyRedirect == "" {
		log.Println("WARNING: Spotify credentials (SPOTIFY_CLIENT_ID/SECRET/REDIRECT_URL) are not fully set")
	}

	app, err := NewApp(cfg)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: app.Router(),
	}

	go func() {
		log.Printf("HTTP server listening on %s\n", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	log.Println("server stopped")
}
