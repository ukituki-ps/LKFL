# T2108 — Dashboard stub

## Веха

M21-frontend-catalog

## Тип

code

## Контекст

Главная страница (Dashboard) для сотрудника — заглушка с приветствием и placeholder-блоками.
Зависит от T2105 (Layout — Shell) и T2109 (API layer).

### Backend endpoint (из `backend/internal/app/server.go`)

```
GET /api/v1/users/me
```

Требует: JWT + tenant middleware.
Response: профиль пользователя.

```json
{
  "id": "uuid",
  "email": "user@example.com",
  "first_name": "Иван",
  "last_name": "Петров",
  "tenant_id": "uuid",
  "keycloak_sub": "keycloak-uuid",
  "status": "active",
  "created_at": "2025-01-01T00:00:00Z"
}
```

## Что сделать

### `src/pages/Dashboard.tsx` — главная страница

```tsx
import { useQuery } from '@tanstack/react-query'
import { getUserProfile } from '@/api/user'
import { Card, Text, Group, Stack } from '@mantine/core'
import { Link } from 'react-router-dom'

export function Dashboard() {
  const { data: user, isLoading } = useQuery({
    queryKey: ['user-profile'],
    queryFn: () => getUserProfile(),
  })

  if (isLoading) return <div>Загрузка...</div>

  const greeting = getGreeting()
  const userName = user?.first_name || 'Сотрудник'

  return (
    <div className="dashboard">
      {/* Greeting */}
      <h1>
        {greeting}, {userName}!
      </h1>

      {/* Stat cards */}
      <Group gap="md" mb="xl">
        <StatCard title="Баланс баллов" value="—" icon="star" />
        <StatCard title="Активные льготы" value="—" icon="check" />
        <StatCard title="Доступные льготы" value="—" icon="grid" />
      </Group>

      {/* Event feed placeholder */}
      <Card withBorder mb="xl">
        <Text fw={600} mb="md">Последние события</Text>
        <Text c="dimmed" size="sm">
          События появятся после активации льгот
        </Text>
      </Card>

      {/* Quick actions */}
      <Card withBorder>
        <Text fw={600} mb="md">Быстрые действия</Text>
        <Stack gap="sm">
          <Link to="/catalog" style={{ textDecoration: 'none' }}>
            <Card withBorder>
              <Text>Просмотреть каталог льгот</Text>
            </Card>
          </Link>
          <Link to="/documents" style={{ textDecoration: 'none' }}>
            <Card withBorder>
              <Text>Мои документы</Text>
            </Card>
          </Link>
          <Link to="/support" style={{ textDecoration: 'none' }}>
            <Card withBorder>
              <Text>Обратиться в поддержку</Text>
            </Card>
          </Link>
        </Stack>
      </Card>
    </div>
  )
}

function getGreeting(): string {
  const hour = new Date().getHours()
  if (hour < 6) return 'Доброй ночи'
  if (hour < 12) return 'Доброе утро'
  if (hour < 18) return 'Добрый день'
  return 'Добрый вечер'
}

function StatCard({ title, value, icon }: { title: string; value: string; icon: string }) {
  return (
    <Card withBorder flex={{}} style={{ flex: 1, minWidth: 180 }}>
      <Text size="xs" c="dimmed" mb="xs">{title}</Text>
      <Text fw={600} size="xl">{value}</Text>
    </Card>
  )
}
```

### Placeholder блоки

| Блок | Описание | Значение |
|------|----------|----------|
| Баланс баллов | Отображение баланса (F2, M23) | «—» |
| Активные льготы | Счётчик активных (F2, M26) | «—» |
| Доступные льготы | Счётчик из каталога (можно подтянуть из M20 API) | «—» |
| Последние события | Event feed (F2) | Заглушка |
| Быстрые действия | Ссылки на /catalog, /documents, /support | Реальные ссылки |

## Требования

- Приветствие с именем пользователя (first_name из `/api/v1/users/me`)
- Временное приветствие (утро/день/вечер/ночь)
- Placeholder stat cards (значения «—», подключаются в F2)
- Placeholder event feed
- Quick actions с реальными ссылками на /catalog, /documents, /support
- Загрузка профиля через React Query

## Критерии приёмки

- [ ] `src/pages/Dashboard.tsx` — главная страница
- [ ] Приветствие: `{greeting}, {first_name}!`
- [ ] Данные профиля из `GET /api/v1/users/me`
- [ ] 3 stat card placeholder (Баланс, Активные, Доступные)
- [ ] Event feed placeholder
- [ ] Quick actions: ссылки на /catalog, /documents, /support
- [ ] Loading state при загрузке профиля
- [ ] Временное приветствие (4 варианта)
