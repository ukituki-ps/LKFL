# T1509 — Mobile + Forms архитектура

## Веха

M15-архитектура-фронтенда

## Контекст

LKFL — платформа для 100 000+ сотрудников. Большинство заходит с мобильных. DS `DESIGN_SYSTEM.md §8` описывает детальную мобильную архитектуру (400+ строк нормативного текста). Формы — критическая часть: регистрация, wizard-ы (DMS, MatCapital), admin CRUD, survey.

**Survey-формы:** модель данных и branching описаны в ADR-025 (Survey Engine, 633 строки). T1509 описывает только frontend-слой (компоненты вопросов, валидация, optimistic submit). Backend-модель — из ADR-025, не изобретать.

Эта задача создаёт отдельный документ `фронтенд-mobile-forms.md` — самостоятельный архитектурный артефакт, на который ссылается T1502 (§§I-J).

## Что сделать

### Mobile-архитектура

1. **`AprilMobileShellBar`** — единая нижняя панель
   - «Один активный контекст» — экран ИЛИ модалка, не оба одновременно
   - Слоты: `leading` (Назад), `center` (табы/действия), поиск
   - `position: fixed` к viewport (root) или `absolute` (внутренний слой)
   - z-index: `APRIL_MOBILE_SHELL_BAR_Z_INDEX` (над sheet оверлеями)

2. **Модальности**
   - Desktop: `AprilModal` (центрированный, scrollable body, headerActions)
   - Mobile: `AprilVaulBottomSheet` + `AprilMobileShellBar` (свайп закрытие, действия в панели)
   - Fallback: `AprilMobileBottomSheet` (Mantine Drawer, без Vaul жестов)

3. **Breakpoints**
   - ≥1280px: sidebar 250px + контент
   - 768-1279px: collapsible sidebar
   - <768px: mobile bottom bar, полноэкранные страницы

4. **Touch-ориентиры**
   - Мин. зона касания: 44×44px
   - Safe area: `env(safe-area-inset-bottom)`
   - Content padding: `aprilMobileShellBarContentPaddingBottom()`
   - Font-size у инпутов ≥16px (iOS Safari zoom prevention)

5. **Жесты**
   - Android back + «Назад» в панели согласованы
   - iOS swipe-back
   - Pull-to-refresh — только осознанно

### Forms-архитектура

1. **Бизнес-формы** (регистрация, wizard шаги, обращения)
   - Zod + react-hook-form (peer deps DS)
   - Каждая форма — Zod schema + `useForm`
   - Валидация: клиентская (Zod) + серверная (API)
   - Error display: Mantine `error` + `description`

2. **Admin-формы** (CRUD карточек, провайдеров, правил)
   - `AprilJsonSchemaForm` (RJSF + Ajv8) для кастомных форм по JSON Schema
   - Или Zod + react-hook-form (если JSON Schema overkill)

3. **Wizard-формы** (DMS upgrade, MatCapital)
   - Каждая шаг — отдельная форма с Zod schema
   - Состояние wizard — Zustand `useWizardsStore`
   - Validation на каждом шаге → блокировка forward
   - Review-шаг → summary всех данных

4. **Survey-формы** (M14)
   - Динамические вопросы (branching)
   - Каждый вопрос — компонент (Select, Radio, Text, Rating)
   - Валидация: required + format
   - Optimistic submit → rollback на error

## Результат

- `архитектура/фронтенд-mobile-forms.md` — отдельный документ (Mobile + Forms)
- Готов к включению в T1502 (или самостоятельному использованию)

## Критерии приёмки

- [ ] Mobile: `AprilMobileShellBar` описан с MUST/MUST NOT правилами
- [ ] Mobile: модальности (AprilModal vs AprilVaulBottomSheet) — когда что
- [ ] Mobile: breakpoints + touch-ориентиры + safe area
- [ ] Mobile: жесты (Android back, iOS swipe-back)
- [ ] Forms: Zod + react-hook-form — стратегия бизнес-форм
- [ ] Forms: AprilJsonSchemaForm — admin-формы (опционально)
- [ ] Forms: wizard — Zustand store + step validation
- [ ] Forms: survey — динамические вопросы + optimistic submit
