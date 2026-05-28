# T2214 — Frontend: прототип → код (header, дизайн, страницы-заглушки)

> **Мета-задача.** Декомпозирована на 4 подзадачи. Реализация — последовательно (см. порядок ниже).

## Веха

M22-f1-hardening

## Тип

code

## Подзадачи

| Код | Название | Зависит от | Файл |
|-----|----------|------------|------|
| **T2214.1** | Brand tokens, тема, шрифт | — | `T2214.1-brand-tokens/brief.md` |
| **T2214.2** | Shell → AprilProductHeader, навигация | T2214.1 | `T2214.2-shell-header-nav/brief.md` |
| **T2214.3** | Страницы-заглушки (Dashboard, Points, Documents, Support) | T2214.1, T2214.2 | `T2214.3-pages-stubs/brief.md` |
| **T2214.4** | Каталог: EngagementCard + FilterBar + Lucide | T2214.1, T2214.3 | `T2214.4-catalog-refactor/brief.md` |

**Порядок выполнения:** T2214.1 → T2214.2 → T2214.3 → T2214.4 (последовательно; параллельно — T2214.3 и T2214.4 после T2214.2).

---

## Контекст

Текущий фронтенд — скелет на Mantine defaults (blue theme, sidebar, системный шрифт, эмодзи-иконки).
Прототип (`артефакты/Прототип ЛК физика(1).html`) — зрелый, детализированный дизайн с продуманной дизайн-системой.

**Цель:** привести фронтенд к визуальному соответствию прототипа — header-навигация, дизайн-система, карточки, страницы-заглушки с моками.

### Что меняется

| Элемент | Сейчас | Станет | Подзадача |
|---------|--------|--------|-----------|
| Layout | `AppShell` (header + sidebar + main) | `AprilProductHeader` + `<main>` без sidebar | T2214.2 |
| Навигация | `EmployeeNav` (sidebar, вертикальные кнопки) | Горизонтальные ссылки в header с underline active | T2214.2 |
| Цвет primary | blue (Mantine default) | `#00B33C` (СДЭК-зелёный) | T2214.1 |
| Шрифт | системный стек | Inter (Google Fonts) | T2214.1 |
| Заголовки | Mantine default | `font-weight: 800` | T2214.1 |
| Иконки | эмодзи (❤️🛡️👥) | Lucide SVG (`lucide-react`) | T2214.3, T2214.4 |
| Карточки льгот | Paper 160px с эмодзи | icon 44×44 bg-gray, название, провайдер, footer | T2214.4 |
| Фильтры каталога | FilterBar с Select | Filter pills (pill-кнопки с зелёным active) | T2214.4 |
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

### 2. Иконочная стратегия: `lucide-react` напрямую

DS v0.1.13 экспортирует ~25 иконок через ре-экспорт. Прототипу нужно ~40. Использовать прямые импорты `lucide-react`.

### 3. StubBadge — только dev

`StubBadge` рендерится только когда `import.meta.env.DEV === true`. В production — `null`. Удаление StubBadge из production-билда не требует отдельной задачи.

### 4. Brand tokens — временный хардкод

`--brand-*` CSS переменные с СДЭК-зелёным — временные. В M22+ будут переопределяться brand CSS из API tenant'а.

### 5. Header height: 56px vs 58px прототип

DS `AprilProductHeader` — 56px (comfortable). Прототип — 58px. Разница 2px — приемлемо. Если критично — кастомный `style={{ height: 58 }}`.

---

## Зависимости

### npm-зависимости
- `lucide-react` — добавить в `package.json` (устанавливается в T2214.3)
- `@ukituki-ps/april-ui` ≥ 0.1.13 (уже в package.json)

### Нет зависимостей от других задач M22

## Критерии приёмки (агрегированные)

См. критерии приёмки в каждой подзадаче. Сводка:

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
- [ ] Filter pills вместо selects (pill-кнопки с зелёным active)
- [ ] Grid 3 колонки (как в прототипе)
- [ ] Icon mapping в EngagementCard

### Страницы-заглушки (T2214.3)
- [ ] Dashboard: моки по прототипу (3 stat-карточки, льготы, события, быстрые действия)
- [ ] Points: моки по прототипу (баланс, прогресс-бары, транзакции)
- [ ] Documents: моки по прототипу (таблица с бейджами)
- [ ] Support: моки по прототипу (FAQ + форма)
- [ ] StubBadge на каждом заглушечном блоке (только dev)

### Общее
- [ ] Все эмодзи заменены на Lucide иконки
- [ ] `npm run dev` → страница визуально соответствует прототипу
- [ ] Unit-тесты не ломаются (или обновлены)
- [ ] E2E тесты не ломаются (или обновлены)

## Модалки (не в этой задаче)

Реализуются через универсальные компоненты дизайн-системы `@ukituki-ps/april-ui`.
ТЗ для DS-агента: `ТЗ-Design-System.md` (в этой же директории).

Компоненты, которые нужны в DS:
- `AprilWizard` — универсальный wizard (progress bar, steps, back/next, footer actions)
- `AprilWizardProgress` — горизонтальный progress bar с номерами шагов, label'ами, линиями, состояниями
- `AprilOptionCard` — selectable option card с border highlight
- `AprilTabbedContent` — табы + контент (для Benefit Detail)
- `AprilPolicyCard` — градиентная карточка полиса

После апгрейда DS — LKFL подключит компоненты и реализует 3 модалки прототипа:
- DMS Wizard (4 шага: опция → оплата → подтверждение → готово)
- Benefit Detail (табы: условия / полис / клиники на карте)
- Mat Capital Wizard (4 шага)
