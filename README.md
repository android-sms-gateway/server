# 📱 SMSGate Server

[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stars][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![License][license-shield]][license-url]

SMSGate Server is the Go backend for the SMSGate ecosystem. It provides a REST API for dispatching SMS through registered Android devices, tracks message state, receives incoming inbox messages, and exposes device, webhook, health, and metrics operations.

## 📚 Table of Contents

- [📱 SMSGate Server](#-smsgate-server)
  - [📚 Table of Contents](#-table-of-contents)
  - [📖 About the Project](#-about-the-project)
    - [Gateway modes](#gateway-modes)
  - [⭐ Features](#-features)
  - [🚀 Getting Started](#-getting-started)
    - [Prerequisites](#prerequisites)
  - [📦 Installation](#-installation)
    - [Docker Compose](#docker-compose)
    - [Standalone container](#standalone-container)
    - [Build from source](#build-from-source)
  - [📖 Usage](#-usage)
    - [Health check](#health-check)
    - [API groups](#api-groups)
    - [Authentication](#authentication)
    - [Documentation](#documentation)
  - [⚙️ Configuration](#️-configuration)
    - [Important settings](#important-settings)
    - [Worker tasks](#worker-tasks)
  - [🚀 Deployment](#-deployment)
  - [🚧 Development](#-development)
  - [🤝 Contributing](#-contributing)
  - [🗺️ Roadmap](#️-roadmap)
  - [📞 Contact](#-contact)
  - [⚖️ License](#️-license)
  - [📜 Legal Notice](#-legal-notice)

## 📖 About the Project

SMSGate Server connects API clients with Android devices. A client submits a message to the server; the server stores it, sends a push event through Firebase Cloud Messaging (FCM) or the configured upstream service, and records the resulting delivery states. Devices pull pending work and report status back to the server.

The server also provides device registration and management, incoming SMS/data-SMS/MMS synchronization, encrypted device uploads, attachment downloads, webhooks, scoped authentication, server-sent events, and Prometheus metrics. It is intended for self-hosted operators, API integrators, Android client developers, and contributors.

### Gateway modes

| Mode      | Device registration                                             | Push delivery               | Required configuration                                                                   |
| --------- | --------------------------------------------------------------- | --------------------------- | ---------------------------------------------------------------------------------------- |
| `public`  | New devices may register anonymously                            | FCM                         | Valid `FCM__CREDENTIALS_JSON`                                                            |
| `private` | New anonymous devices must present the configured private token | Configured upstream service | Non-empty `GATEWAY__PRIVATE_TOKEN`; `GATEWAY__UPSTREAM_URL` defaults to SMSGate upstream |

The third-party API remains authenticated in both modes. Private mode is not an offline mode; its default upstream is the SMSGate upstream service.

## ⭐ Features

- Send text or binary-data SMS with scheduling, TTL, SIM selection, priority, device selection, and delivery reports.
- Track message state and history with date, state, device, content, sorting, and pagination options.
- Cancel pending messages and receive a conflict response when a message is no longer pending.
- Register and manage Android devices, including SIM metadata, online state, and last-seen data.
- Receive encrypted incoming SMS, data-SMS, and MMS batches from devices; list and download inbox messages and attachments.
- Configure user- and device-scoped webhooks for application events.
- Authenticate API clients with Basic auth or scoped JWT access and refresh tokens; use OTP-based device registration where enabled.
- Query liveness, readiness, and startup probes; expose SSE events and Prometheus metrics.
- Run the backend and a separate background worker in public or private deployment mode.
- Store data in MySQL 8.0.13+ or MariaDB 10.2.7+; MariaDB LTS is recommended.

Outgoing enqueue requests support text and data messages. Incoming MMS messages and attachments are supported by the inbox API.

## 🚀 Getting Started

### Prerequisites

Choose one of the following environments:

- **Source:** Go 1.25.8+ and a MySQL 8.0.13+ or MariaDB 10.2.7+ database.
- **Containers:** Docker with Docker Compose and a reachable database, or the bundled MariaDB service.
- **Push delivery:** either valid Firebase service-account JSON for public mode or a private token and reachable upstream service for private mode.

For a first local setup, copy the tracked templates and edit the mode-specific values:

```bash
cp configs/config.example.yml configs/config.yml
cp .env.example .env
```

`configs/config.yml` and `.env` are local files and are ignored by Git. Select exactly one gateway mode before starting the server:

```dotenv
# Private mode
GATEWAY__MODE=private
GATEWAY__PRIVATE_TOKEN=your-private-token

# Public mode
# GATEWAY__MODE=public
# FCM__CREDENTIALS_JSON=compact-service-account-json
```

The environment template leaves mode-dependent credentials blank. Do not use example values as production secrets.

## 📦 Installation

### Docker Compose

The standard Compose deployment starts the backend, background worker, and MariaDB. It mounts `configs/config.yml` and overrides the server listen address for the container network:

```bash
docker compose -f deployments/docker-compose/docker-compose.yml up --build
```

The backend container applies database migrations when it starts with its default command. The worker runs the same binary with the `worker` subcommand and does not apply migrations.

### Standalone container

The standalone image needs a reachable database and a mounted configuration file. Pass `.env` explicitly, and override the repository-local configuration path and loopback listen address for the container:

```bash
docker run --rm \
  --env-file .env \
  -e CONFIG_PATH=/app/config.yml \
  -e HTTP__LISTEN=0.0.0.0:3000 \
  -p 3000:3000 \
  -v "$PWD/configs/config.yml:/app/config.yml:ro" \
  ghcr.io/android-sms-gateway/server:latest
```

### Build from source

```bash
make deps
make build
./tmp/sms-gateway
```

The build target writes the binary to `tmp/sms-gateway`. The `make install` target installs the command on the Go binary path. Run the background worker as a separate process:

```bash
go run ./cmd/sms-gateway/main.go worker
# or, after make build
./tmp/sms-gateway worker
```

## 📖 Usage

### Health check

The backend exposes service probes at the root path:

```bash
curl http://localhost:3000/health/ready
```

Available probes are `/health`, `/health/live`, `/health/ready`, and `/health/startup`. Prometheus metrics are exposed at `/metrics` when the metrics module is enabled by the service.

### API groups

The default public path is `/api`. It can be changed with `HTTP__API__PATH`.

| Group             | Default paths                     | Purpose                                                                                               |
| ----------------- | --------------------------------- | ----------------------------------------------------------------------------------------------------- |
| Third-party API   | `/api/3rdparty/v1`                | Messages, devices, inbox, webhooks, settings, authentication, and health                              |
| Mobile/device API | `/api/mobile/v1`                  | Device registration, pending messages, state reporting, settings, webhooks, events, and inbox uploads |
| Upstream API      | `/api/upstream/v1`                | Public-mode push enqueue endpoint used by the upstream integration                                    |
| Root operations   | `/health*`, `/metrics`            | Probes and metrics                                                                                    |
| API discovery     | `/.well-known/api-catalog`        | Links to the API description, documentation, and status endpoint                                      |
| OpenAPI           | `/api/docs`, `/api/docs/doc.json` | Served when `HTTP__OPENAPI__ENABLED=true`                                                             |

The `GET /api/3rdparty/v1/logs` route is reserved but currently returns `501 Not Implemented` because device logs are not exposed by the cloud server.

### Authentication

Third-party endpoints use Basic authentication or JWT bearer tokens. JWT tokens are issued and refreshed per user and carry scopes. When `JWT__SECRET` is empty, JWT functionality is disabled; a configured secret must contain at least 32 bytes. Access and refresh token lifetimes must be positive, and the refresh lifetime must be longer than the access lifetime.

Relevant token endpoints include:

- `POST /api/3rdparty/v1/auth/token` — issue an access/refresh pair with Basic auth.
- `POST /api/3rdparty/v1/auth/token/refresh` — refresh an access token with a refresh bearer token.
- `DELETE /api/3rdparty/v1/auth/token/{jti}` — revoke a token with Basic auth.

Available scopes include `messages:send`, `messages:list`, `messages:read`, `messages:export`, `messages:cancel`, `devices:list`, `devices:delete`, `inbox:list`, `inbox:refresh`, `logs:read`, `settings:read`, `settings:write`, `tokens:manage`, `tokens:refresh`, `webhooks:list`, `webhooks:write`, and `webhooks:delete`.

### Documentation

- [Private server setup](https://docs.sms-gate.app/getting-started/private-server/)
- [API integration](https://docs.sms-gate.app/integration/api/)
- [Authentication](https://docs.sms-gate.app/integration/authentication/)
- [Webhooks](https://docs.sms-gate.app/features/webhooks/)

## ⚙️ Configuration

The tracked templates are:

- [YAML configuration example](configs/config.example.yml)
- [Environment variable template](.env.example)

The configuration loader reads `.env`, then the YAML file selected by `CONFIG_PATH` (or `config.yml` by default), and then applies tagged environment variables. Effective precedence is:

```text
existing process environment > .env > YAML > code defaults
```

Nested YAML keys use double underscores for environment overrides, for example `DATABASE__HOST`, `MESSAGES__QUEUE__MAX_PENDING`, and `TASKS__INBOX_CLEANUP__MAX_AGE`.

### Important settings

| Area              | Settings                                                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| Gateway           | `GATEWAY__MODE`, `GATEWAY__PRIVATE_TOKEN`, `GATEWAY__UPSTREAM_URL`                                                                               |
| HTTP              | `HTTP__LISTEN`, `HTTP__PROXIES`, `HTTP__API__HOST`, `HTTP__API__PATH`, `HTTP__OPENAPI__ENABLED`                                                  |
| Database          | `DATABASE__HOST`, `DATABASE__PORT`, `DATABASE__USER`, `DATABASE__PASSWORD`, `DATABASE__DATABASE`, `DATABASE__TIMEZONE`, connection-pool settings |
| Push              | `FCM__CREDENTIALS_JSON`, `FCM__DEBOUNCE_SECONDS`, `FCM__TIMEOUT_SECONDS`                                                                         |
| Messages          | Cache, real-time hashing, queue limits, and queue-statistics settings                                                                            |
| Cache and pub/sub | `CACHE__URL`, `PUBSUB__URL`, `PUBSUB__BUFFER_SIZE`                                                                                               |
| Authentication    | `JWT__SECRET`, token lifetimes and issuer, plus OTP settings                                                                                     |
| Worker            | Message hashing/cleanup, device cleanup, token cleanup, and inbox cleanup intervals and retention                                                |

Keep private tokens, database passwords, JWT secrets, FCM credentials, and any Redis credentials out of committed files. `memory://` cache and pub/sub URLs are suitable for a single process; use a shared `redis://` deployment when running multiple backend replicas.

### Worker tasks

The worker runs the following periodic tasks:

| Task               | Default interval | Purpose                                                                         |
| ------------------ | ---------------- | ------------------------------------------------------------------------------- |
| `messages_hashing` | `168h`           | Hash processed message content and phone numbers to remove plaintext.           |
| `messages_cleanup` | `24h`            | Delete messages older than `max_age` (`720h` by default).                       |
| `devices_cleanup`  | `24h`            | Remove inactive devices older than `max_age` (`8760h` by default).              |
| `tokens_cleanup`   | `24h`            | Remove expired tokens after the `1h` grace period.                              |
| `inbox_cleanup`    | `24h`            | Remove inbox messages and attachments older than `max_age` (`720h` by default). |

The backend's `MESSAGES__HASHING_INTERVAL_SECONDS` setting controls real-time message hashing and is separate from the worker's `TASKS__MESSAGES_HASHING__INTERVAL` task schedule.

## 🚀 Deployment

The repository includes these deployment assets:

- [Docker Compose deployment](deployments/docker-compose/docker-compose.yml)
- [Docker Compose reverse-proxy deployment](deployments/docker-compose-proxy/README.md)
- [Helm chart](deployments/helm-chart/README.md)
- [Prometheus alerting rules](deployments/prometheus/server.rules.yml)

Use an external MySQL/MariaDB service or the bundled Compose database. Before exposing the API, configure the public API host, trusted proxies, authentication secrets, and HTTPS termination for the deployment environment.

## 🚧 Development

Common project commands are defined in the [Makefile](Makefile):

```bash
make gen          # generate code and OpenAPI documentation
make fmt          # format Go code
make lint         # run golangci-lint
make test         # run Go tests with race detection and coverage
make test-e2e     # run unit tests and the E2E module
make db-upgrade   # apply database migrations
make all          # generate, format, lint, and test
```

`make init-dev` installs development tools used by the Makefile, including Air, Swag, and Goose. The E2E suite uses its own Compose setup and may require a valid FCM credential when running the public push path.

## 🤝 Contributing

Open an issue before submitting a pull request so the change can be discussed. Keep generated files in sync when changing API annotations, and run the relevant formatting, linting, and test commands before opening the pull request.

## 🗺️ Roadmap

Track planned work and known issues in the [issue tracker](https://github.com/android-sms-gateway/server/issues).

## 📞 Contact

- **Email:** [support@sms-gate.app](mailto:support@sms-gate.app)
- **Documentation:** [docs.sms-gate.app](https://docs.sms-gate.app/)
- **Issues:** [github.com/android-sms-gateway/server/issues](https://github.com/android-sms-gateway/server/issues)

## ⚖️ License

Apache-2.0. See [LICENSE](LICENSE) and the [project changelog](CHANGELOG.md).

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
