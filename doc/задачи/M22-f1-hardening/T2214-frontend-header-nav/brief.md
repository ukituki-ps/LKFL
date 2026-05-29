# T2214 — Frontend: прототип → код (header, дизайн, страницы-заглушки)

> **Мета-задача.** Декомпозирована на 5 подзадач. Реализация — последовательно (см. порядок ниже).

## Веха

M22-f1-hardening

## Тип

code

## Подзадачи

| Код | Название | Зависит от | Файл |
|-----|----------|------------|------|
| **T2214.5** | DS upgrade v0.1.13 → v0.1.16 | — | `T2214.5-ds-upgrade/brief.md` |
| **T2214.1** | Brand tokens, тема, шрифт | T2214.5 | `T2214.1-brand-tokens/brief.md` |
| **T2214.2** | Shell → AprilProductHeader, навигация | T2214.1 | `T2214.2-shell-header-nav/brief.md` |
| **T2214.3** | Страницы-заглушки (Dashboard, Points, Documents, Support) | T2214.1, T2214.2 | `T2214.3-pages-stubs/brief.md` |
| **T2214.4** | Каталог: EngagementCard + FilterBar + AprilIcon | T2214.1, T2214.3 | `T2214.4-catalog-refactor/brief.md` |

**Порядок выполнения:** T2214.5 → T2214.1 → T2214.2 → T2214.3 → T2214.4 (строго последовательно).

---

## Контекст

Текущий фронтенд — скелет на Mantine defaults (blue theme, sidebar, системный шрифт, эмодзи-иконки).
Прототип (`артефакты/Прототип ЛК физика(1).html`) — зрелый, детализированный дизайн с продуманной дизайн-системой.

**Цель:** привести фронтенд к визуальному соответствию прототипа — header-навигация, дизайн-система, карточки, страницы-заглушки с моками.

### DS upgrade: v0.1.13 → v0.1.16

Выделено в отдельную задачу **T2214.5** (выполняется первой, до всех остальных подзадач).

**`@ukituki-ps/april-ui@0.1.16`** решает все блокирующие проблемы v0.1.13:

| Проблема в v0.1.13 | Решение в v0.1.16 |
|--------------------|--------------------|
| 34 иконки (нужно ~40) | **54 иконки** — покрытие прототипа 35/38 |
| Нет `AprilProductHeader` | ✅ **есть** (`left`/`center`/`right`/`sticky`) |
| Нет `AprilFilterPills` | ✅ **есть** |
| Нет компонентов ТЗ-DS | ✅ **все 17 компонентов** реализованы |
| Нет breaking changes | ✅ чистое добавление |

**4 иконки отсутствуют в DS:** `HeartPulse`, `ShieldCheck`, `ShieldPlus`, `SmartSpeaker`.
Решение: использовать nearest-аналоги из DS (см. таблицу ниже).

### Icon fallback mapping

| Прототип | DS v0.1.16 (использовать) | Примечание |
|----------|---------------------------|------------|
| `heart-pulse` | `AprilIconHeart` | достаточный визуальный аналог |
| `shield-check` | `AprilIconSuccess` (=CheckCircle2) | semantic match |
| `shield-plus` | `AprilIconPlusCircle` | semantic match |
| `smart-speaker` | `AprilIconSmartphone` | nearest функциональный аналог |

### Что меняется

| Элемент | Сейчас | Станет | Подзадача |
|---------|--------|--------|-----------|
| Layout | `AppShell` (header + sidebar + main) | `AprilProductHeader` + `<main>` без sidebar | T2214.2 |
| Навигация | `EmployeeNav` (sidebar, вертикальные кнопки) | Горизонтальные ссылки в header с underline active | T2214.2 |
| Цвет primary | blue (Mantine default) | `#00B33C` (СДЭК-зелёный) | T2214.1 |
| Шрифт | системный стек | Inter (Google Fonts) | T2214.1 |
| Заголовки | Mantine default | `font-weight: 800` | T2214.1 |
| Иконки | эмодзи (❤️🛡️👥) | `AprilIcon*` из `@ukituki-ps/april-ui@0.1.16` | T2214.3, T2214.4 |
| Карточки льгот | Paper 160px с эмодзи | icon 44×44 bg-gray, название, провайдер, footer | T2214.4 |
| Фильтры каталога | FilterBar с Select | `AprilFilterPills` из DS | T2214.4 |
| Dashboard | заглушка «—» | моки по прототипу + StubBadge | T2214.3 |
| Баллы / Документы / Поддержка | заглушки | заглушки по прототипу + StubBadge | T2214.3 |
| Balance | нет | Balance-pill в header `right` | T2214.2 |
| Уведомления | нет | Bell icon в header `right` | T2214.2 |
| Avatar | blue | зелёный круг 34px с инициалами | T2214.2 |
| Провайдеры | локальный `AprilProviders` | `LKFLProviders` (обёртка над DS `AprilProviders` + QueryClientProvider) | T2214.1 |

---

## Ключевые решения

### 1. Провайдеры: `LKFLProviders`, не `AprilProviders`

Локальный `AprilProviders` (строит `QueryClientProvider` + `MantineProvider`) переименован в `LKFLProviders`. Внутри использует `AprilProviders` из `@ukituki-ps/april-ui` как обёртку.

**Причина:** DS `AprilProviders` не содержит `QueryClientProvider`. Конфликт имён → rename.

### 2. Иконочная стратегия: `AprilIcon*` из DS v0.1.16

Использовать иконки через ре-экспорт DS: `AprilIconBell`, `AprilIconCoins` и т.д.
Для 3 отсутствующих — nearest-аналоги из DS (таблица выше).

**Прямой импорт `lucide-react` — НЕ требуется.** Все иконки прототипа покрыты.

**Правило (архитектура/фронтенд.md §Иконки):** только `AprilIcon` + `AprilIcon*` из `@april/ui`. Соблюдается.

### 3. StubBadge — только dev

`StubBadge` рендерится только когда `import.meta.env.DEV === true`. В production — `null`. Удаление StubBadge из production-билда не требует отдельной задачи.

### 4. Brand tokens — временный хардкод

`--brand-*` CSS переменные с СДЭК-зелёным — временные. В M22+ будут переопределяться brand CSS из API tenant'а.

### 5. Header height: 56px vs 58px прототип

DS `AprilProductHeader` — 56px (comfortable). Прототип — 58px. Разница 2px — приемлемо. Если критично — кастомный `style={{ height: 58 }}`.

### 6. FilterBar → AprilFilterPills

DS v0.1.16 содержит `AprilFilterPills` — использовать DS-компонент вместо кастомной реализации.

---

## Зависимости

### npm-зависимости
- `@ukituki-ps/april-ui` ≥ 0.1.16 (апгрейд из 0.1.13 в T2214.5)
- `@ukituki-ps/april-tokens` ≥ 0.1.16 (апгрейд в T2214.5)
- `lucide-react` — **НЕ добавлять** (уже транзитивная зависимость через `@ukituki-ps/april-ui`)

### Нет зависимостей от других задач M22

## Критерии приёмки (агрегированные)

См. критерии приёмки в каждой подзадаче. Сводка:

### DS upgrade (T2214.5)
- [ ] `@ukituki-ps/april-ui@0.1.16` установлен
- [ ] `@ukituki-ps/april-tokens@0.1.16` установлен
- [ ] No breaking changes в импортах существующего кода
- [ ] `npm run build` / `npm run test` / `npm run lint` проходят без ошибок

### Header и layout (T2214.2)
- [ ] `Shell.tsx` использует `AprilProductHeader` (не `AppShell`)
- [ ] Sidebar убран (навигация горизонтальная в header'е)
- [ ] 5 ссылок навигации с underline active-индикатором
- [ ] Balance-pill в правой зоне header'а
- [ ] Bell icon button в правой зоне header'а
- [ ] Avatar с инициалами (зелёный круг 34px) + dropdown
- [ ] Sticky header (прилипает при скролле)
- [ ] Фон страницы `#F2F2F2`
- [ ] Content max-width 1100px, margin auto, padding 28px
- [ ] Мобильная навигация: Burger → Drawer (breakpoint 768px)
- [ ] Admin маршруты: собственная навигация в header'е

### Дизайн-система (T2214.1)
- [ ] Brand CSS tokens (`--brand-*`) подключены
- [ ] Primary color → `#00B33C` (СДЭК-зелёный)
- [ ] Шрифт Inter подключён (Google Fonts)
- [ ] Заголовки: `font-weight: 800`
- [ ] Большие числа: `letter-spacing: -0.5px`
- [ ] `radius-card: 14px`, `radius-btn: 6px`

### Каталог (T2214.4)
- [ ] Карточки по дизайну прототипа (icon 44×44, footer с ценой + badge)
- [ ] `AprilFilterPills` из DS вместо кастомных filter pills
- [ ] Grid 3 колонки (как в прототипе)
- [ ] Icon mapping в EngagementCard через `AprilIcon*`

### Страницы-заглушки (T2214.3)
- [ ] Dashboard: моки по прототипу (3 stat-карточки, льготы, события, быстрые действия)
- [ ] Points: моки по прототипу (баланс, прогресс-бары, транзакции)
- [ ] Documents: моки по прототипу (таблица с бейджами)
- [ ] Support: моки по прототипу (FAQ + форма)
- [ ] StubBadge на каждом заглушечном блоке (только dev)

### Общее
- [ ] Все эмодзи заменены на `AprilIcon*` из DS
- [ ] `npm run dev` → страница визуально соответствует прототипу
- [ ] Unit-тесты не ломаются (или обновлены)
- [ ] E2E тесты не ломаются (или обновлены)
- [ ] `lucide-react` НЕ добавлен в `dependencies` (только транзитивная)

## Модалки (не в этой задаче)

Реализуются через компоненты дизайн-системы `@ukituki-ps/april-ui@0.1.16` — все необходимые компоненты уже доступны:
- `AprilWizard` + `AprilWizardProgress` — универсальный wizard
- `AprilOptionCard` — selectable option card
- `AprilCard` — карточка с header + body
- `AprilFaqItem` — аккордеон FAQ
- `AprilStatCard`, `AprilBalanceCard` — статистика
- `AprilFilterPills` — pill-фильтры
- `AprilTransactionRow` — строки транзакций
- `AprilQuickButton` — кнопки быстрых действий
- `AprilSuccessScreen` — экран успеха wizard

После апгрейда DS — LKFL подключит компоненты и реализует 3 модалки прототипа:
- DMS Wizard (4 шага: опция → оплата → подтверждение → готово)
- Benefit Detail (табы: условия / полис / клиники на карте)
- Mat Capital Wizard (4 шага)

ТЗ-Design-System.md — пометить «Выполнено в v0.1.16».
