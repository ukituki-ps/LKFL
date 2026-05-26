# T0709 — Общая проверка: перекрёстные ссылки и консистентность

## Веха

M07-аудит-и-рефакторинг

## Контекст

После выполнения T0701-T0708 вся документация изменена. Необходима финальная проверка:

1. Все cross-references между файлами работают
2. Нет stale references к удалённым модулям (например, `льготы.md` и `активности.md` → redirect на `engagement.md`)
3. Нет circular references
4. ADR список в `архитектура/README.md` полный (ADR-013 → ADR-020)
5. Status-таблицы в `задачи/README.md` актуальны
6. `Контекст/настраиваемость.md` матрица не содержит противоречий с новыми архитектурными решениями

**Чек-лист проверки:**

| Проверка | Файлы |
|---|---|
| Все md файлы в `архитектура/` | прочитать список dir → сверить с README table |
| Все md files в `контекст/` | readme table → actual |
| Все md files в `спецификация/` | readme table → actual |
| Cross-refs `[link](path)` in all files | grep `\]\(` → validate each target exists |
| ADR numbering | ADR-001 до ADR-020 — нет пропусков, нет дубликатов |
| Task numbering | T0101 до T0709 — sequential, no gaps |
| Namespace consistency | `internal/*/` references in all docs |
| Service count | Platform + Billing + Integrations + Payment-gateway = 4 Go services → все refs updated |

### Файлы-мишени

| Действие | Файл |
|---|------|
| Fix stale links | ALL — wherever found |
| Fix ADR numbers | `архитектура/README.md` |
| Fix task refs | wherever found |
| Fix service count | wherever "3 сервис" → "4 сервиса" |
| Final status | `задачи/README.md` → M07 complete |

### Критерии приёмки

- [x] Все ADR-001 → ADR-020 в `архитектура/README.md` — 20 ADR'ов
- [x] Все `[link](path)` в doc/ работают (не broken)
- [x] Service count consistently 4 in all docs
- [x] Namespace `billing.*` uniform в billing-движок.md и modules.md
- [x] Матрица настраиваемости не противоречит архитектурным решениям
- [x] Файлы-мишени все перечислены выше
