# T2702-T2705 — Consent (оставшиеся задачи)

## Веха

M27-consent

## T2702 — internal/consent/ (Lifecycle)
- Service: Sign, Revoke, List, GetByDocument
- Status transitions: signed → revoked
- IP address capture
- Unit tests

## T2703 — API: Consent
```
GET  /api/v1/consents              — список согласий
GET  /api/v1/consents/:id          — детали
POST /api/v1/consents/:documentId/sign — подпись
POST /api/v1/consents/:id/revoke   — отзыв
```

## T2704 — PDF генерация
- gofpdf для генерации PDF полисов/согласий
- Template: название документа, текст, дата, подпись, IP
- MinIO/S3 storage (stub — file system для dev)
- Download endpoint

## T2705 — Consent flow block
- Flow engine integration: если consent required → блокировка активации
- Check: user has signed consent for required documents
- Error: "Подпишите согласие для продолжения"
- Unit tests

## Критерии приёмки

- [ ] Все 4 задачи реализованы
- [ ] Consent lifecycle
- [ ] PDF генерация
- [ ] Flow block при отсутствии согласия
