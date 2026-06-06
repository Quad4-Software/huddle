# Huddle

Self-hosted voice, video, and chat for small groups. Create a room, share a link, talk.

The server handles signaling and access control. Media and encrypted messages go peer-to-peer over WebRTC.

## Features

- Voice and screen sharing
- End-to-end encrypted text chat and file sharing
- Invite links with optional room passwords
- Host controls (kick members)
- Proof-of-work on room create/join to slow down bots
- Single binary deployment

## Quick start

**Docker**

```bash
cp .env.example .env
# set HUDDLE_INVITE_SECRET to a long random string
docker compose up -d --build
```

Open `http://localhost:8080`.

**From source**

Requires Go 1.26+, Node 20+, and pnpm 11+.

```bash
task run
```

For local development with hot reload:

```bash
task dev
```

Backend on `:8080`, frontend on `:5173` (proxied to the API).

## Configuration

| Variable | Description |
|----------|-------------|
| `HUDDLE_INVITE_SECRET` | Secret for signing invite tokens. Required in production. |
| `HUDDLE_TRUST_PROXY` | Trust `X-Forwarded-*` headers when behind a reverse proxy. |
| `HUDDLE_PORT` | Host port for Docker Compose. |

Useful server flags:

| Flag | Default | Description |
|------|---------|-------------|
| `-max-room-size` | `12` | Max peers per room |
| `-pow-difficulty` | `18` | Proof-of-work bits (`0` disables) |
| `-rate-limit-create` | `10` | Room creates per IP per minute |

Put TLS in front of the app in production. See `deploy/reverse-proxy.conf.example`.

## Development

```bash
task test       # Go + frontend tests
task check      # format, lint, test, build
task security   # gosec, trivy, pnpm audit
```

## License

Apache 2.0. See [LICENSE](LICENSE).

Copyright 2026 [Quad4.io](https://quad4.io).
