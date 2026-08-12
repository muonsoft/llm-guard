import argparse
import json
import os
import subprocess
import sys
import tempfile
import unittest
from unittest import mock

import reference
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
from spans import codepoint_span_to_bytes, verify_span_slices


REPO_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
CASES_PATH = os.path.join(REPO_ROOT, "testdata", "natasha", "cases.jsonl")
EXPECTED_PATH = os.path.join(REPO_ROOT, "testdata", "natasha", "expected-python.jsonl")
REFERENCE_PATH = os.path.join(os.path.dirname(__file__), "reference.py")


def person_match(**overrides):
    match = {
        "entity": "PERSON",
        "span": {"start": 0, "end": 11},
        "span_bytes": {"start": 0, "end": 21},
        "matched_text": "Иван Петров",
        "normalized": {
            "first": "иван",
            "last": "петров",
            "middle": None,
            "nick": None,
        },
    }
    match.update(overrides)
    return match


def expected_record(**overrides):
    record = {
        "schema_version": 1,
        "id": "person-first-last-001",
        "entity": "PERSON",
        "input": "Иван Петров",
        "matches": [person_match()],
    }
    record.update(overrides)
    return record


class SpanTests(unittest.TestCase):
    def test_cyrillic_multibyte_conversion(self):
        text = "Иван Петров"
        self.assertEqual(codepoint_span_to_bytes(text, 0, 4), (0, 8))
        self.assertEqual(codepoint_span_to_bytes(text, 5, 11), (9, 21))
        verify_span_slices(text, 0, 11, 0, 21, text)

    def test_ascii_punctuation(self):
        text = "hello, world!"
        self.assertEqual(codepoint_span_to_bytes(text, 0, 5), (0, 5))
        self.assertEqual(codepoint_span_to_bytes(text, 7, 12), (7, 12))
        verify_span_slices(text, 7, 12, 7, 12, "world")

    def test_boundary_error(self):
        with self.assertRaises(ValueError):
            codepoint_span_to_bytes("abc", -1, 2)
        with self.assertRaises(ValueError):
            codepoint_span_to_bytes("abc", 0, 4)

    def test_mismatched_byte_slice(self):
        with self.assertRaises(ValueError):
            verify_span_slices("Иван", 0, 4, 0, 7, "Иван")


class SchemaTests(unittest.TestCase):
    def test_load_committed_fixtures(self):
        cases = load_cases(CASES_PATH)
        expected = load_expected(EXPECTED_PATH)
        verify_case_expected_alignment(cases, expected)
        self.assertEqual(len(cases), 21)

    def test_empty_match_record(self):
        record = {
            "schema_version": 1,
            "id": "empty",
            "entity": "ADDRESS",
            "input": "Москва",
            "matches": [],
        }
        validated = load_expected_from_records([record])
        self.assertEqual(validated[0]["matches"], [])

    def test_duplicate_case_ids_fail(self):
        payload = (
            '{"schema_version":1,"id":"dup","entity":"PERSON","corpus_class":"positive",'
            '"input":"Иван Петров","product_expectation":"match",'
            '"intentional_difference_class":null,"intentional_difference_reason":null}\n'
            '{"schema_version":1,"id":"dup","entity":"PERSON","corpus_class":"positive",'
            '"input":"Петров Иван","product_expectation":"match",'
            '"intentional_difference_class":null,"intentional_difference_reason":null}\n'
        )
        with tempfile.NamedTemporaryFile("w", encoding="utf-8", delete=False) as handle:
            handle.write(payload)
            path = handle.name
        try:
            with self.assertRaises(SchemaError):
                load_cases(path)
        finally:
            os.remove(path)

    def test_unknown_field_fails(self):
        record = {
            "schema_version": 1,
            "id": "x",
            "entity": "PERSON",
            "corpus_class": "positive",
            "input": "Иван Петров",
            "product_expectation": "match",
            "intentional_difference_class": None,
            "intentional_difference_reason": None,
            "extra": True,
        }
        with self.assertRaises(SchemaError):
            load_cases_from_records([record])

    def test_malformed_jsonl_fails(self):
        with tempfile.NamedTemporaryFile("w", encoding="utf-8", delete=False) as handle:
            handle.write("{not json}\n")
            path = handle.name
        try:
            with self.assertRaises(SchemaError):
                load_cases(path)
        finally:
            os.remove(path)

    def test_invalid_schema_version_type_fails(self):
        record = expected_record(schema_version=True)
        with self.assertRaises(SchemaError):
            load_expected_from_records([record])

    def test_bool_offset_fails(self):
        record = expected_record(
            matches=[person_match(span={"start": True, "end": 11})]
        )
        with self.assertRaises(SchemaError):
            load_expected_from_records([record])

    def test_missing_normalized_address_field_fails(self):
        record = expected_record(
            entity="ADDRESS",
            input="ул. Ленина, дом 15, кв. 27",
            matches=[
                {
                    "entity": "ADDRESS",
                    "span": {"start": 0, "end": 26},
                    "span_bytes": {"start": 0, "end": 39},
                    "matched_text": "ул. Ленина, дом 15, кв. 27",
                    "normalized": {
                        "parts": [
                            {
                                "type": "street",
                                "name": "Ленина",
                            }
                        ]
                    },
                }
            ],
        )
        with self.assertRaises(SchemaError):
            load_expected_from_records([record])

    def test_entity_drift_fails(self):
        cases = load_cases(CASES_PATH)
        expected = load_expected(EXPECTED_PATH)
        expected[0] = dict(expected[0])
        expected[0]["entity"] = "ADDRESS"
        with self.assertRaises(SchemaError):
            verify_case_expected_alignment(cases, expected)

    def test_match_entity_drift_fails(self):
        record = expected_record(matches=[person_match(entity="ADDRESS")])
        with self.assertRaises(SchemaError):
            load_expected_from_records([record])

    def test_bad_match_order_fails(self):
        record = expected_record(
            matches=[
                person_match(span={"start": 5, "end": 11}),
                person_match(span={"start": 0, "end": 4}, matched_text="Иван"),
            ]
        )
        with self.assertRaises(SchemaError):
            load_expected_from_records([record])

    def test_empty_match_span_fails(self):
        record = expected_record(matches=[person_match(span={"start": 0, "end": 0})])
        with self.assertRaises(SchemaError):
            load_expected_from_records([record])

    def test_case_expected_input_drift_fails(self):
        cases = load_cases(CASES_PATH)
        expected = load_expected(EXPECTED_PATH)
        expected[0] = dict(expected[0])
        expected[0]["input"] = "changed"
        with self.assertRaises(SchemaError):
            verify_case_expected_alignment(cases, expected)

    def test_reordered_expected_file_fails(self):
        cases = load_cases(CASES_PATH)
        expected = load_expected(EXPECTED_PATH)
        reordered = [expected[1], expected[0]] + expected[2:]
        with self.assertRaises(SchemaError):
            verify_case_expected_alignment(cases, reordered)

    def test_noncanonical_cases_file_fails(self):
        cases = load_cases(CASES_PATH)
        with tempfile.NamedTemporaryFile("w", encoding="utf-8", delete=False) as handle:
            for record in reversed(cases):
                handle.write(json.dumps(record, ensure_ascii=False) + "\n")
            path = handle.name
        try:
            loaded = load_cases(path)
            with self.assertRaises(SchemaError):
                verify_canonical_jsonl_bytes(path, loaded)
        finally:
            os.remove(path)

    def test_deterministic_serialization(self):
        expected = load_expected(EXPECTED_PATH)
        first = dumps_jsonl(expected).encode("utf-8")
        second = dumps_jsonl(expected).encode("utf-8")
        self.assertEqual(first, second)
        with open(EXPECTED_PATH, "rb") as handle:
            committed = handle.read()
        self.assertEqual(first, committed)

    def test_canonicalize_expected_recomputes_byte_spans(self):
        canonical = canonicalize_expected(expected_record())
        self.assertEqual(canonical["entity"], "PERSON")
        self.assertEqual(canonical["matches"][0]["span_bytes"], {"start": 0, "end": 21})

    def test_span_slice_failure_is_schema_error(self):
        record = expected_record(
            matches=[person_match(span_bytes={"start": 0, "end": 20})]
        )
        with self.assertRaises(SchemaError):
            validate_expected(record, 1)

    def test_address_part_type_list_fails_as_schema_error(self):
        record = expected_record(
            entity="ADDRESS",
            input="ул. Ленина, дом 15, кв. 27",
            matches=[
                {
                    "entity": "ADDRESS",
                    "span": {"start": 0, "end": 26},
                    "span_bytes": {"start": 0, "end": 39},
                    "matched_text": "ул. Ленина, дом 15, кв. 27",
                    "normalized": {
                        "parts": [
                            {
                                "type": [],
                                "name": "Ленина",
                                "street_type": "улица",
                            }
                        ]
                    },
                }
            ],
        )
        with self.assertRaises(SchemaError):
            load_expected_from_records([record])

    def test_invalid_utf8_mid_file_fails_as_schema_error(self):
        valid_line = json.dumps(
            {
                "schema_version": 1,
                "id": "x",
                "entity": "PERSON",
                "corpus_class": "positive",
                "input": "Иван Петров",
                "product_expectation": "match",
                "intentional_difference_class": None,
                "intentional_difference_reason": None,
            },
            ensure_ascii=False,
        ).encode("utf-8")
        with tempfile.NamedTemporaryFile("wb", delete=False) as handle:
            handle.write(valid_line + b"\n\xff\xfe\n")
            path = handle.name
        try:
            with self.assertRaises(SchemaError) as ctx:
                load_cases(path)
            self.assertIn("not valid UTF-8", str(ctx.exception))
        finally:
            os.remove(path)

    def test_generate_validates_before_writing(self):
        bad_record = expected_record(
            matches=[person_match(span_bytes={"start": 0, "end": 20})]
        )
        with tempfile.TemporaryDirectory() as tmpdir:
            output_path = os.path.join(tmpdir, "expected.jsonl")
            args = argparse.Namespace(cases=CASES_PATH, output=output_path)
            with mock.patch("extract_live.generate_expected_record", return_value=bad_record):
                with self.assertRaises(SchemaError):
                    reference.cmd_generate(args)
            self.assertFalse(os.path.exists(output_path))


def load_cases_from_records(records):
    with tempfile.NamedTemporaryFile("w", encoding="utf-8", delete=False) as handle:
        for record in records:
            handle.write(json.dumps(record, ensure_ascii=False) + "\n")
        path = handle.name
    try:
        return load_cases(path)
    finally:
        os.remove(path)


def load_expected_from_records(records):
    with tempfile.NamedTemporaryFile("w", encoding="utf-8", delete=False) as handle:
        for record in records:
            handle.write(json.dumps(record, ensure_ascii=False) + "\n")
        path = handle.name
    try:
        return load_expected(path)
    finally:
        os.remove(path)


class ReferenceCliTests(unittest.TestCase):
    def run_reference(self, *args, python_args=None):
        command = [sys.executable]
        if python_args:
            command.extend(python_args)
        command.append(REFERENCE_PATH)
        command.extend(args)
        return subprocess.run(
            command,
            cwd=REPO_ROOT,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            universal_newlines=True,
            check=False,
        )

    def test_offline_verify_passes(self):
        result = self.run_reference(
            "verify",
            "--offline",
            "--cases",
            CASES_PATH,
            "--expected",
            EXPECTED_PATH,
        )
        self.assertEqual(result.returncode, 0, msg=result.stderr)

    def test_live_verify_without_dependencies_fails(self):
        result = self.run_reference(
            "verify",
            "--cases",
            CASES_PATH,
            "--expected",
            EXPECTED_PATH,
            python_args=["-S"],
        )
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Natasha reference dependencies are not installed", result.stderr)


if __name__ == "__main__":
    unittest.main()
