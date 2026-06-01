# T2301 — Report

## Выполнено

- [x] StubBadge удалён со всех страниц
- [x] Avatar — круглый (50%)
- [x] Transaction filters — pills
- [x] Transaction rows — иконки + «б»
- [x] Logout → redirect на /login
- [x] Support form — 2 поля, кнопка uppercase
- [x] FAQ — 6 вопросов

## Изменённые файлы

- `frontend/src/pages/Dashboard.tsx` — удалён импорт и 4 вхождения `<StubBadge />`
- `frontend/src/pages/Points.tsx` — удалён StubBadge, заменён SegmentedControl на button pills, обновлены строки транзакций (иконки 36×36 + суффикс «б»), удалены импорты `SegmentedControl` и `Paper`
- `frontend/src/pages/Documents.tsx` — удалён импорт и 1 вхождение `<StubBadge />`, добавлен `<p>` подзаголовок
- `frontend/src/pages/Support.tsx` — удалён импорт и 3 вхождения `<StubBadge />`, удалено поле «Заголовок», изменено «Описание» → «Сообщение», кнопка «ОТПРАВИТЬ» uppercase, обновлён mockFaq (6 вопросов), добавлен `<p>` подзаголовок
- `frontend/src/components/layout/UserMenu.tsx` — Avatar: убран `radius="xl"`, добавлен `borderRadius: '50%'` в style
- `frontend/src/stores/authStore.ts` — в `logout()` добавлен `window.location.href = '/login'`

## Не изменены

- `frontend/src/components/ui/StubBadge.tsx` — файл сохранён (используется в `CatalogDetail.tsx`)
- `frontend/src/components/auth/RequireAuth.tsx` — редирект на `/login` уже работает корректно

## Валидация

- tsc --noEmit: ✅
- npm run build: ✅
