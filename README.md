# 📱 SMSGate Server

[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stars][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![License][license-shield]][license-url]

Backend for the SMSGate ecosystem: a REST API that dispatches SMS through connected Android devices, with the optional private deployment.

## 📚 Table of Contents

- [📱 SMSGate Server](#-smsgate-server)
  - [📚 Table of Contents](#-table-of-contents)
  - [📖 About](#-about)
  - [⭐ Features](#-features)
  - [📦 Prerequisites](#-prerequisites)
  - [🚀 Quickstart](#-quickstart)
  - [⚙️ Configuration](#️-configuration)
  - [🔐 Authentication](#-authentication)
  - [🔌 API Overview](#-api-overview)
  - [📚 Documentation](#-documentation)
  - [🤝 Contributing](#-contributing)
  - [🗺️ Roadmap](#️-roadmap)
  - [📞 Contact](#-contact)
  - [⚖️ License](#️-license)
  - [📜 Legal Notice](#-legal-notice)

## 📖 About

SMSGate Server is the backend of the SMSGate ecosystem. It accepts SMS dispatch requests through a REST API, routes them to connected Android devices over Firebase Cloud Messaging, and tracks delivery state. It runs in two modes: public (anonymous device registration, used at [api.sms-gate.app](https://api.sms-gate.app)) and private (token-protected registration, push relayed through the upstream). Deep docs: https://docs.sms-gate.app/.

## ⭐ Features

- Text, data, and scheduled SMS dispatch
- Message status tracking and cancellation
- Device management (list, delete, online state)
- Health check endpoints (live, ready, startup)
- JWT authentication with scopes and token refresh
- OTP-based device registration
- Inbox, settings, and logs APIs
- Public and private deployment modes
- MySQL 8.0.13+ / MariaDB 10.2.7+ storage (MariaDB LTS recommended)

## 📦 Prerequisites

- MySQL 8.0.13+ or MariaDB 10.2.7+ database (MariaDB LTS recommended)
- Docker + Docker Compose for container setup
- Go 1.25.8+ for building from source

## 🚀 Quickstart

1. Create `configs/config.yml` from [configs/config.example.yml](configs/config.example.yml).
2. For private mode set `gateway.mode: private` and `gateway.private_token`.
3. Start the server:

```bash
docker run -p 3000:3000 \
  -v ./configs/config.yml:/app/config.yml \
  ghcr.io/android-sms-gateway/server:latest
```

Or with Compose (backend + background worker + MariaDB):

```bash
docker compose -f deployments/docker-compose/docker-compose.yml up --build
```

The worker handles background tasks: message hashing, message/device/token cleanup. Run it as a separate process or via the `worker` subcommand:

```bash
./sms-gateway worker
```

Local development:

```bash
make run        # go run ./cmd/sms-gateway/main.go
make air        # hot-reload dev server (TZ=UTC DEBUG=1)
make db-upgrade # apply migrations
```

## ⚙️ Configuration

Configuration lives in [configs/config.example.yml](configs/config.example.yml); every key can be overridden by env vars using `SECTION__KEY` (e.g. `DATABASE__HOST`, `GATEWAY__MODE`). Key sections: `database`, `gateway`, `http`, `fcm`, `sse`, `messages`, `cache`, `pubsub`, `jwt`, `otp`, `tasks`.

```bash
export GATEWAY__MODE=private
export GATEWAY__PRIVATE_TOKEN=change-me
export DATABASE__HOST=localhost
export HTTP__LISTEN=0.0.0.0:3000
```

### Worker Tasks

The background worker runs periodic tasks controlled by the `tasks` config section:

| Task               | Default Interval | Purpose                                                                 |
| ------------------ | ---------------- | ----------------------------------------------------------------------- |
| `messages_hashing` | `168h` (7 days)  | Hash processed messages to avoid plain text storage                     |
| `messages_cleanup` | `24h`            | Delete messages older than `max_age` (default `720h` / 30 days)         |
| `devices_cleanup`  | `24h`            | Remove inactive devices older than `max_age` (default `8760h` / 1 year) |
| `tokens_cleanup`   | `24h`            | Revoke expired tokens past `max_age` grace (default `1h`)               |

## 🔐 Authentication

The API supports Basic auth and JWT bearer tokens. JWT tokens carry scopes and are issued per user:

- `POST /api/3rdparty/v1/auth/token` - issue access/refresh pair (Basic auth)
- `POST /api/3rdparty/v1/auth/token/refresh` - rotate access token (Bearer refresh)
- `DELETE /api/3rdparty/v1/auth/token/{jti}` - revoke token (Basic auth)

Available scopes: `messages:send`, `messages:list`, `messages:read`, `messages:export`, `messages:cancel`, `devices:list`, `devices:delete`, `inbox:list`, `inbox:refresh`, `logs:read`, `settings:read`, `settings:write`, `tokens:manage`, `tokens:refresh`, `webhooks:list`, `webhooks:write`, `webhooks:delete`.

Full reference: [integration/authentication](https://docs.sms-gate.app/integration/authentication/).

## 🔌 API Overview

| Group    | Base path                                              |
| -------- | ------------------------------------------------------ |
| Messages | `/api/3rdparty/v1/messages`                            |
| Devices  | `/api/3rdparty/v1/devices`                             |
| Webhooks | `/api/3rdparty/v1/webhooks`                            |
| Health   | `/api/3rdparty/v1/health[/live \| /ready \| /startup]` |
| Auth     | `/api/3rdparty/v1/auth/token`                          |

Also: `/api/3rdparty/v1/inbox`, `/settings`, `/logs`. OpenAPI schema is served when `http.openapi.enabled: true`.

## 📚 Documentation

- [Private server setup](https://docs.sms-gate.app/getting-started/private-server/)
- [API integration](https://docs.sms-gate.app/integration/api/)
- [Authentication](https://docs.sms-gate.app/integration/authentication/)
- [Webhooks](https://docs.sms-gate.app/features/webhooks/)

## 🤝 Contributing

Open an issue first, then submit a PR. Run `make lint` and `make test` locally.

## 🗺️ Roadmap

Track planned work and known issues in the [issue tracker](https://github.com/android-sms-gateway/server/issues).

## 📞 Contact

- **Email:** [support@sms-gate.app](mailto:support@sms-gate.app)
- **Documentation:** [docs.sms-gate.app](https://docs.sms-gate.app/)
- **Issues:** [github.com/android-sms-gateway/server/issues](https://github.com/android-sms-gateway/server/issues)

## ⚖️ License

Apache-2.0. See [LICENSE](LICENSE).

## 📜 Legal Notice

Android is a trademark of Google LLC.

[contributors-shield]: https://img.shields.io/github/contributors/android-sms-gateway/server?style=for-the-badge
[contributors-url]: https://github.com/android-sms-gateway/server/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/android-sms-gateway/server?style=for-the-badge
[forks-url]: https://github.com/android-sms-gateway/server/network/members
[stars-shield]: https://img.shields.io/github/stars/android-sms-gateway/server?style=for-the-badge
[stars-url]: https://github.com/android-sms-gateway/server/stargazers
[issues-shield]: https://img.shields.io/github/issues/android-sms-gateway/server?style=for-the-badge
[issues-url]: https://github.com/android-sms-gateway/server/issues
[license-shield]: https://img.shields.io/github/license/android-sms-gateway/server?style=for-the-badge
[license-url]: https://github.com/android-sms-gateway/server/blob/master/LICENSE
