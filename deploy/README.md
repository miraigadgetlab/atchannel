# @channel deployment

Production stack: PostgreSQL 17 + Redis 7 + Go API + nginx (static frontend + reverse proxy).

## Layout

- `podman-compose.yml` — full stack for podman-compose users.
- `Dockerfile.backend` / `Dockerfile.frontend` — image builds (context = repo root).
- `nginx.conf` — nginx server config (also embedded in the frontend image).
- `.env.example` — copy to `.env` for podman-compose.
- `quadlet/` — systemd Quadlet units (pod, network, containers, `.build`, one-shot migration).
  - `atchannel.env.example` — copy to `~/.config/containers/atchannel.env`.

## podman-compose

```sh
cp deploy/.env.example deploy/.env
# edit deploy/.env: POSTGRES_PASSWORD, JWT_SECRET, domain names
podman-compose -f deploy/podman-compose.yml up -d --build
```

## Quadlet (systemd)

Clone the repo to `/opt/atchannel`, then:

```sh
mkdir -p ~/.config/containers/systemd
cp deploy/quadlet/atchannel.env.example ~/.config/containers/atchannel.env
# edit ~/.config/containers/atchannel.env (fill DATABASE_URL in full)
cp deploy/quadlet/*.{pod,network,container,build} ~/.config/containers/systemd/
systemctl --user daemon-reload
systemctl --user start atchannel-migrate.service   # one-shot: run migrations
systemctl --user start atchannel-pod.service       # starts postgres, redis, backend, frontend
```

Notes:

- Rootless Quadlet requires `podman` from the distro; user services persist across
  logins via `loginctl enable-linger $USER`.
- The backend/migrate images are built by the `atchannel-backend.build` unit; the
  frontend by `atchannel-frontend.build`. Builds run once via one-shot services and
  are cached (see `systemctl --user list-units '*build*'`).
- `ATCHANNEL_ENV` secrets live in one env file; systemd does **no** variable
  interpolation, so write `DATABASE_URL` out in full.

## Upgrades

```sh
systemctl --user start atchannel-migrate.service
systemctl --user restart atchannel-pod.service
```

## TLS

The nginx server listens on 80/443 as plain HTTP (TLS termination is left to
external cert management, e.g. a reverse proxy or certbot). `COOKIE_SECURE=true`
requires HTTPS upstream; if terminating elsewhere, keep that front in front.
