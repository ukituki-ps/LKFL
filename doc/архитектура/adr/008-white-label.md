# ADR-008: White-label через CSS переменные (brand override)

**Статус:** Accepted
**Дата:** 2026-05-22
**Контекст:** М01-создание-описания

## Контекст

Прототип — бренд СДЭК (зелёная тема `#00B33C`). April tokens — своя палитра (teal). Платформа multi-tenant: каждая компания — свой бренд.

## Решение

**Brand override file** — отдельный CSS файл, подключаемый **после** `@april/tokens/css`:

```css
/* src/theme/brand-sdek.css */
:root {
  --april-accent-teal-6: #00B33C;
  --april-accent-teal-8: #009A33;
  --april-accent-teal-1: #F0FDF4;
  --april-accent-teal-3: #BBF7D0;
  --april-semantic-success: #00B33C;
  --april-neutral-bg-alt: #F2F2F2;
}
```

Mantine theme: override `primaryColor: 'teal'` + кастомная палитра `green`.

**Для multi-tenant:** `brand-russdragmet.css`, `brand-company-x.css` — переключается через query param или subdomain.

## Альтернативы

| Вариант | Плюсы | Минусы |
|---------|-------|--------|
| Новый preset в DS | Чистый API | Требуется bump DS, merge в upstream |
| Inline styles | Быстрый | Хаки, сложно поддерживать |
| CSS-in-JS override | Динамический | Мantine не поддерживает runtime theme swap без re-render |

## Вердикт

**CSS override file.** Zero dependency на DS bump, работает для любого бренда, легко тестировать.

## Следствия

- `src/theme/brand-*.css` — по одному файлу на бренд
- Tenant resolver в middleware → brand CSS injection
- Нет изменений в `@april/ui` и `@april/tokens`
