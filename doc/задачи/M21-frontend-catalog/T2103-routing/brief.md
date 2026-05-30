# T2103 — Routing

## Веха

M21-frontend-catalog

## Тип

code

## Контекст

Настройка роутинга для двух групп пользователей: сотрудники (employee) и администраторы (admin).
Зависит от T2101 (Vite bootstrap) и T2104 (auth flow — данные ролей из authStore).

Роутинг на базе `react-router-dom` v6.27 (установлен в T2101).

## Что сделать

### `src/App.tsx` — корневой роутер

```tsx
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { RequireAuth } from './components/auth/RequireAuth'
import { EmployeeRoutes } from './routes/employee'
import { AdminRoutes } from './routes/admin'

export function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Employee routes */}
        <Route element={<RequireAuth roles={['employee', 'catalog_manager', 'admin']} />}>
          <Route path="/" element={<Dashboard />} lazy />
          <Route path="/catalog" element={<Catalog />} lazy />
          <Route path="/points" element={<Points />} lazy />
          <Route path="/documents" element={<Documents />} lazy />
          <Route path="/support" element={<Support />} lazy />
        </Route>

        {/* Admin routes */}
        <Route element={<RequireAuth roles={['catalog_manager', 'admin']} />}>
          <Route path="/admin/hr" element={<AdminHR />} lazy />
          <Route path="/admin/catalog" element={<AdminCatalog />} lazy />
          <Route path="/admin/content" element={<AdminContent />} lazy />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
```

### `src/routes/employee.tsx` — маршруты сотрудника

Маршруты для всех авторизованных пользователей:
- `/` — Dashboard (главная страница)
- `/catalog` — Каталог льгот (основная страница M21)
- `/points` — Баланс баллов (stub)
- `/documents` — Документы (stub)
- `/support` — Поддержка (stub)

### `src/routes/admin.tsx` — маршруты администратора

Маршруты для ролей `catalog_manager` и `admin`:
- `/admin/hr` — Управление пользователями (stub)
- `/admin/catalog` — Управление каталогом (stub, CRUD в M22)
- `/admin/content` — Управление контентом (stub)

### `src/components/auth/RequireAuth.tsx` — компонент защиты маршрутов

```tsx
import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'

interface RequireAuthProps {
  roles?: string[] // если пустой массив — любой авторизованный
}

export function RequireAuth({ roles = [] }: RequireAuthProps) {
  const { isAuthenticated, userRoles } = useAuthStore()
  const location = useLocation()

  if (!isAuthenticated) {
    // Сохраняем attempted URL для редиректа после логина
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  // Проверка ролей (RBAC)
  if (roles.length > 0) {
    const hasRole = roles.some(role => userRoles.includes(role))
    if (!hasRole) {
      return <Navigate to="/forbidden" replace />
    }
  }

  return <Outlet />
}
```

### Lazy loading

Все страницы загружаются лениво через `React.lazy` + `Suspense`:

```tsx
import { lazy, Suspense } from 'react'

const Dashboard = lazy(() => import('@/pages/Dashboard'))
const Catalog = lazy(() => import('@/pages/Catalog'))

// В App.tsx:
<Suspense fallback={<div>Загрузка...</div>}>
  <Routes>...</Routes>
</Suspense>
```

### Роли (из M20 backend)

Роли определены в `backend/internal/user/model.go`:
- `employee` — сотрудник (по умолчанию)
- `catalog_manager` — менеджер каталога
- `hr` — HR-менеджер
- `admin` — полный доступ

Проверка ролей в backend: `sharedauth.RBACMiddleware([]string{"catalog_manager", "admin"})`.

## Требования

- react-router-dom ≥ 6.27 (установлен в T2101)
- RequireAuth проверяет isAuthenticated из authStore (T2104)
- RequireAuth проверяет роли из authStore.userRoles
- Lazy loading для всех страниц
- Redirect на /login при отсутствии auth
- Redirect на /forbidden при отсутствии роли

## Критерии приёмки

- [ ] `src/App.tsx` — BrowserRouter с Routes
- [ ] `src/routes/employee.tsx` — 5 маршрутов сотрудника
- [ ] `src/routes/admin.tsx` — 3 маршрута администратора
- [ ] `src/components/auth/RequireAuth.tsx` — проверка auth + ролей
- [ ] Lazy loading с Suspense для всех страниц
- [ ] Redirect на /login при отсутствии токена
- [ ] Redirect на /forbidden при отсутствии роли
- [ ] Роли соответствуют backend: employee, catalog_manager, hr, admin
