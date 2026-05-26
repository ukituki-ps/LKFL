# T0702 — Выделение compliance (cascade + audit) из consent/

## Веха

M07-аудит-и-рефакторинг

## Контекст

Пакет `internal/consent/` документирует 5 публичных API:
- `Grant()`, `Revoke()`, `List()`, `CheckGranted()` — lifecycle согласий
- `CascadeRevoke()` — каскадное удаление при отзыве ПДн

**Проблема:**
`CascadeRevoke()` — это compliance-операция, затрагивающая 3 домена:
1. **Engagement** — деактивация всех активных льгот сотрудника
2. **Notification** — информирование сотрудника о результате каскадного удаления
3. **Audit trail** — логирование всех удалённых записей для ФСТЭК-проверки

Это не «consent lifecycle» — это compliance enforcement. Смешивать grant/revoke с cascade delete в одном пакете нарушает SRP.

**Решение — новый подпакет `compliance/`:**

```
internal/
├── consent/
│   ├── engine.go       # Grant, Revoke, List, CheckGranted — lifecycle only
│   └── templates.go    # consent templates
├── compliance/         # ← НОВЫЙ
│   ├── cascade.go      # CascadeRevoke — delete all user data
│   ├── audit.go        # Audit trail для ФСТЭК (кто/когда/что согласился, удалил)
│   └── retention.go    # Data retention enforcement (3 years, 5 years policies)
```

Consent-engine вызывает compliance.CascadeRevoke() при Revoke().
Compliance зависит от user/ (для деактивации), engagement/ (для деактивации льгот), notification/ (для информирования).

### Файлы-мишени

| Действие | Файл |
|---|---|
| Новый пакет | `архитектура/пакеты-platform.md` — `compliance/` |
| Обновить consent/ | `архитектура/пакеты-platform.md` — убрать CascadeRevoke из consent |
| Обновить DI граф | `архитектура/пакеты-platform.md` — compliance → user, engagement, notification |
| Обновить таблицу пакетов | `архитектура/пакеты-platform.md` — 9 пакетов вместо 8 |
| Обновить модули | `архитектура/модули.md` — Platform: 9 internal packages |
| Обновить безопасность | `архитектура/безопасность.md` — audit trail section → compliance/ |
| Создать ADR | `архитектура/adr/ADR-015-compliance-package.md` |
| Обновить README архитектуры | `архитектура/README.md` — ADR-015 |
| Обновить README вехи | `задачи/README.md` — статус M07 |

### Критерии приёмки

- [x] `архитектура/пакеты-platform.md` — 9 internal пакетов (auth, user, consent, compliance, engagement, eligibility, notification, recommendations, api)
- [x] `compliance/` документирован с 3 публичными API: CascadeRevoke, AuditTrail, EnforceRetention
- [x] `consent/` уменьшен до lifecycle только (Grant, Revoke, List, Check)
- [x] DI граф обновлён (compliance зависит от 3 пакетов + db/)
- [x] `архитектура/безопасность.md` — audit trail section ссылается на compliance/
- [x] Создан ADR-015 (compliance package)
- [x] Файлы-мишени все перечислены выше
