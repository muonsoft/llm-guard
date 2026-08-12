#!/usr/bin/env python3
"""Natasha reference harness CLI."""

import argparse
import sys

from schema import (
    SchemaError,
    canonicalize_expected,
    dumps_jsonl,
    load_cases,
    load_expected,
    validate_expected,
    verify_canonical_jsonl_bytes,
    verify_case_expected_alignment,
)


def _parse_args(argv):
    parser = argparse.ArgumentParser(description="Natasha reference harness")
    subparsers = parser.add_subparsers(dest="command")

    generate = subparsers.add_parser("generate", help="generate expected JSONL from cases")
    generate.add_argument("--cases", required=True)
    generate.add_argument("--output", required=True)

    verify = subparsers.add_parser("verify", help="verify cases against expected output")
    verify.add_argument("--cases", required=True)
    verify.add_argument("--expected", required=True)
    verify.add_argument(
        "--offline",
        action="store_true",
        help="validate committed fixtures without importing Natasha",
    )
    return parser.parse_args(argv)


def cmd_generate(args):
    from extract_live import generate_expected_record

    cases = load_cases(args.cases)
    records = []
    for line_no, case in enumerate(cases, start=1):
        generated = generate_expected_record(case)
        validated = validate_expected(generated, line_no)
        records.append(canonicalize_expected(validated))
    output = dumps_jsonl(records)
    with open(args.output, "w", encoding="utf-8", newline="\n") as handle:
        handle.write(output)


def cmd_verify(args):
    cases = load_cases(args.cases)
    expected = load_expected(args.expected)
    verify_case_expected_alignment(cases, expected)
    verify_canonical_jsonl_bytes(args.cases, cases)
    verify_canonical_jsonl_bytes(args.expected, expected)

    if args.offline:
        return

    from extract_live import generate_expected_record

    regenerated = [canonicalize_expected(generate_expected_record(case)) for case in cases]
    if regenerated != expected:
        raise SchemaError("live regeneration does not match committed expected output")


def main(argv=None):
    args = _parse_args(argv or sys.argv[1:])
    if not args.command:
        print("error: command is required (generate or verify)", file=sys.stderr)
        return 1
    try:
        if args.command == "generate":
            cmd_generate(args)
        elif args.command == "verify":
            cmd_verify(args)
        else:
            raise RuntimeError("unknown command: {}".format(args.command))
    except SchemaError as exc:
        print("schema error: {}".format(exc), file=sys.stderr)
        return 1
    except RuntimeError as exc:
        print("error: {}".format(exc), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
