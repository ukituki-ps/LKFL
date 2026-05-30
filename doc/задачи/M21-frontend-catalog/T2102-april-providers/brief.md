# T2102 — AprilProviders

## Веха

M21-frontend-catalog

## Тип

code

## Контекст

Корневой провайдер для React-приложения на базе @ukituki-ps/april-ui и Mantine.
Зависит от T2101 (Vite bootstrap — package.json, vite.config.ts, структура src/).

Исходник: `doc/архитектура/фронтенд.md` §A (Обзор).

## Что сделать

### `src/main.tsx` — точка входа

```tsx
import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import { AprilProviders } from './lib/providers'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <AprilProviders>
      <App />
    </AprilProviders>
  </React.StrictMode>,
)
```

### `src/lib/providers.tsx` — AprilProviders

Компонент-обёртка, объединяющий:
- `MantineProvider` с темой `createAprilTheme()`
- `QueryClientProvider` из @tanstack/react-query
- `DensityProvider` из @mantine/core

```tsx
import { MantineProvider, MantineTheme, createTheme } from '@mantine/core'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// Создаём тему на основе April tokens
export function createAprilTheme(): MantineTheme {
  return createTheme({
    // April tokens будут подтягиваться из CSS variables
    // (переопределяются brand CSS из API tenant'а)
  })
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 min
      retry: 1,
    },
  },
})

export function AprilProviders({ children }: { children: React.ReactNode }) {
  return (
    <MantineProvider theme={createAprilTheme()}>
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    </MantineProvider>
  )
}
```

### Brand CSS variables

CSS переменные бренда загружаются с backend (endpoint `GET /admin/tenants/{id}/brand` — M20).
В M21 заглушка: используем default April tokens.

## Требования

- @ukituki-ps/april-ui ≥ 0.1.13
- @ukituki-ps/april-tokens ≥ 0.1.13
- @mantine/core ≥ 7.17.8
- @tanstack/react-query ≥ 5.50.0
- React.StrictMode включён

## Критерии приёмки

- [ ] `src/main.tsx` с ReactDOM.createRoot
- [ ] `src/lib/providers.tsx` — AprilProviders с MantineProvider + QueryClientProvider
- [ ] `createAprilTheme()` возвращает MantineTheme
- [ ] App рендерится внутри AprilProviders
- [ ] `npm run dev` → страница загружается без ошибок
