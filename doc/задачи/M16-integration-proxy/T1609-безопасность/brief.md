# T1609 — `архитектура/безопасность.md` — credential isolation

## Веха

M16-integration-proxy

## Контекст

`безопасность.md` описывает безопасность монолита. Proxy добавляет credential isolation.

## Что сделать

1. Добавить секцию "Integration Proxy — credential isolation"
2. Credential storage: proxy only, mono knows nothing
3. Encryption: AES-256-GCM для credentials в DB
4. Master key: Vault (production), env var (development)
5. Attack surface: webhook endpoint на proxy, не на mono
6. gRPC security: localhost, no TLS needed (same host)
7. Update STRIDE analysis: добавлен proxy component

### Файлы-мишени

| Действие | Файл |
|----------|------|
| Обновить | `архитектуру/безопасность.md` |

### Критерии приёмки

- [ ] Credential isolation секция
- [ ] Encryption описание (AES-256-GCM)
- [ ] Master key management (Vault + env)
- [ ] Webhook attack surface → proxy
- [ ] gRPC security (localhost, no TLS)
- [ ] STRIDE update
