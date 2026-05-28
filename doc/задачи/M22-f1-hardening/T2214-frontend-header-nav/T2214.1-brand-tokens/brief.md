# T2214.1 — Brand tokens, тема, шрифт

## Родительская задача

T2214 — Frontend: прототип → код (header, дизайн, страницы-заглушки)

## Веха

M22-f1-hardening

## Тип

code

## Кратко

Подключить CSS-токены бренда из прототипа, обновить Mantine-тему на зелёный `#00B33C`, подключить шрифт Inter, обновить провайдеры.

**Фундамент:** без этой подзадачи остальные визуально не совпадут с прототипом.

---

## Что сделать

### 1. `src/components/ui/BrandTokens.css`

CSS переменные из прототипа. Переменные нейтральные (не привязаны к СДЭК), используются через Mantine `primaryColor`.

```css
:root {
  --brand-green:        #00B33C;
  --brand-green-dark:   #009A33;
  --brand-green-light:  #F0FDF4;
  --brand-green-border: #BBF7D0;
  --brand-bg:           #F2F2F2;
  --brand-card:         #FFFFFF;
  --brand-text:         #1A1A1A;
  --brand-text-muted:   #6B7280;
  --brand-text-subtle:  #9CA3AF;
  --brand-border:       #EBEBEB;
  --brand-row:          #F9FAFB;
  --brand-radius-card:  14px;
  --brand-radius-btn:   6px;
  --brand-shadow-card:  0 1px 4px rgba(0,0,0,0.06);
}
```

> **White-label (TODO M22+):** эти переменные будут переопределяться brand CSS из API tenant'а (`GET /admin/tenants/{id}/brand`). Сейчас — хардкод для прототипа СДЭК.

Импортировать в `main.tsx` (до `@mantine/core/styles.css`).

### 2. `src/lib/theme.ts`

Заменить `primaryColor: 'blue'` → кастомную зелёную шкалу `brand`:

```ts
colors: {
  brand: ['#F0FDF4', '#DCFCE7', '#BBF7D0', '#86EFAC', '#4ADE80',
          '#22C55E', '#16A34A', '#00B33C', '#009A33', '#00651E'],
},
primaryColor: 'brand',
fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, sans-serif',
headings: {
  fontFamily: 'Inter, sans-serif',
  fontWeight: '800',
},
defaultRadius: 'md',  // 14px
```

### 3. `index.html` — Google Fonts Inter

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap" rel="stylesheet">
```

> **Security note:** для ФСТЭК-стенда Google Fonts будет заменён на self-hosted шрифт (задача T2214.1 — не scope).

### 4. `src/main.tsx` — Import BrandTokens.css

Добавить `import '@/components/ui/BrandTokens.css'` перед `@mantine/core/styles.css`.

### 5. `src/lib/providers.tsx` — AprilProviders из DS

Заменить локальный `AprilProviders` на обёртку:

```tsx
import { AprilProviders as AprilProvidersDS } from '@ukituki-ps/april-ui'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createAprilTheme } from './theme'

export function LKFLProviders({ children }: { children: ReactNode }) {
  return (
    <AprilProvidersDS theme={createAprilTheme()}>
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    </AprilProvidersDS>
  )
}
```

> **Naming:** локальный провайдер → `LKFLProviders` (чтобы не конфликтовать с `AprilProviders` из DS). Обновить импорт в `main.tsx`.

## Файлы

### Новые
- `src/components/ui/BrandTokens.css`

### Изменяются
- `src/lib/theme.ts` — primaryColor, fontFamily, headings
- `src/lib/providers.tsx` → `LKFLProviders` (обёртка над DS `AprilProviders`)
- `src/main.tsx` — import BrandTokens, `LKFLProviders` вместо `AprilProviders`
- `index.html` — Google Fonts Inter

## Зависимости

- Нет внешних зависимостей
- Нет зависимостей от других задач M22

## Критерии приёмки

- [ ] `BrandTokens.css` импортирован в `main.tsx`
- [ ] Mantine theme: `primaryColor: 'brand'` (зелёная шкала `#00B33C`)
- [ ] Шрифт Inter подключён (Google Fonts в `index.html`)
- [ ] Заголовки: `font-weight: 800` в теме
- [ ] `LKFLProviders` оборачивает `AprilProviders` (DS) + `QueryClientProvider`
- [ ] `npm run dev` → страница работает без ошибок
- [ ] Тесты не ломаются
- [ ] Фон страницы `#F2F2F2` (через CSS token `--brand-bg` в Shell)
