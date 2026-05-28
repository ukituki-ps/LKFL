# T2105 — Layout

## Веха

M21-frontend-catalog

## Тип

code

## Контекст

Компонент layout (оболочка приложения) с навигацией для двух групп пользователей.
Зависит от T2102 (AprilProviders — MantineProvider, тема) и T2104 (auth flow — роли из authStore).

Компоненты April UI: `@ukituki-ps/april-ui` предоставляет `AprilShellBar` (sidebar + header).

## Что сделать

### `src/components/layout/Shell.tsx` — оболочка приложения

```tsx
import { AprilShellBar, AprilShellBarSection } from '@ukituki-ps/april-ui'
import { useAuthStore } from '@/stores/authStore'
import { Outlet } from 'react-router-dom'
import { EmployeeNav } from './EmployeeNav'
import { AdminNav } from './AdminNav'

export function Shell() {
  const { userRoles } = useAuthStore()
  const isAdmin = userRoles.includes('admin') || userRoles.includes('catalog_manager') || userRoles.includes('hr')

  return (
    <AprilShellBar>
      <AprilShellBarSection position="left">
        {isAdmin ? <AdminNav /> : <EmployeeNav />}
      </AprilShellBarSection>
      <AprilShellBarSection position="main">
        <main className="shell-content">
          <Outlet />
        </main>
      </AprilShellBarSection>
      <AprilShellBarSection position="right">
        {/* User avatar, logout */}
        <UserMenu />
      </AprilShellBarSection>
    </AprilShellBar>
  )
}
```

### `src/components/layout/EmployeeNav.tsx` — навигация сотрудника

Меню для ролей: `employee`, `catalog_manager`, `admin`.

```tsx
import { Link, useLocation } from 'react-router-dom'

const employeeMenu = [
  { path: '/', label: 'Главная', icon: 'home' },
  { path: '/catalog', label: 'Каталог льгот', icon: 'grid' },
  { path: '/points', label: 'Баллы', icon: 'star' },
  { path: '/documents', label: 'Документы', icon: 'file' },
  { path: '/support', label: 'Поддержка', icon: 'help' },
]

export function EmployeeNav() {
  const location = useLocation()

  return (
    <nav>
      {employeeMenu.map(item => (
        <Link
          key={item.path}
          to={item.path}
          className={location.pathname === item.path ? 'active' : ''}
        >
          {item.label}
        </Link>
      ))}
    </nav>
  )
}
```

### `src/components/layout/AdminNav.tsx` — навигация администратора

Меню для ролей: `admin`, `catalog_manager`, `hr`.

```tsx
import { Link, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'

const adminMenu = [
  { path: '/admin/hr', label: 'HR', icon: 'users', roles: ['hr', 'admin'] },
  { path: '/admin/catalog', label: 'Каталог', icon: 'grid', roles: ['catalog_manager', 'admin'] },
  { path: '/admin/content', label: 'Контент', icon: 'file-text', roles: ['admin'] },
]

export function AdminNav() {
  const location = useLocation()
  const { userRoles } = useAuthStore()

  const visibleItems = adminMenu.filter(item =>
    item.roles.some(role => userRoles.includes(role))
  )

  return (
    <nav>
      {visibleItems.map(item => (
        <Link
          key={item.path}
          to={item.path}
          className={location.pathname === item.path ? 'active' : ''}
        >
          {item.label}
        </Link>
      ))}
    </nav>
  )
}
```

### `src/components/layout/UserMenu.tsx` — меню пользователя

```tsx
import { useAuthStore } from '@/stores/authStore'

export function UserMenu() {
  const { user, logout } = useAuthStore()

  return (
    <div>
      <span>{user?.first_name} {user?.last_name}</span>
      <button onClick={logout}>Выйти</button>
    </div>
  )
}
```

### Responsive

Desktop first (ширина контента ≥ 1200px). Мобильная адаптация — в следующих вехах.

## Требования

- AprilShellBar из @ukituki-ps/april-ui
- EmployeeNav: 5 пунктов (Главная, Каталог, Баллы, Документы, Поддержка)
- AdminNav: 3 пункта (HR, Каталог, Контент) с RBAC-фильтрацией
- Active state на текущем маршруте
- UserMenu с именем пользователя (из authStore.user) и кнопкой logout
- Desktop first (≥ 1200px)

## Критерии приёмки

- [ ] `src/components/layout/Shell.tsx` — оболочка с AprilShellBar
- [ ] `src/components/layout/EmployeeNav.tsx` — 5 пунктов навигации
- [ ] `src/components/layout/AdminNav.tsx` — 3 пункта с RBAC-фильтрацией
- [ ] `src/components/layout/UserMenu.tsx` — имя + logout
- [ ] Active state на текущем маршруте
- [ ] Контент рендерится через Outlet
- [ ] Desktop layout ≥ 1200px
