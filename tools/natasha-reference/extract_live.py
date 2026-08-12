"""Pinned Natasha extractor integration for live generate/verify."""

from spans import codepoint_span_to_bytes


def _import_natasha():
    try:
        from natasha import AddressExtractor, NamesExtractor
        from natasha.extractors import normalize_text
    except ImportError as exc:
        raise RuntimeError(
            "Natasha reference dependencies are not installed. "
            "Install tools/natasha-reference/requirements.lock in the pinned "
            "Python 3.6.15 environment documented in README.md."
        ) from exc
    return NamesExtractor, AddressExtractor, normalize_text


def _normalize_name_fact(fact):
    return {
        "first": fact.first,
        "last": fact.last,
        "middle": fact.middle,
        "nick": fact.nick,
    }


def _normalize_address_fact(fact):
    parts = []
    for part in fact.parts:
        cls_name = type(part).__name__
        if cls_name == "Settlement":
            parts.append(
                {
                    "type": "settlement",
                    "name": part.name,
                    "settlement_type": part.type,
                }
            )
        elif cls_name == "Street":
            parts.append(
                {
                    "type": "street",
                    "name": part.name,
                    "street_type": part.type,
                }
            )
        elif cls_name == "Building":
            parts.append(
                {
                    "type": "building",
                    "number": part.number,
                    "building_type": part.type,
                }
            )
        elif cls_name == "Room":
            parts.append(
                {
                    "type": "room",
                    "number": part.number,
                    "room_type": part.type,
                }
            )
        elif cls_name == "Region":
            parts.append(
                {
                    "type": "region",
                    "name": part.name,
                    "region_type": part.type,
                }
            )
        else:
            raise RuntimeError("unsupported address part type: {}".format(cls_name))
    return {"parts": parts}


def _match_record(text, entity, span, fact):
    start, end = span
    byte_start, byte_end = codepoint_span_to_bytes(text, start, end)
    matched_text = text[start:end]
    if entity == "PERSON":
        normalized = _normalize_name_fact(fact)
    elif entity == "ADDRESS":
        normalized = _normalize_address_fact(fact)
    else:
        raise RuntimeError("unsupported entity: {}".format(entity))
    return {
        "entity": entity,
        "span": {"start": start, "end": end},
        "span_bytes": {"start": byte_start, "end": byte_end},
        "matched_text": matched_text,
        "normalized": normalized,
    }


def _assert_normalized_input(text, normalize_text):
    normalized = normalize_text(text)
    if normalized != text:
        raise RuntimeError(
            "Natasha normalize_text changed case input: {!r} -> {!r}".format(text, normalized)
        )


def generate_expected_record(case):
    NamesExtractor, AddressExtractor, normalize_text = _import_natasha()
    text = case["input"]
    entity = case["entity"]
    _assert_normalized_input(text, normalize_text)
    if entity == "PERSON":
        extractor = NamesExtractor()
    else:
        extractor = AddressExtractor()
    matches = [
        _match_record(text, entity, match.span, match.fact) for match in extractor(text)
    ]
    return {
        "schema_version": 1,
        "id": case["id"],
        "entity": entity,
        "input": text,
        "matches": matches,
    }
