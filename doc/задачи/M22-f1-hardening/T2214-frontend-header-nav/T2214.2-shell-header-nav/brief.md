# T2214.2 — Shell → AprilProductHeader, навигация, header-right

## Родительская задача

T2214 — Frontend: прототип → код (header, дизайн, страницы-заглушки)

## Веха

M22-f1-hardening

## Тип

code

## Кратко

Заменить Mantine `AppShell` (sidebar-layout) на горизонтальный layout с `AprilProductHeader`. Реализовать HeaderNav, HeaderRight, мобильную навигацию (Burger → Drawer).

---

## Зависимости

- **T2214.1** (Brand tokens) — нужен `--brand-*` для цветов навигации

## Что сделать

### 1. `src/components/layout/Shell.tsx` — AprilProductHeader

```tsx
<div style={{ minHeight: '100vh', backgroundColor: 'var(--brand-bg)' }}>
  <AprilProductHeader
    left={<Logo onClick={() => navigate('/')} />}
    center={<HeaderNav isAdmin={isAdminRoute} />}
    right={<HeaderRight />}
    sticky
  />
  <main style={{ maxWidth: 1100, margin: '0 auto', padding: '28px 28px 56px' }}>
    <Outlet />
  </main>
</div>
```

**Header height:** DS `AprilProductHeader` → 56px (comfortable) / 48px (compact). Прототип — 58px. Разница ≤2px — приемлемо. Если критично — кастомный `style={{ height: 58 }}` на корневом div header'а.

**Sidebar убран** полностью (desktop). Мобильная навигация — Burger → Drawer.

### 2. `src/components/layout/HeaderNav.tsx` — горизонтальная навигация

5 ссылок из `employeeRoutes` (Главная, Каталог льгот, Баллы, Документы, Поддержка).

Для admin — `adminRoutes` (HR, Каталог, Контент).

Стили:
- `padding: 0 14px`, `font-size: 13px`, `font-weight: 500`
- inactive: `color: var(--brand-text-muted)`
- active: `color: var(--brand-text)`, `border-bottom: 2px solid var(--brand-green)`, `font-weight: 600`
- hover: `color: var(--brand-green)`

### 3. `src/components/layout/HeaderRight.tsx`

```tsx
<Group gap={10}>
  <BalancePill />   {/* bg: var(--brand-green-light), border: var(--brand-green-border) */}
  <BellIcon />      {/* 34x34, bg: var(--brand-row), border: 1px solid var(--brand-border) */}
  <UserMenu />      {/* avatar: 34px circle, bg: var(--brand-green) */}
</Group>
```

**BalancePill:** mock-значение «1 250» (до F2). Иконка `Coins` (Lucide).

**BellIcon:** кнопка-заглушка (переход на `/notifications` — TODO F2).

**UserMenu:** обновить avatar → `bg: var(--brand-green)`, size 34px.

### 4. Мобильная навигация

```tsx
const isMobile = useMediaQuery('(max-width: 768px)')

// Desktop: горизонтальная навигация в HeaderNav
// Mobile: Burger → Drawer с теми же ссылками
{isMobile && (
  <Drawer opened={opened} onClose={close} position="left" title="Меню">
    <EmployeeNav onClose={close} />
  </Drawer>
)}
```

Breakpoint: `768px` (как в текущем Shell).

### 5. `src/components/layout/EmployeeNav.tsx` — упрощение

Оставить только для мобильного Drawer. Убрать sidebar-специфичные стили.

### 6. `src/components/layout/AdminNav.tsx` — упрощение

Аналогично EmployeeNav — только для мобильного Drawer.

### 7. `src/components/layout/UserMenu.tsx` — адаптация

- Avatar: `bg: var(--brand-green)`, size 34px circle, инициалы `font-size: 11px`
- Dropdown: email + «Выйти» (без изменений)

## Файлы

### Новые
- `src/components/layout/HeaderNav.tsx`
- `src/components/layout/HeaderRight.tsx`

### Изменяются
- `src/components/layout/Shell.tsx` — замена `AppShell` на `AprilProductHeader`
- `src/components/layout/EmployeeNav.tsx` — упрощение (только Drawer)
- `src/components/layout/AdminNav.tsx` — упрощение (только Drawer)
- `src/components/layout/UserMenu.tsx` — avatar → зелёный

## Критерии приёмки

### Header и layout
- [ ] `Shell.tsx` использует `AprilProductHeader` (не `AppShell`)
- [ ] Sidebar убран (навигация горизонтальная в header'е)
- [ ] 5 ссылок навигации с underline active-индикатором
- [ ] Balance-pill в правой зоне header'а
- [ ] Bell icon button в правой зоне header'а
- [ ] Avatar с инициалами (зелёный круг 34px) + dropdown
- [ ] Sticky header (прилипает при скролле)
- [ ] Фон страницы `#F2F2F2`
- [ ] Content max-width 1100px, margin auto, padding 28px

### Мобильная навигация
- [ ] Burger кнопка видна на `≤ 768px`
- [ ] Drawer с теми же ссылками
- [ ] Нажатие на ссылку → закрытие Drawer

### Admin маршруты
- [ ] Admin навигация в header'е (HR, Каталог, Контент)
- [ ] Переключение employee/admin по `isAdminRoute`

### Тесты
- [ ] E2E тесты не ломаются (или обновлены для нового layout'а)
