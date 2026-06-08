# Roket.Chat-GeekBot 🤖

> A Go-powered Rocket.Chat bot for team daily standups. Manage teams, collect standup reports, and keep everyone aligned — all from your Rocket.Chat channels with slash commands.

![Go version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Rocket.Chat](https://img.shields.io/badge/Rocket.Chat-bot-red?logo=rocket.chat)

## Features

- **Team daily standups** — collect and report standup updates via slash commands
- **Multi-team support** — run standups for multiple teams independently
- **Role-based access** — main admin, team leads, and team members with scoped permissions
- **Configurable per team** — schedule (cron), questions, timezone, channel
- **SQLite persistence** — no external database required
- **GitHub Codespaces ready** — one-click dev environment with prebuilds

## Commands

### Main Admin
| Command | Description |
|---------|-------------|
| `/standup team create <name>` | Create a new team |
| `/standup team delete <name>` | Delete a team |
| `/standup team set-lead <team> @user` | Set team lead |

### Team Lead
| Command | Description |
|---------|-------------|
| `/standup team add @user` | Add member to team |
| `/standup team remove @user` | Remove member from team |
| `/standup team set schedule <cron>` | Set standup schedule |
| `/standup team set channel #channel` | Set report channel |
| `/standup team set questions <q1\|q2\|q3>` | Set custom questions |
| `/standup team set timezone <tz>` | Set team timezone |
| `/standup team members` | List team members |

### Team Members
| Command | Description |
|---------|-------------|
| `/standup submit <answer1\|answer2\|answer3>` | Submit daily standup |
| `/standup status` | Check if you've submitted today |

### All Users
| Command | Description |
|---------|-------------|
| `/standup help` | Show available commands |
| `/standup report` | View latest standup report for your team |

## Quick Start

### Prerequisites
- A Rocket.Chat server with bot user credentials
- Go 1.22+ (or use the Codespaces devcontainer)

### Configuration

Create a `config.yaml` file or set environment variables:

```yaml
server_url: "https://chat.yourcompany.com"
bot_username: "geekbot"
bot_password: "your-password"
main_admin: "admin_username"  # Rocket.Chat username of the main admin
```

### Run

```bash
go run ./cmd/bot
```

## Development with Codespaces

Click the button below to start a preconfigured dev environment:

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://github.com/codespaces/new?hide_repo_select=true&ref=main&repo=sunba91-su/Roket.Chat-GeekBot)

The devcontainer includes:
- Go 1.22+ toolchain
- VS Code Go extensions (linting, debugging, test explorer)
- Preconfigured 4-core machine with 30min idle timeout

## Project Structure

```
.
├── .devcontainer/          # Codespaces devcontainer config
├── .github/workflows/      # CI pipeline
├── cmd/bot/                # Application entry point
├── internal/
│   ├── config/             # Configuration loading
│   ├── rocket/             # Rocket.Chat SDK client
│   ├── commands/           # Slash command framework
│   ├── standup/            # Standup business logic
│   └── store/              # SQLite persistence
├── .env.example            # Environment variable template
└── go.mod                  # Go module definition
```

## Architecture

```
Rocket.Chat Server
      ↕ WebSocket + REST
Roket.Chat-GeekBot
  ├── Command Router
  ├── Team Manager (roles, members)
  ├── Standup Collector
  ├── Report Generator
  └── SQLite Store
```

## Roadmap

- [x] Team daily standup reports
- [ ] Scheduled standup reminders
- [ ] Standup history and trends
- [ ] Web dashboard
- [ ] Integration with GitHub, Jira, and other tools
- [ ] Standup notifications via DM

## License

MIT
