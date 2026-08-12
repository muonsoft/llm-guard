"""JSONL schema version 1 validation and canonical serialization."""

import json
from typing import Any, Dict, List, Mapping, Sequence, Set, Tuple

from spans import codepoint_span_to_bytes, verify_span_slices

SCHEMA_VERSION = 1

ENTITIES = frozenset({"PERSON", "ADDRESS"})
CORPUS_CLASSES = frozenset({"positive", "negative", "ambiguous"})
PRODUCT_EXPECTATIONS = frozenset({"match", "no_match"})
DIFFERENCE_CLASSES = frozenset(
    {"regression", "intentional_difference", "unsupported_out_of_scope"}
)

CASE_REQUIRED = frozenset(
    {
        "schema_version",
        "id",
        "entity",
        "corpus_class",
        "input",
        "product_expectation",
        "intentional_difference_class",
        "intentional_difference_reason",
    }
)

EXPECTED_REQUIRED = frozenset(
    {"schema_version", "id", "entity", "input", "matches"}
)
MATCH_REQUIRED = frozenset(
    {"entity", "span", "span_bytes", "matched_text", "normalized"}
)


class SchemaError(ValueError):
    """Raised when JSONL records fail schema validation."""


def _reject_unknown_fields(record, allowed, context):
    unknown = set(record.keys()) - allowed
    if unknown:
        raise SchemaError(
            "{} contains unknown fields: {}".format(context, ", ".join(sorted(unknown)))
        )


def _require_type(value, expected_type, field):
    if not isinstance(value, expected_type):
        raise SchemaError(
            "{} must be {}, got {}".format(field, expected_type.__name__, type(value).__name__)
        )


def _require_int(value, field):
    if isinstance(value, bool) or not isinstance(value, int):
        raise SchemaError("{} must be int, got {}".format(field, type(value).__name__))


def _require_schema_version(value, field):
    _require_int(value, field)
    if value != SCHEMA_VERSION:
        raise SchemaError("{} must be {}, got {}".format(field, SCHEMA_VERSION, value))


def _require_entity(value, field):
    _require_type(value, str, field)
    if value not in ENTITIES:
        raise SchemaError("{} must be one of {}".format(field, sorted(ENTITIES)))


def _validate_span(span, field, text_len, require_non_empty):
    _require_type(span, dict, field)
    _reject_unknown_fields(span, {"start", "end"}, field)
    for key in ("start", "end"):
        if key not in span:
            raise SchemaError("{} missing {!r}".format(field, key))
        _require_int(span[key], "{}.{}".format(field, key))
    start, end = span["start"], span["end"]
    if start < 0 or end < 0 or start > end or end > text_len:
        raise SchemaError(
            "{} has invalid range ({}, {}) for text length {}".format(field, start, end, text_len)
        )
    if require_non_empty and start == end:
        raise SchemaError("{} must be non-empty".format(field))
    return start, end


def _validate_normalized(value, entity, field):
    _require_type(value, dict, field)
    if entity == "PERSON":
        allowed = {"first", "last", "middle", "nick"}
        _reject_unknown_fields(value, allowed, field)
        for key in allowed:
            if key not in value:
                raise SchemaError("{} missing PERSON field {!r}".format(field, key))
            item = value[key]
            if item is not None:
                _require_type(item, str, "{}.{}".format(field, key))
        return {key: value[key] for key in sorted(allowed)}
    if entity == "ADDRESS":
        allowed = {"parts"}
        _reject_unknown_fields(value, allowed, field)
        if "parts" not in value:
            raise SchemaError("{} missing ADDRESS field 'parts'".format(field))
        parts = value["parts"]
        _require_type(parts, list, "{}.parts".format(field))
        normalized_parts = []
        for index, part in enumerate(parts):
            part_field = "{}.parts[{}]".format(field, index)
            _require_type(part, dict, part_field)
            if "type" not in part:
                raise SchemaError("{} missing field 'type'".format(part_field))
            part_type = part["type"]
            _require_type(part_type, str, "{}.type".format(part_field))
            if part_type not in {"settlement", "street", "building", "room", "region"}:
                raise SchemaError("{} has invalid type {!r}".format(part_field, part_type))
            if part_type == "settlement":
                required_part = ("type", "name", "settlement_type")
            elif part_type == "street":
                required_part = ("type", "name", "street_type")
            elif part_type == "building":
                required_part = ("type", "number", "building_type")
            elif part_type == "room":
                required_part = ("type", "number", "room_type")
            else:
                required_part = ("type", "name", "region_type")
            allowed_part = set(required_part)
            _reject_unknown_fields(part, allowed_part, part_field)
            for key in required_part:
                if key not in part:
                    raise SchemaError("{} missing field {!r}".format(part_field, key))
            normalized_part = {}
            for key in sorted(required_part):
                value_item = part[key]
                if key.endswith("_type"):
                    if value_item is not None:
                        _require_type(value_item, str, "{}.{}".format(part_field, key))
                elif key == "number":
                    _require_int(value_item, "{}.{}".format(part_field, key))
                else:
                    _require_type(value_item, str, "{}.{}".format(part_field, key))
                normalized_part[key] = value_item
            normalized_parts.append(normalized_part)
        return {"parts": normalized_parts}
    raise SchemaError("unknown entity {!r}".format(entity))


def validate_case(record, line_no):
    context = "case line {}".format(line_no)
    _require_type(record, dict, context)
    _reject_unknown_fields(record, CASE_REQUIRED, context)
    for key in CASE_REQUIRED:
        if key not in record:
            raise SchemaError("{} missing required field {!r}".format(context, key))
    _require_schema_version(record["schema_version"], "{}.schema_version".format(context))
    _require_type(record["id"], str, "{}.id".format(context))
    if not record["id"]:
        raise SchemaError("{} id must be non-empty".format(context))
    _require_entity(record["entity"], "{}.entity".format(context))
    corpus_class = record["corpus_class"]
    _require_type(corpus_class, str, "{}.corpus_class".format(context))
    if corpus_class not in CORPUS_CLASSES:
        raise SchemaError("{} corpus_class must be one of {}".format(context, sorted(CORPUS_CLASSES)))
    _require_type(record["input"], str, "{}.input".format(context))
    product_expectation = record["product_expectation"]
    _require_type(product_expectation, str, "{}.product_expectation".format(context))
    if product_expectation not in PRODUCT_EXPECTATIONS:
        raise SchemaError(
            "{} product_expectation must be one of {}".format(
                context, sorted(PRODUCT_EXPECTATIONS)
            )
        )
    diff_class = record["intentional_difference_class"]
    diff_reason = record["intentional_difference_reason"]
    if diff_class is not None:
        _require_type(diff_class, str, "{}.intentional_difference_class".format(context))
        if diff_class not in DIFFERENCE_CLASSES:
            raise SchemaError("{} invalid intentional_difference_class {!r}".format(context, diff_class))
        _require_type(diff_reason, str, "{}.intentional_difference_reason".format(context))
        if not diff_reason:
            raise SchemaError("{} intentional_difference_reason required when class is set".format(context))
    else:
        if diff_reason is not None:
            raise SchemaError(
                "{} intentional_difference_reason must be null when class is null".format(context)
            )
    return canonicalize_case(record)


def _validate_match_order(matches, line_no):
    previous = None
    for index, match in enumerate(matches):
        start = match["span"]["start"]
        end = match["span"]["end"]
        current = (start, end)
        if previous is not None and current < previous:
            raise SchemaError(
                "expected line {} matches must be in nondecreasing span order".format(line_no)
            )
        previous = current


def validate_match(match, text, line_no, index, case_entity):
    context = "expected line {} match[{}]".format(line_no, index)
    _require_type(match, dict, context)
    _reject_unknown_fields(match, MATCH_REQUIRED, context)
    for key in MATCH_REQUIRED:
        if key not in match:
            raise SchemaError("{} missing required field {!r}".format(context, key))
    _require_entity(match["entity"], "{}.entity".format(context))
    if match["entity"] != case_entity:
        raise SchemaError(
            "{} entity {!r} must equal case entity {!r}".format(
                context, match["entity"], case_entity
            )
        )
    start, end = _validate_span(
        match["span"], "{}.span".format(context), len(text), require_non_empty=True
    )
    byte_start, byte_end = _validate_span(
        match["span_bytes"],
        "{}.span_bytes".format(context),
        len(text.encode("utf-8")),
        require_non_empty=True,
    )
    _require_type(match["matched_text"], str, "{}.matched_text".format(context))
    try:
        verify_span_slices(text, start, end, byte_start, byte_end, match["matched_text"])
    except ValueError as exc:
        raise SchemaError("{} span validation failed: {}".format(context, exc)) from exc
    normalized = _validate_normalized(match["normalized"], case_entity, "{}.normalized".format(context))
    return {
        "entity": case_entity,
        "span": {"end": end, "start": start},
        "span_bytes": {"end": byte_end, "start": byte_start},
        "matched_text": match["matched_text"],
        "normalized": normalized,
    }


def validate_expected(record, line_no):
    context = "expected line {}".format(line_no)
    _require_type(record, dict, context)
    _reject_unknown_fields(record, EXPECTED_REQUIRED, context)
    for key in EXPECTED_REQUIRED:
        if key not in record:
            raise SchemaError("{} missing required field {!r}".format(context, key))
    _require_schema_version(record["schema_version"], "{}.schema_version".format(context))
    _require_type(record["id"], str, "{}.id".format(context))
    if not record["id"]:
        raise SchemaError("{} id must be non-empty".format(context))
    _require_entity(record["entity"], "{}.entity".format(context))
    _require_type(record["input"], str, "{}.input".format(context))
    matches = record["matches"]
    _require_type(matches, list, "{}.matches".format(context))
    case_entity = record["entity"]
    normalized_matches = [
        validate_match(match, record["input"], line_no, index, case_entity)
        for index, match in enumerate(matches)
    ]
    _validate_match_order(normalized_matches, line_no)
    return {
        "entity": case_entity,
        "id": record["id"],
        "input": record["input"],
        "matches": normalized_matches,
        "schema_version": SCHEMA_VERSION,
    }


def canonicalize_case(record):
    return {
        "corpus_class": record["corpus_class"],
        "entity": record["entity"],
        "id": record["id"],
        "input": record["input"],
        "intentional_difference_class": record["intentional_difference_class"],
        "intentional_difference_reason": record["intentional_difference_reason"],
        "product_expectation": record["product_expectation"],
        "schema_version": SCHEMA_VERSION,
    }


def canonicalize_expected(record):
    text = record["input"]
    entity = record["entity"]
    matches = []
    for match in record["matches"]:
        start = match["span"]["start"]
        end = match["span"]["end"]
        byte_start, byte_end = codepoint_span_to_bytes(text, start, end)
        matches.append(
            {
                "entity": entity,
                "matched_text": match["matched_text"],
                "normalized": match["normalized"],
                "span": {"end": end, "start": start},
                "span_bytes": {"end": byte_end, "start": byte_start},
            }
        )
    return {
        "entity": entity,
        "id": record["id"],
        "input": text,
        "matches": matches,
        "schema_version": SCHEMA_VERSION,
    }


def dumps_canonical(record):
    return json.dumps(record, ensure_ascii=False, sort_keys=True, separators=(",", ":"))


def dumps_jsonl(records):
    lines = [dumps_canonical(record) for record in records]
    if not lines:
        return ""
    return "\n".join(lines) + "\n"


def verify_canonical_jsonl_bytes(path, records):
    with open(path, "rb") as handle:
        committed = handle.read()
    canonical = dumps_jsonl(records).encode("utf-8")
    if committed != canonical:
        raise SchemaError("{} is not canonical UTF-8 JSONL".format(path))


def load_jsonl(path):
    records = []
    try:
        handle = open(path, encoding="utf-8")
    except UnicodeDecodeError as exc:
        raise SchemaError("{} is not valid UTF-8: {}".format(path, exc)) from exc
    with handle:
        try:
            for line_no, line in enumerate(handle, start=1):
                stripped = line.strip()
                if not stripped:
                    raise SchemaError("{}:{} empty line is not allowed".format(path, line_no))
                try:
                    record = json.loads(stripped)
                except json.JSONDecodeError as exc:
                    raise SchemaError("{}:{} invalid JSON: {}".format(path, line_no, exc)) from exc
                records.append(record)
        except UnicodeDecodeError as exc:
            raise SchemaError("{} is not valid UTF-8: {}".format(path, exc)) from exc
    return records


def load_cases(path):
    records = load_jsonl(path)
    seen = set()
    validated = []
    for line_no, record in enumerate(records, start=1):
        case = validate_case(record, line_no)
        if case["id"] in seen:
            raise SchemaError("duplicate case id {!r}".format(case["id"]))
        seen.add(case["id"])
        validated.append(case)
    return validated


def load_expected(path):
    records = load_jsonl(path)
    seen = set()
    validated = []
    for line_no, record in enumerate(records, start=1):
        expected = validate_expected(record, line_no)
        if expected["id"] in seen:
            raise SchemaError("duplicate expected id {!r}".format(expected["id"]))
        seen.add(expected["id"])
        validated.append(expected)
    return validated


def verify_case_expected_alignment(cases, expected):
    if len(cases) != len(expected):
        raise SchemaError(
            "expected record count {} does not match case count {}".format(len(expected), len(cases))
        )
    if [record["id"] for record in expected] != [case["id"] for case in cases]:
        raise SchemaError("expected record order must match case order")
    for case, record in zip(cases, expected):
        if record["input"] != case["input"]:
            raise SchemaError(
                "input drift for case {!r}: case has {!r}, expected has {!r}".format(
                    case["id"], case["input"], record["input"]
                )
            )
        if record["entity"] != case["entity"]:
            raise SchemaError(
                "entity drift for case {!r}: case has {!r}, expected has {!r}".format(
                    case["id"], case["entity"], record["entity"]
                )
            )
