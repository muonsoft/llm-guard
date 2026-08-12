---
name: backend-structure
description: Go backend layout convention — cmd/ entrypoints, internal/ layering, bootstrap wiring, storage behind interfaces.
---

# Go backend structure

Concrete package maps live in each project's `<project>-backend-golang` skill; this covers the shared convention only.

## Repository layout

```text
/
├── cmd/
│   └── <binary>/        # HTTP server, CLI, or other entrypoints
├── internal/
│   ├── api/             # HTTP handlers, routing
│   ├── auth/            # Session / OAuth middleware (when used)
│   ├── bootstrap/       # Config, application wiring
│   ├── <domain>/        # Business logic, domain types, storage
│   └── ...              # Other internal packages
└── web/                 # Frontend sources (when applicable)
```

- **`cmd/<binary>/`** — thin `main`: read config, call bootstrap, register runnables, start the server.
- **`internal/`** — all application code; not importable by external modules.

## Layering

| Layer | Role |
|-------|------|
| `internal/<domain>` | Domain types, invariants, storage interfaces and implementations |
| Service layer (in domain or dedicated package) | Orchestration, transactions, validation of commands |
| `internal/api` | HTTP transport: decode JSON, auth, map errors to status codes |
| `internal/bootstrap` | Wire config → stores → services → handlers → mux |

**Dependency direction:** handlers → services → storage interfaces. Domain packages do not import `internal/api`.

Storage is behind an interface (e.g. `Store` accepting `afero.Fs` for tests). SQLite, filesystem, or remote backends are implementation details.

## When adding a feature

1. Domain rules and persistence → `internal/<domain>` (types, store, service methods)
2. HTTP surface → `internal/api` handler + route registration
3. Long-running work → runnable worker or job manager pattern (see [runnable-background-processes](../runnable-background-processes/SKILL.md))
4. Tests at each layer — handler tests with `apitest`, domain tests with mocks or `afero`
5. Update project spec when behavior is formally specified
