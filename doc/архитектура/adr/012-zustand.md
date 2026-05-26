# ADR-012: Zustand для state management

**Статус:** Accepted
**Дата:** 2026-05-22
**Контекст:** М01-создание-описания

## Контекст

Frontend — 5 страниц, 3 модалки, относительно простой state: баланс, каталог, уведомления, user profile.

## Решение

**Zustand** — лёгкий, typed, devtools, no boilerplate:

```ts
import { create } from 'zustand';

interface BalanceStore {
  balance: number;
  categories: Record<string, number>;
  transactions: Transaction[];
  fetchBalance: () => Promise<void>;
}

export const useBalanceStore = create<BalanceStore>((set) => ({
  balance: 0,
  categories: {},
  transactions: [],
  fetchBalance: async () => {
    const data = await api.get('/billing/v1/accounts/me/balance');
    set({ balance: data.balance, categories: data.categories });
  },
}));
```

## Альтернативы

| Вариант | Плюсы | Минусы |
|---------|-------|--------|
| Redux Toolkit | Mature, DevTools, ecosystem | Overkill для 5 страниц, boilerplate |
| React Context | Встроен | Re-render всех consumers, медленный |
| Jotai | Atomic, лёгкий | Меньше ecosystem, сложнее debug |
| Recoil | Facebook, atomic | Deprecated-ish, сложная миграция |

## Вердикт

**Zustand.** Лёгкий (1KB), TypeScript-first, devtools middleware, достаточно для текущего объёма. Если вырастет → миграция на Redux Toolkit без боли.

## Следствия

- 4 stores: `balance`, `catalog`, `user`, `notifications`
- API client (`fetch`) → store update
- No global state для модалок (local state через `useState`)
