#!/usr/bin/env bash
set -euo pipefail

COMPOSE="podman-compose -f $(dirname "$0")/podman-compose.yml"

usage() {
  cat <<EOF
Usage: $(basename "$0") <command> [service]

Commands:
  up          Start all services (no rebuild)
  down        Stop all services
  restart     Restart all services (no rebuild)

  build       Rebuild a service (or all with --all)
  deploy      Rebuild + restart a service (or all with --all)

  logs        Tail logs (or specific service)
  status      Show container status

Examples:
  $(basename "$0") deploy frontend    # rebuild frontend + restart
  $(basename "$0") deploy backend     # rebuild backend + restart
  $(basename "$0") deploy             # rebuild everything + restart
  $(basename "$0") build frontend     # rebuild frontend only
  $(basename "$0") up                 # start everything
  $(basename "$0") logs backend       # tail backend logs
  $(basename "$0") down               # stop everything
EOF
}

cmd_up()       { $COMPOSE up -d; }
cmd_down()     { $COMPOSE down; }
cmd_restart()  { $COMPOSE down && $COMPOSE up -d; }
cmd_status()   { $COMPOSE ps; }
cmd_logs()     { $COMPOSE logs -f --tail=100 "${1:-}"; }

cmd_build() {
  local svc="${1:-}"
  if [[ -z "$svc" || "$svc" == "--all" ]]; then
    $COMPOSE build
  else
    $COMPOSE build --no-cache "$svc"
  fi
}

cmd_deploy() {
  local svc="${1:-}"
  if [[ -z "$svc" || "$svc" == "--all" ]]; then
    $COMPOSE build && $COMPOSE up -d
  else
    $COMPOSE build --no-cache "$svc" && $COMPOSE up -d
  fi
}

[[ $# -lt 1 ]] && { usage; exit 1; }

case "$1" in
  up)       cmd_up ;;
  down)     cmd_down ;;
  restart)  cmd_restart ;;
  build)    cmd_build "${2:-}" ;;
  deploy)   cmd_deploy "${2:-}" ;;
  logs)     cmd_logs "${2:-}" ;;
  status)   cmd_status ;;
  *)        usage; exit 1 ;;
esac
