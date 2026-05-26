# ADR-034: i18n — YAGNI check

**Статус:** Accepted
**Дата:** 2026-05-26
**Контекст:** M15-архитектура-фронтенда, T1506

---

## Контекст

Первый tenant — СДЭК (русская компания). 152-ФЗ. В спецификации нет требования на мультиязычность.

| Фактор | Оценка |
|--------|--------|
| Первый tenant | СДЭК — русский язык |
| 152-ФЗ | Русский язык интерфейса обязателен |
| White-label | Следующие tenant'ы могут быть не русскоязычными |
| Стоимость i18next | +10KB bundle, сложность во все компоненты |

---

## Рассмотренные варианты

### Вариант А: i18next (полноценная i18n)

| Плюсы | Минусы |
|-------|--------|
| Готов к multi-language | +10KB bundle, learning curve |
| ICU messages, pluralization | Все строки → `t('key')` — ~25-30 компонентов затронуты |
| Industry standard | **YAGNI:** нет реального use case сейчас |

### Вариант Б: `lib/translations/ru.ts` (объект строк)

| Плюсы | Минусы |
|-------|--------|
| 0KB overhead | Переход на i18next — refactoring (~3-5 часов) |
| Простой: `translations[key]` | Нет pluralization, ICU messages |
| Подготовка к future i18n | — |

### Вариант В: Хардкод строк в компонентах

| Плюсы | Минусы |
|-------|--------|
| 0 сложности | Невозможно добавить i18n later без refactoring всех компонентов |
| — | Не поддерживаемость для multi-tenant |

---

## Решение

**Вариант Б: `lib/translations/ru.ts`** — объект строк. Не добавлять i18next сейчас.

**Обоснование:**
1. **YAGNI:** нет реального use case для мультиязычности. СДЭК — русскоязычный.
2. **Подготовка:** строки вынесены в отдельный файл — future i18next transition тривиален.
3. **0KB overhead:** нет дополнительной dependency.

**Структура:**
```ts
// lib/translations/ru.ts
export const translations = {
  dashboard: {
    greeting: 'Привет, {name}',
    balanceLabel: 'Баланс баллов',
    activeBenefits: 'Активных льгот',
    periodEnd: 'До конца периода',
  },
  catalog: {
    title: 'Каталог льгот',
    search: 'Поиск льготы...',
    filters: { all: 'Все', dms: 'ДМС', fitness: 'Спорт', ... },
  },
  points: {
    title: 'Мои баллы',
    subtitle: 'История начислений и списаний',
    balance: 'Доступный баланс',
    categories: { health: 'Здоровье и спорт', education: 'Обучение', other: 'Мерч и прочее' },
  },
  // ... ~200-300 ключей на 8 страниц + 3 модалки
} as const;
```

**Использование:**
```tsx
import { translations } from '../lib/translations/ru'

<h1>{translations.dashboard.greeting.replace('{name}', userName)}</h1>
```

---

## Оценка стоимости добавления i18n later

| Этап | Стоимость |
|------|-----------|
| Текущий подход: `lib/translations/ru.ts` — объект `{ key: string }` | ~200-300 строк на 8 страниц + 3 модалки |
| Переход на i18next: `translations[key]` → `useTranslation().t(key)` | ~3-5 часов refactoring |
| Затронутые компоненты | ~25-30 компонентов (все, где есть строки) |
| Структурные изменения | Нет — только замена способа доступа к строкам |
| Риск | Минимальный: `lib/translations/ru.ts` как источник ключей → i18next resources |

---

## Следствия

- `lib/translations/ru.ts` — объект строк (источник ключей)
- При добавлении нового tenant с другим языком → переход на i18next (3-5 часов)
- Не хардкодить строки в компоненты — всегда через `translations`
- Ключи структурированы по страницам: `dashboard.*`, `catalog.*`, `points.*` и т.д.
