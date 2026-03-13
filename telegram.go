package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	api *tgbotapi.BotAPI
	app *App
}

func NewTelegramBot(token string, app *App) (*TelegramBot, error) {
	if token == "" {
		return nil, errors.New("telegram bot token is empty")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &TelegramBot{
		api: bot,
		app: app,
	}, nil
}

func (t *TelegramBot) HandleUpdate(upd tgbotapi.Update) {
	if upd.Message == nil {
		return
	}

	if upd.Message.IsCommand() {
		t.handleCommand(upd.Message)
		return
	}
}

func (t *TelegramBot) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		t.handleStart(msg)
	default:
		resp := tgbotapi.NewMessage(msg.Chat.ID, "Неизвестная команда")
		if _, err := t.api.Send(resp); err != nil {
			log.Printf("telegram send error: %v", err)
		}
	}
}

func (t *TelegramBot) handleStart(msg *tgbotapi.Message) {
	webAppURL := t.app.cfg.AppBaseURL
	if webAppURL == "" {
		webAppURL = "/app"
	}

	webApp := tgbotapi.WebAppInfo{URL: webAppURL}
	button := tgbotapi.KeyboardButton{
		Text:   "Открыть музыкальный плеер",
		WebApp: &webApp,
	}
	markup := tgbotapi.ReplyKeyboardMarkup{
		Keyboard:       [][]tgbotapi.KeyboardButton{{button}},
		ResizeKeyboard: true,
	}

	text := "Привет! Нажми кнопку ниже, чтобы открыть музыкальный плеер."
	resp := tgbotapi.NewMessage(msg.Chat.ID, text)
	resp.ReplyMarkup = markup

	if _, err := t.api.Send(resp); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

func (a *App) handleTelegramWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if a.tgBot == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	var upd tgbotapi.Update
	if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
		log.Printf("telegram webhook decode error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	a.tgBot.HandleUpdate(upd)
	w.WriteHeader(http.StatusOK)
}
