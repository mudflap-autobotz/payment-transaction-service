# CLAUDE

## Code Style Guidelines

* **No Comments:** Do not write any comments (inline or block) in the code.
* **Self-Documenting Code:** Write clean, readable code with descriptive variable and function names instead of adding explanatory text.
* **Exceptions:** Only include comments if absolutely required by compiler directives or specific framework configurations (e.g., eslint-disable).

## Principles

* **12 Factor App**
* **SOLID Principles**
* **ACID Properties**

## Architecture

* **hexagonal architecture**

## Project Structure

* `cmd/api/` — composition root (`main.go`, `wire.go`, `wire_gen.go`)
* `internal/domain/` — entities + repository/service interfaces (ports)
* `internal/service/` — application/use-case layer, implements the domain service port, depends only on the repository port
* `internal/repository/<entity>/` — Postgres adapters implementing the domain repository port, external API
* `internal/handler/` — inbound HTTP adapter (Fiber handlers)
* `internal/router/` — route registration
* `internal/middleware/` — auth, logger, recover, ratelimit
* `internal/dto/` — request/response transfer objects
* `internal/config/` — env-based config loading (cleanenv)
* `pkg/utils/` — shared helpers

## Tech Stack

* Go 1.26
* **Fiber v3** — HTTP framework
* **Bun + pgx** — Postgres ORM/driver
* **google/wire** — dependency injection
* **zerolog** — logging
* **OpenTelemetry** — tracing
* **cleanenv** — environment/config loading
* **go-playground/validator** — request validation
* **payment-common** (`github.com/mudflap-autobotz/payment-common`)
