#!/bin/bash

# ==============================================================================
# VENTO.AI - Producer Backend Core (Hexagonal Architecture)
# ==============================================================================
# Script para levantar el entorno de desarrollo del Backend Core.
# Incluye validaciones de dependencias y estética corporativa.
# ==============================================================================

# --- Setup Colors & Formatting (Vento.ai Brand Palette) ---
C_RESET="\e[0m"
C_BOLD="\e[1m"
C_DIM="\e[2m"

COLOR_PRIMARY="\e[38;5;63m"
COLOR_SUCCESS="\e[38;5;41m"
COLOR_WARNING="\e[38;5;214m"
COLOR_ERROR="\e[38;5;196m"
COLOR_MUTED="\e[38;5;60m"

# --- Helper Functions ---

log_step() {
  echo -e "${COLOR_PRIMARY}${C_BOLD}▶${C_RESET} ${C_BOLD}$1${C_RESET}"
}

log_success() {
  echo -e "${COLOR_SUCCESS}${C_BOLD}✔${C_RESET}  ${C_DIM}$1${C_RESET}"
}

log_warning() {
  echo -e "${COLOR_WARNING}${C_BOLD}⚠${C_RESET}  ${C_WARNING}$1${C_RESET}"
}

log_error() {
  echo -e "${COLOR_ERROR}${C_BOLD}✖${C_RESET}  ${C_BOLD}$1${C_RESET}"
}

header() {
  clear
  echo -e "${COLOR_PRIMARY}${C_BOLD}"
  echo "  ██╗   ██╗███████╗███╗   ██╗████████╗██████╗     █████╗ ██╗"
  echo "  ██║   ██║██╔════╝████╗  ██║╚══██╔══╝██╔═══██╗  ██╔══██╗██║"
  echo "  ██║   ██║█████╗  ██╔██╗ ██║   ██║   ██║   ██║  ███████║██║"
  echo "  ╚██╗ ██╔╝██╔══╝  ██║╚██╗██║   ██║   ██║   ██║  ██╔══██║██║"
  echo "   ╚████╔╝ ███████╗██║ ╚████║   ██║   ╚██████╔╝  ██║  ██║██║"
  echo "    ╚═══╝  ╚══════╝╚═╝  ╚═══╝   ╚═╝    ╚═════╝   ╚═╝  ╚═╝╚═╝"
  echo -e "${C_RESET}"
  echo -e "  ${C_BOLD}Producer Backend Core - Development Environment${C_RESET}"
  echo -e "  ${COLOR_MUTED}Initializing hexagonal architecture services...${C_RESET}"
  echo ""
}

check_command() {
  if ! $1 version &> /dev/null; then
    log_error "Dependency missing: $1 is not installed or not working."
    echo -e "  Please install $1 to continue."
    exit 1
  fi
}

# --- Main Script ---

header

log_step "Validating system requirements..."

check_command "go"
GO_VERSION=$(go version | awk '{print $3}')
log_success "Go detected: $GO_VERSION"

check_command "docker compose"
log_success "Docker Compose detected: $(docker compose version --short)"

echo ""
log_step "Starting infrastructure (PostgreSQL)..."
if docker compose -f ../../devops/docker-compose.yml up -d; then
  log_success "Database is up and running."
else
  log_error "Failed to start infrastructure."
  exit 1
fi

echo ""
log_step "Starting Vento.ai Producer Backend (Go)..."
echo -e "${COLOR_MUTED}The server will be available at: ${COLOR_PRIMARY}http://localhost:8080${C_RESET}"
echo ""

# Run the Go server
go run cmd/server/main.go
