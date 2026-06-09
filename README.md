# Roket.Chat-GeekBot

> A Go-powered Rocket.Chat bot for team daily standups. Manage teams, collect standup reports conversationally via DM, and post formatted summaries to team channels — all with slash commands.

![Go version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Rocket.Chat](https://img.shields.io/badge/Rocket.Chat-bot-red?logo=rocket.chat)
![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker)

## Features

- **Conversational standups** — bot asks one question at a time via DM, collects answers in a natural flow
- **Multi-team support** — run independent standups for multiple teams
- **Role-based access** — main admin > team leads > team members with scoped permissions
- **Configurable per team** — custom questions, schedule (cron), timezone, and report channel
- **Formatted reports** — standup summaries posted to the team's configured channel with @mentions and emoji labels
- **SQLite persistence** — no external database required, single-file storage
- **Docker deployment** — multi-stage distroless image, docker-compose ready
- **Automatic reconnect** — exponential backoff on WebSocket disconnection

## Commands

| Command | Role | Description |
|---------|------|-------------|
| `/standup help` | All | Show available commands |
| `/standup team create <name>` | Admin | Create a new team |
| `/standup team delete <name>` | Admin | Delete a team |
| `/standup team set-lead <name> @user` | Admin | Set team lead |
| `/standup team add @user` | Lead | Add member to team |
| `/standup team remove @user` | Lead | Remove member from team |
| `/standup team set schedule <cron>` | Lead | Set standup schedule |
| `/standup team set channel #channel` | Lead | Set report channel |
| `/standup team set questions <q1\|q2\|q3>` | Lead | Set custom questions |
| `/standup team set timezone <tz>` | Lead | Set team timezone |
| `/standup team members` | Lead | List team members |
| `/standup submit` | Member | Start a standup submission (DM) |
| `/standup status` | Member | Check submission status |
| `/standup report` | Member | View latest team report |

## Quick Start

### Prerequisites

- A Rocket.Chat server with a bot user credential set (server URL, username, password)
- Go 1.22+ (for native development) or Docker (for containerized deployment)

### Configuration

Copy the environment template and fill in your Rocket.Chat bot credentials:

```bash
cp .env.example .env
```

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ROCKETCHAT_SERVER_URL` | Yes | — | Rocket.Chat server URL |
| `ROCKETCHAT_BOT_USERNAME` | Yes | — | Bot account username |
| `ROCKETCHAT_BOT_TOKEN` | One of | — | Personal Access Token (recommended — no special char issues) |
| `ROCKETCHAT_BOT_USER_ID` | One of | — | User ID for the PAT |
| `ROCKETCHAT_BOT_PASSWORD` | One of | — | Bot account password (alternative auth method) |
| `ROCKETCHAT_MAIN_ADMIN` | Yes | — | Rocket.Chat username of the main bot administrator |
| `STANDUP_DB_PATH` | No | `~/standup-bot.db` | Path to the SQLite database file |

### Run with Go

```bash
go run ./cmd/bot
```

### Run with Docker

```bash
# Build and start
make docker-run

# Or manually:
docker compose up -d --build

# View logs
docker compose logs -f

# Stop
docker compose down
```

The SQLite database is persisted in a named Docker volume (`bot-data`).

## Deployment

### Docker (recommended)

```bash
# Build the production image (~10 MB)
docker build -t geekbot .

# Run with your .env file
docker run -d \
  --name geekbot \
  --restart unless-stopped \
  --env-file .env \
  -v bot-data:/data \
  geekbot
```

### docker-compose

```bash
docker compose up -d --build
```

The compose file includes resource limits, automatic restart, and a persistent volume for the database.

### System requirements

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 0.1 core | 0.5 core |
| RAM | 32 MB | 128 MB |
| Disk | 100 MB | 500 MB |

The bot only makes outbound connections (WebSocket + REST) — no inbound ports are required.

## Project Structure

```
.
├── .devcontainer/          # GitHub Codespaces devcontainer
├── .github/workflows/      # CI pipeline (vet, build, test)
├── cmd/bot/                # Application entry point
├── internal/
│   ├── commands/           # Slash command registry and handlers
│   ├── config/             # Environment variable loading
│   ├── convstate/          # Conversation state manager (DM flow)
│   ├── rocket/             # Rocket.Chat realtime + REST client
│   └── store/              # SQLite persistence layer
├── Dockerfile              # Multi-stage distroless build
├── docker-compose.yml      # Docker Compose deployment
├── Makefile                # Build automation targets
└── SECURITY.md             # Security policy and disclosure
```

## Architecture

```
┌─────────────────────┐
│  Rocket.Chat Server │
└──────────┬──────────┘
           │ WebSocket + REST
┌──────────▼──────────┐
│   Roket.Chat-GeekBot  │
│  ┌──────────────┐   │
│  │ Command      │   │
│  │ Router       │   │
│  └──────┬───────┘   │
│  ┌──────▼───────┐   │
│  │ Team Manager │   │
│  │ (roles,      │   │
│  │  members)    │   │
│  └──────┬───────┘   │
│  ┌──────▼───────┐   │
│  │ Standup      │   │
│  │ Collector    │   │
│  └──────┬───────┘   │
│  ┌──────▼───────┐   │
│  │ Report       │   │
│  │ Generator    │   │
│  └──────┬───────┘   │
│  ┌──────▼───────┐   │
│  │ SQLite Store  │   │
│  └──────────────┘   │
└─────────────────────┘
```

## Development

```bash
# Build
make build

# Run tests
make test

# Lint
make vet

# Clean artifacts
make clean
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for branch strategy, commit conventions, and PR guidelines.

## Security

See [SECURITY.md](SECURITY.md) for the security policy and vulnerability disclosure process.

## License

MIT
