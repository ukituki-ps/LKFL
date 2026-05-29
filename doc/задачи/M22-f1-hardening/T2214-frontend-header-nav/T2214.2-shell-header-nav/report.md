# T2214.2 — Отчёт о выполнении

## Статус

выполнено

## Что сделано

1. **Shell.tsx** — полная замена `AppShell` (sidebar-layout) на горизонтальный layout с `AprilProductHeader` из DS.
2. **HeaderNav.tsx** — горизонтальная навигация в header'е. 5 ссылок для employee, 3 для admin. Active-индикатор (зелёная подчёркивающая линия), hover — зелёный цвет.
3. **HeaderRight.tsx** — правая зона: BalancePill (mock «1 250»), колокольчик (заглушка), UserMenu.
4. **EmployeeNav.tsx** — упрощён (только для мобильного Drawer).
5. **AdminNav.tsx** — упрощён (только для мобильного Drawer).
6. **UserMenu.tsx** — avatar зелёный (34px), инициалы.
7. **Мобильная навигация** — Burger → Drawer с теми же ссылками. Breakpoint: 768px.

## Файлы

| Файл | Действие |
|------|----------|
| `src/components/layout/HeaderNav.tsx` | создан |
| `src/components/layout/HeaderRight.tsx` | создан |
| `src/components/layout/Shell.tsx` | изменён (полный реврайт) |
| `src/components/layout/EmployeeNav.tsx` | изменён |
| `src/components/layout/AdminNav.tsx` | изменён |
| `src/components/layout/UserMenu.tsx` | изменён |

## Критерии приёмки

- [x] Shell использует `AprilProductHeader` из DS (не `AppShell`)
- [x] Sidebar убран (навигация горизонтальная)
- [x] 5 ссылок навигации с underline active-индикатором
- [x] Balance-pill в правой зоне header'а
- [x] Bell icon button в правой зоне header'а
- [x] Avatar с инициалами (зелёный круг 34px) + dropdown
- [x] Sticky header
- [x] Фон страницы `#F2F2F2`
- [x] Content max-width 1100px, margin auto, padding 28px
- [x] Burger кнопка на ≤768px
- [x] Drawer с навигацией
- [x] Admin навигация (HR, Каталог, Контент)
- [x] Прямые импорты из `lucide-react` отсутствуют (PanelLeft → AprilIconPanelLeft)

## Примечания

- Мобильная навигация реализована инлайн в Shell.tsx (Drawer)
- EmployeeNav/AdminNav сохранены как компоненты (для потенциального reuse)
- `AprilIconPanelLeft` (hamburger icon) из `@ukituki-ps/april-ui@0.1.16` (был `lucide-react` → заменён в ходе аудита)
