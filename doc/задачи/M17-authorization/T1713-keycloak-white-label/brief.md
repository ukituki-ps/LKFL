# T1713 — Keycloak white-label: лого + CSS variables + dark mode + brand sync

## Веха

M17-authorization

## Тип

code

## Проблема

Тема Keycloak `lkfl` (T1712) работает, но не готова к production:
- Placeholder лого — teal rect + текст "LKFL", не соответствует бренду
- CSS переменные `--lkfl-*` не синхронизированы с фронтендом (`--brand-*`)
- Нет dark mode — страница логина светлая при тёмной системе
- Нет white-label: при добавлении tenant'а нужно копировать всю тему
- `tenant_brand_config` в БД и Keycloak тема — два независимых источника правды
- Фронтенд Shell.tsx — текстовый "ЛКФЛ" вместо лого

**Следствия:**
- Нельзя добавить нового tenant'а без изменения Keycloak темы вручную
- Нет brand consistency между Keycloak и фронтендом (разные CSS переменные)
- Пользователи с dark mode видят светлую страницу логина
- Лого в Keycloak и фронтенде — разные (placeholder vs текстовый)

## Что делать

### Фаза A: Синхронизация CSS переменных (Keycloak ↔ фронтенд)

**Цель:** единое пространство переменных `--brand-*` между Keycloak и фронтендом.

1. **`login.css` — переименование токенов**
   - Все `--lkfl-*` → `var(--brand-*, <default>)`
   - Синхронизация имён с фронтендом: `--brand-primary`, `--brand-bg`, `--brand-text` и т.д.
   - Fallback значения — April teal (совпадает с `@ukituki-ps/april-tokens`)

2. **`brand-default.css` — April teal значения**
   - Отдельный CSS файл только с переопределением переменных
   - Загружается ПЕРЕД `login.css` в `theme.properties`
   - 18 токенов: primary (3 шт), bg (2 шт), border (2 шт), text (3 шт), radius (2 шт), shadow (1 шт), semantic (3 шт)

3. **`brand-sdek.css` — СДЭК зелёная**
   - Цвета из прототипа: `--brand-primary: #00B33C`, `--brand-bg-alt: #F2F2F2`, `--brand-text: #1A1A1A`
   - Для переключения: изменить `theme.properties` → `styles=css/brand-sdek.css, css/login.css`

4. **`switch-tenant-theme.sh` — скрипт переключения**
   - Принимает имя tenant'а: `switch-tenant-theme.sh sdek`
   - Валидация: CSS файл существует
   - Обновление `theme.properties` + перезапуск Keycloak

### Фаза B: Лого

**Цель:** извлечь лого из прототипа, использовать в Keycloak и фронтенде.

5. **Лого Keycloak**
   - Извлечь `<path d="...">` из `Прототип ЛК физика(1).html` (строка 533-534)
   - `fill="#1a1a1a"` → `fill="currentColor"` (наследуется из CSS)
   - `infra/keycloak/theme/lkfl/login/resources/lkfl-theme/logo.svg`

6. **Favicon Keycloak**
   - Тот же path, resized для favicon
   - `theme.properties`: `favicon=/resources/lkfl-theme/favicon.svg`

7. **Лого фронтенд — компонент `Logo.tsx`**
   - `frontend/src/components/common/Logo.tsx` — переиспользуемый SVG-компонент
   - `fill="currentColor"` — цвет из CSS родителя
   - `size`: sm (14px height, header), md (48px), lg (64px)

8. **Shell.tsx — заменить текст "ЛКФЛ" на `<Logo />`**
   - Строки 43-52: `<Text>ЛКФЛ</Text>` → `<Logo size="sm" />`
   - Стиль: `color: var(--brand-text)`, `height: 14px` (как в прототипе)

9. **`brand-default.css` для фронтенда**
   - `frontend/src/styles/brand-default.css` — единый CSS файл с переменными
   - Подключить в `main.tsx`: `import '@/styles/brand-default.css'`
   - Включает backward-compat: `--brand-green`, `--brand-green-light`, `--brand-green-border`

### Фаза C: Dark mode + переключатель

**Цель:** тёмная тема в Keycloak, автоматическая (system preference) + ручное переключение.

10. **`@media (prefers-color-scheme: dark)` — automatic baseline**
    - Добавить секцию в конец `login.css`
    - Dark токены: `--brand-bg: #1a1b1e`, `--brand-text: #c1c2c5`, `--brand-border: #373a40`
    - Конtrast ratio: teal на dark = 6.7:1 ✅ (WCAG AA)

11. **Переключатель dark/light — cookie-based toggle**
    - Mustache template override: `header-content.ftl` → inject `<script>` + `<style>`
    - `theme-toggle.css` — стили кнопки (правый верхний угол, фиксированный)
    - `theme-toggle.js` — логика: cookie read/write, toggle, system listener
    - Priority: cookie > system preference
    - Кнопка: ☀️ / 🌙, `z-index: 9999`

12. **Account console dark mode**
    - `@import login.css` + `@media` секция
    - Переключатель не нужен (временная страница)

### Фаза D: Brand sync (DB → Keycloak)

**Цель:** `tenant_brand_config` в БД → генерация CSS файлов для Keycloak.

13. **Seed `css_variables` — полная карта токенов**
    - Обновить seed: `cmd/seed/main.go` + `cmd/server/main.go`
    - `css_variables` JSONB → 18 токенов (синхронизировано с `--brand-*`)
    - Пример: `{"primary": "#00B33C", "primary_hover": "#009A33", ...}`

14. **`cmd/brand-sync/main.go` — on-demand генерация CSS**
    - Подключиться к PostgreSQL → `SELECT * FROM tenant_brand_config`
    - Для каждой записи → сгенерировать `brand-{slug}.css`
    - Записать в output directory
    - Логировать: какие файлы созданы/обновлены
    - On-demand: не автоматический, запускается вручную/в CI

## Требования

- `--brand-*` переменные едины между Keycloak и фронтендом
- `brand-default.css` + `brand-sdek.css` — переключаются через `theme.properties`
- Лого из прототипа в Keycloak (SVG) и фронтенде (компонент Logo.tsx)
- Dark mode: automatic (`@media`) + manual (cookie toggle)
- `tenant_brand_config` → CSS генерация через `cmd/brand-sync`
- Seed `css_variables` — полная карта токенов

## Критерии приёмки

- [ ] `login.css` — все токены через `var(--brand-*, <default>)`
- [ ] `brand-default.css` — April teal, загружается перед `login.css`
- [ ] `brand-sdek.css` — цвета СДЭК из прототипа
- [ ] `switch-tenant-theme.sh` — переключает тему по имени tenant'а
- [ ] Лого Keycloak — SVG path из прототипа, `fill="currentColor"`
- [ ] Favicon Keycloak — отображается в tab браузера
- [ ] `Logo.tsx` — компонент с size variants (sm/md/lg)
- [ ] Shell.tsx — `<Logo size="sm" />` вместо текста "ЛКФЛ"
- [ ] `brand-default.css` фронтенд — подключён в `main.tsx`
- [ ] Dark mode `@media` — работает при system preference
- [ ] Переключатель — ☀️/🌙, cookie `lkfl-theme`, приоритет над system
- [ ] Account console — dark mode через `@media`
- [ ] Seed `css_variables` — полная карта 18 токенов
- [ ] `cmd/brand-sync/main.go` — генерирует CSS из БД
- [ ] `go build ./...` без ошибок
- [ ] `npm run build` без ошибок

## Зависимости

- **depends_on:** T1712 (Keycloak тема lkfl)
- **touches:** `infra/keycloak/theme/`, `frontend/src/components/common/`, `frontend/src/styles/`, `cmd/brand-sync/`, `cmd/seed/main.go`
- **risk:** Mustache template override — зависит от версии Keycloak 25+; logo path — бренд СДЭК, не нейтральный
