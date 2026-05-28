# Отчёт T2103 — Routing

## Статус
выполнено

## Что сделано

### 1. `src/components/auth/RequireAuth.tsx` — компонент защиты маршрутов
- Проверка `isAuthenticated` из `useAuthStore`
- Redirect на `/login` с сохранением attempted URL в `state.from`
- RBAC: проверка ролей через `userRoles.includes(role)`
- Redirect на `/forbidden` при отсутствии нужной роли
- `roles = []` — любой авторизованный пользователь

### 2. `src/App.tsx` — корневой роутер
- `BrowserRouter` (перенесён из main.tsx в App.tsx)
- `Suspense` с fallback «Загрузка...» для всех lazy компонентов
- Auth routes без защиты: `/login`, `/callback`, `/forbidden`
- Employee routes через `RequireAuth` + `Shell`: `/`, `/catalog`, `/points`, `/documents`, `/support`
- Admin routes через `RequireAuth` + `Shell`: `/admin/hr`, `/admin/catalog`, `/admin/content`
- Catch-all `*` → redirect на `/`
- Lazy loading всех страниц через `React.lazy` + `.then(m => ({ default: m.Component }))`

### 3. `src/routes/employee.tsx` — определения employee маршрутов
- 5 маршрутов: Главная, Каталог льгот, Баллы, Документы, Поддержка
- Тип `EmployeeRoute` для TypeScript

### 4. `src/routes/admin.tsx` — определения admin маршрутов
- 3 маршрута: HR, Каталог, Контент
- Каждый маршрут содержит `roles` для RBAC проверки
- Тип `AdminRoute` для TypeScript

### 5. Stub страницы администратора
- `src/pages/AdminHR.tsx`
- `src/pages/AdminCatalog.tsx`
- `src/pages/AdminContent.tsx`

## Результаты проверки
- `npm run build` — ✅ успешно (1.26s, 808 modules)
- `npm run lint` — ✅ без ошибок

## Замечания
- Shell.tsx остаётся stub-ом (будет реализован в T2105)
- Все employee страницы (Dashboard, Catalog, Points, Documents, Support) остаются stub-ами
- Admin страницы созданы как stub-ы
- RequireAuth использует `useAuthStore` из T2104 (auth flow)
