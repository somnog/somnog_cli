# Somnog CLI

The official CLI for the [Somnog](https://github.com/somnog/somnog) full-stack monorepo framework.

Somnog CLI manages the entire lifecycle of a Somnog project — from scaffolding a new app to generating resources, running migrations, and starting all dev servers.

---

## Installation

**macOS / Linux (recommended):**

```sh
curl -fsSL https://raw.githubusercontent.com/somnog/somnog_cli/main/install.sh | sh
```

**Build from source:**

```sh
git clone https://github.com/somnog/somnog_cli.git
cd somnog_cli/tools
go build -o somnog .
sudo mv somnog /usr/local/bin/
```

**Verify installation:**

```sh
somnog version
```

---

## Quick Start

```sh
# 1. Scaffold a new project
somnog new my-app

# 2. Enter the project directory
cd my-app

# 3. Configure your environment
cp .env.example .env

# 4. Start all development servers
somnog start
```

`somnog new` clones the Somnog template, initializes a fresh git repository, and runs `pnpm install` automatically.

---

## Commands

### `somnog new <project-name>`

Scaffold a new full-stack Somnog project.

```sh
somnog new my-app
somnog new my-app --template https://github.com/your-org/custom-template.git
```

| Flag         | Default                                | Description                                   |
| ------------ | -------------------------------------- | --------------------------------------------- |
| `--template` | `https://github.com/somnog/somnog.git` | Git repository to use as the project template |

---

### `somnog start`

Start all development servers in parallel.

```sh
somnog start
```

Runs `pnpm dev` across all apps concurrently:

- **Go API** — hot-reload via `air`
- **Next.js web** — frontend dev server
- **Next.js admin** — admin panel dev server

---

### `somnog generate resource <Name>`

Generate a complete full-stack resource from a single command.

```sh
somnog generate resource Invoice
somnog generate resource ProductCategory
```

The resource name must start with an uppercase letter. Multi-word names use PascalCase.

**What gets generated:**

| File                                           | Description                                                                                             |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| `apps/api/internal/models/<name>.go`           | GORM model with UUID primary key and timestamps                                                         |
| `apps/api/internal/services/<name>_service.go` | CRUD service with pagination, search, and sort                                                          |
| `apps/api/internal/handlers/<name>_handler.go` | Gin HTTP handler (List, GetByID, Create, Update, Delete)                                                |
| `apps/api/internal/routes/routes.go`           | Routes injected at `// somnog:handlers`, `// somnog:routes:protected`, `// somnog:routes:admin` markers |
| `packages/shared/types/<name>.ts`              | TypeScript interface                                                                                    |
| `packages/shared/schemas/<name>.ts`            | Zod validation schemas (base, create, update)                                                           |
| `packages/shared/types/index.ts`               | Export barrel updated                                                                                   |
| `packages/shared/schemas/index.ts`             | Export barrel updated                                                                                   |
| `apps/admin/resources/<name>.ts`               | Admin panel resource config (table, form, columns)                                                      |
| `apps/admin/resources/index.ts`                | Resource registry updated                                                                               |

**After generating a resource:**

```sh
# 1. Edit the model fields
vim apps/api/internal/models/invoice.go

# 2. Run the migration
somnog migrate

# 3. Update the admin table columns if needed
vim apps/admin/resources/invoice.ts
```

---

### `somnog migrate`

Run database migrations.

```sh
somnog migrate
```

Executes `go run ./cmd/migrate/...` from the `apps/api/` directory.

---

### `somnog seed`

Run database seeders.

```sh
somnog seed
```

Executes `go run ./cmd/seed/...` from the `apps/api/` directory.

---

### `somnog routes`

List all registered API routes by parsing `apps/api/internal/routes/routes.go`.

```sh
somnog routes
```

Detects routes registered on `r`, `protected`, and `admin` route groups.

---

### `somnog version`

Print the installed CLI version.

```sh
somnog version
```

---

## Project Structure

All commands (except `somnog new`) require a `somnog.config.ts` file to be present somewhere in the directory tree. The CLI walks up from the current directory to find the project root.

The expected monorepo layout:

```
my-app/
├── apps/
│   ├── api/                    # Go — Gin + GORM
│   │   ├── cmd/
│   │   │   ├── migrate/        # Migration entrypoint
│   │   │   └── seed/           # Seeder entrypoint
│   │   └── internal/
│   │       ├── handlers/
│   │       ├── models/
│   │       ├── routes/
│   │       │   └── routes.go   # somnog:* markers required
│   │       └── services/
│   ├── web/                    # Next.js — public frontend
│   └── admin/                  # Next.js — admin panel
│       └── resources/
├── packages/
│   └── shared/
│       ├── schemas/            # Zod schemas
│       └── types/              # TypeScript interfaces
└── somnog.config.ts            # Project root marker
```

---

## Code Generation Markers

The `generate resource` command injects code at marker comments in existing files. These markers must be present in your project for injection to work.

| Marker                       | File                 | Purpose                               |
| ---------------------------- | -------------------- | ------------------------------------- |
| `// somnog:handlers`         | `routes.go`          | Handler instantiation injection point |
| `// somnog:routes:protected` | `routes.go`          | Protected route injection point       |
| `// somnog:routes:admin`     | `routes.go`          | Admin route injection point           |
| `// somnog:types`            | `types/index.ts`     | Type export injection point           |
| `// somnog:schemas`          | `schemas/index.ts`   | Schema export injection point         |
| `// somnog:resources`        | `resources/index.ts` | Resource import injection point       |
| `// somnog:resource-list`    | `resources/index.ts` | Resource list injection point         |

---

## Requirements

| Tool    | Version |
| ------- | ------- |
| Go      | 1.24+   |
| Node.js | 22+     |
| pnpm    | 9+      |
| git     | Any     |

---

## Supported Platforms

| OS      | Architecture |
| ------- | ------------ |
| macOS   | amd64, arm64 |
| Linux   | amd64, arm64 |
| Windows | amd64        |

---

## Contributing

The CLI source lives in `tools/`.

```sh
git clone https://github.com/somnog/somnog_cli.git
cd somnog_cli/tools

# Build
go build -o somnog .

# Run locally
./somnog version
```

**Releasing a new version:**

Push a version tag to trigger the release workflow. GitHub Actions will cross-compile binaries for all supported platforms and publish a GitHub Release automatically.

```sh
git tag v1.1.0
git push origin v1.1.0
```

---

## License

MIT
