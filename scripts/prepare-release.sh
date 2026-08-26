#!/usr/bin/env bash
# Prepare CHANGELOG.md for a SemVer release (Keep a Changelog).
#
# Supports planned sections (## [0.1.0] — planned), future nonempty [Unreleased]
# promotion, --check-only, --require-final, --semver-only, and idempotent re-runs.
# Release SemVer rejects leading-zero core identifiers, empty or dotted-invalid
# prerelease identifiers, and build metadata (+...); prerelease numeric identifiers
# must not have leading zeroes.
#
# Usage:
#   scripts/prepare-release.sh 0.1.2 [YYYY-MM-DD]
#   scripts/prepare-release.sh 0.1.2 --check-only
#   scripts/prepare-release.sh 0.1.2 --require-final
#   scripts/prepare-release.sh 0.1.2 --semver-only

set -euo pipefail

VERSION="${1:?version required (for example 0.1.2)}"
CHANGELOG="${CHANGELOG_FILE:-CHANGELOG.md}"

DATE=""
CHECK_ONLY=false
REQUIRE_FINAL=false
SEMVER_ONLY=false

shift
while [ "$#" -gt 0 ]; do
	case "$1" in
	--check-only)
		CHECK_ONLY=true
		;;
	--require-final)
		REQUIRE_FINAL=true
		;;
	--semver-only)
		SEMVER_ONLY=true
		;;
	--*)
		echo "unknown option: $1" >&2
		exit 1
		;;
	*)
		if [ -n "$DATE" ]; then
			echo "unexpected extra argument: $1" >&2
			exit 1
		fi
		DATE="$1"
		;;
	esac
	shift
done

if [ -z "$DATE" ]; then
	DATE="$(date -u +%Y-%m-%d)"
fi

validate_numeric_identifier() {
	local id="$1"
	local label="$2"
	if [[ ! "$id" =~ ^(0|[1-9][0-9]*)$ ]]; then
		echo "invalid ${label} identifier: ${id}" >&2
		return 1
	fi
	return 0
}

validate_prerelease_identifiers() {
	local prerelease="$1"
	local part

	if [ -z "$prerelease" ]; then
		echo "empty prerelease identifier" >&2
		return 1
	fi
	if [[ "$prerelease" =~ \.\. ]] || [[ "$prerelease" == .* ]] || [[ "$prerelease" == *. ]]; then
		echo "invalid prerelease identifiers: ${prerelease}" >&2
		return 1
	fi

	IFS=. read -ra parts <<<"$prerelease"
	for part in "${parts[@]}"; do
		if [ -z "$part" ]; then
			echo "empty prerelease identifier in: ${prerelease}" >&2
			return 1
		fi
		if [[ "$part" =~ ^[0-9]+$ ]]; then
			validate_numeric_identifier "$part" prerelease || return 1
		elif [[ ! "$part" =~ ^[0-9A-Za-z-]+$ ]]; then
			echo "invalid prerelease identifier: ${part}" >&2
			return 1
		fi
	done
	return 0
}

validate_release_semver() {
	local version="$1"
	local core prerelease

	if [[ "$version" == *"+"* ]]; then
		echo "build metadata not supported: ${version}" >&2
		return 1
	fi

	case "$version" in
	*-*)
		core="${version%%-*}"
		prerelease="${version#*-}"
		if [ -z "$prerelease" ]; then
			echo "empty prerelease suffix: ${version}" >&2
			return 1
		fi
		;;
	*)
		core="$version"
		prerelease=""
		;;
	esac

	local major minor patch
	IFS=. read -r major minor patch <<<"$core" || true
	if [ -z "${major:-}" ] || [ -z "${minor:-}" ] || [ -z "${patch:-}" ] || [[ "$core" == *.*.*.* ]]; then
		echo "invalid semver core: ${version}" >&2
		return 1
	fi

	validate_numeric_identifier "$major" major || return 1
	validate_numeric_identifier "$minor" minor || return 1
	validate_numeric_identifier "$patch" patch || return 1

	if [ -n "$prerelease" ]; then
		validate_prerelease_identifiers "$prerelease" || return 1
	fi

	return 0
}

validate_release_semver "$VERSION" || exit 1

if [ "$SEMVER_ONLY" = true ]; then
	exit 0
fi

if [ ! -f "$CHANGELOG" ]; then
	echo "changelog file not found: $CHANGELOG" >&2
	exit 1
fi

cleanup() {
	if [ -n "${tmp:-}" ] && [ -f "$tmp" ]; then
		rm -f "$tmp"
	fi
}
trap cleanup EXIT INT HUP TERM

finalized_header_prefix() {
	printf '## [%s] - ' "$VERSION"
}

is_finalized() {
	local prefix
	prefix="$(finalized_header_prefix)"
	awk -v prefix="$prefix" '
		index($0, prefix) == 1 {
			date = substr($0, length(prefix) + 1)
			if (date ~ /^[0-9]{4}-[0-9]{2}-[0-9]{2}$/) {
				found = 1
				exit
			}
		}
		END { exit(found ? 0 : 1) }
	' "$CHANGELOG"
}

planned_header() {
	printf '## [%s] — planned' "$VERSION"
}

is_planned() {
	awk -v header="$(planned_header)" '
		$0 == header { found = 1; exit }
		END { exit(found ? 0 : 1) }
	' "$CHANGELOG"
}

planned_has_content() {
	local header
	header="$(planned_header)"
	awk -v header="$header" '
		$0 == header { in_planned = 1; next }
		in_planned && /^## \[/ { exit }
		in_planned && /^### / { found = 1; exit }
		END { exit(found ? 0 : 1) }
	' "$CHANGELOG"
}

unreleased_has_content() {
	awk '
		/^## \[Unreleased\]/ { in_unreleased = 1; next }
		in_unreleased && /^## \[/ { exit }
		in_unreleased && /^### / { found = 1; exit }
		END { exit(found ? 0 : 1) }
	' "$CHANGELOG"
}

finalize_planned() {
	local header
	header="$(planned_header)"
	tmp="$(mktemp)"
	awk -v header="$header" -v ver="$VERSION" -v dt="$DATE" '
		$0 == header {
			print "## [" ver "] - " dt
			skip_planned_note = 1
			next
		}
		skip_planned_note && $0 == "" { next }
		skip_planned_note && /^> / { next }
		skip_planned_note { skip_planned_note = 0 }
		{ print }
	' "$CHANGELOG" >"$tmp"
	mv "$tmp" "$CHANGELOG"
	tmp=""
}

promote_unreleased() {
	tmp="$(mktemp)"
	awk -v ver="$VERSION" -v dt="$DATE" '
		/^## \[Unreleased\]/ {
			print "## [Unreleased]"
			print ""
			print "## [" ver "] - " dt
			next
		}
		{ print }
	' "$CHANGELOG" >"$tmp"
	mv "$tmp" "$CHANGELOG"
	tmp=""
}

if is_finalized; then
	if [ "$REQUIRE_FINAL" = true ] || [ "$CHECK_ONLY" = true ]; then
		echo "changelog section [${VERSION}] is finalized"
		exit 0
	fi
	echo "changelog section [${VERSION}] already finalized"
	exit 0
fi

if [ "$REQUIRE_FINAL" = true ]; then
	echo "changelog section [${VERSION}] is not finalized" >&2
	exit 1
fi

ready=false
if is_planned; then
	if planned_has_content; then
		ready=true
	fi
elif unreleased_has_content; then
	ready=true
fi

if [ "$ready" = false ]; then
	if is_planned; then
		echo "planned section [${VERSION}] has no release notes (expected at least one ### section)" >&2
	elif ! grep -qF '## [Unreleased]' "$CHANGELOG"; then
		echo "missing [Unreleased] section and no planned section for ${VERSION} in $CHANGELOG" >&2
	else
		echo "[Unreleased] has no release notes (expected at least one ### section)" >&2
	fi
	exit 1
fi

if [ "$CHECK_ONLY" = true ]; then
	if is_planned; then
		echo "planned section [${VERSION}] is ready for release"
	else
		echo "[Unreleased] is ready for version ${VERSION}"
	fi
	exit 0
fi

if is_planned; then
	finalize_planned
	echo "finalized planned section [${VERSION}] - ${DATE} in ${CHANGELOG}"
else
	promote_unreleased
	echo "promoted [Unreleased] to [${VERSION}] - ${DATE} in ${CHANGELOG}"
fi

trap - EXIT INT HUP TERM
cleanup
