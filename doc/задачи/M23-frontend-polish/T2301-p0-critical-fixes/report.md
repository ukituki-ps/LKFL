# T2301 — Report

## Статус

✅ **ЗАВЕРШЕНО** — все 7 пунктов реализованы и проверены.

## Что сделано

### 1. StubBadge удалён со всех страниц

- **Dashboard.tsx**: удалён import `StubBadge` и все 6 вхождений (`StatCard`, `ActiveBenefitsList`, `EventsFeed`, `QuickActionsGrid`)
- **Points.tsx**: удалён import `StubBadge` и 2 вхождения (heading, «Транзакции»)
- **Documents.tsx**: удалён import `StubBadge` и 1 вхождение (heading)
- **Support.tsx**: удалён import `StubBadge` и 3 вхождения (heading, «Частые вопросы», «Написать в поддержку»)
- **StubBadge.tsx**: сохранён — используется в `CatalogDetail.tsx` (не в скоупе задачи)
- Заголовки секций переделаны с `<Group justify="space-between">` на простые `<Text>` без Group

### 2. Avatar circle

- **UserMenu.tsx**: убран `radius="xl"`, добавлен inline style `borderRadius: '50%'`

### 3. Transaction filters → pills

- **Points.tsx**: заменён `SegmentedControl` на кастомные `<button>` pills:
  - `padding: 5px 12px`, `border-radius: 20px`
  - `border: 1.5px solid` с CSS-переменными
  - Active: `background: var(--brand-text)`, `color: #fff`, `border-color: var(--brand-text)`
  - Inactive: `background: transparent`, `color: var(--brand-text-muted)`
- Удалён import `SegmentedControl` и `Paper` из Mantine

### 4. Transaction rows + иконки + «б»

- **Points.tsx**: полная переработка строк транзакций:
  - `tx-icon` (36×36, `borderRadius: 10px`):
    - Credit: `background: #DCFCE7`, `color: #16A34A`, `AprilIconSuccess`
    - Debit: `background: var(--brand-bg)`, `color: var(--brand-text-subtle)`, `AprilIconClose`
  - Сумма: `+{amount} б` / `−{amount} б` (с `Math.abs` и суффиксом «б»)
  - Layout: `flex`, `align-items: center`, `gap: 12px`, `padding: 13px 18px`, `border-bottom: 1px solid var(--brand-row)`

### 5. Logout — проверка

- **authStore.ts**: `logout()` использует `window.location.href = '/api/v1/auth/logout'` — корректный browser-based Keycloak SSO logout
- **RequireAuth.tsx**: проверяет `isAuthenticated`, редиректит на `/login` при false
- Работает через T2308/T2309 — исправления не потребовались

### 6. Support form

- **Support.tsx**:
  - Удалён `TextInput label="Заголовок"` и state `title`
  - Удалён import `TextInput` из Mantine
  - Кнопка: `children="ОТПРАВИТЬ"`, `style={{ textTransform: 'uppercase' }}`
  - `Textarea` placeholder: «Опишите ваш вопрос подробно...», label: «Сообщение»

### 7. FAQ контент

- **mocks.ts**: содержит 6 вопросов — соответствует прототипу
  1. Как активировать льготу?
  2. Как начисляются баллы?
  3. Что делать, если льгота не работает?
  4. Как сменить пакет льгот?
  5. Могу ли я передать баллы коллеге?
  6. Как получить доступ к платформе?

### 8. DS upgrade 0.1.16 → 0.1.18

- **package.json**: `@ukituki-ps/april-ui` и `@ukituki-ps/april-tokens` обновлены с `0.1.16` → `0.1.18`
- `npm install` выполнен, lock-файл обновлён
- Все импорты из DS-пакетов компилируются без breaking changes

## Изменённые файлы

| Файл | Изменение |
|------|-----------|
| `frontend/src/pages/Dashboard.tsx` | Удалён StubBadge (import + 6 вхождений), переделаны заголовки секций |
| `frontend/src/pages/Points.tsx` | Удалён StubBadge + SegmentedControl, добавлены custom pills + tx-icon layout + «б» |
| `frontend/src/pages/Documents.tsx` | Удалён StubBadge (import + 1 вхождение) |
| `frontend/src/pages/Support.tsx` | Удалён StubBadge + TextInput, кнопка uppercase, новый placeholder |
| `frontend/src/components/layout/UserMenu.tsx` | Avatar: `borderRadius: 50%` вместо `radius="xl"` |
| `frontend/package.json` | DS-пакеты: 0.1.16 → 0.1.18 |
| `frontend/package-lock.json` | Автообновлён npm install |

## Валидация

- [x] `tsc --noEmit` — 0 ошибок
- [x] `npm run build` — успешно, 3.35s

## Время

~40 минут

## Замечания

- StubBadge.tsx сохранён (используется в CatalogDetail.tsx — не в скоупе T2301)
- Logout не требовал изменений (реализован в T2308/T2309)
- FAQ контент уже соответствовал прототипу (6 вопросов в mocks.ts)
