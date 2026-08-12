---
name: api-conventions
description: House HTTP API style — POST-action/RPC routing (not REST), action vocabulary, snake_case JSON, request/response body shape, error envelope, status mapping. Use when designing or implementing API handlers.
---

# HTTP API conventions (POST-action)

Source of truth for routes: `internal/api/router.go` (or equivalent mux setup).
Full JSON examples: [reference.md](reference.md).

**This is the house style: POST-action / RPC, deliberately NOT REST.** Agents tend to
drift toward REST (resource path params, `PUT`/`DELETE`) from habit — do not. Follow the
rules below.

## Routing

- **Mutations and non-trivial queries** → `POST /api/<resource>/<action>` with a JSON body.
  Examples: `POST /api/entries/approve`, `POST /api/skill-sources/create`,
  `POST /api/memory-entries/find`, `POST /api/github-connections/upsert`.
- **Simple reads** → `GET /api/<resource>` (no path params for state; identifiers go in the
  JSON body of a `POST .../find` or `.../get` when there are filters).
- **Complex filters/queries** → `POST /api/<resource>/find` with the filter as a JSON body —
  never encode complex filters into the URL query string.
- No `PUT`, no `DELETE`, no `/{id}` path parameters for resource state.
- HTTP methods used: **`POST` and `GET` only**. `POST` by default; `GET` only for the
  browser/HTTP exceptions listed below.

`<resource>` is a noun (entity or aggregate); `<action>` is a verb. Collections are
**plural** (`entries`, `skill-sources`); a per-user/per-scope singleton is **singular**
(`subscription`, `settings`).

### Why POST-action (keep this rationale in mind — it's the reason, not an accident)

- **Uniform handling:** every request/response is a JSON body decoded/encoded the same way —
  no split between path params, query params, and body.
- **No URL-encoding pain:** complex filters (arrays, nested objects, ranges) go in a JSON body
  instead of being hand-encoded into query strings.
- **Metrics/monitoring:** a fixed path per action (`/api/entries/find`) aggregates cleanly;
  `/{id}` path variables fragment dashboards and make rate/latency-by-route useless.

### GET exceptions — only where the browser or HTTP layer requires a URL

Use `GET` (with path/query) **only** for:

- File / binary **downloads** (`<a download>`, direct navigation to a URL)
- **Bookmarkable / shareable deep-link** URLs the browser must encode
- Resources referenced by `<img src>`, `<a href>`, `<link>`
- **Server-Sent Events / streaming** (`EventSource` is GET)
- **OAuth redirect callbacks** (the provider dictates the method)
- Health / readiness probes (`GET /healthz`, `GET /readyz`), static assets
- Cache/CDN-sensitive reads where HTTP caching matters

### ❌ Do not (REST drift) → ✅ house style

```text
❌ PUT    /api/entries/{id}              ✅ POST /api/entries/update      {"id":"…", …}
❌ DELETE /api/entries/{id}              ✅ POST /api/entries/delete      {"ids":["…"]}
❌ GET    /api/entries?status=draft&tag=a,b   ✅ POST /api/entries/find   {"status":"draft","tags":["a","b"]}
❌ PATCH  /api/entries/{id}              ✅ POST /api/entries/update      {"id":"…","fields":{…}}
```

## Action vocabulary

Prefer these standard `<action>` names before inventing one:

| action         | Purpose |
|----------------|---------|
| `find`         | Search/query a collection (filters + pagination) |
| `find-by-ids`  | Load a collection by identifiers; returns a map keyed by id |
| `get`          | Single resource by id. **Rarely needed** — use only to load heavy attributes (large text, binary) impractical in `find`/`find-by-ids` |
| `create`       | Create one entity |
| `create-batch` | Create many in one call |
| `update`       | Update one entity (target id in the body) |
| `set`          | Upsert a singleton (create + update in one) |
| `delete`       | Delete by `ids` (bulk-capable, capped) |

For expressiveness a **Rich API** with domain verbs is encouraged where it reads better:
`POST /api/users/register`, `POST /api/users/register-by-invitation`,
`POST /api/entries/approve` instead of only `.../create`/`.../update`.

**`get` vs `find-by-ids`:** in most cases `find-by-ids` is enough — the client passes the ids
it needs and gets a map back. Add `get` only for one entity's heavy attributes.

## JSON naming

- Request and response fields use **snake_case**: `project_id`, `source_url`, `has_changes`, `created_by`.
- Match tags on Go structs: `` `json:"project_id"` ``.
- snake_case handles acronyms and abbreviations unambiguously — `oauth_url`, `http_status`, `entry_id`, `repo_url` — where camelCase forces awkward, inconsistent choices (`oAuthUrl`? `oauthURL`? `httpUrl`?).
- The frontend client maps snake_case at the API boundary (one place), keeping payloads consistent server-side.

## Request / response body shape

- **Root is always an object `{}`.** A primitive or bare array at the root is forbidden, both
  for the request and the response — it leaves no room to add meta fields later without a
  breaking change. Elements of an `items` array are objects too, never primitives.
- **Symmetry:** for CRUD, prefer the same field names and types in request and response. The
  response may add fields (`id`, `created_at`); the request may carry write-only fields. This
  lets the client work like a repository — `find-by-ids` → edit → `update` with no reshaping.
- **Empty body:** for a parameter-less action (`logout`, a no-filter `get`) a `POST` with no
  body or `{}` is fine.

### `find`

**Request** — pagination `page` (starts at 1) and `items_per_page` (default `page = 1`,
`items_per_page = 20`; cap the page size, typically 100). Optional `order` — an array of
`{"field": "...", "direction": "asc"|"desc"}` where array order is sort priority. All other
filters are free-form snake_case fields.

**Response** — required `items` (array of objects) and `count`. For large collections an
adaptive count is allowed: `count_relation` (`=` | `>=` | `~`) with a `count_limit`; return
422 when the real count exceeds the limit rather than doing an expensive full count.

### `find-by-ids`

**Request** — a single `ids` array (cap the length, e.g. 1000). **Response** — `items` as a
**map** keyed by id (not an array); no `count`.

### `delete`

**Request** — a single `ids` array (capped). **Response** — `204 No Content`.

## Handler vs domain responsibility

- **API handler:** auth/session, decode + parse the request, map it to a command/query, call
  the use case, encode the response. Nothing more.
- **Use case / domain:** all product rules and entity validation (limits, uniqueness,
  cross-field/cross-entity rules).
- **Anti-pattern:** business validation in the handler before the use case runs.

**Parse strictly — never swallow parse errors.** When decoding a string into a typed value
(`decimal.NewFromString`, `uuid.FromString`, time layouts), do not ignore the error: invalid
input like `"abc"` silently becomes a zero value and can pass domain checks meant for a real
value. Return `400` (bad request) with a per-field message. An *omitted* optional field may
default (empty string → `decimal.Zero`); a *present but invalid* value is always an error.

## Error format

```json
{
  "error": "Validation error",
  "violations": [
    { "error": "is_blank", "message": "Value must not be blank.", "property_path": "name" }
  ]
}
```

- `error` (required) — short description, safe for UI display.
- `violations` — array for 422 (and similar) field-level errors. Each element:
  - `error` — stable machine code for programmatic handling (changing it is a breaking change);
  - `message` — required, user-facing, localized to the request locale;
  - `property_path` — optional field path.
- **User-facing `message` text carries no internal artifacts** — no requirement IDs
  (`FR-xxx`, `ADR-xxx`), spec/catalog names, or role codes. Keep it short and actionable
  ("choose from the list", "specify a value"). Requirement IDs belong in GoDoc / OpenSpec /
  developer docs, not in responses.

### Response helpers

```go
writeJSON(w, payload)           // 200 + application/json
writeError(w, code, msg string) // {"error":"<msg>"} + status code
```

Keep error messages short and safe for UI display; log details with `clog.Errorf` server-side.

## Status mapping

| Code | When to use |
|------|-------------|
| 400 Bad Request | Malformed body / deserialization / string-parse failure; `errors.Is(err, ErrInvalidInput)` |
| 401 Unauthorized | Not authenticated (auth middleware / handler, when auth is enabled) |
| 403 Forbidden | Authenticated but insufficient permissions |
| 404 Not Found | `errors.Is(err, ErrNotFound)`, or unknown route |
| 405 Method Not Allowed | Unsupported HTTP verb for the route |
| 409 Conflict | **Version/lock conflict only** — optimistic/pessimistic lock, ETag/revision mismatch, concurrent write of the *same* entity. **Not** for business-rule violations (those are 422). `errors.Is(err, ErrConflict)` |
| 415 Unsupported Media Type | Unsupported `Content-Type` |
| 422 Unprocessable Entity | Validation errors and **business-rule violations the user can fix** — field format, missing references, duplicate unique field. Return `violations` + `property_path`; build via `validator.CreateViolation` (skill **golang-validation**), no HTTP constants in `domain/` |
| 429 Too Many Requests | Rate limit exceeded |
| 502 / 503 | Upstream fetch failure / feature disabled or dependency unavailable |
| 500 Internal Server Error | Unexpected internal error (log before returning) |

## Request bodies

- `Content-Type: application/json` for every `POST`; the action lives in the path, the data in the body.
- Empty body allowed only where the handler explicitly accepts it.
- Updates carry the target id in the body (`{"id":"…"}`), not the URL.

## Adding a new endpoint

1. Add a handler method on the API handler type.
2. Register `POST /api/<resource>/<action>` (or a GET only if it hits a listed exception) in the mux.
3. Document the JSON shape (snake_case, root object, symmetric where CRUD).
4. Add an `api_test` with `apitest` — see [golang-tests](../golang-tests/SKILL.md).
5. Keep the web client's TypeScript types in sync when the project has a frontend client.
6. Update OpenSpec (or project spec) if the behavior is specified there.

## Checklist

- [ ] Mutation/query is `POST /api/<resource>/<action>` (no `PUT`/`DELETE`/`{id}` path params)
- [ ] Action uses the standard vocabulary (`find`/`find-by-ids`/`create`/`update`/`set`/`delete`/`get`) or a clear domain verb
- [ ] Any `GET` is justified by a listed browser/HTTP exception
- [ ] JSON fields snake_case; request/response roots are objects; CRUD shapes are symmetric
- [ ] `find` returns `items` + `count`; `find-by-ids` returns a map; `delete` takes capped `ids`
- [ ] String parses (decimal/uuid/time) return 400 on bad input, never a silent zero value
- [ ] Business validation lives in the use case/domain, not the handler
- [ ] Domain errors mapped via `errors.Is` to correct status; 409 only for version/lock conflicts
- [ ] Errors logged with `clog.Errorf` before 5xx
- [ ] Test in `internal/api/*_test.go`
