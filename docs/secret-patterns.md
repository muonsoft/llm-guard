# Secret pattern snapshot (2026-08-12)

This document pins the conservative offline secret shapes implemented in M6.
Patterns are intentionally small and versioned; unknown provider formats remain
undetected rather than guessed heuristically.

## Official sources

| Provider | Reference |
|----------|-----------|
| GitHub token prefixes | [About authentication to GitHub](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/about-authentication-to-github) |
| GitHub credential revoke | [Revoke a credential](https://docs.github.com/en/rest/credentials/revoke) |
| GitLab token prefixes | [GitLab token prefixes](https://docs.gitlab.com/security/tokens/) |
| AWS access key IDs | [AWS unique identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html) |
| OpenAI-like keys | **Heuristic only** — no stable official exact-shape contract was found at snapshot time |

## Supported shapes

### JWT (`SECRET_JWT`)

- Exactly three non-empty base64url segments separated by `.`
- Header JSON object with non-empty string `alg`; `alg: none` is rejected
- Only the header segment is decoded; payload and signature are validated by
  base64url alphabet and unpadded length rules (`length % 4` must not be `1`)
  without decoding or allocating decoded bytes
- Token boundaries: not embedded in longer alphanumeric runs

### PEM private key (`SECRET_PRIVATE_KEY`)

Supported `BEGIN`/`END` labels (matching pair required):

- `PRIVATE KEY`
- `RSA PRIVATE KEY`
- `EC PRIVATE KEY`
- `OPENSSH PRIVATE KEY`
- `PGP PRIVATE KEY BLOCK`

Blocks require a line break after the `BEGIN` line and before the `END` line
(LF or CRLF). Concatenated `BEGIN...body...END` without line boundaries is
rejected. Body must be non-empty, syntactically valid base64. Public/certificate
blocks are rejected.

### Provider API keys (`SECRET_API_KEY`)

| Shape | Pattern | Notes |
|-------|---------|-------|
| GitHub classic PAT | `ghp_[A-Za-z0-9]{36}` | 40 characters total |
| GitHub fine-grained PAT | `github_pat_[A-Za-z0-9]{22}_[A-Za-z0-9]{59}` | Revoke-example shape; 93 characters total |
| GitLab PAT | `glpat-[A-Za-z0-9\-_]{20}` | 26 characters total |
| AWS access key ID | `(AKIA\|ASIA)[A-Z0-9]{16}` | Exactly 20 characters |
| OpenAI-like | `sk-[A-Za-z0-9]{20,64}` | Heuristic min/max body after prefix |
| OpenAI project-like | `sk-proj-[A-Za-z0-9\-_]{20,128}` | Heuristic min/max body after prefix |

Strict token boundaries apply; truncated, embedded, or invalid-alphabet
candidates are rejected. Multi-shape inputs return findings in stable textual
order independent of internal pattern-table order.

### Connection strings (`CONNECTION_STRING`)

Allowlisted schemes: `postgres`, `postgresql`, `mysql`, `mongodb`, `mongodb+srv`,
`redis`, `rediss`, `amqp`, `amqps`.

Requirements:

- Absolute URL with non-empty host
- Non-empty username **and** password in authority (`user:password@host`)
- `http`/`https`, passwordless DSN, and query-only credentials are rejected
- Trailing punctuation trimming is bracket-aware; IPv6 closing `]` is preserved
  when it terminates the authority (for example `postgres://u:p@[::1]`)

Validation uses `net/url` after bounded candidate extraction.

## Policy defaults

| Entity family | Default `Mask` action |
|---------------|----------------------|
| PII / custom entities | `mask` |
| All four secret entities | `block` |

Use `WithSecretAction(ActionMask)` to restore mask-like behavior for secrets.
Use `WithEntityAction` for per-entity overrides. Duplicate `WithSecretAction`
or duplicate exact-entity overrides are rejected at construction.

## Limitations

- No cryptographic JWT signature verification or claims inspection
- No PEM key parsing beyond base64 syntax
- No entropy-based generic secret detection
- No live provider verification or network calls
- OpenAI-like shapes are conservative heuristics and may miss or (rarely) false-positive on prefix-shaped prose

## Update procedure

1. Confirm shape change against official provider documentation.
2. Update detector patterns in `secret_apikey.go` / related files.
3. Add synthetic positive, negative, and malformed cases to `testdata/secrets/cases.jsonl`.
4. Bump the snapshot date in this document.
5. Run focused secret tests and fuzz targets from the M6 verification checklist.

Never commit live credentials to source, fixtures, or test output.
