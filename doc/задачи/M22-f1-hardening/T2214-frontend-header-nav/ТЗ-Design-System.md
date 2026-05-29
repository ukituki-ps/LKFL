# ТЗ для Design System Agent — @ukituki-ps/april-ui

> **Статус:** ✅ **Выполнено в v0.1.16.**
> Все 17 компонентов из этого ТЗ реализованы и экспортированы в `@ukituki-ps/april-ui@0.1.16`.
> Прототип-референс: `артефакты/Прототип ЛК физика(1).html`

---

## Что реализовано в v0.1.16

### Компоненты — все 17 из ТЗ ✅

| Компонент | Статус v0.1.16 | Примечание |
|-----------|----------------|------------|
| `AprilFilterPills` | ✅ экспортирован | filter pills |
| `AprilWizard` + `AprilWizardProgress` | ✅ экспортированы | wizard + progress bar |
| `AprilStatCard` | ✅ экспортирован | stat card с label/value/hint |
| `AprilCard` | ✅ экспортирован | card с header + scrollable body |
| `AprilBenefitRow` | ✅ экспортирован | benefit row (icon + name + meta + badge) |
| `AprilEventRow` | ✅ экспортирован | event row (icon + text + time) |
| `AprilQuickButton` | ✅ экспортирован | quick action button (icon + text) |
| `AprilBalanceCard` | ✅ экспортирован | balance card с progress bars |
| `AprilTransactionRow` | ✅ экспортирован | transaction row (icon + type + name + amount) |
| `AprilFaqItem` | ✅ экспортирован | FAQ accordion item |
| `AprilOptionCard` | ✅ экспортирован | selectable option card |
| `AprilPayOptionCard` | ✅ экспортирован | payment option card |
| `AprilFormInput` / `AprilFormTextarea` / `AprilFormSelect` | ✅ экспортированы | form controls |
| `AprilConfirmCheckbox` | ✅ экспортирован | confirmation checkbox |
| `AprilSuccessScreen` | ✅ экспортирован | success screen for wizard |
| `AprilConfirmDoc` | ✅ экспортирован | document preview for confirmation |
| `AprilProductHeader` | ✅ экспортирован | product header (left/center/right/sticky) |

### Иконки — 54 штуки (было 34 в v0.1.13)

Новые иконки в v0.1.16: `ArrowUpCircle`, `Baby`, `Brain`, `CheckCircle`, `Circle`, `Clock`, `Coffee`, `Coins`, `Download`, `Dumbbell`, `Gift`, `GraduationCap`, `Heart`, `Languages`, `MapPin`, `PlusCircle`, `ShoppingBag`, `Smartphone`, `Sparkles`, `UserPlus`.

### Дополнительные компоненты (beyond ТЗ)

v0.1.16 также содержит компоненты, не входившие в ТЗ:

| Компонент | Назначение |
|-----------|-----------|
| `AchievementBadge` | Бейдж ачивки |
| `BenefitCard` | Карточка льготы (доменный компонент) |
| `ClinicMapList` | Список клиник с картой |
| `DocumentRow` | Строка документа |
| `EmptyStateIllustration` | Иллюстрация пустого состояния |
| `EventsFeed` | Лента событий |
| `FAQItem` | FAQ item |
| `FacetedSearch` | Faceted search |
| `GamifiedProgress` | Геймифицированный прогресс |
| `PolicyCard` | Карточка полиса (градиент) |
| `QuickActionsGrid` | Сетка быстрых действий |
| `SmartBundle` | Smart bundle |
| `StatCard` | Stat card (альтернатива AprilStatCard) |
| `StatusChip` | Status chip |
| `SupportFAQ` | Support FAQ |
| `TopTabNavigation` | Top tab navigation |
| `TransactionList` | Список транзакций |
| `WizardContainer` + `WizardFooter` | Wizard обёртка + footer |

### Breaking changes v0.1.13 → v0.1.16

**Нет.** Чистое добавление. Все экспорты v0.1.13 сохранены.

### Peer dependencies v0.1.16

```json
{
  "@emotion/react": "^11.0.0",
  "@mantine/core": "^7.0.0",
  "@mantine/hooks": "^7.9.0",
  "react": "^18.0.0",
  "react-dom": "^18.0.0",
  "react-hook-form": "^7.0.0",     // optional
  "zod": "^3.22.0"                   // optional
}
```

---

## Не добавлять в DS (продуктовые компоненты)

Эти компоненты остаются в продукте LKFL (не в DS):

| Компонент | Почему не в DS |
|-----------|----------------|
| BenefitCard (каталог) | Привязан к домену льгот (provider, price, badge) — **НО** `BenefitCard` есть в v0.1.16 как обобщённая версия |
| PolicyCard | **ЕСТЬ в v0.1.16** как `PolicyCard` |
| ClinicMap | **ЕСТЬ в v0.1.16** как `ClinicMapList` |

LKFL может использовать как DS-версии, так и доменные обёртки.

---

## Icon fallback для LKFL

3 иконки прототипа отсутствуют в DS (даже в v0.1.16):

| Прототип | DS v0.1.16 (использовать) | Примечание |
|----------|---------------------------|------------|
| `heart-pulse` | `AprilIconHeart` | достаточный визуальный аналог |
| `shield-check` | `AprilIconSuccess` (=CheckCircle2) | semantic match |
| `shield-plus` | `AprilIconPlusCircle` | semantic match |
| `smart-speaker` | `AprilIconSmartphone` | nearest функциональный аналог |

Эти иконки можно добавить в DS в будущем (v0.1.17+).

---

## KPI (из оригинального ТЗ)

- [x] Все 17 компонентов экспортированы из `@ukituki-ps/april-ui` (v0.1.16)
- [x] TypeScript types (`index.d.ts`) — полные пропсы
- [x] Все компоненты используют `AprilIcon` (не сырой Lucide)
- [x] Все размеры/цвета — через CSS tokens
- [x] Primary color — через Mantine `primaryColor`
- [x] Компоненты работают в `comfortable` и `compact` density
- [x] `data-testid` проп на каждом компоненте
- [x] semver bump: minor (0.1.13 → 0.1.16)
