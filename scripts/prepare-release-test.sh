#!/usr/bin/env bash
# Isolated tests for scripts/prepare-release.sh (does not touch CHANGELOG.md).

set -euo pipefail

ROOT="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
PREPARE="$ROOT/scripts/prepare-release.sh"
PASS=0
FAIL=0

pass() {
	printf 'PASS: %s\n' "$1"
	PASS=$((PASS + 1))
}

fail() {
	printf 'FAIL: %s\n' "$1" >&2
	FAIL=$((FAIL + 1))
}

run_prepare() {
	local changelog="$1"
	local version="$2"
	shift 2
	CHANGELOG_FILE="$changelog" "$PREPARE" "$version" "$@" 2>&1
}

new_tmpdir() {
	mktemp -d "${TMPDIR:-/tmp}/prepare-release-test.XXXXXX"
}

test_invalid_version() {
	dir="$(new_tmpdir)"
	trap 'rm -rf "$dir"' RETURN
	printf '%s\n' '## [Unreleased]' >"$dir/CHANGELOG.md"
	if run_prepare "$dir/CHANGELOG.md" bad-version --check-only >/dev/null 2>&1; then
		fail "invalid version should fail"
	else
		pass "invalid version rejected"
	fi
}

test_leading_zero_semver_rejected() {
	dir="$(new_tmpdir)"
	trap 'rm -rf "$dir"' RETURN
	printf '%s\n' '## [Unreleased]' >"$dir/CHANGELOG.md"
	if run_prepare "$dir/CHANGELOG.md" 01.2.3 --check-only >/dev/null 2>&1; then
		fail "leading-zero semver should fail"
	else
		pass "leading-zero semver rejected"
	fi
}

test_malformed_prerelease_rejected() {
	dir="$(new_tmpdir)"
	trap 'rm -rf "$dir"' RETURN
	printf '%s\n' '## [Unreleased]' >"$dir/CHANGELOG.md"
	if run_prepare "$dir/CHANGELOG.md" 1.2.3-alpha..1 --check-only >/dev/null 2>&1; then
		fail "dotted-empty prerelease should fail"
	else
		pass "dotted-empty prerelease rejected"
	fi
	if run_prepare "$dir/CHANGELOG.md" 1.2.3-01 --check-only >/dev/null 2>&1; then
		fail "leading-zero prerelease should fail"
	else
		pass "leading-zero prerelease rejected"
	fi
	if CHANGELOG_FILE=/dev/null "$PREPARE" 1.2.3- --semver-only >/dev/null 2>&1; then
		fail "trailing empty prerelease should fail"
	else
		pass "trailing empty prerelease rejected"
	fi
}

test_planned_header_requires_exact_line() {
	dir="$(new_tmpdir)"
	trap 'rm -rf "$dir"' RETURN
	cat >"$dir/CHANGELOG.md" <<'EOF'
## [Unreleased]

## [0.8.0] — planned extra suffix

### Added
- should not match
EOF
	if run_prepare "$dir/CHANGELOG.md" 0.8.0 --check-only >/dev/null 2>&1; then
		fail "planned header suffix should not match"
	else
		pass "planned header suffix rejected"
	fi
}

test_valid_prerelease_accepted() {
	dir="$(new_tmpdir)"
	trap 'rm -rf "$dir"' RETURN
	cat >"$dir/CHANGELOG.md" <<'EOF'
## [Unreleased]

### Added
- rc item
EOF
	if run_prepare "$dir/CHANGELOG.md" 1.2.3-rc.1 --check-only >/dev/null; then
		pass "valid prerelease accepted"
	else
		fail "valid prerelease accepted"
	fi
}

test_require_final_rejects_regex_false_positive() {
	dir="$(new_tmpdir)"
	trap 'rm -rf "$dir"' RETURN
	cat >"$dir/CHANGELOG.md" <<'EOF'
## [Unreleased]

## [0x1x0] - 2026-08-26

### Added
- wrong header
EOF
	if run_prepare "$dir/CHANGELOG.md" 0.1.0 --require-final >/dev/null 2>&1; then
		fail "require-final should reject 0x1x0 false positive"
	else
		pass "require-final rejects 0x1x0 false positive"
	fi
}

test_require_final_exact_header() {
	dir="$(new_tmpdir)"
	trap 'rm -rf "$dir"' RETURN
	cat >"$dir/CHANGELOG.md" <<'EOF'
## [Unreleased]

## [0.7.0] - 2026-08-26

### Added
- exact
EOF
	if run_prepare "$dir/CHANGELOG.md" 0.7.0 --require-final >/dev/null; then
		pass "require-final accepts exact finalized header"
	else
		fail "require-final accepts exact finalized header"
	fi
}

test_planned_check_only() {
	dir="$(new_tmpdir)"
	trap 'rm -rf "$dir"' RETURN
	cat >"$dir/CHANGELOG.md" <<'EOF'
## [Unreleased]

## [0.2.0] — planned

### Added
- smoke feature
EOF
	if run_prepare "$dir/CHANGELOG.md" 0.2.0 --check-only >/dev/null; then
		pass "planned section --check-only"
	else
		fail "planned section --check-only"
	fi
}

test_planned_finalize_idempotent() {
	dir="$(new_tmpdir)"
	trap 'rm -rf "$dir"' RETURN
	cat >"$dir/CHANGELOG.md" <<'EOF'
## [Unreleased]

## [0.2.1] — planned

> **Note:** planned only.

### Changed
- item
EOF
	run_prepare "$dir/CHANGELOG.md" 0.2.1 2026-01-15 >/dev/null
	if grep -qF '## [0.2.1] - 2026-01-15' "$dir/CHANGELOG.md" \
		&& ! grep -q 'planned' "$dir/CHANGELOG.md" \
		&& ! grep -q '^> ' "$dir/CHANGELOG.md"; then
		pass "planned section finalized"
	else
		fail "planned section finalized"
	fi
	if run_prepare "$dir/CHANGELOG.md" 0.2.1 2026-01-15 >/dev/null; then
		pass "planned finalize idempotent"
	else
		fail "planned finalize idempotent"
	fi
}

test_unreleased_promotion() {
	dir="$(new_tmpdir)"
	trap 'rm -rf "$dir"' RETURN
	cat >"$dir/CHANGELOG.md" <<'EOF'
## [Unreleased]

### Added
- future item

## [0.1.0] - 2026-01-01

### Added
- old
EOF
	run_prepare "$dir/CHANGELOG.md" 0.3.0 2026-02-02 >/dev/null
	if grep -qF '## [0.3.0] - 2026-02-02' "$dir/CHANGELOG.md" \
		&& awk '/^## \[0.3.0\]/ { in_v=1; next } in_v && /^## \[/ { exit } in_v && /future item/ { found=1 } END { exit(found ? 0 : 1) }' "$dir/CHANGELOG.md"; then
		pass "unreleased promotion"
	else
		fail "unreleased promotion"
	fi
}

test_require_final_on_planned() {
	dir="$(new_tmpdir)"
	trap 'rm -rf "$dir"' RETURN
	cat >"$dir/CHANGELOG.md" <<'EOF'
## [Unreleased]

## [0.4.0] — planned

### Added
- pending
EOF
	if run_prepare "$dir/CHANGELOG.md" 0.4.0 --require-final >/dev/null 2>&1; then
		fail "require-final should reject planned"
	else
		pass "require-final rejects planned"
	fi
}

test_require_final_accepts_finalized() {
	dir="$(new_tmpdir)"
	trap 'rm -rf "$dir"' RETURN
	cat >"$dir/CHANGELOG.md" <<'EOF'
## [Unreleased]

## [0.5.0] - 2026-03-03

### Added
- shipped
EOF
	if run_prepare "$dir/CHANGELOG.md" 0.5.0 --require-final >/dev/null; then
		pass "require-final accepts finalized"
	else
		fail "require-final accepts finalized"
	fi
}

test_empty_unreleased_fails() {
	dir="$(new_tmpdir)"
	trap 'rm -rf "$dir"' RETURN
	cat >"$dir/CHANGELOG.md" <<'EOF'
## [Unreleased]

## [0.1.0] - 2026-01-01

### Added
- prior
EOF
	if run_prepare "$dir/CHANGELOG.md" 0.6.0 --check-only >/dev/null 2>&1; then
		fail "empty unreleased should fail check-only"
	else
		pass "empty unreleased rejected"
	fi
}

test_invalid_version
test_leading_zero_semver_rejected
test_malformed_prerelease_rejected
test_planned_header_requires_exact_line
test_valid_prerelease_accepted
test_require_final_rejects_regex_false_positive
test_require_final_exact_header
test_planned_check_only
test_planned_finalize_idempotent
test_unreleased_promotion
test_require_final_on_planned
test_require_final_accepts_finalized
test_empty_unreleased_fails

printf '\nprepare-release-test: %d passed, %d failed\n' "$PASS" "$FAIL"
if [ "$FAIL" -ne 0 ]; then
	exit 1
fi
