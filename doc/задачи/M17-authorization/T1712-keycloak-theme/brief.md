# T1712 — Keycloak тема lkfl: April дизайн-система v0.1.16

## Веха

M17-authorization

## Тип

code

## Проблема

Keycloak на стенде отображается с дефолтной белым стилем (`loginTheme: keycloak`, `accountTheme: keycloak`). Страница логина и account console не соответствуют визуальной идентичности платформы — нет цветов April дизайн-системы, нет карточек, кнопок, радиусов, теней.

**Следствия:**
- Страница логина Keycloak выглядит «чужой» на фоне стилизованного фронтенда
- Нет brand consistency: пользователь видит стандартный Keycloak → ощущение «не доделанности»
- Невозможно white-label: каждый tenant не может задать свои цвета/лого через тему

## Что делать

### 1. Создать freeform тему `lkfl`

**Каталог:** `infra/keycloak/theme/lkfl/`

Freeform theme (Keycloak 17+) — обычный каталог, монтируется как volume. Не требует JAR, не требует пересборки.

Структура:
```
infra/keycloak/theme/lkfl/
├── login/
│   ├── theme.properties          # parent=keycloak, import=keycloak.v2
│   └── resources/
│       ├── css/login.css         # ~478 строк стилей (April tokens)
│       └── lkfl-theme/logo.svg   # Лого placeholder (teal)
└── account/
    ├── theme.properties          # parent=keycloak, import=keycloak.v2
    └── resources/css/login.css   # @import login.css + account overrides
```

### 2. CSS-стили на основе April tokens

**Файл:** `infra/keycloak/theme/lkfl/login/resources/css/login.css`

Токены из `@ukituki-ps/april-tokens@0.1.16`:

| Токен | Значение | Элементы |
|-------|----------|----------|
| `--lkfl-primary` | `#12b886` | кнопки primary, ссылки, checkbox accent |
| `--lkfl-primary-hover` | `#0ca678` | hover primary |
| `--lkfl-primary-active` | `#099268` | active primary |
| `--lkfl-primary-light` | `#c3fae8` | фокус инпута, selected list |
| `--lkfl-bg` | `#ffffff` | карточка, инпуты |
| `--lkfl-bg-alt` | `#f8f9fa` | фон страницы |
| `--lkfl-border` | `#EBEBEB` | бордер карточки, разделители |
| `--lkfl-border-input` | `#dee2e6` | бордер инпутов |
| `--lkfl-text` | `#212529` | текст основной |
| `--lkfl-text-muted` | `#868e96` | текст вспомогательный |
| `--lkfl-radius-card` | `14px` | `.login-card` |
| `--lkfl-radius-btn` | `6px` | кнопки, инпуты |
| `--lkfl-shadow-card` | `0 1px 4px rgba(0,0,0,0.06)` | `.login-card` |

Стилизуемые элементы:
- `.login-pf` — фон страницы
- `.login-card` — карточка с бордером, тенью, радиусом
- `input[type=text/password/email]` — инпуты с фокусом
- `.button.primary` — primary кнопка (100% width)
- `.button.default` — secondary кнопка
- `.alert-error` / `.alert-success` — алерты
- `.user-list` — список пользователей (selector)
- `.auth-selector` — 2FA/WebAuthn
- `.social-link` — social providers
- `.login-header`, `.login-footer` — header/footer

### 3. Обновить realm config

**Файл:** `infra/keycloak/realm-lkfl-sdek.json`

```diff
-  "loginTheme": "keycloak",
-  "accountTheme": "keycloak",
+  "loginTheme": "lkfl",
+  "accountTheme": "lkfl",
```

### 4. Монтировать тему во все compose-файлы

**Файлы:** `docker-compose.yml`, `docker-compose.dev.yml`, `docker-compose.staging.yml`, `docker-compose.prod.yml`

```yaml
volumes:
  - ./infra/keycloak/theme:/opt/keycloak/themes/lkfl:ro
```

### 5. Лого placeholder

**Файл:** `infra/keycloak/theme/lkfl/login/resources/lkfl-theme/logo.svg`

SVG 180×48, teal прямоугольник с текстом «LKFL». Заменить на реальный логотип бренда tenant'а.

## Требования

- Тема `lkfl` загружается Keycloak при старте (freeform, volume mount)
- Страница логина использует April tokens (теал, карточки, бордеры)
- Account console наследует стили логина
- Realm config использует тему `lkfl`
- Все 4 compose-файла монтируют тему
- Responsive: корректное отображение на мобильных (< 480px)
- White-label ready: тему можно скопировать → переименовать → изменить CSS variables

## Критерии приёмки

- [ ] `docker compose up -d keycloak` — Keycloak стартует без ошибок
- [ ] Страница логина (`/realms/lkfl-sdek/login/`) — teal кнопки, карточка 14px, тени
- [ ] Инпуты с фокусом (border teal + box-shadow)
- [ ] Алерты стилизованы (error red, success green)
- [ ] Account console стилизован
- [ ] Responsive (< 480px) — карточка full-width
- [ ] Realm config: `loginTheme: lkfl`, `accountTheme: lkfl`
- [ ] Все compose-файлы монтируют тему

## Зависимости

- **depends_on:** T1701 (infra bootstrap), T1906 (Keycloak realm)
- **touches:** `infra/keycloak/theme/`, `infra/keycloak/realm-lkfl-sdek.json`, `docker-compose*.yml`
- **risk:** Минимальный — freeform тема, не затрагивает код бэкенда/фронтенда
