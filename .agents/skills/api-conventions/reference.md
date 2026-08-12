# API conventions — reference (JSON examples)

Companion to [SKILL.md](SKILL.md). All examples use the house style: `POST /api/<resource>/<action>`,
snake_case fields, object at the root.

## Root and items

**Not allowed:** root is an array or primitive; `items` elements are primitives.

**Allowed:** root is an object; `items` is an array of objects.

```json
{ "items": [ { "id": "…", "name": "value" } ], "count": 1 }
```

## Symmetry (request ↔ response)

`POST /api/entries/create` request:

```json
{
  "project_id": "01234567-0123-0123-0123-0123456789ab",
  "source_url": "https://example.com/doc",
  "status": "draft"
}
```

`POST /api/entries/find` response — same fields plus service fields:

```json
{
  "items": [
    {
      "id": "01234567-0123-0123-0123-0123456789ab",
      "project_id": "01234567-0123-0123-0123-0123456789ab",
      "source_url": "https://example.com/doc",
      "status": "draft",
      "created_at": "2026-07-04T14:15:22Z",
      "updated_at": "2026-07-04T14:15:22Z"
    }
  ],
  "count": 123
}
```

## find: request

```json
{
  "status": "draft",
  "tags": ["go", "backend"],
  "page": 1,
  "items_per_page": 20,
  "order": [
    { "field": "created_at", "direction": "desc" },
    { "field": "name", "direction": "asc" }
  ]
}
```

Missing/zero pagination defaults to `page = 1`, `items_per_page = 20`; cap the page size
(typically 100). Adaptive count when the collection is large:

```json
{ "items": [ { "id": "…" } ], "count": 10000, "count_relation": ">=", "count_limit": 10000 }
```

Return 422 when the real count exceeds `count_limit` instead of running an expensive full count.

## find-by-ids: request and response

Request — a single `ids` array (cap the length, e.g. 1000):

```json
{ "ids": ["01234567-0123-0123-0123-0123456789ab", "01234567-0123-0123-0123-0123456789ac"] }
```

Response — `items` is a **map keyed by id** (no `count`):

```json
{
  "items": {
    "01234567-0123-0123-0123-0123456789ab": {
      "id": "01234567-0123-0123-0123-0123456789ab",
      "project_id": "01234567-0123-0123-0123-0123456789ab",
      "name": "Ring",
      "created_at": "2026-07-04T14:15:22Z",
      "updated_at": "2026-07-04T14:15:22Z"
    }
  }
}
```

## delete: request and cap

Request — a single `ids` array; cap the count (default 100), return 422 when exceeded:

```json
{ "ids": ["01234567-0123-0123-0123-0123456789ab"] }
```

Response — `204 No Content`.

## Error envelope

```json
{
  "error": "Validation error",
  "violations": [
    {
      "error": "is_blank",
      "message": "Value must not be blank.",
      "property_path": "name"
    }
  ]
}
```

- `error` — short, UI-safe description.
- `violations[].error` — stable machine code (changing it is a breaking change).
- `violations[].message` — required, user-facing, localized; no `FR-xxx`/`ADR-xxx`/role codes.
- `violations[].property_path` — optional field path.
