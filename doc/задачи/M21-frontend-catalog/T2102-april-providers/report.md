# Отчёт T2102 — AprilProviders

## Статус
выполнено

## Дата
2026-05-26

## Что сделано

### 1. `src/lib/theme.ts` — April тема
- Реализован `createAprilTheme()` → `MantineThemeOverride`
- Тема использует CSS переменные April tokens (`--april-font-family`, `--april-font-mono`)
- Дефолтные fallback для шрифтов: `-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`
- Глобальные стили для Button (`radius: md`) и Card (`padding: lg, radius: md`)
- `primaryColor: blue` (заглушка до brand config)

### 2. `src/lib/providers.tsx` — AprilProviders
- Компонент `AprilProviders` объединяет `MantineProvider` + `QueryClientProvider`
- Тема создаётся один раз на модуле (не каждый рендер)
- `queryClient` с настройками: staleTime 5min, retry 1, refetchOnWindowFocus false
- Mutations: retry 0 (не повторять мутации)
- `queryClient` экспортирован для тестов и devtools

### 3. `src/main.tsx` — точка входа
- `ReactDOM.createRoot` → `React.StrictMode` → `AprilProviders` → `App`
- Подключён `@mantine/core/styles.css`
- Import April tokens CSS закомментирован (пакет может быть недоступен в dev)

### 4. `src/App.tsx` — корневой компонент
- Обёртка `BrowserRouter` из react-router-dom
- Placeholder: routes добавятся в T2103

## Результаты проверки

| Команда | Результат |
|---------|-----------|
| `npm run build` | ✅ Успех (1.15s, 791 модуль) |
| `npm run lint` | ✅ Без ошибок |

## Замечания

- Тип `createTheme()` в Mantine v7 возвращает `MantineThemeOverride`, а не `MantineTheme` — исправлено
- Brand CSS variables override отложен на M22+ (когда будет доступен endpoint `/admin/tenants/{id}/brand`)

## Изменённые файлы

| Файл | Действие |
|------|----------|
| `frontend/src/lib/theme.ts` | Переписан (был placeholder) |
| `frontend/src/lib/providers.tsx` | Переписан (был placeholder) |
| `frontend/src/main.tsx` | Обновлён (импорт AprilProviders + Mantine CSS) |
| `frontend/src/App.tsx` | Обновлён (BrowserRouter обёртка) |
