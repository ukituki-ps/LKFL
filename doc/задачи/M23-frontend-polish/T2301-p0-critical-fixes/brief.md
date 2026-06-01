# T2301 — P0 Critical Fixes: StubBadge, Avatar, Transactions, Logout, Support, FAQ

## Веха

M23 — Frontend Polish

## Тип

code

## Проблема

6 критических расхождений между прототипом и фронтендом, блокирующих визуальную приемку.

### 1. StubBadge на каждой странице

`StubBadge` компонент отображается в заголовках всех страниц. Его нет в прототипе — это временный артефакт разработки.

**Файлы:** Dashboard, Catalog, Points, Documents, Support — все заголовки содержат `<StubBadge />`.

### 2. Avatar `radius="xl"` вместо `circle`

Прототип: `border-radius: 50%` (круг). Фронтенд: `radius="xl"` (скруглённый квадрат).

**Файл:** `UserMenu.tsx` строка 35.

### 3. Transaction filters — SegmentedControl вместо pills

Прототип: filter pills `border-radius: 20px`, `border: 1.5px solid`, active = black bg + white text.
Фронтенд: Mantine `SegmentedControl` — совершенно другой вид.

**Файл:** `Points.tsx` строка 132.

### 4. Transaction rows — нет иконок + нет суффикса «б»

Прототип: `tx-icon 36×36` (plus/minus), сумма `+500 б`, `−600 б`.
Фронтенд: нет иконок, сумма `+500`, `-300` без суффикса.

**Файл:** `Points.tsx` строка 145-177.

### 5. Logout не работает

`authStore.logout()` вызывает `POST /api/v1/auth/logout`, но после очистки localStorage страница не перенаправляет на login. Нет redirect logic.

**Файл:** `authStore.ts` строка 97-110, `App.tsx` / `RequireAuth.tsx`.

### 6. Support form — лишнее поле «Заголовок», кнопка не uppercase

Прототип: только «Тема» + «Сообщение», кнопка «ОТПРАВИТЬ» uppercase.
Фронтенд: «Тема» + «Заголовок» + «Описание», кнопка «Отправить обращение» titlecase.

**Файл:** `Support.tsx` строка 166-200.

### 7. FAQ контент не совпадает

Прототип: 6 вопросов с конкретными ответами. Фронтенд: 5 вопросов с другим текстом.

**Файл:** `Support.tsx` строка 23-49.

## Что делать

### 1. Убрать StubBadge

- Удалить `<StubBadge />` из Dashboard, Points, Documents, Support
- Если `StubBadge.tsx` больше не используется — удалить компонент
- Заменить на `<p>` подзаголовок где нужно (прототип имеет `<p>` под h1)

### 2. Avatar circle

- `UserMenu.tsx`: `radius="xl"` → `radius="xl"` → `circle`
- Проверить что Mantine Avatar поддерживает `radius="xl"` → `radius="xl"` это не circle, нужен `radius="xl"` → убрать radius prop, использовать CSS `border-radius: 50%`

### 3. Transaction filters → pills

- Заменить `SegmentedControl` на кастомные buttons:
  - `padding: 5px 12px`, `border-radius: 20px`
  - `border: 1.5px solid var(--brand-border)`
  - Active: `background: var(--brand-text)`, `color: #fff`, `border-color: var(--brand-text)`
  - Inactive: `background: transparent`, `color: var(--brand-text-muted)`

### 4. Transaction rows + иконки + «б»

- Добавить `tx-icon` (36×36, borderRadius 10px):
  - Credit: `background: #DCFCE7`, `color: #16A34A`, icon `plus-circle`
  - Debit: `background: var(--brand-bg)`, `color: var(--brand-text-subtle)`, icon по типу
- Сумма: `{t.type === 'credit' ? '+' : '−'}{Math.abs(t.amount)} б`
- Layout: `flex`, `align-items: center`, `gap: 12px`, `padding: 13px 18px`, `border-bottom: 1px solid var(--brand-row)`

### 5. Logout fix

- `RequireAuth.tsx`: после `logout()` → `navigate('/login')`
- `authStore.logout()`: добавить `window.location.href = '/login'` после clearAuth
- Или: `RequireAuth` проверяет `isAuthenticated` и делает redirect

### 6. Support form

- Удалить `TextInput label="Заголовок"`
- Кнопка: `children="ОТПРАВИТЬ"`, `textTransform: uppercase` в style
- `Textarea` placeholder: «Опишите ваш вопрос подробно...» (как в прототипе)

### 7. FAQ контент

- Заменить `mockFaq` на 6 вопросов из прототипа (строки 976-998)

## Требования

- Все 7 пунктов исправлены
- Визуальное соответствие прототипу на уровне P0
- `tsc --noEmit` без ошибок
- `npm run build` без ошибок

## Критерии приёмки

- [ ] StubBadge удалён со всех страниц
- [ ] Avatar — круглый (50%)
- [ ] Transaction filters — pills border-radius 20px, active = black bg
- [ ] Transaction rows — иконки 36×36, суффикс «б»
- [ ] Logout → redirect на /login
- [ ] Support form — 2 поля (Тема + Сообщение), кнопка uppercase «ОТПРАВИТЬ»
- [ ] FAQ — 6 вопросов из прототипа
- [ ] `tsc --noEmit` ✅
- [ ] `npm run build` ✅

## Зависимости

- **depends_on:** M21 (базовый фронтенд)
- **touches:** `Dashboard.tsx`, `Points.tsx`, `Documents.tsx`, `Support.tsx`, `UserMenu.tsx`, `authStore.ts`, `RequireAuth.tsx`
