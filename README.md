## Telegram Mini App Spotify Player (Go)

Мини‑приложение для Telegram (c использованием WebApp) с музыкальным плеером, которое управляет вашим Spotify через Web API и Spotify Connect.

### Архитектура

- **Go‑бэкенд**
  - HTTP‑сервер (`main.go`, `app.go`) с REST‑эндпоинтами для mini app и обработкой вебхука Telegram
  - Интеграция с Telegram Bot API (`telegram.go`) — команда `/start` и кнопка WebApp
  - Интеграция со Spotify (`spotify.go`, `auth.go`, `api.go`) — OAuth2, плейлисты, поиск, управление плеером
- **Frontend mini app**
  - Простая страница `static/index.html`, открываемая внутри Telegram через WebApp‑кнопку. Страница полностью тестовая, мб позже доработаю, пока не до этого --(0)

### Требования


- Spotify аккаунт и доступ к Spotify Developer Dashboard(для подключения айди из спотика)
- Telegram Bot Token (через `@BotFather`)
- Публично доступный HTTPS‑URL (прод) или туннель (`ngrok`, `cloudflared`) для вебхука Telegram и Spotify redirect

### Переменные окружения

Обязательные для нормальной работы:

- **`TELEGRAM_BOT_TOKEN`** — токен бота от BotFather
- **`SPOTIFY_CLIENT_ID`** — Client ID из Spotify Developer Dashboard
- **`SPOTIFY_CLIENT_SECRET`** — Client Secret из Spotify
- **`SPOTIFY_REDIRECT_URL`** — redirect URI, который нужно указать в настройках Spotify‑приложения
- **`APP_BASE_URL`** — публичный URL mini app (страницы `/app`), который будет открыт из Telegram



### Установка и запуск

```bash
cd /home/chealix/GolandProjects/chealixjj
go mod tidy      
go run ./...
```

По умолчанию сервер будет слушать `http://localhost:8080`


### Настройка Spotify приложения

1. Зайдите в [Spotify Developer Dashboard](https://developer.spotify.com/dashboard).
2. Создайте новое приложение
3. В настройках приложения, сохраните **Client ID** и **Client Secret** и и нужно прописать их в переменные окружения:


### Настройка Telegram бота и WebApp

1. Создайте бота через `@BotFather` в Telegram.
2. Получите `TELEGRAM_BOT_TOKEN` и задайте его в окружении:



### Маршруты бэкенда

- `GET /healthz` — проверка работоспособности сервера.
- `POST /telegram/webhook` — вебхук Telegram (на него указывает `setWebhook`).
- `GET /auth/spotify/login?tg_user_id=...` — начало Spotify OAuth2 для Telegram‑пользователя.
- `GET /auth/spotify/callback` — callback от Spotify, сохраняет токены и закрывает вкладку.
- `GET /api/me?tg_user_id=...` — статус авторизации: привязан ли Spotify.
- `GET /api/playlists?tg_user_id=...` — плейлисты пользователя.
- `GET /api/playlist_tracks?id=...&tg_user_id=...` — треки плейлиста.
- `GET /api/search?q=...&tg_user_id=...` — поиск треков.
- `GET /api/player/current?tg_user_id=...` — текущий трек.
- `POST /api/player/pause|next|previous?tg_user_id=...` — управление плеером через Spotify Connect.
- `GET /app` — страница mini app (`static/index.html`).
- `GET /static/*` — статика мини‑приложения.

Привязка к Telegram‑пользователю сейчас реализована через `tg_user_id` в query‑параметрах. В mini app он берётся из `Telegram.WebApp.initDataUnsafe.user.id`.

### Как работает mini app

- В Telegram пользователь пишет `/start` боту.
- Бот отвечает сообщением с клавиатурой, где есть кнопка WebApp “Открыть музыкальный плеер”.
- При нажатии на кнопку открывается встроенный браузер Telegram с URL `/app`.
- Страница `static/index.html`:
  - Инициализирует `Telegram.WebApp`, читает `user.id` → `tg_user_id`.
  - Делает запрос к `/api/me`, чтобы понять, привязан ли Spotify.
  - Если Spotify не привязан — показывает кнопку “Войти через Spotify”, которая открывает `/auth/spotify/login?tg_user_id=...` во внешнем окне.
  - После успешной авторизации в Spotify, вкладка закрывается, и пользователь видит обновлённый статус в mini app.
  - Mini app может:
    - Загружать плейлисты (`/api/playlists`).
    - Делать поиск треков (`/api/search`).
    - Показывать текущий трек и управлять им (pause/next/previous) через Spotify Connect.

    


