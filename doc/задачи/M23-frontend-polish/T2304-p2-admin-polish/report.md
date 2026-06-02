# T2304 — Отчёт

## Статус

✅ Выполнено

## Дата

2026-06-02

## Что сделано

### Documents (`Documents.tsx`)

1. **Heading** — заголовок `<Title order={1}>Документы</Title>` без иконки, подзаголовок «Заявления, согласия и сформированные документы»
2. **Download button** — кнопка «Скачать» по стилю прототипа: `variant="default"`, `padding: 6px 12px`, `background: var(--brand-row)`, `border: 1px solid var(--brand-border)`, `borderRadius: 6`, `fontSize: 12`, `fontWeight: 600`, `color: var(--brand-text-muted)`
3. **Table header** — `textTransform: uppercase`, `letterSpacing: 0.5px`, `fontSize: 11`, `fontWeight: 600`, `color: var(--brand-text-subtle)`, `background: var(--brand-row)`

### Support (`Support.tsx`)

4. **Layout** — правая колонка: `flex: 0 0 380px` (было `flex: 1 1 45%`)
5. **FAQ accordion** — заменён Mantine `Accordion` на кастомный `FaqItem`:
   - `border-bottom: 1px solid var(--brand-row)`
   - Question: `padding: 16px 20px`, `fontSize: 13`, `fontWeight: 600`, chevron (AprilIconChevronRight + rotate 90deg)
   - Open state: `color: var(--brand-green, #00B33C)`, chevron rotate
   - Answer: `padding: 0 20px 16px`, `fontSize: 13`, `color: var(--brand-text-muted)`, `lineHeight: 1.6`
   - Состояние через `useState` для каждого элемента

### Admin HR (`AdminHR.tsx`)

6. Таблица пользователей (mock): 5 записей, столбцы id/email/имя/фамилия/статус/роли
7. Поиск по email и имени
8. Модальное окно редактирования пользователя (имя, фамилия, роль)
9. Периоды начислений (mock): 3 периода с датами и статусами
10. Геймификация (stub): информационный блок

### Admin Catalog (`AdminCatalog.tsx`)

11. Таблица карточек льгот (mock): 5 записей, столбцы название/провайдер/категория/стоимость/статус
12. Метрики: карточки с количеством активных/всего карточек, конверсия (F2 placeholder)
13. CRUD: добавление карточки через модальную форму, редактирование, удаление
14. Форма: название, провайдер, категория (Select), описание, стоимость (NumberInput), статус

### Admin Content (`AdminContent.tsx`)

15. FAQ: список вопросов/ответов (mock, 5 записей), CRUD через модальные окна
16. Баннеры: список (mock, 3 записи) со статусами и позициями, кнопки редактирования/удаления
17. Описания карточек (stub): информационный блок

## Валидация

- `tsc --noEmit` — ✅ без ошибок
- `npm run build` — ✅ без ошибок

## Замечания

- Иконка `AprilIconChevronDown` отсутствует в `@ukituki-ps/april-ui` — использован `AprilIconChevronRight` с CSS transform `rotate(90deg)` / `rotate(-90deg)` для имитации chevron-down/up
- Иконка `AprilIconGrid` отсутствует в `@ukituki-ps/april-ui` — использован `AprilIconFileText` для Admin Catalog
- App.tsx уже содержал правильные импорты admin-страниц, изменений не потребовалось
