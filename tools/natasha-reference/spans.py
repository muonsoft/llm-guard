"""UTF-8 byte span conversion for Python codepoint offsets."""


def codepoint_span_to_bytes(text, start, end):
    """Convert a Python codepoint half-open span to UTF-8 byte offsets."""
    if start < 0 or end < 0 or start > end or end > len(text):
        raise ValueError(
            "invalid codepoint span ({}, {}) for text length {}".format(start, end, len(text))
        )
    byte_start = len(text[:start].encode("utf-8"))
    byte_end = len(text[:end].encode("utf-8"))
    return byte_start, byte_end


def verify_span_slices(text: str, start: int, end: int, byte_start: int, byte_end: int, matched_text: str) -> None:
    """Verify codepoint and byte slices both equal matched_text."""
    codepoint_slice = text[start:end]
    byte_slice = text.encode("utf-8")[byte_start:byte_end].decode("utf-8")
    if codepoint_slice != matched_text:
        raise ValueError(
            "codepoint slice {!r} != matched_text {!r}".format(codepoint_slice, matched_text)
        )
    if byte_slice != matched_text:
        raise ValueError(
            "byte slice {!r} != matched_text {!r}".format(byte_slice, matched_text)
        )
    expected_bytes = codepoint_span_to_bytes(text, start, end)
    if (byte_start, byte_end) != expected_bytes:
        raise ValueError(
            "byte span ({}, {}) != expected ({}, {})".format(
                byte_start, byte_end, expected_bytes[0], expected_bytes[1]
            )
        )
