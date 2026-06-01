# T2304 — P2 Fixes + Admin Pages Polish

## Веха

M23 — Frontend Polish

## Тип

code

## Проблема

P2 расхождения в Documents, Support + admin-страницы не доделаны.

### Documents

- **Heading:** «Мои документы» + иконка вместо простого «Документы»
- **Download button:** `Button variant="subtle"` вместо кастомного `btn-download` из прототипа (`padding: 6px 12px`, `background: var(--row)`, `border: 1px solid var(--border)`, hover green)
- **Table header:** не uppercase, не `letter-spacing: 0.5px` как в прототипе

### Support

- **Layout:** `flex: 1 1 55% / 45%` вместо `1fr 380px` (правая колонка фиксированная 380px в прототипе)
- **FAQ accordion:** Mantine `Accordion` вместо кастомного с chevron-down rotate
- **Success state:** layout отличается от прототипа (иконка в круге, padding, текст)

### Admin страницы

- **Admin HR (`/admin/hr`):** заглушка, нет контента
- **Admin Catalog (`/admin/catalog`):** заглушка, нет CRUD карточек
- **Admin Content (`/admin/content`):** заглушка, нет FAQ/баннеров

## Что делать

### 1. Documents heading

- Убрать иконку + StubBadge
- Заголовок: `<h1>Документы</h1>` + `<p>Заявления, согласия и сформированные документы</p>`

### 2. Documents download button

```tsx
<Button
  variant="default"
  size="xs"
  leftSection={<AprilIconDownload size={13} />}
  style={{
    padding: '6px 12px',
    background: 'var(--brand-row)',
    border: '1px solid var(--brand-border)',
    borderRadius: 6,
    fontSize: 12,
    fontWeight: 600,
    color: 'var(--brand-text-muted)',
  }}
>
  Скачать
</Button>
```

### 3. Documents table header

- Uppercase, `letter-spacing: 0.5px`, `font-size: 11px`, `font-weight: 600`, `color: var(--brand-text-subtle)`
- Background: `var(--brand-row)`

### 4. Support layout

- Правая колонка: `flex: 0 0 380px`

### 5. Support FAQ — кастомный accordion

- Каждый item: `border-bottom: 1px solid var(--brand-row)`
- Question: `padding: 16px 20px`, `font-size: 13px`, `font-weight: 600`, chevron-down справа
- Open state: `color: var(--green)`, chevron rotate 180°
- Answer: `padding: 0 20px 16px`, `font-size: 13px`, `color: var(--text-muted)`, `line-height: 1.6`

### 6. Admin страницы — базовый контент

**Admin HR:**
- Таблица пользователей (mock)
- Периоды начислений (mock)
- Геймификация (stub)

**Admin Catalog:**
- CRUD карточек (mock): список + форма добавления/редактирования
- Метрики (stub)

**Admin Content:**
- FAQ: список вопросов/ответов (mock)
- Баннеры: список (stub)
- Описания карточек (stub)

## Требования

- Documents, Support визуально соответствуют прототипу на уровне P2
- Admin страницы содержат базовый контент (не пустые заглушки)
- `tsc --noEmit` без ошибок

## Критерии приёмки

- [ ] Documents heading: «Документы» + подзаголовок
- [ ] Documents download button как в прототипе
- [ ] Documents table header uppercase + letter-spacing
- [ ] Support layout: 1fr 380px
- [ ] Support FAQ: кастомный accordion с chevron rotate
- [ ] Support success state: как в прототипе
- [ ] Admin HR: таблица пользователей + периоды
- [ ] Admin Catalog: CRUD карточек
- [ ] Admin Content: FAQ + баннеры stub
- [ ] `tsc --noEmit` ✅
- [ ] `npm run build` ✅

## Зависимости

- **depends_on:** T2301 (StubBadge removal)
- **touches:** `Documents.tsx`, `Support.tsx`, `AdminHR.tsx`, `AdminCatalog.tsx`, `AdminContent.tsx`
