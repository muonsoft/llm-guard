#!/bin/sh
# Side-effect-free release readiness checks for llm-guard v0.1.0.
# Modes: consumer | license | fuzz | evidence | (default) full dry-run.
# No tag, push, upload, publish, or release mutation.
set -eu

ROOT="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

FUZZ_TIME="${FUZZ_TIME:-2s}"
MODULE_MANIFEST="${MODULE_MANIFEST:-docs/dependency-license-modules.txt}"
INVENTORY="${INVENTORY:-docs/dependency-license-inventory.md}"
EXTERNAL_BASELINE="${EXTERNAL_BASELINE:-docs/evaluation/external-baseline.json}"
GENERATED_SMOKE="${GENERATED_SMOKE:-./testdata/evaluation/generated/smoke.jsonl}"

log() {
	printf 'release-check: %s\n' "$1"
}

die() {
	printf 'release-check: error: %s\n' "$1" >&2
	exit 1
}

require_cmd() {
	command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

run_hygiene() {
	log 'hygiene: gofmt and whitespace'
	unformatted="$(gofmt -l . 2>/dev/null | grep -v '^$' || true)"
	if [ -n "$unformatted" ]; then
		printf '%s\n' "$unformatted" >&2
		die 'gofmt required on listed files'
	fi
	git diff --check
}

run_tests() {
	log 'tests'
	go test ./...
}

run_vet() {
	log 'vet'
	go vet ./...
}

run_race() {
	log 'race'
	go test -race ./...
}

run_consumer() {
	log 'external consumer'
	require_cmd go
	require_cmd mktemp
	require_cmd cp

	tmpdir="$(mktemp -d "${TMPDIR:-/tmp}/llm-guard-consumer.XXXXXX")"
	cleanup() {
		rm -rf "$tmpdir"
	}
	trap cleanup EXIT INT HUP TERM

	cp testdata/external-consumer/main.go "$tmpdir/main.go"
	(
		cd "$tmpdir"
		go mod init example.com/llm-guard-external-consumer
		go mod edit -require=github.com/muonsoft/llm-guard@v0.0.0
		go mod edit -replace=github.com/muonsoft/llm-guard="$ROOT"
		go mod tidy
		go run .
	)
	trap - EXIT INT HUP TERM
	cleanup
}

run_evaluation() {
	log 'evaluation regression (schema v1 corpus)'
	go run ./cmd/llmguard-eval \
		-corpus ./testdata/evaluation/cases.jsonl \
		-format json \
		-fail-on-regression
}

run_generated_smoke() {
	log 'generated evaluation smoke (contract + lifecycle)'
	go run ./cmd/llmguard-eval \
		-suite "$GENERATED_SMOKE" \
		-profile contract \
		-format json \
		-fail-on-regression
	go run ./cmd/llmguard-eval \
		-suite "$GENERATED_SMOKE" \
		-profile lifecycle \
		-format json \
		-fail-on-regression
}

require_external_baseline() {
	log 'external baseline metadata'
	if [ ! -f "$EXTERNAL_BASELINE" ]; then
		die "missing external baseline: $EXTERNAL_BASELINE"
	fi
	for key in measured_commit source_id mapping_version exposure_summary; do
		if ! grep -Fq "\"$key\"" "$EXTERNAL_BASELINE"; then
			die "external baseline missing required key: $key"
		fi
	done
}

run_evidence() {
	require_external_baseline
	measured_commit="$(grep -o '"measured_commit"[[:space:]]*:[[:space:]]*"[^"]*"' "$EXTERNAL_BASELINE" | sed 's/.*"\([^"]*\)"$/\1/')"
	head_commit="$(git rev-parse HEAD)"
	if [ "$measured_commit" != "$head_commit" ]; then
		die "external baseline measured_commit ($measured_commit) does not match HEAD ($head_commit); refresh baseline or use default dry-run"
	fi
	log "external baseline commit matches HEAD ($head_commit)"
}

run_license() {
	log 'license inventory consistency'
	require_cmd go

	if [ ! -f "$MODULE_MANIFEST" ]; then
		die "missing module manifest: $MODULE_MANIFEST"
	fi
	if [ ! -f "$INVENTORY" ]; then
		die "missing dependency inventory: $INVENTORY"
	fi
	if [ ! -f THIRD_PARTY_NOTICES ]; then
		die 'missing THIRD_PARTY_NOTICES'
	fi

	tmp_actual="$(mktemp "${TMPDIR:-/tmp}/llm-guard-modules-actual.XXXXXX")"
	tmp_expected="$(mktemp "${TMPDIR:-/tmp}/llm-guard-modules-expected.XXXXXX")"
	cleanup_license() {
		rm -f "$tmp_actual" "$tmp_expected"
	}
	trap cleanup_license EXIT INT HUP TERM

	go list -m all | LC_ALL=C sort >"$tmp_actual"
	grep -v '^[[:space:]]*#' "$MODULE_MANIFEST" | grep -v '^[[:space:]]*$' | LC_ALL=C sort >"$tmp_expected"

	if ! diff -u "$tmp_expected" "$tmp_actual" >/dev/null 2>&1; then
		diff -u "$tmp_expected" "$tmp_actual" >&2 || true
		die "go list -m all does not match $MODULE_MANIFEST"
	fi

	while IFS= read -r line; do
		module_path="${line%% *}"
		case "$module_path" in
		github.com/muonsoft/llm-guard) ;;
		*)
			module_version="${line#"$module_path "}"
			if [ -z "$module_version" ] || [ "$module_version" = "$line" ]; then
				die "unable to parse module version from: $line"
			fi
			inventory_row="| \`${module_path}\` | ${module_version} |"
			if ! grep -Fq "$inventory_row" "$INVENTORY"; then
				die "inventory missing exact module row: ${module_path} ${module_version}"
			fi
			;;
		esac
	done <"$tmp_actual"

	for module in \
		github.com/muonsoft/errors \
		github.com/muonsoft/go-razdel; do
		if ! grep -Fq "$module" THIRD_PARTY_NOTICES; then
			die "THIRD_PARTY_NOTICES missing shipped dependency: $module"
		fi
	done

	for needle in \
		'github.com/muonsoft/errors v0.5.0' \
		'Copyright (c) 2022 MuonSoft' \
		'github.com/muonsoft/go-razdel v0.1.0' \
		'Copyright (c) 2026 MuonSoft' \
		'668dbe191a5cfd94bebf9155e2ffa5f94ff3fe33' \
		'Copyright (c) 2017'; do
		if ! grep -Fq "$needle" THIRD_PARTY_NOTICES; then
			die "THIRD_PARTY_NOTICES missing required notice marker: $needle"
		fi
	done

	if ! grep -Fq 'docs/dependency-license-inventory.md' THIRD_PARTY_NOTICES; then
		die 'THIRD_PARTY_NOTICES must reference docs/dependency-license-inventory.md'
	fi

	trap - EXIT INT HUP TERM
	cleanup_license
}

run_fuzz() {
	log "fuzz smoke (${FUZZ_TIME} per target)"
	fuzz_cache="$(mktemp -d "${TMPDIR:-/tmp}/llm-guard-fuzz.XXXXXX")"
	cleanup_fuzz() {
		rm -rf "$fuzz_cache"
	}
	trap cleanup_fuzz EXIT INT HUP TERM
	export GOCACHE="$fuzz_cache"

	for target in \
		FuzzMaskRestoreRoundTrip \
		FuzzResolveInvariants \
		FuzzCustomRegexpDetectorInvariants; do
		log "fuzz target: $target"
		go test . -run '^$' -fuzz "^${target}$" -fuzztime="$FUZZ_TIME"
	done

	trap - EXIT INT HUP TERM
	cleanup_fuzz
}

run_benchmark_smoke() {
	log 'benchmark smoke'
	go test ./... -run '^$' -bench . -benchmem -count=1
}

run_full() {
	run_hygiene
	run_tests
	run_vet
	run_race
	run_consumer
	run_evaluation
	run_generated_smoke
	require_external_baseline
	run_license
	(run_fuzz)
	run_benchmark_smoke
	log 'full dry-run complete'
}

usage() {
	cat <<'EOF'
Usage: ./scripts/release-check.sh [mode]

Modes:
  consumer   compile and run external-module fixture via local replace
  license    verify go list -m all against committed inventory/notices
  fuzz       bounded smoke for required fuzz targets
  evidence   verify committed external-baseline.json measured_commit equals HEAD
  (none)     full side-effect-free release dry-run (network-free; does not require
             measured_commit == HEAD)

Environment:
  FUZZ_TIME           per-target fuzz duration (default: 2s)
  MODULE_MANIFEST     module list file (default: docs/dependency-license-modules.txt)
  INVENTORY           detailed inventory file (default: docs/dependency-license-inventory.md)
  EXTERNAL_BASELINE   safe external baseline JSON (default: docs/evaluation/external-baseline.json)
  GENERATED_SMOKE     generated suite v2 smoke path (default: ./testdata/evaluation/generated/smoke.jsonl)
EOF
}

main() {
	require_cmd go
	require_cmd git
	require_cmd gofmt

	case "${1:-}" in
	'')
		run_full
		;;
	consumer)
		run_consumer
		;;
	license)
		run_license
		;;
	fuzz)
		run_fuzz
		;;
	evidence)
		run_evidence
		;;
	-h | --help | help)
		usage
		;;
	*)
		usage
		die "unknown mode: $1"
		;;
	esac
}

main "$@"
