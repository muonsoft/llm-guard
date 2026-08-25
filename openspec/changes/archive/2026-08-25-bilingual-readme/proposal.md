## Why

Текущий README содержит корректную MVP-информацию, но смешивает публичное
позиционирование, внутреннюю milestone-историю, подробности отдельных detectors и
release evidence. Новому consumer нужен короткий и проверяемый вход в проект на
английском и русском языках без потери security boundary и известных
ограничений.

## What Changes

- `README.md` становится основным англоязычным OSS entry point.
- Добавляется эквивалентный `README.ru.md`; оба файла содержат заметный language
  switch и одинаковые продуктовые, security и limitation claims.
- Quick start сокращается до одного компилируемого `Mask → LLM → Restore` flow,
  согласованного с public examples и canonical import path.
- Built-in coverage, defaults, operational properties и deeper documentation
  группируются в компактные таблицы и ссылки вместо milestone-секций.
- Публичная документация продолжает уточнять исходный scope из
  `docs/light_llm_guard_go_mvp_plan.md`: precision-oriented prefilter, RU-first,
  pure Go, caller-owned state и отсутствие обязательного gateway/API.
- Runtime, exported Go API и detector behavior не меняются. **BREAKING** изменений
  нет.

## Capabilities

### New Capabilities

Нет.

### Modified Capabilities

- `oss-distribution`: public entry point получает две эквивалентные языковые
  версии и проверяемый минимум сведений для безопасной consumer-интеграции.

## Impact

Затрагиваются `README.md`, новый `README.ru.md`, OpenSpec capability
`oss-distribution` и milestone/evidence документы. Новых runtime dependencies,
API symbols, CI services или release publication действий нет.
