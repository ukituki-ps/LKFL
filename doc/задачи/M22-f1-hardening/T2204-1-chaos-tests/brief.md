# T2204.1 — Браузерные негативные / хаос-тесты (Playwright)

## Веха

M22-f1-hardening

## Тип

code

## Контекст

100 автоматизированных хаос-тестов, запускающихся в реальном браузере (Playwright),
имитирующих поведение пользователя, который «тыкает на всё подряд».
Цель — найти UI баги, race conditions, краши, неправильные состояния,
которые не ловят обычные тесты.

## Что сделать

### Сценарии хаос-тестов

Каждый тест запускает реальный браузер (Chromium), авторизуется, и выполняет **случайную последовательность действий**:

1. **Random navigation** — переключение между всеми доступными страницами в случайном порядке (10–50 переходов за сессию)
2. **Random button clicks** — клик по всем кликабельным элементам на странице в случайном порядке
3. **Random filter changes** — быстрая смена фильтров каталога (5–10 переключений подряд)
4. **Random search input** — ввод случайных строк (unicode, emoji, SQL injection, XSS, 1000+ символов)
5. **Rapid pagination** — переключение страниц 1→10→1→5→3 быстро (race condition)
6. **Double-click / triple-click** — двойные и тройные клики на кнопках
7. **Form chaos** — заполнение полей случайными данными, отправка пустых форм
8. **Concurrent navigation** — переход на другую страницу во время загрузки текущей
9. **Tab close/reopen** — эмуляция закрытия и reopening вкладки
10. **Network interruption** — отключение сети на 5 сек, затем восстановление
11. **Slow network simulation** — throttling до 3G
12. **Viewport changes** — переключение mobile/tablet/desktop в процессе сессии
13. **Scroll chaos** — рандомный скролл во время рендера
14. **Keyboard chaos** — случайные нажатия (Tab, Enter, Escape, Backspace, Ctrl+Z, Ctrl+A)
15. **Context menu** — правый клик, копирование через контекстное меню

### Структура тестов (100 тестов)

| Категория | Кол-во | Описание |
|-----------|-------|----------|
| Random navigation chaos | 10 | Случайные переходы, проверка что ни один не крашит |
| Random button click chaos | 10 | Клик по всем элементам, проверка на JS errors |
| Rapid filter chaos | 10 | Быстрая смена фильтров, UI не ломается |
| Search input chaos | 10 | Случайные строки (unicode, emoji, XSS, SQL injection, 1000+ chars) |
| Rapid pagination chaos | 10 | Быстрое переключение страниц, race conditions |
| Double/triple click chaos | 10 | Двойные/тройные клики на кнопках и ссылках |
| Form chaos | 10 | Случайное заполнение и отправка форм (admin CRUD) |
| Network chaos | 10 | Отключение/восстановление сети во время операций |
| Viewport chaos | 10 | Переключение viewport в процессе сессии |
| Keyboard chaos | 10 | Случайные нажатия клавиш |

### Требования

- **Playwright** — реальный браузер (headed mode для CI с видео)
- **Воспроизводимость** — фиксированный seed, `--seed=123`
- **Video recording** — каждый тест записывает видео
- **Console errors capture** — fail при unhandled errors
- **Screenshot on failure** — автоматический скриншот
- **Trace recording** — Playwright trace для провалившихся тестов
- **CI integration** — все 100 тестов в CI, failure любого = fail CI
- **Минимум 100 тестов, каждый с уникальным seed и сценарием**

## Критерии приёмки

- [ ] Random navigation chaos (10 тестов)
- [ ] Random button click chaos (10 тестов)
- [ ] Rapid filter chaos (10 тестов)
- [ ] Search input chaos (10 тестов)
- [ ] Rapid pagination chaos (10 тестов)
- [ ] Double/triple click chaos (10 тестов)
- [ ] Form chaos (10 тестов)
- [ ] Network chaos (10 тестов)
- [ ] Viewport chaos (10 тестов)
- [ ] Keyboard chaos (10 тестов)
- [ ] **Все 100 тестов воспроизводимы (seed)**
- [ ] **Video recording для каждого теста**
- [ ] **Console errors capture (fail on unhandled)**
- [ ] **Screenshot + trace on failure**
- [ ] **CI integration (100 хаос-тестов в pipeline)**
- [ ] **0 unhandled console errors во время хаос-тестов**
