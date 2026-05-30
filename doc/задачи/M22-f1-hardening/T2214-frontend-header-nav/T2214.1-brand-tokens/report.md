# T2214.1 — Отчёт о выполнении

## Статус

выполнено

## Что сделано

1. **BrandTokens.css** — CSS-переменные бренда (зелёная палитра, нейтралы, радиусы, тени).
2. **theme.ts** — кастомная шкала `brand` (#00B33C), `primaryColor: 'brand'`, шрифт Inter, заголовки fw 800.
3. **index.html** — Google Fonts Inter (400/500/600/700/800). TODO M42: self-hosted для ФСТЭК.
4. **providers.tsx** — замена на `AprilProviders` из DS (v0.1.13) + `QueryClientProvider`. Локальный провайдер переименован в `LKFLProviders`.
5. **main.tsx** — импорт BrandTokens.css перед Mantine CSS. `LKFLProviders` вместо `AprilProviders`.

## Файлы

| Файл | Действие |
|------|----------|
| `src/components/ui/BrandTokens.css` | создан |
| `src/lib/theme.ts` | изменён |
| `src/lib/providers.tsx` | изменён |
| `src/main.tsx` | изменён |
| `index.html` | изменён |

## Критерии приёмки

- [x] BrandTokens.css импортирован в main.tsx
- [x] Mantine theme: `primaryColor: 'brand'` (зелёная шкала)
- [x] Шрифт Inter подключён (Google Fonts)
- [x] Заголовки: `font-weight: 800`
- [x] `LKFLProviders` оборачивает `AprilProviders` (DS) + `QueryClientProvider`
- [x] `AprilProductHeader` импортируется из `@ukituki-ps/april-ui` (проверка)

## Примечания

- DS пакеты обновлены в T2214.5 (`@ukituki-ps/april-ui@0.1.16`, `@ukituki-ps/april-tokens@0.1.16`)
- `AprilProductHeader` и `AprilProviders` доступны в v0.1.16
- `AprilFilterPills` и новые иконки (Coins, Coffee, Brain, Gift, Heart, Dumbbell) — доступны в v0.1.16
