# T1712 — Report

## Дата

2026-05-31

## Что сделано

### 1. Freeform тема `lkfl` (login)

**`infra/keycloak/theme/lkfl/login/`**

- `theme.properties` — `parent=keycloak`, `import=keycloak.v2`, logo, favicon, CSS
- `resources/lkfl-theme/logo.svg` — placeholder teal лого 180×48
- `resources/css/login.css` — **478 строк** стилей на основе April tokens

**April tokens в CSS:**

| Токен | Значение | Применение |
|-------|----------|-----------|
| `--lkfl-primary` | `#12b886` | primary кнопки, ссылки, checkbox accent |
| `--lkfl-primary-hover` | `#0ca678` | hover primary |
| `--lkfl-primary-active` | `#099268` | active primary |
| `--lkfl-primary-light` | `#c3fae8` | фокус инпута, selected list item |
| `--lkfl-bg` | `#ffffff` | карточка, инпуты |
| `--lkfl-bg-alt` | `#f8f9fa` | фон страницы |
| `--lkfl-border` | `#EBEBEB` | карточка, разделители |
| `--lkfl-border-input` | `#dee2e6` | инпуты default |
| `--lkfl-text` | `#212529` | текст основной |
| `--lkfl-text-muted` | `#868e96` | текст вспомогательный, footer |
| `--lkfl-text-subtle` | `#9CA3AF` | placeholder, footer |
| `--lkfl-radius-card` | `14px` | `.login-card` |
| `--lkfl-radius-btn` | `6px` | кнопки, инпуты, алерты |
| `--lkfl-radius-input` | `6px` | инпуты |
| `--lkfl-shadow-card` | `0 1px 4px rgba(0,0,0,0.06)` | `.login-card` |

### 2. Freeform тема `lkfl` (account console)

**`infra/keycloak/theme/lkfl/account/`**

- `theme.properties` — наследует keycloak base
- `resources/css/login.css` — `@import` login.css + account-specific overrides

### 3. Realm config

**`infra/keycloak/realm-lkfl-sdek.json`**

```diff
-  "loginTheme": "keycloak",
-  "accountTheme": "keycloak",
+  "loginTheme": "lkfl",
+  "accountTheme": "lkfl",
```

### 4. Docker Compose — все 4 файла

Volume mount добавлен в каждый:

```yaml
volumes:
  - ./infra/keycloak/theme:/opt/keycloak/themes/lkfl:ro
```

| Файл | Статус |
|------|--------|
| `docker-compose.yml` | ✅ theme mount добавлен |
| `docker-compose.dev.yml` | ✅ theme mount добавлен |
| `docker-compose.staging.yml` | ✅ theme mount добавлен |
| `docker-compose.prod.yml` | ✅ theme mount добавлен |

### 5. Коммит и PR

- Коммит: `f225cb4` → squash merge → `be39a20` на `main`
- PR: https://github.com/ukituki-ps/LKFL/pull/5

## Изменённые файлы

| Файл | Действие |
|------|---------|
| `infra/keycloak/theme/lkfl/login/theme.properties` | создан (25 строк) |
| `infra/keycloak/theme/lkfl/login/resources/lkfl-theme/logo.svg` | создан (4 строки) |
| `infra/keycloak/theme/lkfl/login/resources/css/login.css` | создан (478 строк) |
| `infra/keycloak/theme/lkfl/account/theme.properties` | создан (16 строк) |
| `infra/keycloak/theme/lkfl/account/resources/css/login.css` | создан (13 строк) |
| `infra/keycloak/realm-lkfl-sdek.json` | `loginTheme`/`accountTheme` → `lkfl` |
| `docker-compose.yml` | +theme mount +import-realm |
| `docker-compose.dev.yml` | +theme mount |
| `docker-compose.staging.yml` | +theme mount |
| `docker-compose.prod.yml` | +theme mount |

## Статус

✅ Все критерии приёмки выполнены:
- [x] Freeform тема `lkfl` создана (login + account)
- [x] CSS-стили на основе April tokens v0.1.16
- [x] Realm config: `loginTheme: lkfl`, `accountTheme: lkfl`
- [x] Все 4 compose-файла монтируют тему
- [x] Responsive (< 480px)
- [x] Коммит + PR #5 merged на main

## Следующие шаги (опционально)

1. Заменить placeholder лого SVG на реальный логотип бренда tenant'а
2. Per-tenant темы: `theme/sdek/` — копия `lkfl/` с кастомными цветами
3. Dark mode в Keycloak: добавить `[data-mantine-color-scheme='dark']` секцию в `login.css`
4. ADR-039: документировать white-label стратегию тем Keycloak
