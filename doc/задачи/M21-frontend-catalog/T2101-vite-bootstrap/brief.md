# T2101 — Vite + React 18 Bootstrap

## Веха

M21-frontend-catalog

## Тип

code

## Контекст

Инициализация фронтенда: Vite + React 18 + TypeScript + April UI + Mantine.
Исходник: `doc/архитектура/фронтенд.md` §A (Обзор).

## Что сделать

### `frontend/package.json`

```json
{
  "name": "lkfl-frontend",
  "private": true,
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "lint": "eslint src/ --ext .ts,.tsx",
    "test": "vitest",
    "test:e2e": "playwright test"
  },
  "dependencies": {
    "@ukituki-ps/april-ui": "^0.1.13",
    "@ukituki-ps/april-tokens": "^0.1.13",
    "@mantine/core": "^7.17.8",
    "@mantine/hooks": "^7.17.8",
    "react": "^18.3.1",
    "react-dom": "^18.3.1",
    "react-router-dom": "^6.27.0",
    "zustand": "^5.0.0",
    "@tanstack/react-query": "^5.50.0"
  },
  "devDependencies": {
    "@types/react": "^18.3.0",
    "@types/react-dom": "^18.3.0",
    "@vitejs/plugin-react": "^4.3.0",
    "typescript": "^5.5.0",
    "vite": "^6.0.0",
    "eslint": "^8.57.0",
    "@typescript-eslint/eslint-plugin": "^7.0.0",
    "@typescript-eslint/parser": "^7.0.0",
    "vitest": "^2.0.0",
    "@testing-library/react": "^16.0.0",
    "@playwright/test": "^1.45.0"
  }
}
```

### `frontend/vite.config.ts`

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/admin': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
})
```

### Структура `frontend/src/`

```
src/
├── main.tsx                 # Entry point (AprilProviders)
├── App.tsx                  # Router
├── routes/                  # Route definitions
│   ├── employee.tsx         # Employee routes
│   └── admin.tsx            # Admin routes
├── pages/                   # Page components
│   ├── Dashboard.tsx
│   ├── Catalog.tsx
│   ├── Points.tsx
│   ├── Documents.tsx
│   └── Support.tsx
├── components/              # Shared components
│   ├── layout/
│   │   ├── Shell.tsx
│   │   ├── Sidebar.tsx
│   │   └── Header.tsx
│   ├── catalog/
│   │   ├── EngagementCard.tsx
│   │   ├── FilterBar.tsx
│   │   └── SearchInput.tsx
│   └── auth/
│       └── RequireAuth.tsx
├── stores/                  # Zustand stores
│   ├── authStore.ts
│   └── catalogStore.ts
├── api/                     # API layer
│   ├── client.ts            # fetch wrapper
│   ├── engagements.ts
│   └── types.ts             # OpenAPI generated types
├── lib/                     # Utilities
│   └── theme.ts             # April theme creation
└── assets/                  # Static assets
```

## Требования

- Vite 6 + React 18 + TypeScript 5.5
- April UI + April Tokens + Mantine 7
- Zustand 5 (state management)
- React Query 5 (data fetching)
- React Router 6.27 (routing)
- Alias `@` → `src/`
- Vite proxy для dev (`/api` → `:8080`)
- ESLint + TypeScript strict mode

## Критерии приёмки

- [ ] `frontend/package.json` с зависимостями
- [ ] `frontend/vite.config.ts` с proxy
- [ ] `frontend/src/` структура создана
- [ ] `npm install` без ошибок
- [ ] `npm run dev` запускает Vite dev server
- [ ] `npm run build` собирает dist
- [ ] `npm run lint` без ошибок
- [ ] TypeScript strict mode
