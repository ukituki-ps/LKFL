# T2204.1 — Отчёт: Браузерные негативные / хаос-тесты (Playwright)

## Выполнено

Создано 100 хаос-тестов в 10 категориях, расположенных в `frontend/e2e/chaos/`.

## Структура файлов

| Файл | Описание | Кол-во тестов |
|------|----------|--------------|
| `frontend/e2e/chaos/helpers.ts` | Вспомогательные функции: детерминированный RNG (mulberry32), проверка краша, хаос-данные | — |
| `frontend/e2e/chaos/random-navigation.spec.ts` | Случайная навигация между страницами (20-40 переходов) | 10 |
| `frontend/e2e/chaos/random-button-click.spec.ts` | Случайные клики по всем кликабельным элементам | 10 |
| `frontend/e2e/chaos/rapid-filter.spec.ts` | Быстрая смена фильтров каталога (5-10 переключений) | 10 |
| `frontend/e2e/chaos/search-input.spec.ts` | Ввод хаос-строк (XSS, SQLi, emoji, unicode, 10K+ символов) | 10 |
| `frontend/e2e/chaos/rapid-pagination.spec.ts` | Быстрое переключение страниц пагинации | 10 |
| `frontend/e2e/chaos/double-click.spec.ts` | Двойные/тройные/быстрые клики на кнопках | 10 |
| `frontend/e2e/chaos/form-chaos.spec.ts` | Случайное заполнение и отправка форм | 10 |
| `frontend/e2e/chaos/network-chaos.spec.ts` | Отключение сети, throttling, случайные ошибки API | 10 |
| `frontend/e2e/chaos/viewport-chaos.spec.ts` | Переключение viewport mobile/tablet/desktop | 10 |
| `frontend/e2e/chaos/keyboard-chaos.spec.ts` | Случайные нажатия клавиш + комбинации | 10 |

**Итого: 100 тестов**

## Критерии приёмки

- ✅ Random navigation chaos (10 тестов)
- ✅ Random button click chaos (10 тестов)
- ✅ Rapid filter chaos (10 тестов)
- ✅ Search input chaos (10 тестов)
- ✅ Rapid pagination chaos (10 тестов)
- ✅ Double/triple click chaos (10 тестов)
- ✅ Form chaos (10 тестов)
- ✅ Network chaos (10 тестов)
- ✅ Viewport chaos (10 тестов)
- ✅ Keyboard chaos (10 тестов)
- ✅ Воспроизводимость через seed (mulberry32 PRNG)
- ✅ Video recording (`video: 'on'` в проекте chaos)
- ✅ Console errors capture (fail при unhandled errors)
- ✅ Screenshot on failure (наследуется из базовой конфигурации)
- ✅ Trace recording (`trace: 'on'` в проекте chaos)
- ✅ CI integration (проект `chaos` в playwright.config.ts, `testMatch: '**/chaos/**/*.spec.ts'`)

## Детали реализации

### helpers.ts
- **chaosSeed(seed)** — детерминированный PRNG на основе mulberry32
- **randomFrom(rng, arr)** — случайный элемент массива
- **randomInt(rng, min, max)** — случайное число в диапазоне
- **expectNoCrash(page)** — проверка: root не пустой, URL не about:blank
- **setupChaosTest(page, startUrl)** — настройка: API моки, сбор console/page errors, навигация
- **CHAOS_INPUTS** — 16 хаос-строк (XSS, SQLi, emoji, unicode, длинные строки, control chars)
- **VIEWPORT_SIZES** — 8 viewport размеров (iPhone 8 → 4K)
- **CHAOS_KEYS** — 16 клавиш для хаос-теста
- **CHAOS_KEY_COMBOS** — 7 комбинаций клавиш (Ctrl+Z/Y/A/C/V/X, Ctrl+Shift+Z)

### playwright.config.ts
- Добавлен проект `chaos` с `testMatch: '**/chaos/**/*.spec.ts'`
- Video recording: `on` (всегда)
- Trace recording: `on` (всегда)
- Обычные проекты исключают chaos-тесты через `testIgnore`

### Запуск

```bash
# Все хаос-тесты
npx playwright test --project=chaos

# Конкретная категория
npx playwright test --project=chaos chaos/random-navigation.spec.ts

# Headed mode для отладки
npx playwright test --project=chaos --headed
```

## Замечания

- Тесты требуют запущенный Vite dev server (как и обычные E2E тесты)
- Network chaos тесты используют `page.route()` для симуляции отключения/ошибок сети
- Все тесты используют моки API (backend не требуется)
- Тесты не зависят от конкретного контента страниц — проверяют только отсутствие крашей
