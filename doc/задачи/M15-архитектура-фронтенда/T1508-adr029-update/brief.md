# T1508 — Перенос актуализации DS gap analysis в ADR-029

> **Тип задачи:** перенос готового анализа из brief в ADR-файл. Не исследование — решение уже принято.

## Веха

M15-архитектура-фронтенда

## Контекст

ADR-029 (`архитектура/adr/029-ds-components-gap-tz.md`) составлен на основе ранней версии DS. С тех пор `@ukituki-ps/april-ui` значительно вырос: появились `ProductHeaderToolbar`, `ProductSidebarNavigation`, `CardListColumn`, `AprilMobileShellBar`, `AprilVaulBottomSheet`, `AprilJsonTreeEditor`, `AprilJsonSchemaForm`, `AprilGradientSegmentedControl` и другие компоненты.

Текущий ADR-029 содержит 11 «недостающих» компонентов (DS-001 → DS-011). Проверка показала: из 11 — только 2 действительно требуют решения.

## Что сделать

### 1. Пересмотреть все 11 компонентов

| DS-ID | Компонент | Статус ADR-029 | Реальность | Решение |
|-------|-----------|----------------|------------|---------|
| DS-001 | `WizardContainer` | 🔴 критический | Mantine `Stepper` + `AprilModal` покрывают 80%. Wizard в LKFL — JSON-driven renderer (ADR-019), не DS-компонент | **Обёртка в LKFL:** `modals/Wizard.tsx` над `Mantine Stepper` + `AprilModal`. Не добавлять в DS — специфичен для LKFL |
| DS-002 | `TransactionList` | 🔴 критический | `Mantine Table` + `CardListColumn` + токены покрывают | **Компонент в LKFL:** `components/TransactionRow.tsx` + `Mantine Table`. Простой — не нужен в DS |
| DS-003 | `StatCard` | 🔴 критический | `Mantine Paper` + токены. Прототип — простая карточка с числом | **YAGNI:** `Mantine Paper` + CSS. Не нужен отдельный компонент |
| DS-004 | `EventsFeed` | 🟡 средний | `CardListColumn` покрывает список событий | **Использовать `CardListColumn`** из DS |
| DS-005 | `PolicyCard` | 🟡 средний | `Mantine Card` + токены. Модалка ДМС — таб «Мой полис» | **`Mantine Card`** — достаточно |
| DS-006 | `ClinicMapList` | 🟡 средний | iframe карта (ADR-029 стр. 418). Не DS-компонент | **iframe + `Mantine Card`** — не DS |
| DS-007 | `TopTabNavigation` | 🟡 средний | `ProductHeaderToolbar` + `AprilGradientSegmentedControl` покрывают | **Использовать DS компоненты** |
| DS-008 | `BalancePill` | 🟢 низкий | `Mantine Badge`/`Pill` + токены | **`Mantine Badge`** — достаточно |
| DS-009 | `DocumentRow` | 🟢 низкий | `Mantine Table` + токены | **`Mantine Table` + `components/DocumentRow.tsx`** |
| DS-010 | `SupportFAQ` | 🟢 низкий | `Mantine Accordion` (ADR-007) | **`Mantine Accordion`** — уже в прототипе |
| DS-011 | `FilterPills` | 🟢 низкий | `Mantine Badge` + `FacetedSearch` (ADR-029 §1.1) | **`Mantine Badge` + DS `FacetedSearch`** |

### 2. Обновить ADR-029

- Заменить таблицу §1.1 (Покрытие) — добавить новые DS компоненты
- Заменить таблицу §1.2 (Промежуточное) — обновить статус
- Заменить таблицу §1.3 (Полное отсутствие) — сократить до 2 реальных gap'ов
- Добавить раздел: **Компоненты LKFL vs DS** — критерий «добавлять в DS или делать в LKFL»
  - В DS: если используется в 2+ продуктах April
  - В LKFL: если специфичен для льготной платформы

### 3. Обновить ссылки

- `архитектура/фронтенд.md` (T1502) — ссылка на обновлённый ADR-029
- `NAVIGATION.md` — обновить ссылку на ADR-029 (или делегировать T1507)

## Результат

- `архитектура/adr/029-ds-components-gap-tz.md` — обновлён
- 11 → 2 реальных gap'а (WizardContainer как LKFL-обёртка, TransactionList как LKFL-компонент)
- 9 компонентов покрываются существующей DS или Mantine

## Критерии приёмки

- [ ] Все 11 компонентов пересмотрены
- [ ] Таблицы §1.1, §1.2, §1.3 обновлены
- [ ] Добавлен раздел «Компоненты LKFL vs DS»
- [ ] Ссылки в NAVIGATION.md обновлены (или делегированы T1507 — финальное обновление)
