# Vento.AI — Backend (Go)

Backend del ecosistema Vento.AI construido en **Go** con **Arquitectura Hexagonal**.

## Módulos

| Módulo | Descripción |
|---|---|
| `producer/vento_core` | Servicio principal: Auth, Users, API REST |

## Inicio Rápido

```bash
# 1. Levantar PostgreSQL
cd devops && docker compose up -d

# 2. Correr el servidor
cd producer/vento_core
go run cmd/server/main.go
```

El servidor arranca en `http://localhost:8080`.

## Endpoints (Phase 1)

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/v1/auth/register` | Registro de usuario |
| POST | `/api/v1/auth/login` | Login |
| GET | `/api/v1/auth/me` | Info del usuario autenticado |
