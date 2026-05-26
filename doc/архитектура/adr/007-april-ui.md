# ADR-007: April UI + Mantine как frontend foundation

**Статус:** Accepted
**Дата:** 2026-05-22
**Контекст:** М01-создание-описания

## Контекст

Прототип ЛК — чистый HTML/CSS. Для production нужен React + UI kit. April экосистема уже использует `@april/ui` + `@april/tokens` + Mantine.

## Решение

**`@ukituki-ps/april-ui 0.1.13`** — корпоративные компоненты (Modal, Cards, Buttons, Tables).
**`@ukituki-ps/april-tokens 0.1.13`** — CSS variables, цвета, typography.
**`@mantine/core 7.17.8`** — production-ready компоненты (Stepper, Tabs, Accordion, Progress).

**Маппинг прототипа → April UI / Mantine:**

| Прототип | Компонент |
|----------|-----------|
| Cards, badges | `@april/ui` Card, Badge |
| Buttons (primary, outline, ghost) | `@april/ui` Button |
| Forms | Mantine TextInput, Select, Textarea |
| Modal dialogs | Mantine Modal + `@april/ui` AprilModal |
| Wizard (4 шага) | Mantine Stepper |
| Tabs (ДМС: 3 вкладки) | Mantine Tabs |
| Accordion (FAQ) | Mantine Accordion |
| Progress bar | Mantine Progress |
| Tables | Mantine Table |
| Navigation | Mantine Header |

## Следствия

- Подключение через GPR (GitHub Packages Registry)
- `ds-prepare.sh` для загрузки тарбаллов (как в April Profile)
- White-label через CSS переменные (ADR-008)
- Единая DS с April Profile и Worker
