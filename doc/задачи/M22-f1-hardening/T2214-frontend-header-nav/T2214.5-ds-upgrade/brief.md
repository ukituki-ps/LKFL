# T2214.5 — Апгрейд дизайн-системы v0.1.13 → v0.1.16

## Родительская задача

T2214 — Frontend: прототип → код (header, дизайн, страницы-заглушки)

## Веха

M22-f1-hardening

## Тип

code

## Кратко

Выделенная задача для апгрейда пакетов дизайн-системы `@ukituki-ps/april-ui` и `@ukituki-ps/april-tokens` из v0.1.13 в v0.1.16. Проверка обратной совместимости, обновление импортов, верификация сборки.

**Выполняется первой** — все остальные подзадачи T2214.* зависят от новой версии DS.

---

## Что делает v0.1.16

### Новые компоненты (17 из ТЗ-DS)

| Компонент | Назначение | Где используется |
|-----------|-----------|-----------------|
| `AprilProductHeader` | Product header (left/center/right/sticky) | T2214.2 |
| `AprilFilterPills` | Pill-фильтры каталога | T2214.4 |
| `AprilWizard` + `AprilWizardProgress` | Универсальный wizard | TODO F3 |
| `AprilStatCard` | Стат-карточка (label/value/hint) | T2214.3 (Dashboard) |
| `AprilCard` | Карточка с header + scrollable body | TODO |
| `AprilBenefitRow` | Строка льготы (icon + name + meta + badge) | T2214.3 |
| `AprilEventRow` | Строка события (icon + text + time) | T2214.3 |
| `AprilQuickButton` | Кнопка быстрого действия (icon + text) | T2214.3 |
| `AprilBalanceCard` | Карточка баланса с progress bars | T2214.3 (Points) |
| `AprilTransactionRow` | Строка транзакции (icon + type + name + amount) | T2214.3 |
| `AprilFaqItem` | FAQ accordion item | T2214.3 (Support) |
| `AprilOptionCard` | Selectable option card | TODO F3 |
| `AprilPayOptionCard` | Payment option card | TODO F3 |
| `AprilFormInput` / `AprilFormTextarea` / `AprilFormSelect` | Form controls | TODO F3 |
| `AprilConfirmCheckbox` | Confirmation checkbox | TODO F3 |
| `AprilSuccessScreen` | Success screen для wizard | TODO F3 |
| `AprilConfirmDoc` | Document preview для confirmation | TODO F3 |

### Новые иконки (20 штук)

`ArrowUpCircle`, `Baby`, `Brain`, `CheckCircle`, `Circle`, `Clock`, `Coffee`, `Coins`, `Download`, `Dumbbell`, `Gift`, `GraduationCap`, `Heart`, `Languages`, `MapPin`, `PlusCircle`, `ShoppingBag`, `Smartphone`, `Sparkles`, `UserPlus`.

**Итого:** 34 → 54 иконки. Покрытие прототипа: 35/38 (3 иконки требуют fallback, см. ниже).

### Breaking changes

**Нет.** Чистое добавление. Все экспорты v0.1.13 сохранены.

---

## Что сделать

### 1. Обновить зависимости в `frontend/package.json`

```diff
-    "@ukituki-ps/april-ui": "0.1.13",
-    "@ukituki-ps/april-tokens": "0.1.13",
+    "@ukituki-ps/april-ui": "0.1.16",
+    "@ukituki-ps/april-tokens": "0.1.16",
```

```bash
cd frontend && npm install
```

### 2. Проверить обратную совместимость импортов

Проверить все файлы, импортирующие из `@ukituki-ps/april-ui` и `@ukituki-ps/april-tokens`:

```bash
# Проверить какие компоненты используются
rg "from '@ukituki-ps/april" frontend/src/ --json
```

Убедиться, что все используемые компоненты v0.1.13:
- Экспортируются в v0.1.16 (проверить `index.ts` пакета)
- Сохраняют те же пропсы (проверить TypeScript-типы)
- Не требуют изменения CSS/стилей

### 3. Обновить `node_modules` и зачистить кеш

```bash
cd frontend
rm -rf node_modules/.vite
npm install
```

### 4. Верифицировать сборку

```bash
cd frontend
npm run build    # должна завершиться без ошибок
npm run lint     # должно пройти без новых варнингов
npm run test     # unit-тесты должны проходить
```

### 5. Верифицировать dev-режим

```bash
cd frontend
npm run dev
```

- Открыть `http://localhost:5173`
- Проверить, что существующие компоненты рендерятся корректно
- Проверить консоль браузера — нет runtime ошибок

### 6. Обновить lock-файл

```bash
cd frontend && npm install --package-lock-only
```

Проверить `package-lock.json` — версии пакетов обновлены до 0.1.16.

---

## Icon fallback mapping (для справки)

4 иконки прототипа отсутствуют в DS v0.1.16. Использовать nearest-аналоги:

| Прототип | DS v0.1.16 (использовать) | Примечание |
|----------|---------------------------|------------|
| `heart-pulse` | `AprilIconHeart` | достаточный визуальный аналог |
| `shield-check` | `AprilIconSuccess` (=CheckCircle2) | semantic match |
| `shield-plus` | `AprilIconPlusCircle` | semantic match |
| `smart-speaker` | `AprilIconSmartphone` | nearest функциональный аналог |

> **Прямой импорт `lucide-react` — НЕ разрешён.** Архитектура (фронтенд.md §Иконки, NAVIGATION.md правило #8): только `AprilIcon*` из `@ukituki-ps/april-ui`.

---

## Файлы

### Изменяются
- `frontend/package.json` — версии `@ukituki-ps/april-ui` и `@ukituki-ps/april-tokens`
- `frontend/package-lock.json` — автогенерация при `npm install`

### Не изменяются
- `src/` — код приложения не меняется (нет breaking changes в v0.1.16)
- `node_modules/` — обновляется автоматически

## Зависимости

- **Нет зависимостей от других задач** — выполняется первой
- **T2214.1–T2214.4 зависят от этой задачи** — используют новые компоненты v0.1.16

## Критерии приёмки

- [ ] `@ukituki-ps/april-ui@0.1.16` установлен в `frontend/package.json`
- [ ] `@ukituki-ps/april-tokens@0.1.16` установлен в `frontend/package.json`
- [ ] `package-lock.json` обновлён
- [ ] `npm run build` завершается без ошибок
- [ ] `npm run lint` проходит без новых варнингов
- [ ] `npm run test` — все unit-тесты проходят
- [ ] `npm run dev` — приложение запускается без runtime ошибок
- [ ] Консоль браузера — нет ошибок при рендеринге
- [ ] Существующие компоненты из `@ukituki-ps/april-ui` рендерятся корректно
- [ ] Новые компоненты (`AprilProductHeader`, `AprilFilterPills`, иконки `AprilIcon*`) импортируются без ошибок (проверка compile-time)
- [ ] `lucide-react` НЕ добавлен в `dependencies` package.json
