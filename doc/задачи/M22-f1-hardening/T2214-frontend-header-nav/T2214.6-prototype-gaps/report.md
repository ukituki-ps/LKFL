# T2214.6 — Отчёт о выполнении

## Статус: ✅ выполнено

**Дата:** 29 мая 2026

## Выполненные гэпы

| # | ГЭп | Описание | Файл |
|---|-----|----------|------|
| 1 | 🔴 Критичный | Страница детализации `/catalog/:slug` с моками BENEFIT_DATA, табы DMS | `CatalogDetail.tsx` (new), `App.tsx`, `routes/employee.tsx` |
| 2 | 🔴 Критичный | Быстрые действия: 5 кнопок из прототипа, grid 2 колонки, toast | `Dashboard.tsx` |
| 3 | 🔴 Критичный | Dashboard layout: 2 колонки (льготы+события \| быстрые действия 292px) | `Dashboard.tsx` |
| 4 | 🔴 Критичный | Points layout: 2 колонки (баланс+категории 320px \| транзакции) | `Points.tsx` |
| 5 | ⚠️ Умеренный | Balance pill: кастомная зелёная пилюля `#F0FDF4` + «N баллов» | `HeaderRight.tsx` |
| 6 | ⚠️ Умеренный | Stat card «Баланс баллов»: зелёный фон, белый текст | `Dashboard.tsx` |
| 7 | ⚠️ Умеренный | События: иконки 30×30, временные метки, bold текст | `Dashboard.tsx` |
| 8 | ⚠️ Умеренный | Карточки каталога: hover `translateY(-2px)` + shadow | `EngagementCard.tsx` |
| 9 | ⚠️ Умеренный | Документы: secondary строка `docMeta` | `Documents.tsx` |
| 10 | ⚠️ Умеренный | Документы: кнопка «Скачать» Button с текстом | `Documents.tsx` |
| 11 | ⚠️ Умеренный | Support: success state после сабмита формы | `Support.tsx` |
| 12 | ⚠️ Умеренный | Stat card 3: число + «дн» мелким шрифтом | `Dashboard.tsx` |

## Изменённые файлы

| Файл | Изменение | Гэпы |
|------|-----------|------|
| `src/pages/CatalogDetail.tsx` | **Создан** | ГЭП-1 |
| `src/pages/Dashboard.tsx` | Переработан layout + 5 гэпов | ГЭП-2,3,6,7,12 |
| `src/pages/Points.tsx` | Переработан layout | ГЭП-4 |
| `src/pages/Documents.tsx` | Обновлены моки + кнопка | ГЭП-9,10 |
| `src/pages/Support.tsx` | Добавлен success state | ГЭП-11 |
| `src/components/layout/HeaderRight.tsx` | Кастомная balance pill | ГЭП-5 |
| `src/components/catalog/EngagementCard.tsx` | Hover эффект | ГЭП-8 |
| `src/App.tsx` | Добавлен маршрут `/catalog/:slug` | ГЭП-1 |
| `src/routes/employee.tsx` | Добавлен entry для detail route | ГЭП-1 |

## Критерии приёмки

- [x] **ГЭП-1:** `/catalog/:slug` рендерится, не падает в 404. Показывает название, провайдер, описание, стоимость, кнопку. DMS с табами.
- [x] **ГЭП-2:** Быстрые действия — 5 кнопок из прототипа, grid 2 колонки, при клике — toast.
- [x] **ГЭП-3:** Dashboard — 2 колонки: слева льготы + события, справа быстрые действия (292px).
- [x] **ГЭП-4:** Points — 2 колонки: слева баланс + категории (внутри зелёной карточки), справа транзакции.
- [x] **ГЭП-5:** Balance pill — зелёный фон, border, текст «N баллов».
- [x] **ГЭП-6:** Stat card «Баланс баллов» — зелёный фон, белый текст.
- [x] **ГЭП-7:** События — иконки в цветных квадратах, временные метки, bold.
- [x] **ГЭП-8:** Карточки каталога — hover с translateY(-2px) + shadow.
- [x] **ГЭП-9:** Документы — secondary строка под названием.
- [x] **ГЭП-10:** Документы — кнопка «Скачать» с текстом.
- [x] **ГЭП-11:** Support — success state после сабмита.
- [x] **ГЭП-12:** Stat card 3 — число + «дн» мелким шрифтом.
- [x] TypeScript — 0 ошибок (`tsc --noEmit` clean)
- [x] Unit-тесты — 113 passed, 0 failed

## Замечания

- Toast реализован через inline DOM-элемент (без `@mantine/notifications`) — достаточно для stub'а, будет заменён на полноценную систему нотификаций в F2.
- StubBadge остаётся видимым только в dev-режиме (`!import.meta.env.DEV`).
- `BENEFIT_DATA` моки соответствуют прототипу `Прототип ЛК физика(1).html`.
