# Отчёт T2105 — Layout

## Статус
выполнено

## Дата
2026-05-26

## Что сделано

### Созданные/изменённые файлы

| Файл | Описание |
|------|----------|
| `src/components/layout/Shell.tsx` | Оболочка приложения на Mantine AppShell |
| `src/components/layout/EmployeeNav.tsx` | Навигация сотрудника (5 пунктов) |
| `src/components/layout/AdminNav.tsx` | Навигация администратора (3 пункта, RBAC) |
| `src/components/layout/UserMenu.tsx` | Меню пользователя (аватар, имя, роль, logout) |

### Реализация

**Shell.tsx** — оболочка приложения:
- Mantine `AppShell` с header (60px), navbar (250px), main content
- Burger-меню для мобильных устройств (< 768px) через `useMediaQuery`
- Автоопределение контекста: `/admin/*` маршруты → AdminNav, остальное → EmployeeNav
- RBAC-проверка: admin/catalog_manager/hr роли получают доступ к AdminNav
- Контент рендерится через `<Outlet />`

**EmployeeNav.tsx** — навигация сотрудника:
- 5 пунктов из `@/routes/employee` констант (T2103)
- Active state на текущем маршруте (подсветка синим)
- `onClose` callback для закрытия мобильного меню

**AdminNav.tsx** — навигация администратора:
- 3 пункта из `@/routes/admin` констант (T2103)
- RBAC-фильтрация: показывает только пункты, доступные по ролям пользователя
- Active state на текущем маршруте
- `onClose` callback для закрытия мобильного меню

**UserMenu.tsx** — меню пользователя:
- Аватар с инициалами из `authStore.user`
- Имя и роль (desktop, ≥ 768px)
- Dropdown: email + кнопка «Выйти» (вызывает `authStore.logout`)
- Роли отображаются на русском (Сотрудник, Менеджер каталога, HR, Администратор)

### Технические решения

- **MediaQuery**: компонент `MediaQuery` отсутствует в `@mantine/core` v7. Использован `useMediaQuery` хук из `@mantine/hooks` с breakpoint `(max-width: 768px)` для мобильного меню.
- **AprilShellBar**: не используется (placeholder в `@ukituki-ps/april-ui`). Полностью на Mantine `AppShell`.
- **RBAC**: проверка ролей через `authStore.userRoles` + `adminRoutes[].roles` фильтрация.

## Результаты проверки

- `npm run build` — ✅ успешно (tsc + vite build, 1.71s)
- `npm run lint` — ✅ без ошибок
- Shell chunk: 87.99 kB (gzip: 28.82 kB)

## Замечания

- Мобильная адаптация базовая (burger menu + useMediaQuery). Детальная адаптация под мобильные — в следующих вехах.
- Иконки в навигации не реализованы (предусмотрены в route-константах, но UI без иконок).
