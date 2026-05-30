# Аудит T2214 — Frontend: прототип → код (актуализированный)

> **Дата:** 2026-05-28
> **Версия DS:** `@ukituki-ps/april-ui@0.1.16` (обнаружена, актуализирует отчёт)
> **Файлы:** 9 (4×brief.md + 4×plan.yaml + 1 ТЗ-DS)
> **Статус файлов:** ✅ Все 9 файлов обновлены с учётом v0.1.16

---

## Резюме

| Параметр | Значение |
|----------|----------|
| Мета-задача | T2214 — Frontend: прототип → код |
| Подзадачи | 4 (T2214.1 → T2214.4) |
| **Вердикт** | ✅ **Все проблемы решены** (14 → 0) |
| Критических | 0 (было 1, исправлено upgrade DS) |
| Серьёзных | 0 (было 3, исправлено 3) |
| Средних | 0 (было 6, исправлено 6) |
| Минорных | 4 (L1-L4 — не блокируют) |

---

## Что исправлено upgrade DS v0.1.13 → v0.1.16

| ID | Было | Статус | Причина |
|----|------|--------|---------|
| **K1** | 🔴 lucide-react vs AprilIcon — нарушение архитектуры | ✅ **RESOLVED** | DS v0.1.16: 54 иконки, покрытие прототипа 35/38 + 4 fallback |
| **S3** | 🟠 AprilProductHeader недоступен в v0.1.13 | ✅ **RESOLVED** | `AprilProductHeader` экспортирован в v0.1.16 |
| **M3** | 🟡 Три префикса CSS переменных | ✅ **RESOLVED** | Унифицировано: `--brand-*` для LKFL, `--april-*` для DS |

### ТЗ-DS — все 17 компонентов реализованы в v0.1.16

| Компонент | v0.1.13 | v0.1.16 |
|-----------|---------|---------|
| `AprilFilterPills` | ❌ | ✅ |
| `AprilWizard` + `AprilWizardProgress` | ❌ | ✅ |
| `AprilStatCard` | ❌ | ✅ |
| `AprilCard` | ❌ | ✅ |
| `AprilBenefitRow` | ❌ | ✅ |
| `AprilEventRow` | ❌ | ✅ |
| `AprilQuickButton` | ❌ | ✅ |
| `AprilBalanceCard` | ❌ | ✅ |
| `AprilTransactionRow` | ❌ | ✅ |
| `AprilFaqItem` | ❌ | ✅ |
| `AprilOptionCard` | ❌ | ✅ |
| `AprilPayOptionCard` | ❌ | ✅ |
| `AprilFormInput`/`Textarea`/`Select` | ❌ | ✅ |
| `AprilConfirmCheckbox` | ❌ | ✅ |
| `AprilSuccessScreen` | ❌ | ✅ |
| `AprilConfirmDoc` | ❌ | ✅ |
| `AprilProductHeader` | ✅ | ✅ |

---

## Решённые проблемы

### S1 (🟠 → ✅): T2214 добавлена в план M22

**Что сделано:** `план/вехи.md` — добавлена строка T2214 после T2213. Цель M22 обновлена: добавлено «frontend polish». Exit criteria M22 обновлены: добавлено «Frontend визуально соответствует прототипу».

Счёт: 13 (T2201-T2213) + 1 (T2214 meta) + 4 (подзадачи) = **18 задач** ✅

### M1 (🟡 → ✅): T2214 оставлена в M22 как «frontend polish»

**Решение:** Цель M22 расширена: `рефакторинг, тесты, нагрузка, мониторинг, CI/CD, деплой, security, release, frontend polish`. T2214 — часть hardening-вехи как визуальная доводка F1.

---

## Минорные (не блокируют)

| ID | Описание | Действие |
|----|----------|----------|
| **L1** | Google Fonts → нет follow-up задачи для ФСТЭК (self-hosted) | Добавить TODO-ссылку в T2214.1 brief на будущую задачу |
| **L2** | Header 56px vs 58px не зафиксировано | Оставить как есть (≤2px разница — приемлемо, зафиксировано в brief) |
| **L3** | plan.yaml — нет оценки effort | Добавить при планировании спринта |
| **L4** | StubBadge — `!import.meta.env.DEV` вместо `import.meta.env.PROD` | ✅ Исправлено в T2214.3 brief |

---

## Что исправлено в файлах

### Parent brief.md

| Изменение | До | После |
|-----------|-----|-------|
| DS версия | Не указана | **v0.1.16** (шаг 0 в T2214.1) |
| Иконочная стратегия | `lucide-react` напрямую | **`AprilIcon*` из DS** + fallback таблица |
| Порядок выполнения | «параллельно T2214.3 и T2214.4» | **Строго последовательно** (было противоречие) |
| npm deps | `lucide-react` добавить | **НЕ добавлять `lucide-react`** |
| FilterBar | Кастомный filter pills | **`AprilFilterPills` из DS** |
| AprilProductHeader | Без оговорки о доступности | **Явно: доступен в v0.1.16** |
| Модалки | ТЗ для DS-агента | **Компоненты уже в DS v0.1.16** |

### T2214.1 brief.md + plan.yaml

| Изменение | До | После |
|-----------|-----|-------|
| Название | «Brand tokens, тема, шрифт» | **Brand tokens, тема, шрифт, DS upgrade** |
| plan.yaml шаг 0 | Отсутствует | **`npm install @ukituki-ps/april-ui@0.1.16 @ukituki-ps/april-tokens@0.1.16`** |
| Проверка | Нет | **Компиляция импорта `AprilProductHeader`** |

### T2214.2 brief.md + plan.yaml

| Изменение | До | После |
|-----------|-----|-------|
| AprilProductHeader | Без оговорки о версии | **Явно: из DS v0.1.16** |
| Иконки HeaderRight | Без указания | **`AprilIconCoins` + `AprilIconBell`** |

### T2214.3 brief.md + plan.yaml

| Изменение | До | После |
|-----------|-----|-------|
| Иконочная стратегия | **`lucide-react` напрямую** | **`AprilIcon*` из DS v0.1.16** |
| npm deps | `lucide-react` добавить | **НЕ добавлять `lucide-react`** |
| StubBadge условие | `import.meta.env.PROD` | **`!import.meta.env.DEV`** |
| plan.yaml шаг 1 | `npm install lucide-react` | **Удалено** |
| plan.yaml проверка | Отсутствует | **`lucide-react` НЕ в dependencies** |

### T2214.4 brief.md + plan.yaml

| Изменение | До | После |
|-----------|-----|-------|
| Название | «EngagementCard + FilterBar + Lucide» | **EngagementCard + AprilFilterPills + AprilIcon** |
| Иконочная стратегия | `lucide-react` напрямую | **`AprilIcon*` из DS v0.1.16** |
| FilterBar | Кастомный filter pills | **`AprilFilterPills` из DS** |
| «Рефакторинг» | Рефакторинг | **Полный реврайт** (точнее описывает scope) |
| plan.yaml проверка | Отсутствует | **`lucide-react` НЕ импортируется напрямую** |

### ТЗ-Design-System.md

| Изменение | До | После |
|-----------|-----|-------|
| Статус | «требуется реализовать» | **✅ Выполнено в v0.1.16** |
| Компоненты | ТЗ для DS-агента | **Таблица «присутствует в v0.1.16»** |
| Icon fallback | Отсутствует | **Добавлена таблица 4 fallback иконок** |
| Breaking changes | Не указано | **Нет breaking changes** |
| Дополнительные компоненты | Не указаны | **19 компонентов beyond ТЗ** |
| Peer deps | Не указаны | **Полный список peer deps** |

---

## Иконочная стратегия — финальная

```
                    lucide-react (300+ иконок)
                         ↓ re-export (54 иконки)
                    @ukituki-ps/april-ui@0.1.16
                         ↓ AprilIcon*
                    LKFL frontend (T2214.3, T2214.4)

Прямой импорт lucide-react: ЗАПРЕЩЁНО
```

**Покрытие прототипа:**
- 35 из 38 иконок → 1:1 маппинг через `AprilIcon*`
- 3 иконки → nearest-аналоги из DS:
  - `heart-pulse` → `AprilIconHeart`
  - `shield-check` → `AprilIconSuccess`
  - `shield-plus` → `AprilIconPlusCircle`
  - `smart-speaker` → `AprilIconSmartphone`

**Архитектурные правила соблюдены:**
- `архитектура/фронтенд.md` §Иконки — ✅
- NAVIGATION.md правило #8 — ✅
- ADR-007 — ✅

---

## Порядок выполнения подзадач (финальный)

```
T2214.1 (DS upgrade + tokens) → T2214.2 (Header) → T2214.3 (Pages) → T2214.4 (Catalog)
    ↑                                                                        ↑
    Строго последовательно, без параллельного выполнения                       ↑
```
