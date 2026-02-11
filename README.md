<p align="center">
  <img src="static/images/logo.svg" alt="SimpleDoc logo" width="120" />
</p>

<h1 align="center">SimpleDoc</h1>

<p align="center">
  A lightweight, self-hosted documentation platform for teams.<br/>
  Write in Markdown. Collaborate with roles. Deploy anywhere.
</p>

<p align="center">
  <a href="#features">Features</a> &middot;
  <a href="#quick-start">Quick Start</a> &middot;
  <a href="#deployment">Deployment</a> &middot;
  <a href="#configuration">Configuration</a> &middot;
  <a href="#license">License</a>
</p>

<p align="center">
  <img src="static/images/screenshot.png" alt="SimpleDoc screenshot" width="720" />
</p>

---

## Why SimpleDoc?

Most documentation tools are either too complex to self-host or too simple to use with a team. SimpleDoc sits in the sweet spot: a **single Go binary** backed by PostgreSQL that gives you collaborative documentation editing with Markdown, role-based access control, and a polished UI out of the box.

No JavaScript build step. No external dependencies beyond Postgres. Just deploy and start writing.

## Features

### Collaborative Markdown Editing
- Full **Markdown editor** with live preview powered by [goldmark](https://github.com/yuin/goldmark) (GitHub Flavored Markdown)
- Tables, task lists, strikethrough, code blocks, blockquotes, and inline HTML
- Built-in **Markdown help reference** in the editor
- **Image management** — upload, replace, and embed images directly from the editor

### Content Organization
- **Sections and pages** — organize documentation into logical groups
- **Section rows** — visually group sections on the home page
- **Drag-and-drop reordering** — rearrange sections and rows with Sortable.js
- **Soft delete** — accidentally deleted content can be recovered from the database

### Role-Based Access Control
- **Admin role** — full access to all features, user management, and site settings
- **Editor role** — create, edit, and delete documentation content
- **Custom roles** — create any role and restrict specific sections to users who have it
- **Section-level permissions** — lock sections so only users with the required role can view them

### User Management
- **Admin panel** for creating and managing users and roles
- **Password reset** via email (SMTP integration) or admin-set
- **Session-based authentication** with secure, HTTP-only cookies
- **Brute-force protection** — math challenge after repeated failed login attempts

### Theming
- **4 built-in themes**: Midnight (dark), Slate, Silver, and Daylight (light)
- **7 accent colors**: Blue, Purple, Green, Orange, Red, Teal, Pink
- All customizable from the admin UI — no code changes required

### Version History
- Every edit to pages, sections, images, and site settings is tracked in history tables
- See the current version number while editing

### Production-Ready
- **Single binary** — compiles to a static Go binary with zero runtime dependencies
- **Minimal Docker image** — multi-stage build from `scratch` (< 20 MB)
- **Database migrations** built-in with golang-migrate
- **Non-root container** — runs as UID 65534

## Quick Start

### Prerequisites
- Go 1.24+
- PostgreSQL 16+
- Make

### Development

```bash
# Clone the repository
git clone https://github.com/simpledochub/simple-doc.git
cd simple-doc

# Start database, seed sample data, and run the server
make dev
```

This will:
1. Start a PostgreSQL container
2. Run database migrations
3. Seed sample documentation content
4. Start the server at `http://localhost:8080`

### Other Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the Go binary |
| `make run` | Run the server |
| `make seed` | Seed the database with sample content |
| `make db-up` | Start the PostgreSQL container |
| `make db-down` | Stop the PostgreSQL container |
| `make db-reset` | Reset the database (removes all data) |
| `make db-psql` | Open a psql shell to the database |
| `make build-docker` | Build the Docker image |

## Deployment

### Docker

```bash
# Build the image
docker build -t simpledochub/simple-doc .

# Run with a PostgreSQL instance
docker run -d \
  -p 8080:8080 \
  -e DATABASE_URL="postgres://user:pass@db:5432/simpledoc?sslmode=disable" \
  simpledochub/simple-doc
```

### Docker Hub

Pre-built images are available on Docker Hub:

```bash
docker pull simpledochub/simple-doc:latest
```

## Configuration

All settings are configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable` | PostgreSQL connection string |
| `PORT` | `8080` | Server port |
| `MIGRATIONS_DIR` | `migrations` | Path to SQL migration files |
| `TEMPLATES_DIR` | `templates` | Path to HTML templates |
| `CONTENT_DIR` | `content` | Path to seed content |
| `STATIC_DIR` | `static` | Path to static assets |
| `SMTP_HOST` | `localhost` | SMTP server for password reset emails |
| `SMTP_PORT` | `25` | SMTP port |
| `SMTP_USER` | *(empty)* | SMTP username (optional) |
| `SMTP_PASS` | *(empty)* | SMTP password (optional) |
| `SMTP_FROM` | `noreply@example.com` | From address for emails |
| `BASE_URL` | `http://localhost:8080` | Public URL of the site |

## Tech Stack

- **Go** — HTTP server, templating, and business logic
- **PostgreSQL** — data storage with full migration support
- **goldmark** — Markdown to HTML rendering (GFM)
- **pgx** — PostgreSQL driver
- **bcrypt** — password hashing
- **golang-migrate** — database schema migrations
- **htmx** — live preview in the editor
- **Sortable.js** — drag-and-drop reordering

## Project Structure

```
simple-doc/
├── cmd/
│   ├── server/       # Main server entrypoint
│   └── seed/         # Database seed script
├── handlers/         # HTTP handlers
├── internal/
│   ├── db/           # Database queries
│   └── markdown/     # Markdown rendering
├── migrations/       # SQL migration files
├── templates/        # HTML templates
├── static/           # Static assets
├── content/          # Seed markdown content
├── config/           # Configuration
├── Dockerfile
└── Makefile
```

## License

MIT
