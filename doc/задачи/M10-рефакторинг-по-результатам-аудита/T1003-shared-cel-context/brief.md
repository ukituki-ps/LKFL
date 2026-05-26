# T1003 — Shared CELContext pkg (Platform ↔ Billing)

## Веха

M10-рефакторинг-по-результатам-аудита

## Контекст

`CELContext struct` сейчас задокументирован как часть `platform/internal/cel/context.go`.
Billing имеет собственный `rule_engine/` с embedded cel-go evaluation.

**Проблема:**
ADR-021 говорит: "изменение CEL context schema → синхронный update platform + billing (cel.go types)".
Это fragile contract между двумя Go modules с разными go.mod.

Таблица из модули.md:
```
cel/      → Platform internal, LLM Proxy, Redis DB 4
rule_engine/ → Billing internal, cel-go (своя копия)
```

Если Platform добавит поле в CELContext (например `game.loyalty_level` в M09) — Billing не увидит изменения. Schema drift = runtime error при evaluation.

**Решение — shared/pkg/cel-context:**
Вынести `CELContext struct` и helper `BuildCELContext()` в shared Go package:
```
shared/
  pkg/
    cel-context/
      context.go     # CELContext struct
      builder.go     # BuildCELContext(ctx, user, offer, extra) → *CELContext
      tag_resolver.go # TagResolver (P1)
```

Обе стороны импортируют `shared/pkg/cel-context`. Go compiler гарантирует consistency.

**Биллинг может импортировать shared/:**
Monorepo: billing/ импортирует shared/ через `replace` directive в go.mod. ADR-011 monorepo это поддерживает.

### Файлы-мишени

| Действие | Файл |
|---|-|-|
| shared/pkg/cel-context | `архитектура/пакеты-platform.md` — новый shared pkg |
| cel/ уменьшен | `архитектура/пакеты-platform.md` — из buildContext → import |
| rule_engine/ | `архитектура/модули.md` — Billing → shared/pkg/cel-context |
| ADR-021 update | `архитектура/adr/021-cel-unified-rule-engine.md` — shared schema |
| ADR-011 update | `архитектура/adr/011-monorepo.md` — shared/ для cross-service types |
| Обновить README | `архитектура/README.md` — ссылка на shared/pkg |
| Зависимости | `архитектура/модули.md` — "синхронный update" → "compiler guaranteed" |
| Создать ADR | `архитектура/adr/025-shared-cel-context.md` |

### Критерии приёмки

- [ ] `архитектура/пакеты-platform.md` — shared/pkg/cel-context описан
- [ ] CELContext struct (User, Tags, Benefit, Game поля) задокументирован в shared
- [ ] cel/ package использует import (не собственную копию)
- [ ] Billing rule_engine/ использует import (не собственную копию)
- [ ] "синхронный update platform + billing" → "Go compiler guarantees schema consistency"
- [ ] ADR-021 обновлён — shared schema section добавлена
- [ ] Создан ADR-025: shared cross-service types (cel-context как pattern)
- [ ] `архитектура/модули.md` — зависимости cel/ и rule_engine/ → shared/
- [ ] `архитектура/cel-engine.md` обновлён — references shared/pkg вместо platform/cel/
