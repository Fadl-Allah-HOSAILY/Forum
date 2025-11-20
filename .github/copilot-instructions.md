**Project Overview**
- **Type:** Go web application (single binary) using SQLite for persistence.
- **Entry point:** `main.go` — opens `forum.db`, initializes schema/tables, loads templates and registers HTTP routes.
- **Package layout:** `backend/` holds HTTP handlers, DB initialization and helper functions; `templates/` and `static/` hold frontend assets.

**Big Picture Architecture**
- **Routing:** Routes are registered in `main.go` with handlers provided by `backend` (e.g. `http.HandleFunc("/login", backend.LoginHandler(DB))`). Handlers are closures that capture a `*sql.DB`.
- **Persistence:** Uses SQLite file `forum.db`. Schema and PRAGMA settings are created in `backend/init.go`. The code sets `DB.SetMaxOpenConns(1)` and enables `PRAGMA foreign_keys = ON` and a busy timeout — do not remove these without understanding SQLite locking constraints.
- **Sessions:** Session state is stored in the `sessions` table. Cookie name: `session_token`. Session expiration is enforced in SQL queries (see `backend/login.go` and `backend/getUserIDFromSession.go`).
- **Templates & Static:** Templates are loaded via `backend.LoadTemplates("templates/*.html")`. Static files are served under `/static/` from the `static/` folder.

**Key Files to Read First**
- `main.go` — server boot, route registration, listen address (currently `10.1.4.6:8081`).
- `backend/init.go` — DB schema, PRAGMA, connection settings.
- `backend/login.go`, `backend/signup.go` — show typical handler patterns and session creation.
- `backend/middleware.go` / `backend/getUserIDFromSession.go` — auth helpers: `AuthRequired`, `NotAuthRequired`, and `GetUserIDFromRequest`.
- `backend/templates.go` / `backend/help.go` — template rendering helpers and common utilities.

**Common Patterns & Conventions**
- **Handler factory:** Handlers take `*sql.DB` and return `http.HandlerFunc`. Example: `func LoginHandler(DB *sql.DB) http.HandlerFunc { return func(w http.ResponseWriter, r *http.Request) { ... } }`.
- **Route registration:** Register routes in `main.go` with `http.HandleFunc("/path", backend.YourHandler(DB))`.
- **Template rendering:** Use `templates.ExecuteTemplate(w, "name.html", data)`; many handlers render `map[string]string{"Error": "..."}` for error messages.
- **DB operations:** Prefer prepared queries and check for `sql.ErrNoRows` explicitly (pattern already used in `login.go`).
- **Security checks already present:** URL path cleaning, request size/character checks, and bcrypt for password verification. Follow these patterns when adding new endpoints.

**Developer Workflows / Useful Commands**
- Build and run locally: `go run main.go` or `go build -o forum .` then `./forum`.
- Inspect DB: `sqlite3 forum.db` (or use a GUI SQLite browser) — schema created automatically by `backend/init.go` on startup.
- Module tidying: `go mod tidy` if you add/remove imports.
- Lint / vet: run `go vet ./...` or any preferred linter before committing.

**Integration & External Dependencies**
- SQLite driver: `github.com/mattn/go-sqlite3` (CGO required). Building on some systems may need `gcc` installed.
- Password hashing: `golang.org/x/crypto/bcrypt`.

**When Changing DB Code**
- Preserve PRAGMA and connection settings found in `backend/init.go` (`foreign_keys`, `busy_timeout`, `SetMaxOpenConns(1)`) unless you understand SQLite concurrency implications.
- Schema changes: Add migrations carefully — this project creates tables at startup. If you change a table schema, add a migration path or a safe ALTER flow.

**Examples**
- Add a new route/handler:
  - Create `backend/yourhandler.go` with `func YourHandler(DB *sql.DB) http.HandlerFunc { ... }`.
  - Register in `main.go`: `http.HandleFunc("/yourpath", backend.YourHandler(DB))`.
- Check logged-in user inside a handler: `uid := backend.GetUserIDFromRequest(DB, r)` (returns 0 when unauthenticated).

**Notes / Gotchas**
- The listen address and port are hard-coded in `main.go` (`10.1.4.6:8081`). Change it consciously (and update any docs) if you need `localhost` or a different host/port.
- No test suite detected — be conservative with structural changes and validate manually using the running server and `sqlite3` to inspect data.

If anything here looks incorrect or you'd like more examples (e.g., how to add API endpoints, a sample migration, or recommended linting commands), tell me which area to expand and I'll update this file.
