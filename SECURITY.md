# Security Policy

## Supported scope

Security reports are accepted for the current main branch of
`github.com/muonsoft/llm-guard` while the project is in pre-1.0 MVP stabilization.

In scope:

- PII or secret leakage through public APIs, observers, errors, or logging defaults
- Bypass of default secret blocking or policy enforcement
- Mask/restore correctness issues that expose caller-owned sensitive values
- Supply-chain disclosure gaps for shipped dependencies or project-authored data

Out of scope for this repository:

- Third-party LLM provider security
- Deployment/network perimeter controls outside the library
- Issues requiring real PII, live credentials, or copied upstream dictionaries to
  reproduce

## Reporting a vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Report privately to the maintainers through your organization's agreed secure
channel for MuonSoft open-source projects. If you do not have a contact path,
open a minimal **private** security advisory request through GitHub Security
Advisories for `muonsoft/llm-guard` and omit sensitive payload values.

Include:

- affected version or commit
- reproduction steps with **synthetic** fixtures only
- impact assessment
- suggested fix if available

Allow reasonable time for triage and remediation before public disclosure.

## Safe reporting rules

- Never include real PII, credentials, private keys, or production text in
  public issues, pull requests, comments, or attached logs.
- Use synthetic tokens and fictional names/addresses from project test patterns.
- Do not paste `TokenSet` contents or masked mappings into public channels.

## Supported secure defaults

- Default observer is a no-op; safe observers expose bounded metadata only.
- Secrets block `Mask` by default (`ErrBlocked`); masking requires explicit opt-in.
- `WithUnsafeDevelopmentObserver` is for local debugging only and must not ship in
  production paths.

Maintainers may request additional evidence under the same synthetic-data rule.
