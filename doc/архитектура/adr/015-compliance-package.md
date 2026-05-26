# ADR-015 — Выделение compliance (cascade + audit + retention) из `consent/`

## Контекст

Пакет `internal/consent/` выполняет две разнородные обязанности:

1. **PDn lifecycle** — Grant, Revoke, List, CheckGranted (простая бизнес-логика)
2. **Compliance enforcement** — CascadeRevoke (каскадное удаление), AuditTrail (ФСТЭК-логирование), data retention policies (3 года, 5 лет)

CascadeRevoke затрагивает 3 домена:
- `engagement/flow/` — деактивация всех льгот пользователя
- `notification/` — информирование о результате каскадного удаления
- `security/` — audit trail logging для ФСТЭК

Смешивать grant/revoke с cascade delete нарушает **Single Responsibility Principle**.

## Решение

Вынести compliance enforcement в отдельный пакет `internal/compliance/`:

```
internal/
├── consent/           # остаётся только PDn lifecycle: Grant, Revoke, List, CheckGranted
└── compliance/        # ← НОВЫЙ (T0702)
    ├── cascade.go     # CascadeRevoke: delete all user data
    ├── audit.go       # AuditTrail: ФСТЭК-логирование compliance-событий
    └── retention.go   # EnforceRetention: 3 года, 5 лет политики
```

`ComplianceEngine` зависит от:
- `user/` — для деактивации пользователя
- `engagement/flow/` — для деактивации всех льгот
- `notification/` — для информирования о результате
- `db/` — для audit trail storage

## Последствия

- ✅ SRP восстановлен: consent содержит только lifecycle согласий
- ✅ Compliance изолирован от grant/revoke логики
- ✅ ФСТЭК-требования (audit trail, data retention) централизованы в одном месте
- ⚠️ Count internal packages: 8 → 9

## Статус

✅ Accepted (M07, T0702)
